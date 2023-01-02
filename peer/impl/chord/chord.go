package chord

import (
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/message"
)

func NewChordModule(conf *peer.Configuration, message *message.MessageModule) *ChordModule {
	chord := ChordModule{
		address: conf.Socket.GetAddress(),
		conf:    conf,
		message: message,
	}
	return &chord
}

type ChordModule struct {
	address     string
	conf        *peer.Configuration    // The configuration contains Socket and MessageRegistry
	message     *message.MessageModule // Messaging used to communicate among nodes
	chordId     int                    // ID of this chord node
	predecessor string                 // predecessor of this node
	successor   string                 // successors of this chord node
	fingers     []string               // Finger tables
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
