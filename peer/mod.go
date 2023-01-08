package peer

import (
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/registry"
	"go.dedis.ch/cs438/storage"
	"go.dedis.ch/cs438/transport"
	"time"
)

// Peer defines the interface of a peer in the Peerster system. It embeds all
// the interfaces that will have to be implemented.
type Peer interface {
	Service
	Messaging
	DataSharing
	Chord
	IBlockchain
}

// Factory is the type of function we are using to create new instances of
// peers.
type Factory func(Configuration) Peer

// Configuration if the struct that will contain the configuration argument when
// creating a peer. This struct will evolve.
type Configuration struct {
	Socket          transport.Socket
	MessageRegistry registry.Registry

	// AntiEntropyInterval is the interval at which the peer sends a status
	// message to a random neighbor. 0 means no status messages are sent.
	// Default: 0
	AntiEntropyInterval time.Duration

	// HeartbeatInterval is the interval at which a rumor with an EmptyMessage
	// is sent. At startup a rumor with EmptyMessage should always be sent. Note
	// that sending a rumor is expensive as it involve the
	// ack+status+continueMongering mechanism, which generates a lot of
	// messages. Having a low value can flood the system. A value of 0 means the
	// heartbeat mechanism is not activated, ie. no rumors with EmptyMessage are
	// sent at all.
	// Default: 0
	HeartbeatInterval time.Duration

	// AckTimeout is the timeout after which a peer consider a message lost. A
	// value of 0 represents an infinite timeout.
	// Default: 3s
	AckTimeout time.Duration

	// ContinueMongering defines the chance to send the rumor to a random peer
	// in case both peers are synced. 1 means it will continue, 0.5 means there
	// is a 50% chance, and 0 no chance.
	// Default: 0.5
	ContinueMongering float64

	// ChunkSize defines the size of chunks when storing data.
	// Default: 8192
	ChunkSize uint

	// Backoff parameters used for DataRequests.
	// Default: {2s 2 5}
	BackoffDataRequest Backoff

	Storage storage.Storage

	// TotalPeers is the total number of peers in Peerster. If it is <= 1 then
	// there is no use of Paxos/TLC/Blockchain.
	// Default: 1
	TotalPeers uint

	// PaxosThreshold is a function that return the threshold of peers needed to
	// have a consensus. Default value is N/2 + 1
	PaxosThreshold func(uint) int

	// PaxosID is the starting ID of a Paxos proposer. It is distributed from 1
	// to peers.
	// Default: 0
	PaxosID uint

	// PaxosProposerRetry is the amount of time a proposer waits before it
	// retries to send a prepare when it doesn't get enough promises or accepts.
	// Default: 5s.
	PaxosProposerRetry time.Duration

	// ChordBits is the number of hash bits in chord.
	// This value should be the same as the salt bytes (which is the key inside the Chord)
	ChordBytes int

	// ChordTimeout is the timeout that a chord peer wait, until it considers the remote
	// peer won't reply to the message it sends out
	ChordTimeout time.Duration

	// ChordStabilizeInterval is the interval the chord node wait until it checks its
	// successor pointer is up-to-date
	ChordStabilizeInterval time.Duration

	// ChordFixFingerInterval is the interval the chord node wait until it checks its
	// finger table entry is up-to-date
	ChordFixFingerInterval time.Duration

	// ChordPingInterval is the interval that a chord node wait until it checks the liveness
	// of a finger table entry
	ChordPingInterval time.Duration

	// BlockchainAccountAddress is the account address used in the DCracker blockchain
	BlockchainAccountAddress string

	// BlockchainDifficulty is the proof-of-work difficulty for the DCracker blockchain
	// Such difficulty MUST BE THE SAME for all participants
	BlockchainDifficulty uint

	// BlockchainBlockSize is maximum number of transactions one block contains
	BlockchainBlockSize uint

	// BlockchainBlockTimeout is time to create a new block
	// when #processed_txs >= 1 but < BlockchainBlockSize
	BlockchainBlockTimeout time.Duration

	// BlockchainInitialState is the initial world state of the blockchain
	BlockchainInitialState map[string]common.State

	//// BlockchainAddressToSocket is the mapping from blockchain address to network socket address
	//BlockchainAddressToSocket map[string]string
}

// Backoff describes parameters for a backoff algorithm. The initial time must
// be multiplied by "factor" a maximum of "retry" time.
//
//	for i := 0; i < retry; i++ {
//	  wait(initial)
//	  initial *= factor
//	}
type Backoff struct {
	Initial time.Duration
	Factor  uint
	Retry   uint
}
