package chord

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

func (c *Chord) execChordQueryMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordQueryMsg, ok := msg.(*types.ChordQueryMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// The queried key should be within the range of chord bits, if it is, ignore the packet
	if !c.validRange(chordQueryMsg.Key) {
		return nil
	}

	// Check that whether we are the direct predecessor of the key, if we are, return our successor
	isPredecesor := c.isPredecessor(chordQueryMsg.Key)
	if isPredecesor {
		// If we are the predecessor, return our successor directly to the source of query using Unicast
		c.successorLock.Lock()
		defer c.successorLock.Unlock()

		replySuccessor := c.successor
		if replySuccessor == "" {
			// The initial state of the chord ring, we are the only node inside the ring, therefore, we
			// should return our address as the successor in the reply
			replySuccessor = c.address
		}

		// Prepare the new chord reply message
		chordReplyMsg := types.ChordReplyMessage{
			ReplyPacketID: chordQueryMsg.RequestID,
			Successor:     replySuccessor,
		}
		chordReplyMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordReplyMsg)
		if err != nil {
			return err
		}
		return c.message.Unicast(chordQueryMsg.Source, chordReplyMsgTrans)
	} else {
		// If we are not the predecessor, continue asking other nodes
		// First, we have to find in our finger table, which node has the closest preceding ID
		closestPrecedingFinger := c.closestPrecedingFinger(chordQueryMsg.Key)

		// The chord query message should be kept the same as before, but we forward it to the
		// closest preceding finger node
		chordQueryMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordQueryMsg)
		if err != nil {
			return err
		}
		return c.message.Unicast(closestPrecedingFinger, chordQueryMsgTrans)
	}
}

func (c *Chord) execChordReplyMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordReplyMsg, ok := msg.(*types.ChordReplyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// We receive a reply for our queries. Since the reply is sent directly via Unicast, we are sure that
	// we are the correct receptor of the message. Upon receiving the packet, we should notify the thread
	// that is waiting for our reply by loading the channel from the map and send the successor, if we are
	// still waiting for it.
	queryChan, ok := c.queryChan.Load(chordReplyMsg.ReplyPacketID)
	if ok {
		queryChan.(chan string) <- chordReplyMsg.Successor
	}

	return nil
}
