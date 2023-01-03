package chord

import (
	"github.com/rs/xid"
	"go.dedis.ch/cs438/types"
)

// querySuccessor queries a remote node about the successor of the given key, it can be used
// either when a new node joins the ring, or the node queries the object
func (c *Chord) querySuccessor(remoteNode string, key uint) (string, error) {
	// Prepare the new chord query message
	chordQueryMsg := types.ChordQueryMessage{
		RequestID: xid.New().String(),
		Source:    c.address,
		Key:       key,
	}
	chordQueryMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordQueryMsg)
	if err != nil {
		return "", err
	}

	replyChan := make(chan string, 1)
	c.queryChan.Store(chordQueryMsg.RequestID, replyChan)

	err = c.message.Unicast(remoteNode, chordQueryMsgTrans)
	if err != nil {
		return "", err
	}

	return "", nil
}
