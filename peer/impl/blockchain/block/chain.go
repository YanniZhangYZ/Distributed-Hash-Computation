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
	HashToTxs       map[string]*transaction.SignedTransaction
	AllTxs          map[transaction.Transaction]struct{}
}

func (c *Chain) HasTransactionHash(hash string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.HashToTxs[hash]
	return ok
}

func (c *Chain) HasTransaction(tx *transaction.SignedTransaction) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.AllTxs[tx.TX]
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
		HashToTxs:       make(map[string]*transaction.SignedTransaction),
		AllTxs:          make(map[transaction.Transaction]struct{}),
	}
	c.Blocks[c.Tail.BlockHash] = c.Tail

	return &c
}

func (c *Chain) NextBlock() *Block {
	c.mu.Lock()
	defer c.mu.Unlock()

	b := c.Tail
	return &Block{
		Timestamp: uint64(time.Now().UnixMicro()),
		Nonce:     rand.Uint32(),
		ID:        b.ID + 1,
		Creator:   c.address,
		PrevHash:  b.BlockHash,
		TXs:       make([]*transaction.SignedTransaction, 0),
		State:     common.NewKVStore[common.State](),
	}
}

func (c *Chain) CheckNewBlock(b *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if b.ID != c.Tail.ID+1 {
		return fmt.Errorf("blockID mismatch, expect %d but got %d", c.Tail.ID+1, b.ID)
	}

	if b.PrevHash != c.Tail.BlockHash {
		return fmt.Errorf("block's PrevHash mismatch")
	}

	// TODO : check block hash
	// TODO : replay the txs from this block, check the state

	return nil
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
		c.HashToTxs[tx.HashCode()] = tx
		c.AllTxs[tx.TX] = struct{}{}
	}

	return nil
}

func (c *Chain) GetBlockCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.Blocks)
}

func (c *Chain) GetTransactionCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.HashToTxs)
}

func (c *Chain) GetLastBlock() *Block {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Tail
}

func (c *Chain) PrintChain() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	currBlock := c.Tail.BlockHash

	m := fmt.Sprintf("Blockchain of account %s", c.address.String())
	s := ""
	s += strings.Repeat("#", (96-len(m))/2) + "  "
	s += m
	s += "  " + strings.Repeat("#", 100-len(s)-2) + "\n"

	for {
		s += c.Blocks[currBlock].PrintBlock()
		currBlock = c.Blocks[currBlock].PrevHash
		if currBlock == c.GenesisPrevHash {
			break
		}
		s += strings.Repeat(" ", 50) + "*\n"
		s += strings.Repeat(" ", 50) + "|\n"
		s += strings.Repeat(" ", 50) + "|\n"
		s += strings.Repeat(" ", 50) + "|\n"
		s += strings.Repeat(" ", 50) + "|\n"
	}

	return s
}

// ValidateChain does a full validation on the entire blockchain, which includes
// 1. hashes check of each block,
// 2. txs replay of each block,
// 3. number of blocks on the chain
func (c *Chain) ValidateChain() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	currBlockHash := c.Tail.BlockHash

	blockCnt := len(c.Blocks)
	validated := 0

	for currBlockHash != c.GenesisPrevHash {
		b := c.Blocks[currBlockHash]

		// Not the first block, do the full validation
		if b.PrevHash != c.GenesisPrevHash {
			prevBlock, ok := c.Blocks[b.PrevHash]
			if !ok {
				return fmt.Errorf("PrevHash doesn't exist%s", b.PrevHash)
			}
			err := b.ValidateBlock(&prevBlock.State)
			if err != nil {
				return err
			}
		}

		validated++
		currBlockHash = b.PrevHash
	}

	if validated != blockCnt {
		return fmt.Errorf("all blocks do not form a single chain, %d blocks in total but %d blocks on the chain", blockCnt, validated)
	} else {
		return nil
	}
}
