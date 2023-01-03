package chord

import (
	"crypto"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/message"
	"math/big"
	"sync"
)

func NewChord(conf *peer.Configuration, message *message.Message) *Chord {
	var queryChan sync.Map
	chord := Chord{
		address:   conf.Socket.GetAddress(),
		conf:      conf,
		message:   message,
		queryChan: &queryChan,
	}
	// Compute the ID of this node inside the Chord Ring
	chord.chordID = chord.name2ID()
	// Create the initial topology of the chord ring
	chord.Create()
	return &chord
}

type Chord struct {
	address     string
	conf        *peer.Configuration // The configuration contains Socket and MessageRegistry
	message     *message.Message    // Messaging used to communicate among nodes
	chordID     uint                // ID of this chord node
	predecessor string              // predecessor of this node
	successor   string              // successors of this chord node
	fingers     []string            // Finger tables
	queryChan   *sync.Map           // The sync map stores the channel that used for query results
}

// name2ID computes from the address to the chordID, with the given ChordBits limit
func (c *Chord) name2ID() uint {
	h := crypto.SHA256.New()
	h.Write([]byte(c.address))
	hashSlice := h.Sum(nil)

	// Crop the hashSlice to only the specified chord bits, which is the size of the salt value, i.e.,
	// if the salt is 16 bits, then conf.ChordBytes = 2
	hashSlice = hashSlice[:c.conf.ChordBytes]
	return uint(big.NewInt(0).SetBytes(hashSlice).Uint64())
}

// Create creates a new chord ring topology
func (c *Chord) Create() {
	c.predecessor = ""
	c.successor = ""
	c.fingers = make([]string, c.conf.ChordBytes*8)
}

// Join joins an existing chord ring topology, this is done by asking an existing remote
// node about the successor of the current node's chordID
func (c *Chord) Join(remoteNode string) error {
	c.predecessor = ""
	successor, err := c.querySuccessor(remoteNode, c.chordID)
	if err != nil {
		return err
	}
	c.successor = successor
	return nil
}
