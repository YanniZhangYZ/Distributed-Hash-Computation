package block

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/blockchain/miner"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
	"strconv"
	"strings"
	"time"
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

func NewGenesisBlock(m *miner.Miner) *Block {
	b := Block{}
	b.Timestamp = 0
	b.Nonce = 0
	b.ID = 0
	b.Creator = common.Address{}

	b.PrevHash = strings.Repeat("0", 256/4)
	b.TXHash = strings.Repeat("0", 256/4)
	b.StateHash = strings.Repeat("0", 256/4)

	b.TXs = make([]*transaction.SignedTransaction, 0)

	b.State = common.NewKVStore[common.State]()
	for addr, state := range m.GetConf().BlockchainInitialState {
		b.State.Set(addr, state)
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

func (b *Block) String() string {
	//TODO implement me
	panic("implement me")
}

func (b *Block) ProofOfWork(zeros uint, m *miner.Miner) error {
	start := time.Now()
	for {
		select {
		case <-m.CTX.Done():
			return fmt.Errorf("stopped")
		default:
			{
				b.Nonce++
				hash := b.Hash()
				var i uint
				for i = 0; i < zeros; i++ {
					if hash[i] != 0 {
						continue
					}
				}

				log.Info().Str("address", b.Creator.String()).
					Msgf("Proof of Work finished in %f seconds", time.Now().Sub(start).Seconds())

				return nil
			}
		}

	}
}
