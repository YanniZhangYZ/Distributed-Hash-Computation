package consensus

import (
	"github.com/rs/xid"
	"go.dedis.ch/cs438/types"
	"time"
)

type proposeResult struct {
	isOurs bool
	err    error
}

type phase1Result struct {
	isOurs    bool
	tlcChange bool
	err       error
}

type phase2Result struct {
	isOurs    bool
	tlcChange bool
	success   bool
	err       error
}

type Paxos struct {
	proposeID            uint
	phase                uint
	maxID                uint
	AcceptedID           uint
	AcceptedValue        *types.PaxosValue
	alreadyBroadcast     bool
	promiseCnt           int
	collectEnoughPromise *chan bool
	acceptCnt            map[string]int
	collectEnoughAccept  *chan *types.PaxosValue
}

func (c *Consensus) paxosPropose(name string, mh string) proposeResult {
	for {
		phase1Res := c.paxosPhase1(name, mh)
		if phase1Res.err != nil {
			return proposeResult{false, phase1Res.err}
		}
		if phase1Res.tlcChange {
			return proposeResult{phase1Res.isOurs, nil}
		}

		/* If no error occurs, we can now try stage 2 */
		phase2Res := c.paxosPhase2(name, mh)
		if phase2Res.err != nil {
			return proposeResult{false, phase2Res.err}
		}
		if phase2Res.tlcChange {
			return proposeResult{phase2Res.isOurs, nil}
		}
		if phase2Res.success {
			return proposeResult{phase2Res.isOurs, nil}
		}
		/* Else retry phase 1 */
	}
}

func (c *Consensus) paxosPhase1(name string, mh string) phase1Result {
	for {
		/* Broadcast a PaxosPrepareMessage, collect the promises */
		c.Lock()
		c.paxos.phase = 1
		paxosPrepareMsg := types.PaxosPrepareMessage{
			Step:   c.tlcStep,
			ID:     c.paxos.proposeID,
			Source: c.address,
		}
		paxosPrepareMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(paxosPrepareMsg)
		if err != nil {
			c.Unlock()
			return phase1Result{false, false, err}
		}

		c.paxos.promiseCnt = 0
		promiseChan := make(chan bool, 50) // buffered channel
		c.paxos.collectEnoughPromise = &promiseChan
		c.Unlock()

		err = c.message.Broadcast(paxosPrepareMsgTrans)
		if err != nil {
			return phase1Result{false, false, err}
		}

		select {
		case block := <-c.tlcChangeChan:
			isOurs := block.Value.Filename == name && block.Value.Metahash == mh
			c.Lock()
			c.paxos.phase = 0
			c.paxos.proposeID += c.totalPeers
			c.Unlock()
			return phase1Result{isOurs, true, nil}
		case <-promiseChan:
			return phase1Result{false, false, nil}
		case <-time.After(c.conf.PaxosProposerRetry):
			/* Retry next propose ID */
			c.Lock()
			c.paxos.proposeID += c.totalPeers
			c.Unlock()
		}
	}
}

func (c *Consensus) paxosPhase2(name string, mh string) phase2Result {
	c.Lock()
	c.paxos.phase = 2
	/* If we have received nothing from peers, we set the value to ours */
	proposeValue := c.paxos.AcceptedValue
	if proposeValue == nil {
		proposeValue = &types.PaxosValue{
			UniqID:   xid.New().String(),
			Filename: name,
			Metahash: mh,
		}
	}
	paxosProposeMsg := types.PaxosProposeMessage{
		Step:  c.tlcStep,
		ID:    c.paxos.proposeID,
		Value: *proposeValue,
	}
	paxosProposeMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(paxosProposeMsg)
	if err != nil {
		c.Unlock()
		return phase2Result{false, false, false, err}
	}

	c.paxos.acceptCnt[proposeValue.UniqID] = 0
	acceptedChan := make(chan *types.PaxosValue, 50) // buffered channel
	c.paxos.collectEnoughAccept = &acceptedChan
	c.Unlock()

	err = c.message.Broadcast(paxosProposeMsgTrans)
	if err != nil {
		return phase2Result{false, false, false, err}
	}

	select {
	case block := <-c.tlcChangeChan:
		isOurs := block.Value.Filename == name && block.Value.Metahash == mh
		c.Lock()
		c.paxos.phase = 0
		c.paxos.proposeID += c.totalPeers
		c.Unlock()
		return phase2Result{isOurs, true, false, nil}
	case acceptedValue := <-acceptedChan:
		isOurs := acceptedValue.Filename == name && acceptedValue.Metahash == mh
		return phase2Result{isOurs, false, true, nil}
	case <-time.After(c.conf.PaxosProposerRetry):
		/* Retry next propose ID from phase 1 */
		c.Lock()
		c.paxos.proposeID += c.totalPeers
		c.Unlock()
		return phase2Result{false, false, false, nil}
	}
}
