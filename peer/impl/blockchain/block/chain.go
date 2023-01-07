package block

import (
	"fmt"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Chain struct {
	mu              sync.Mutex
	address         common.Address
	GenesisPrevHash string // 0s
	Difficulty      uint
	Blocks          map[string]*Block
	Tail            *Block
	AllTxs          map[string]*transaction.SignedTransaction
}

func (c *Chain) HasTransaction(hash string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.AllTxs[hash]
	return ok
}

func NewChain(addr common.Address, difficulty uint, initState map[string]common.State) *Chain {
	c := Chain{
		mu:              sync.Mutex{},
		address:         addr,
		GenesisPrevHash: strings.Repeat("0", 64),
		Difficulty:      difficulty,
		Blocks:          make(map[string]*Block),
		Tail:            NewGenesisBlock(initState),
		AllTxs:          make(map[string]*transaction.SignedTransaction),
	}
	c.Blocks[c.Tail.BlockHash] = c.Tail

	return &c
}

func (c *Chain) NextBlock() *Block {
	c.mu.Lock()
	defer c.mu.Unlock()

	b := c.Tail
	return &Block{
		Timestamp: uint64(time.Now().Unix()),
		Nonce:     rand.Uint32(),
		ID:        b.ID + 1,
		Creator:   c.address,
		PrevHash:  b.BlockHash,
		TXs:       make([]*transaction.SignedTransaction, 0),
		State:     common.NewKVStore[common.State](),
	}
}

func (c *Chain) CheckNewBlock(b *Block) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if b.ID != c.Tail.ID+1 {
		return false
	}

	// TODO : check block hash
	// TODO : replay the txs from this block, check the state

	return true
}

func (c *Chain) AppendBlock(b *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if b.ID != c.Tail.ID+1 {
		return fmt.Errorf("block to be appended has a wrong ID")
	}

	c.Tail = b
	c.Blocks[b.BlockHash] = b

	for _, tx := range b.TXs {
		c.AllTxs[tx.HashCode()] = tx
	}

	return nil
}
