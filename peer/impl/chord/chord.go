package chord

import (
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/types"
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
	chord.chordID = chord.name2ID(chord.address)
	// Create the initial topology of the chord ring
	chord.Create()

	/* Chord callbacks */
	conf.MessageRegistry.RegisterMessageCallback(types.ChordQueryMessage{}, chord.execChordQueryMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordReplyMessage{}, chord.execChordReplyMessage)
	return &chord
}

type Chord struct {
	address         string
	conf            *peer.Configuration // The configuration contains Socket and MessageRegistry
	message         *message.Message    // Messaging used to communicate among nodes
	chordID         uint                // ID of this chord node
	predecessor     string              // predecessor of this node
	predecessorLock sync.Mutex          // The mutex to protect concurrent read write to the predecessor
	successor       string              // successor of this chord node
	successorLock   sync.Mutex          // The mutex to protect concurrent read write to the successor
	fingers         []string            // Finger tables
	fingersLock     sync.Mutex          // Finger table lock
	queryChan       *sync.Map           // The sync map stores the channel that used for query results
}

// GetPredecessor gets the predecessor of the current node
func (c *Chord) GetPredecessor() string {
	c.predecessorLock.Lock()
	defer c.predecessorLock.Unlock()
	predecessor := c.predecessor
	return predecessor
}

// GetSuccessor gets the successor of the current node
func (c *Chord) GetSuccessor() string {
	c.successorLock.Lock()
	defer c.successorLock.Unlock()
	successor := c.successor
	return successor
}

// GetFingerTable gets the finger table of the current node
func (c *Chord) GetFingerTable() []string {
	c.fingersLock.Lock()
	defer c.fingersLock.Unlock()
	fingers := make([]string, len(c.fingers))
	copy(fingers, c.fingers)
	return fingers
}

// Create creates a new chord ring topology
func (c *Chord) Create() {
	c.predecessorLock.Lock()
	defer c.predecessorLock.Unlock()
	c.successorLock.Lock()
	defer c.successorLock.Unlock()

	c.predecessor = ""
	c.successor = ""
	c.fingers = make([]string, c.conf.ChordBytes*8)
}

// Join joins an existing chord ring topology, this is done by asking an existing remote
// node about the successor of the current node's chordID
func (c *Chord) Join(remoteNode string) error {
	successor, err := c.querySuccessor(remoteNode, c.chordID)
	if err != nil {
		return err
	}

	c.predecessorLock.Lock()
	defer c.predecessorLock.Unlock()
	c.successorLock.Lock()
	defer c.successorLock.Unlock()

	c.predecessor = ""
	c.successor = successor
	c.fingers[0] = successor
	return nil
}
