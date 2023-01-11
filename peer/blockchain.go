package peer

import (
	"go.dedis.ch/cs438/peer/impl/blockchain/block"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"time"
)

// IBlockchain is the interface that describes functions of a distributed password cracker
type IBlockchain interface {

	// TransferMoney transfers amount of money from your own account to another
	TransferMoney(dst common.Address, amount int64, timeout time.Duration) error

	// ProposeContract proposes a new password-cracking contract
	// It submits a transaction of type CONTRACT_DEPLOYMENT_TX which deploys
	// a new smart contract to the blockchain to verify the outcome from the recipient.
	// Reward is debited from the job proposer's balance in advance and will be transferred to the job recipient
	// upon successful completion of the job.
	ProposeContract(hash string, salt string, reward int64, recipient string) error

	// ExecuteContract execute a password-cracking contract whose address is contractAddr by providing
	// the cracked password.
	// Reward for this contract will be issued to the executor is the password is verified by the contract.
	ExecuteContract(password string, contractAddr string) error

	// GetAccountAddress returns the account address of the peer.
	// This address is different from the network socket address.
	GetAccountAddress() string

	// GetBalance returns the balance of the peer's DCracker account
	GetBalance() int64

	// GetChain returns the chain from its miner
	GetChain() *block.Chain
}
