package peer

import (
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"time"
)

// IDCracker is the interface that describes functions of a distributed password cracker
type IDCracker interface {

	// TransferMoney transfers amount of money from your own account to another
	TransferMoney(dst common.Address, amount int64, timeout time.Duration) error

	// ProposeContract proposes a new password-cracking contract
	// 1. This function uses unicast to send the password-cracking job to a particular recipient
	// 2. In the meanwhile, it submits a transaction of type CONTRACT_DEPLOYMENT_TX which deploys
	// a new smart contract to the blockchain to verify the outcome from that recipient.
	// 3. Reward is debited from the job proposer's balance in advance and will be transferred to the job recipient
	// upon successful completion of the job.
	ProposeContract(password string, reward int64, recipient string) error

	// ExecuteContract instructs the peer to execute some smart contract upon receiving
	// TODO: Now the peer will accept and execute every job assigned to it. Make it configurable later.
	ExecuteContract(todo int, timeout time.Duration) bool

	// GetAccountAddress returns the account address of the peer.
	// This address is different from the network socket address.
	GetAccountAddress() string

	// GetBalance returns the balance of the peer's DCracker account
	GetBalance() int64
}
