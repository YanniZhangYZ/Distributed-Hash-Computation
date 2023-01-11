package contract

// "go.dedis.ch/cs438/blockchain/storage"

// SmartContract describes functions to manipulate the code section in smart contract
type SmartContract interface {
	// Execute directly runs the contract code segment,
	// and ensures the desired property.
	// implemented in miner/execute.go
	// Execute() (error)

	ToString() string

	Marshal() ([]byte, error)

	// ValidateAssumptions(storage.KV) (bool, error)

	// CollectActions(storage.KV) ([]parser.Action, error)

	GetPublisherAccount() string

	GetFinisherAccount() string
}
