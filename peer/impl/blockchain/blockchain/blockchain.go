package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/blockchain/block"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/blockchain/miner"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
	"go.dedis.ch/cs438/peer/impl/consensus"
	"go.dedis.ch/cs438/peer/impl/contract/impl"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/storage"
	"go.dedis.ch/cs438/types"
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

	// numContract is the number of contracts this account has published
	numContract int

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
	d.numContract = 0
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
		//Str("code", string(signedTx.TX.Code)).
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
				//Str("code", string(signedTx.TX.Code)).
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
		//Str("code", string(signedTx.TX.Code)).
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

func (a *Blockchain) ProposeContract(hash string, salt string, reward int64, recipient string, timeout time.Duration) error {
	// plainContract := fmt.Sprintf(
	// 	`
	// 	ASSUME publisher.balance > 5
	// 	IF finisher.crackedPwd.hash == "%s" THEN
	// 	smartAccount.transfer("finisher_ID", %d)
	// `, hash, reward)
	plainContract := impl.BuildPlainContract(hash, recipient, reward)

	// Create a contract instance
	a.numContract++
	contractAddress := fmt.Sprintf("%s_%d", a.address.String(), a.numContract)
	contract := impl.NewContract(
		contractAddress,      // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		a.address.String(),   // publisher
		recipient,            // finisher
	)

	a.nonce++
	rawTx := transaction.NewContractDeploymentTX(a.address, common.StringToAddress(contractAddress), reward, contract, a.nonce)

	// Sign the transaction
	signedTx, err := rawTx.Sign(a.privateKey)
	if err != nil {
		return err
	}

	// Broadcast the transaction to the network
	err = a.broadcastTransaction(&signedTx)
	if err != nil {
		return err
	}

	// Wait for the transaction to be included in a block
	err = a.checkTransaction(&signedTx, timeout)
	if err != nil {
		return err
	}

	return nil
}

func (a *Blockchain) ExecuteContract(password string, hash string, salt string, contractAddr string, timeout time.Duration) error {
	a.nonce++
	rawTx := transaction.NewContractExecutionTX(a.address, common.StringToAddress(contractAddr), password, hash, salt, a.nonce)

	// Sign the transaction
	signedTx, err := rawTx.Sign(a.privateKey)
	if err != nil {
		return err
	}

	// Broadcast the transaction to the network
	err = a.broadcastTransaction(&signedTx)
	if err != nil {
		return err
	}

	// Wait for the transaction to be included in a block
	err = a.checkTransaction(&signedTx, timeout)
	if err != nil {
		return err
	}

	return nil
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
