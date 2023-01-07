package dcracker

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/blockchain/miner"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/types"
	"sync"
	"time"
)

type DCracker struct {
	logger   zerolog.Logger
	message  *message.Message
	miner    *miner.Miner
	peerConf *peer.Configuration

	// address is the account address of the DCracker EOA
	address common.Address

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// TODO : keys not used
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

func NewDCracker(conf *peer.Configuration, message *message.Message) *DCracker {
	d := DCracker{}
	d.message = message
	d.peerConf = message.GetConf()
	d.address = common.Address{HexString: d.peerConf.BlockchainAccountAddress}
	d.logger = log.With().Str("address", d.address.String()).Logger()
	d.miner = miner.NewMiner(message)

	return &d
}

func (a *DCracker) Start() {
	a.logger.Info().Msg("Starting DCracker")
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.miner.Start()
	a.logger.Info().Msg("Started DCracker")
}

func (a *DCracker) Stop() {
	a.cancel()
	a.wg.Wait()
	a.miner.Stop()

	a.logger.Info().Msg("Stopped DCracker")
}

func (a *DCracker) BroadcastTransaction(signedTx *transaction.SignedTransaction) error {
	txMsg := types.TransactionMessage{SignedTX: *signedTx}
	txTransMsg, err := a.message.GetConf().MessageRegistry.MarshalMessage(txMsg)
	if err != nil {
		return err
	}
	err = a.message.Broadcast(txTransMsg)
	if err != nil {
		return err
	}

	return nil
}

func (a *DCracker) CheckTransaction(txHash string, timeout time.Duration) error {
	// TODO : use channel instead of for loop
	start := time.Now()
	for {
		if time.Now().Sub(start) > timeout {
			return fmt.Errorf("transaction verification timeout")
		}

		if a.miner.HasTransaction(txHash) {
			return nil
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (a *DCracker) TransferMoney(dst common.Address, amount int64, timeout time.Duration) error {
	// 0. Do you have enough money?
	balance := a.GetBalance()
	if balance < amount {
		return fmt.Errorf("TransferMoney failed : don't have enough balance")
	}

	// 1. Generate a transaction with type TRANSFER_TX
	rawTx := transaction.NewTransferTX(a.address, dst, amount)

	// 2. Sign the transaction
	signedTx, err := rawTx.Sign(a.privateKey)
	txHash := signedTx.HashCode()
	if err != nil {
		return err
	}

	// 3. Broadcast the transaction to the network
	err = a.BroadcastTransaction(&signedTx)
	if err != nil {
		return err
	}

	// 4. Wait for the transaction to be included in a block
	err = a.CheckTransaction(txHash, timeout)
	if err != nil {
		return err
	}

	return nil
}

func (a *DCracker) ProposeContract(password string, reward int64, recipient string) error {
	//TODO implement me
	panic("implement me")
}

func (a *DCracker) ExecuteContract(todo int, timeout time.Duration) bool {
	//TODO implement me
	panic("implement me")
}

func (a *DCracker) GetAccountAddress() string {
	return a.address.String()
}

func (a *DCracker) GetBalance() int64 {
	worldState := a.miner.GetWorldState()
	state, _ := worldState.Get(a.GetAccountAddress())
	return state.Balance
}
