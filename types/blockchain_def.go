package types

import (
	"go.dedis.ch/cs438/peer/impl/blockchain/block"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
)

// TransactionMessage describes a message that contains a transaction and is broadcast in the blockchain network
// - implements types.Message
type TransactionMessage struct {
	SignedTX transaction.SignedTransaction
}

// BlockMessage describes a message that contains a block to be broadcast to the blockchain network
// - implements types.Message
type BlockMessage struct {
	//Block block.Block
	TransBlock block.TransBlock
}
