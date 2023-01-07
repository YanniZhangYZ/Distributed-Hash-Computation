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

	// for starting and ending daemons
	CTX    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewMiner(message *message.Message) *Miner {
	m := Miner{}
	m.chain = block.NewChain(&m)
	m.address = common.Address{HexString: message.GetConf().BlockchainAccountAddress}
	m.logger = log.With().Str("address", m.address.String()).Logger()
	m.message = message

	m.tmpWorldState = common.NewKVStore[common.State]()

	m.txPending = common.NewSafeQueue[*transaction.SignedTransaction]()
	m.txProcessed = common.NewSafeQueue[*transaction.SignedTransaction]()
	m.txInvalid = common.NewSafeQueue[*transaction.SignedTransaction]()
	m.blockInCh = make(chan *types.BlockMessage, 10)
	m.blockBuffer = sync.Map{}

	m.message.GetConf().MessageRegistry.RegisterMessageCallback(types.TransactionMessage{}, m.execTransactionMessage)
	m.wg = sync.WaitGroup{}

	return &m
}

func (m *Miner) Start() {
	m.logger.Info().Msg("starting miner")
	m.CTX, m.cancel = context.WithCancel(context.Background())

	m.wg.Add(1)
	go m.txProcessingDaemon()

	//m.wg.Add(1)
	//go m.blockProcessingDaemon()
	m.logger.Info().Msg("started miner")
}

func (m *Miner) Stop() {
	m.logger.Info().Msg("stopping miner")
	m.cancel()
	m.wg.Wait()
	m.logger.Info().Msg("stopped miner")
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
