package chord

import (
	"crypto"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/message"
	"math/big"
)

func NewChordModule(conf *peer.Configuration, message *message.MessageModule) *ChordModule {
	chord := ChordModule{
		address: conf.Socket.GetAddress(),
		conf:    conf,
		message: message,
	}

	// Compute the ID of this node inside the Chord Ring
	chord.name2ID()

	return &chord
}

type ChordModule struct {
	address     string
	conf        *peer.Configuration    // The configuration contains Socket and MessageRegistry
	message     *message.MessageModule // Messaging used to communicate among nodes
	chordId     uint                   // ID of this chord node
	predecessor string                 // predecessor of this node
	successor   string                 // successors of this chord node
	fingers     []string               // Finger tables
}

// name2ID computes from the address to the chordID, with the given ChordBits limit
func (c *ChordModule) name2ID() {
	h := crypto.SHA256.New()
	h.Write([]byte(c.address))
	hashSlice := h.Sum(nil)

	// Crop the hashSlice to only the specified chord bits, which is the size of the salt value, i.e.,
	// if the salt is 16 bits, then conf.ChordBytes = 2
	hashSlice = hashSlice[:c.conf.ChordBytes]
	c.chordId = uint(big.NewInt(0).SetBytes(hashSlice).Uint64())
}

// Create creates a new chord ring topology
func (c *ChordModule) Create() {
	c.predecessor = ""
	c.successor = ""
}

func (c *ChordModule) Join(remoteNode string) {
	c.predecessor = ""
	c.successor = c.querySuccessors(remoteNode, c.chordId)
}
