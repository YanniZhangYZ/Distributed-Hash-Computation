package miner

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/blockchain/block"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/types"
	"os"
	"sync"
)

type Miner struct {
	mu sync.Mutex

	chain *block.Chain

	address common.Address

	logger zerolog.Logger

	message *message.Message

	tmpWorldState common.WorldState

	// Txs that are not verified and executed
	txPending common.SafeQueue[*transaction.SignedTransaction]

	// Txs that are verified and executed. These txs will be included in the next block.
	txProcessed common.SafeQueue[*transaction.SignedTransaction]

	// Txs that are currently invalid. These txs will be processed again later.
	txInvalid common.SafeQueue[*transaction.SignedTransaction]

	blockInCh chan *types.BlockMessage

	// blockBuffer is a buffer map for blocks that are still not appended : block.id -> BlockMsg
	blockBuffer sync.Map

	// blockNotificationCh is a map from blockID to its corresponding channel,
	// used to notify and terminate unnecessary block forming and mining
	blockNotificationCh map[int]chan struct{}

	// for starting and ending daemons
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewMiner(message *message.Message) *Miner {
	m := Miner{}
	m.message = message
	m.address = common.Address{HexString: message.GetConf().BlockchainAccountAddress}
	m.chain = block.NewChain(m.address, m.GetConf().BlockchainDifficulty, m.GetConf().BlockchainInitialState)
	m.logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str("account", m.address.String()).Logger()

	m.tmpWorldState = common.NewKVStore[common.State]()

	m.txPending = common.NewSafeQueue[*transaction.SignedTransaction]()
	m.txProcessed = common.NewSafeQueue[*transaction.SignedTransaction]()
	m.txInvalid = common.NewSafeQueue[*transaction.SignedTransaction]()
	m.blockInCh = make(chan *types.BlockMessage, 10)
	m.blockBuffer = sync.Map{}
	m.blockNotificationCh = make(map[int]chan struct{})
	m.blockNotificationCh[1] = make(chan struct{})

	m.message.GetConf().MessageRegistry.RegisterMessageCallback(types.TransactionMessage{}, m.execTransactionMessage)
	m.message.GetConf().MessageRegistry.RegisterMessageCallback(types.BlockMessage{}, m.execBlockMessage)
	m.wg = sync.WaitGroup{}

	return &m
}

func (m *Miner) Start() {
	m.logger.Debug().Msg("starting miner")
	m.ctx, m.cancel = context.WithCancel(context.Background())

	m.wg.Add(1)
	go m.txProcessingDaemon()

	//m.wg.Add(1)
	//go m.blockProcessingDaemon()
	m.logger.Debug().Msg("started miner")
}

func (m *Miner) Stop() {
	m.logger.Debug().Msg("stopping miner")
	m.cancel()
	m.wg.Wait()
	m.logger.Debug().Msg("stopped miner")
}

func (m *Miner) GetConf() *peer.Configuration {
	return m.message.GetConf()
}

func (m *Miner) GetAddress() common.Address {
	return m.address
}

func (m *Miner) resetTmpWorldState() {
	m.tmpWorldState = m.chain.Tail.State.Copy()
}

func (m *Miner) HasTransaction(txHash string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.chain.HasTransaction(txHash)
}

func (m *Miner) GetWorldState() common.WorldState {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.chain.Tail.State.Copy()
}

func (m *Miner) GetContext() *context.Context {
	return &m.ctx
}

func (m *Miner) GetChain() *block.Chain {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.chain
}
