package transaction

import (
	"crypto/ecdsa"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
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

	// Code is the smart contract code
	// Only used for CONTRACT_DEPLOYMENT_TX
	Code string

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

func NewTransferTX(src common.Address, dst common.Address, amount int64) Transaction {
	return Transaction{
		Type:  TRANSFER_TX,
		Src:   src,
		Dst:   dst,
		Value: amount,
	}
}

func NewContractDeploymentTX( /* TODO */ ) Transaction {
	//TODO implement me
	panic("implement me")
}

func NewContractExecutionTX( /* TODO */ ) Transaction {
	//TODO implement me
	panic("implement me")
}

func (tx *SignedTransaction) String() string {
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (tx *SignedTransaction) HashCode() string {
	//TODO implement me
	panic("implement me")
}
