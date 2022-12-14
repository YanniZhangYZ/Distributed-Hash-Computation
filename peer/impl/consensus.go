package impl

import (
	"crypto"
	"encoding/hex"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/storage"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"strconv"
	"sync"
)

type consensusModule struct {
	sync.RWMutex
	cond          *sync.Cond
	address       string
	conf          *peer.Configuration
	message       *messageModule
	threshold     int
	totalPeers    uint
	tlcStep       uint
	paxos         Paxos
	tlcCnt        map[uint]int
	tlcValue      map[uint]*types.BlockchainBlock
	tlcChangeChan chan *types.BlockchainBlock
}

func (c *consensusModule) createNewPaxos() {
	c.paxos = Paxos{
		proposeID: c.conf.PaxosID,
		acceptCnt: make(map[string]int),
	}
}

func (c *consensusModule) buildTLCMsg() types.TLCMessage {
	/* To be called from ExecPaxosAcceptMessage when the paxos reaches consensus */
	h := crypto.SHA256.New()
	h.Write([]byte(strconv.Itoa(int(c.tlcStep))))
	h.Write([]byte(c.paxos.AcceptedValue.UniqID))
	h.Write([]byte(c.paxos.AcceptedValue.Filename))
	h.Write([]byte(c.paxos.AcceptedValue.Metahash))
	prevHash := c.conf.Storage.GetBlockchainStore().Get(storage.LastBlockKey)
	if prevHash == nil {
		prevHash = make([]byte, 32)
	}
	h.Write(prevHash)
	hash := h.Sum(nil)

	block := types.BlockchainBlock{
		Index:    c.tlcStep,
		Hash:     hash,
		Value:    *c.paxos.AcceptedValue,
		PrevHash: prevHash,
	}
	tlcMsg := types.TLCMessage{
		Step:  c.tlcStep,
		Block: block,
	}
	return tlcMsg
}

func (c *consensusModule) advanceTLC(catchup bool) error {
	/* To be called from ExecTLCMessage or itself when catchup*/
	/* Add block to the blockchain */
	block := c.tlcValue[c.tlcStep]
	hashHex := hex.EncodeToString(block.Hash)
	buf, err := block.Marshal()
	if err != nil {
		return err
	}
	c.conf.Storage.GetBlockchainStore().Set(hashHex, buf)
	c.conf.Storage.GetBlockchainStore().Set(storage.LastBlockKey, block.Hash)

	/* Set the name metahash association */
	c.conf.Storage.GetNamingStore().Set(block.Value.Filename, []byte(block.Value.Metahash))

	/* Broadcast if not catchup or already broadcast */
	if !catchup && !c.paxos.alreadyBroadcast {
		tlcMsg := types.TLCMessage{
			Step:  c.tlcStep,
			Block: *block,
		}
		c.paxos.alreadyBroadcast = true
		tlcMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(tlcMsg)
		if err != nil {
			return err
		}

		err = c.message.broadcast(tlcMsgTrans)
		if err != nil {
			return err
		}
	}
	/* Before changing paxos, we should notify any thread that is proposing values */
	c.tlcChangeChan <- block

	/* Increase tlc and update paxos */
	c.tlcStep++
	c.createNewPaxos()

	/* Wake up threads that are waiting for the paxos to finish */
	c.cond.Broadcast()

	/* Catchup if any */
	if c.tlcCnt[c.tlcStep] >= c.threshold {
		c.tlcCnt[c.tlcStep] = 0
		return c.advanceTLC(true)
	}
	return nil
}

func (c *consensusModule) tag(name string, mh string) error {
	/* Check if the name already exists in the name store */
	if c.conf.Storage.GetNamingStore().Get(name) != nil {
		return xerrors.Errorf("Tag name already exists!")
	}

	c.Lock()
	if c.paxos.phase != 0 {
		/* If already proposing, wait */
		c.cond.Wait()
		c.Unlock()
		/* Start again */
		return c.tag(name, mh)
	}
	/* If it is not proposing, start proposing */
	c.Unlock()
	proposeRes := c.paxosPropose(name, mh)
	if proposeRes.err != nil {
		return proposeRes.err
	}
	/* Check if it is our value */
	if proposeRes.isOurs {
		return nil
	}
	return c.tag(name, mh)
}
