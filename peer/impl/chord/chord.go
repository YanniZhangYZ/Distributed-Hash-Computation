package chord

import (
	"fmt"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/types"
	"sync"
	"sync/atomic"
	"time"
)

func NewChord(conf *peer.Configuration, message *message.Message) *Chord {
	var queryChan sync.Map
	var pingChan sync.Map

	chord := Chord{
		address:           conf.Socket.GetAddress(),
		conf:              conf,
		message:           message,
		queryChan:         &queryChan,
		pingChan:          &pingChan,
		ringLenChan:       make(chan uint, 1),
		stopStabilizeChan: make(chan bool, 1),
		stopFixFingerChan: make(chan bool, 1),
		stopPingChan:      make(chan bool, 1),
	}
	// Compute the ID of this node inside the Chord Ring
	chord.chordID = chord.name2ID(chord.address)
	// Create the initial topology of the chord ring
	chord.Create()

	/* Chord callbacks */
	conf.MessageRegistry.RegisterMessageCallback(types.ChordQuerySuccessorMessage{}, chord.execChordQuerySuccMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordReplySuccessorMessage{}, chord.execChordReplySuccMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordQueryPredecessorMessage{}, chord.execChordQueryPredMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordReplyPredecessorMessage{}, chord.execChordReplyPredMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordNotifyMessage{}, chord.execChordNotifyMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordRingLenMessage{}, chord.execChordRingLenMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordClearPredecessorMessage{}, chord.execChordClearPredMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordSkipSuccessorMessage{}, chord.execChordSkipSuccMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordPingMessage{}, chord.execChordPingMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.ChordPingReplyMessage{}, chord.execChordPingReplyMessage)

	return &chord
}

type Chord struct {
	address           string
	conf              *peer.Configuration // The configuration contains Socket and MessageRegistry
	message           *message.Message    // Messaging used to communicate among nodes
	alive             atomic.Int32        // Whether this chord node is alive or not
	chordID           uint                // ID of this chord node
	predecessor       string              // predecessor of this node
	predecessorLock   sync.RWMutex        // The mutex to protect concurrent read write to the predecessor
	successor         string              // successor of this chord node
	successorLock     sync.RWMutex        // The mutex to protect concurrent read write to the successor
	fingerIdx         int                 // Update fingers in round-robin fashion
	fingers           []string            // Finger tables
	fingersLock       sync.RWMutex        // Finger table lock
	queryChan         *sync.Map           // The sync map stores the channel that used for query results
	pingChan          *sync.Map           // The sync map stores the channel that used for ping results
	ringLenChan       chan uint           // The channel is used for the query RingLen
	stopStabilizeChan chan bool           // Communication channel about whether we should stop the node
	stopFixFingerChan chan bool
	stopPingChan      chan bool
}

// GetChordID gets the chordID of the current node
func (c *Chord) GetChordID() uint {
	return c.chordID
}

// GetPredecessor gets the predecessor of the current node
func (c *Chord) GetPredecessor() string {
	c.predecessorLock.RLock()
	defer c.predecessorLock.RUnlock()
	predecessor := c.predecessor
	return predecessor
}

// GetSuccessor gets the successor of the current node
func (c *Chord) GetSuccessor() string {
	c.successorLock.RLock()
	defer c.successorLock.RUnlock()
	successor := c.successor
	return successor
}

// GetFingerTable gets the finger table of the current node
func (c *Chord) GetFingerTable() []string {
	c.fingersLock.RLock()
	defer c.fingersLock.RUnlock()
	fingers := make([]string, len(c.fingers))
	copy(fingers, c.fingers)
	return fingers
}

// Create creates a new chord ring topology
func (c *Chord) Create() {
	c.alive.Store(1)
	c.predecessor = ""
	c.successor = ""
	c.fingers = make([]string, c.conf.ChordBytes*8)
	for i := 0; i < c.conf.ChordBytes*8; i++ {
		c.fingers[i] = ""
	}
}

// Join joins an existing chord ring topology, this is done by asking an existing remote
// node about the successor of the current node's chordID
func (c *Chord) Join(remoteNode string) error {
	c.alive.Store(1)
	successor, err := c.querySuccessor(remoteNode, c.chordID)
	if err != nil {
		return err
	}

	c.predecessorLock.Lock()
	defer c.predecessorLock.Unlock()
	c.successorLock.Lock()
	defer c.successorLock.Unlock()
	c.fingersLock.Lock()
	defer c.fingersLock.Unlock()

	c.predecessor = ""
	c.successor = successor
	c.fingers[0] = successor
	return nil
}

// RingLen returns the length of the ring, i.e., the number of nodes inside the ring
func (c *Chord) RingLen() uint {
	c.successorLock.RLock()
	defer c.successorLock.RUnlock()

	// If we are the only node inside the Chord ring, returns 1
	if c.successor == "" || c.successor == c.address {
		return 1
	}

	// If we are not, prepare a new chord ring length message
	chordRingLenMsg := types.ChordRingLenMessage{
		Source: c.address,
		Length: 1,
	}
	chordRingLenMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordRingLenMsg)
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf("[%s] RingLen failed!", c.address))
	}

	// Send the message to the remote peer
	err = c.message.Unicast(c.successor, chordRingLenMsgTrans)
	if err != nil {
		log.Error().Err(err).Msg(
			fmt.Sprintf("[%s] RingLen failed!", c.address))
	}

	// Either we wait until the timeout, or we receive a response from the reply channel
	select {
	case ringLen := <-c.ringLenChan:
		// We receive an answer before the timeout, return the ring length
		return ringLen
	case <-time.After(c.conf.ChordTimeout):
		// Timeout, return 0, to indicate the failure
		return 0
	}
}

// Leave allows the chord node to leave an existing chord ring gracefully
func (c *Chord) Leave() error {
	c.predecessorLock.Lock()
	defer c.predecessorLock.Unlock()
	c.successorLock.Lock()
	defer c.successorLock.Unlock()
	c.fingersLock.Lock()
	defer c.fingersLock.Unlock()

	// In order for us to leave, we should inform our successor that it should remove us
	// from the predecessor field.
	if c.successor != "" {
		chordClearPredecessorMsg := types.ChordClearPredecessorMessage{}
		chordClearPredecessorMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordClearPredecessorMsg)
		if err != nil {
			return err
		}
		err = c.message.Unicast(c.successor, chordClearPredecessorMsgTrans)
		if err != nil {
			return err
		}
	}

	// We should also inform our predecessor that it should remove us from the successor
	// field, and use our successor as the new successor.
	if c.predecessor != "" {
		chordSkipSuccessorMsg := types.ChordSkipSuccessorMessage{
			Successor: c.successor,
		}
		chordSkipSuccessorMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordSkipSuccessorMsg)
		if err != nil {
			return err
		}
		err = c.message.Unicast(c.predecessor, chordSkipSuccessorMsgTrans)
		if err != nil {
			return err
		}
	}

	// Clear the state inside the Chord node
	c.predecessor = ""
	c.successor = ""
	for i := 0; i < c.conf.ChordBytes*8; i++ {
		c.fingers[i] = ""
	}
	c.alive.Store(0)
	c.StopDaemon()
	return nil
}

// querySuccessor queries a remote node or self about the successor of the given key, it can be used
// either when a new node joins the ring, or the node queries the object
func (c *Chord) querySuccessor(remoteNode string, key uint) (string, error) {
	// Prepare the new chord query message
	chordQueryMsg := types.ChordQuerySuccessorMessage{
		RequestID: xid.New().String(),
		Source:    c.address,
		Key:       key,
	}
	chordQueryMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordQueryMsg)
	if err != nil {
		return "", err
	}

	// Prepare a reply channel that receives the reply from the remote peer, if any response is ready
	replyChan := make(chan string, 1)
	c.queryChan.Store(chordQueryMsg.RequestID, replyChan)

	// Send the message to the remote peer
	err = c.message.Unicast(remoteNode, chordQueryMsgTrans)
	if err != nil {
		return "", err
	}

	// Either we wait until the timeout, or we receive a response from the reply channel
	select {
	case successor := <-replyChan:
		/* Delete the entry in the query reply channels, and return the result */
		c.queryChan.Delete(chordQueryMsg.RequestID)
		return successor, nil
	case <-time.After(c.conf.ChordTimeout):
		/* We are timeout here */
		c.queryChan.Delete(chordQueryMsg.RequestID)
		return "", nil
	}
}
