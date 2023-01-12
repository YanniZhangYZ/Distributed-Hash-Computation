package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/blockchain/block"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/blockchain/miner"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
	"go.dedis.ch/cs438/peer/impl/consensus"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/storage"
	"go.dedis.ch/cs438/types"
	"os"
	"sync"
	"time"
)

type Blockchain struct {
	logger   zerolog.Logger
	message  *message.Message
	miner    *miner.Miner
	peerConf *peer.Configuration

	// address is the account address of the Blockchain EOA
	address common.Address

	// nonce is number of transactions this account has sent
	nonce int

	// submittedTxs keeps record of all submitted txs
	submittedTxs map[string]*transaction.SignedTransaction

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// TODO : keys not used
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

func NewBlockchain(conf *peer.Configuration, message *message.Message, consensus *consensus.Consensus, storage storage.Storage) *Blockchain {
	d := Blockchain{}
	d.message = message
	d.peerConf = message.GetConf()
	d.address = common.Address{HexString: d.peerConf.BlockchainAccountAddress}
	d.nonce = 0
	d.submittedTxs = make(map[string]*transaction.SignedTransaction)
	d.logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str("account", d.address.String()).Logger()
	d.miner = miner.NewMiner(conf, message, consensus, storage)

	return &d
}

func (a *Blockchain) Start() {
	a.logger.Debug().Msg("starting Blockchain")
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.miner.Start()
	a.logger.Debug().Msg("started Blockchain")
}

func (a *Blockchain) Stop() {
	a.logger.Debug().Msg("stopping Blockchain")
	a.cancel()
	a.wg.Wait()
	a.miner.Stop()
	a.logger.Debug().Msg("stopped Blockchain")
}

func (a *Blockchain) broadcastTransaction(signedTx *transaction.SignedTransaction) error {
	txMsg := types.TransactionMessage{SignedTX: *signedTx}
	txTransMsg, err := a.message.GetConf().MessageRegistry.MarshalMessage(txMsg)
	if err != nil {
		return err
	}
	err = a.message.Broadcast(txTransMsg)
	if err != nil {
		return err
	}

	a.submittedTxs[signedTx.HashCode()] = signedTx

	a.logger.Debug().
		Int("type", signedTx.TX.Type).
		Str("src", signedTx.TX.Src.String()).
		Str("dst", signedTx.TX.Dst.String()).
		Int("nonce", signedTx.TX.Nonce).
		Int64("value", signedTx.TX.Value).
		Uint64("timestamp", signedTx.TX.Timestamp).
		Str("code", signedTx.TX.Code).
		Str("data", signedTx.TX.Data).
		Msg("submitted a transaction")

	return nil
}

func (a *Blockchain) checkTransaction(signedTx *transaction.SignedTransaction, timeout time.Duration) error {
	// Now it check if the transaction is confirmed by only querying the blockchain of itself.
	// TODO : Use channel
	// TODO : Send TransactionVerifyMessage to the network to verify the transaction.

	start := time.Now()
	for {
		if time.Now().Sub(start) > timeout {
			a.logger.Debug().
				Int("type", signedTx.TX.Type).
				Str("src", signedTx.TX.Src.String()).
				Str("dst", signedTx.TX.Dst.String()).
				Int("nonce", signedTx.TX.Nonce).
				Int64("value", signedTx.TX.Value).
				Uint64("timestamp", signedTx.TX.Timestamp).
				Str("code", signedTx.TX.Code).
				Str("data", signedTx.TX.Data).
				Msg("submitted transaction verification timeout")

			return fmt.Errorf("transaction verification timeout")
		}

		if a.miner.HasTransaction(signedTx) {
			break
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}

	a.logger.Debug().
		Int("type", signedTx.TX.Type).
		Str("src", signedTx.TX.Src.String()).
		Str("dst", signedTx.TX.Dst.String()).
		Int("nonce", signedTx.TX.Nonce).
		Int64("value", signedTx.TX.Value).
		Uint64("timestamp", signedTx.TX.Timestamp).
		Str("code", signedTx.TX.Code).
		Str("data", signedTx.TX.Data).
		Msg("submitted transaction verified")

	return nil
}

func (a *Blockchain) TransferMoney(dst common.Address, amount int64, timeout time.Duration) error {
	balance := a.GetBalance()

	// 0. For non-declaration transaction, do you have enough money?
	if dst.String() != a.address.String() && balance < amount {
		a.logger.Debug().Int64("balance", balance).Int64("debit", amount).
			Msg("no enough balance for TransferMoney")
		return fmt.Errorf("TransferMoney failed : don't have enough balance")
	}

	// 1. Generate a transaction with type TRANSFER_TX
	a.nonce++
	rawTx := transaction.NewTransferTX(a.address, dst, amount, a.nonce)

	// 2. Sign the transaction
	signedTx, err := rawTx.Sign(a.privateKey)
	if err != nil {
		return err
	}

	// 3. Broadcast the transaction to the network
	err = a.broadcastTransaction(&signedTx)
	if err != nil {
		return err
	}

	// 4. Wait for the transaction to be included in a block
	err = a.checkTransaction(&signedTx, timeout)
	if err != nil {
		return err
	}

	return nil
}

func (a *Blockchain) ProposeContract(hash string, salt string, reward int64, recipient string) error {
	//TODO implement me
	panic("implement me")
}

func (a *Blockchain) ExecuteContract(password string, contractAddr string) error {
	//TODO implement me
	panic("implement me")
}

func (a *Blockchain) GetAccountAddress() string {
	return a.address.String()
}

func (a *Blockchain) GetBalance() int64 {
	worldState := a.miner.GetWorldState()
	state, ok := worldState.Get(a.GetAccountAddress())
	if !ok {
		return 0
	} else {
		return state.Balance
	}
}

func (a *Blockchain) GetMiner() *miner.Miner {
	return a.miner
}

func (a *Blockchain) GetChain() *block.Chain {
	return a.miner.GetChain()
}
