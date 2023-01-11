package contract

import (
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

// SmartContract describes functions to manipulate the code section in smart contract
type SmartContract interface {
	// Execute directly runs the contract code segment,
	// and ensures the desired property.
	// implemented in miner/execute.go
	// Execute() (error)

	ToString() string

	Marshal() ([]byte, error)

	CheckAssumptions(*common.WorldState) (bool, error)

	GatherActions(*common.WorldState) ([]parser.Action, error)

	GetPublisherAccount() string

	GetFinisherAccount() string
}
