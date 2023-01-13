package transaction

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/contract"
	"time"
)

// Design references : Ethereum Yellow Paper ETHEREUM: A SECURE DECENTRALISED GENERALISED TRANSACTION LEDGER
// https://ethereum.github.io/yellowpaper/paper.pdf] Ch.4
// https://takenobu-hs.github.io/downloads/ethereum_evm_illustrated.pdf
// https://ethereum.org/en/developers/docs/transactions/

const (
	// TRANSFER_TX is the type of regular transactions
	// A transaction that transfers Cracker from one account to another
	TRANSFER_TX = iota

	// CONTRACT_DEPLOYMENT_TX is the type of contract deployment transactions
	// A transaction without a 'dest' address, where the data field is used for the contract code
	CONTRACT_DEPLOYMENT_TX = iota

	// CONTRACT_EXECUTION_TX is the type of execution of a contract
	// A transaction that interacts with a deployed smart contract. In this case, 'dest' address is the smart contract address
	CONTRACT_EXECUTION_TX = iota
)

type Transaction struct {

	// Type of transaction
	Type int

	// Dst is the Destination (recipient) of this transaction
	// For TRANSFER_TX, Dst is the account to receive money
	// For CONTRACT_DEPLOYMENT_TX, Dst is set empty
	// For CONTRACT_EXECUTION_TX, Dst is the smart contract address
	Dst common.Address

	// Src is the initiator of this transaction
	Src common.Address

	// Nonce is a sequentially incrementing counter which indicates the transaction number from the account
	Nonce int

	// Value is the amount of Cracker to transfer
	Value int64

	// Data is an optional field to include arbitrary data
	// For CONTRACT_EXECUTION_TX, Data field is interpreted by the smart contract code as the execution argument
	Data string

	// Contract is the smart contract code
	// Only used for CONTRACT_DEPLOYMENT_TX
	Contract []byte

	// Signature from the sender
	// This is generated when the sender's private key signs the transaction and confirms the sender has authorized this transaction
	Signature string

	// timestamp given by the sender
	Timestamp uint64

	// Comment is an optional field to include any additional information by the sender, primarily used for debug
	Comment string
}

type SignedTransaction struct {
	TX        Transaction
	Signature []byte
	TXHash    []byte
}

func NewTransferTX(src common.Address, dst common.Address, amount int64, nonce int) Transaction {
	return Transaction{
		Type:      TRANSFER_TX,
		Src:       src,
		Dst:       dst,
		Value:     amount,
		Timestamp: uint64(time.Now().UnixMicro()),
		Nonce:     nonce,
	}
}

func NewContractDeploymentTX(src common.Address, contractAddr common.Address, reward int64, contract contract.SmartContract, nonce int) Transaction {
	contractBytes, _ := contract.Marshal()
	return Transaction{
		Type:      CONTRACT_DEPLOYMENT_TX,
		Src:       src,
		Dst:       contractAddr,
		Value:     reward,
		Timestamp: uint64(time.Now().UnixMicro()),
		Nonce:     nonce,
		Contract:  contractBytes,
	}
}

func NewContractExecutionTX(src common.Address, contractAddr common.Address, password string, hash string, salt string, nonce int) Transaction {
	return Transaction{
		Type:      CONTRACT_EXECUTION_TX,
		Src:       src,
		Dst:       contractAddr,
		Value:     0,
		Timestamp: uint64(time.Now().UnixMicro()),
		Nonce:     nonce,
		Data:      fmt.Sprintf("%s,%s,%s", password, hash, salt),
	}
}

func (tx *SignedTransaction) String() string {
	str := ""
	str += fmt.Sprintf("type:%d, ", tx.TX.Type)
	str += fmt.Sprintf("Src:%s, ", tx.TX.Src.String())
	str += fmt.Sprintf("Dst:%s, ", tx.TX.Dst.String())
	str += fmt.Sprintf("Nonce:%d, ", tx.TX.Nonce)
	str += fmt.Sprintf("Value:%d, ", tx.TX.Value)
	str += fmt.Sprintf("Data:%s, ", tx.TX.Data)
	str += fmt.Sprintf("Contract:%s, ", string(tx.TX.Contract))
	str += fmt.Sprintf("Signature:%s, ", tx.TX.Signature)
	str += fmt.Sprintf("Timestamp:%d, ", tx.TX.Timestamp)
	str += fmt.Sprintf("Comment:%s, ", tx.TX.Comment)

	return str
}

func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) (SignedTransaction, error) {
	//TODO : Implement signature
	return SignedTransaction{
		TX:        *tx,
		Signature: nil,
		TXHash:    nil,
	}, nil
}

func (tx *SignedTransaction) Hash() []byte {
	hash := sha256.New()
	hash.Write([]byte(tx.String()))
	return hash.Sum(nil)
}

func (tx *SignedTransaction) HashCode() string {
	return hex.EncodeToString(tx.Hash())
}
