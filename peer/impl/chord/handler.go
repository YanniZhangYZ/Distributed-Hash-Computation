package chord

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

func (c *Chord) execChordQuerySuccMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordQueryMsg, ok := msg.(*types.ChordQuerySuccessorMessage)
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
		c.successorLock.RLock()
		defer c.successorLock.RUnlock()

		replySuccessor := c.successor
		if replySuccessor == "" {
			// The initial state of the chord ring, we are the only node inside the ring, therefore, we
			// should return our address as the successor in the reply
			replySuccessor = c.address
		}

		// Prepare the new chord reply message
		chordReplyMsg := types.ChordReplySuccessorMessage{
			ReplyPacketID: chordQueryMsg.RequestID,
			Successor:     replySuccessor,
		}
		chordReplyMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordReplyMsg)
		if err != nil {
			return err
		}
		return c.message.Unicast(chordQueryMsg.Source, chordReplyMsgTrans)
	}

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

func (c *Chord) execChordReplySuccMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordReplyMsg, ok := msg.(*types.ChordReplySuccessorMessage)
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

func (c *Chord) execChordQueryPredMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	_, ok := msg.(*types.ChordQueryPredecessorMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	c.predecessorLock.RLock()
	defer c.predecessorLock.RUnlock()

	predecessor := c.predecessor
	chordReplyMsg := types.ChordReplyPredecessorMessage{
		Predecessor: predecessor,
	}
	chordReplyMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordReplyMsg)
	if err != nil {
		return err
	}
	return c.message.Unicast(pkt.Header.Source, chordReplyMsgTrans)
}

func (c *Chord) execChordReplyPredMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordReplyMsg, ok := msg.(*types.ChordReplyPredecessorMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	if chordReplyMsg.Predecessor == "" {
		// If our successor has no predecessor set, we should directly notify our successor
	} else {

		// If our successor already has one predecessor, we should check whether our successor has
		// a new predecessor, and the new predecessor is within the range between our chordID, and
		// our successor's ID
		c.successorLock.Lock()
		defer c.successorLock.Unlock()
		predecessorID := c.name2ID(chordReplyMsg.Predecessor)
		successorID := c.name2ID(c.successor)
		within := false

		if successorID <= c.chordID {
			within = c.chordID < predecessorID || predecessorID < successorID
		} else {
			within = c.chordID < predecessorID && predecessorID < successorID
		}

		// If the predecessor has a key that is between us and our previous successor, then we should update our
		// successor to the new predecessor
		if within {
			c.successor = chordReplyMsg.Predecessor
			c.fingersLock.Lock()
			c.fingers[0] = chordReplyMsg.Predecessor
			c.fingersLock.Unlock()
		}
	}

	// Notify our successor the existence of us
	chordNotifyMsg := types.ChordNotifyMessage{}
	chordNotifyMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordNotifyMsg)
	if err != nil {
		return err
	}
	return c.message.Unicast(c.successor, chordNotifyMsgTrans)
}

func (c *Chord) execChordNotifyMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	_, ok := msg.(*types.ChordNotifyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	c.predecessorLock.Lock()
	defer c.predecessorLock.Unlock()

	if c.predecessor == "" {
		// If we don't have a predecessor yet
		c.predecessor = pkt.Header.Source
	} else {
		// If we already have a predecessor, check that the new coming one has an ID that is within
		// the range (oldPredecessorID, chordID)
		oldPredecessorID := c.name2ID(c.predecessor)
		newPredecessorID := c.name2ID(pkt.Header.Source)
		within := false

		if c.chordID < oldPredecessorID {
			within = oldPredecessorID < newPredecessorID || newPredecessorID < c.chordID
		} else {
			within = oldPredecessorID < newPredecessorID && newPredecessorID < c.chordID
		}

		// If the new predecessor has a key that is between our previous predecessor and us, then we should
		// update our predecessor to the new predecessor
		if within {
			c.predecessor = pkt.Header.Source
		}
	}

	c.successorLock.Lock()
	defer c.successorLock.Unlock()
	// If we haven't had a successor set, we should set our successor to the source
	// of the packet as well
	if c.successor == "" || c.successor == c.address {
		c.successor = pkt.Header.Source
		c.fingersLock.Lock()
		c.fingers[0] = pkt.Header.Source
		c.fingersLock.Unlock()
	}

	return nil
}
