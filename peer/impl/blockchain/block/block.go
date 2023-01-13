package block

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
	"sort"
	"strconv"
	"strings"
)

type Block struct {
	Timestamp uint64
	Nonce     uint32
	ID        uint32
	Creator   common.Address

	PrevHash  string
	TXHash    string
	StateHash string
	BlockHash string

	TXs   []*transaction.SignedTransaction
	State common.WorldState
}

type TransBlock struct {
	Timestamp uint64
	Nonce     uint32
	ID        uint32
	Creator   common.Address

	PrevHash  string
	TXHash    string
	StateHash string
	BlockHash string

	TXs        []*transaction.SignedTransaction
	StateInMap map[string]common.State
}

func NewGenesisBlock(initState map[string]common.State) *Block {
	b := Block{}
	b.Timestamp = 0
	b.Nonce = 0
	b.ID = 0
	b.Creator = common.Address{}

	b.PrevHash = strings.Repeat("0", 256/4)
	b.TXHash = strings.Repeat("0", 256/4)
	b.StateHash = strings.Repeat("0", 256/4)

	b.TXs = make([]*transaction.SignedTransaction, 0)

	b.State = common.NewWorldState()

	if initState != nil {
		for addr, state := range initState {
			b.State.Set(addr, state)
		}
	}

	b.BlockHash = b.HashCode()

	return &b
}

// HashCode calculates and returns the hashcode string of this block using all its information
// This function will also update the field Block.BlockHash
func (b *Block) HashCode() string {
	return hex.EncodeToString(b.Hash())
}

// Hash calculates and returns the hash of this block using all its information
// This function will also update the field Block.BlockHash
func (b *Block) Hash() []byte {
	h := sha256.New()

	h.Write([]byte(strconv.Itoa(int(b.Timestamp))))
	h.Write([]byte(strconv.Itoa(int(b.Nonce))))
	h.Write([]byte(strconv.Itoa(int(b.ID))))
	h.Write([]byte(b.Creator.String()))

	h.Write([]byte(b.PrevHash))
	h.Write([]byte(b.TXHash))
	h.Write([]byte(b.StateHash))

	for _, tx := range b.TXs {
		h.Write(tx.Hash())
	}

	h.Write(b.State.Hash())

	hash := h.Sum(nil)
	b.BlockHash = hex.EncodeToString(hash)
	return hash
}

func (b *Block) ProofOfWork(zeros uint, ctx *context.Context, notifyCh chan struct{}) error {
	for {
		select {
		case <-(*ctx).Done():
			return fmt.Errorf("stopped")
		case <-notifyCh:
			return fmt.Errorf("early stop POW for the already mined block-%d", b.ID)
		default:
			{
				b.Nonce++
				hash := b.Hash()
				var i uint
				for i = 0; i < zeros; i++ {
					if hash[i] != 0 {
						break
					}
				}

				if i == zeros {
					return nil
				} else {
					continue
				}
			}
		}
	}
}

func (b *Block) GetTransBlock() *TransBlock {
	return &TransBlock{
		Timestamp: b.Timestamp,
		Nonce:     b.Nonce,
		ID:        b.ID,
		Creator:   b.Creator,

		PrevHash:   b.PrevHash,
		TXHash:     b.TXHash,
		StateHash:  b.StateHash,
		BlockHash:  b.BlockHash,
		TXs:        b.TXs,
		StateInMap: b.State.GetSimpleMap(),
	}
}

func (b *TransBlock) GetBlock() *Block {
	bb := Block{
		Timestamp: b.Timestamp,
		Nonce:     b.Nonce,
		ID:        b.ID,
		Creator:   b.Creator,

		PrevHash:  b.PrevHash,
		TXHash:    b.TXHash,
		StateHash: b.StateHash,
		BlockHash: b.BlockHash,
		TXs:       b.TXs,
		State:     common.NewWorldState(),
	}

	for k, v := range b.StateInMap {
		bb.State.Set(k, v)
	}

	return &bb
}

func (b *Block) PrintBlock() string {
	s := strings.Repeat("=", 100) + "\n"
	s += fmt.Sprintf("| Block #%d, Hash %s \n", b.ID, b.BlockHash[:8])
	s += "|" + strings.Repeat("-", 99) + "\n"
	s += fmt.Sprintf("| Timestamp: %d\n", b.Timestamp)
	s += fmt.Sprintf("| Nonce: %d\n", b.Nonce)
	s += fmt.Sprintf("| Creator: %s\n", b.Creator.String())
	s += fmt.Sprintf("| PrevHash: %s\n", b.PrevHash[:8])

	s += "| Transactions:\n"
	for _, tx := range b.TXs {
		s += fmt.Sprintf("| \t%s\n", tx.String())
	}

	s += "| World State:\n"
	stateKeys := b.State.Keys()
	sort.Strings(stateKeys)
	for _, key := range stateKeys {
		state, _ := b.State.Get(key)
		s += fmt.Sprintf("| \t%s : %s\n", key, state.String())
	}

	s += strings.Repeat("=", 100) + "\n"
	return s
}

func (b *Block) String() string {
	s := b.PrintBlock()
	s = strings.Replace(s, "\n", ", ", -1)
	s = strings.Replace(s, "\t", ", ", -1)
	s = strings.Replace(s, "=", "", -1)
	s = strings.Replace(s, "-", "", -1)
	return s
}

func (b *TransBlock) String() string {
	s := strings.Repeat("=", 64) + "\n"
	s += fmt.Sprintf("Block #%d, Hash %s \n", b.ID, b.BlockHash)
	s += strings.Repeat("-", 64) + "\n"
	s += fmt.Sprintf("Timestamp: %d\n", b.Timestamp)
	s += fmt.Sprintf("Nonce: %d\n", b.Nonce)
	s += fmt.Sprintf("Creator: %s\n", b.Creator.String())
	s += fmt.Sprintf("PrevHash: %s\n", b.PrevHash)

	s += "Transactions:\n"
	for _, tx := range b.TXs {
		s += fmt.Sprintf("\t%s", tx.String())
	}

	s += "World State:\n"
	for key, value := range b.StateInMap {
		s += fmt.Sprintf("\t%s : %s\n", key, value.String())
	}

	s += strings.Repeat("=", 64) + "\n"
	return s
}

// ValidateBlock validates the correctness of this block
// It replays all txs within this block on the given prevWorldState
// and check the hashes
func (b *Block) ValidateBlock(prevWorldState *common.WorldState) error {

	// Check hashes
	givenHash := b.BlockHash
	if givenHash != b.HashCode() {
		return fmt.Errorf("block hash %s does not match expected hash %s", givenHash, b.HashCode())
	}

	// Replay transactions
	tmpWorldState := (*prevWorldState).Copy()
	for _, tx := range b.TXs {
		err := transaction.VerifyAndExecuteTransaction(tx, tmpWorldState)
		if err != nil {
			return err
		}
	}

	if !tmpWorldState.Equal(&b.State) {
		return fmt.Errorf("prev state after txs replay does not match block's state")
	} else {
		return nil
	}
}
