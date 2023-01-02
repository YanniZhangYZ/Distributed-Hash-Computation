package chord

import (
	"go.dedis.ch/cs438/peer"
)

type ChordModule struct {
	address     string
	conf        *peer.Configuration // The configuration contains Socket and MessageRegistry
	message     *peer.Messaging     // Messaging used to communicate among nodes
	successors  []string            // List of successors of this chord node
	fingers     []string            // Finger tables
	predecessor string              // Predecessor of this node
}

// Create creates a new chord ring topology
func (c *ChordModule) Create() {
	c.successors = make([]string, c.conf.ChordNumSuccessors)
	c.predecessor = ""
}
