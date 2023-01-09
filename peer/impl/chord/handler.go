package chord

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

// execChordQuerySuccMessage is the callback function to handle ChordQuerySuccessorMessage
func (c *Chord) execChordQuerySuccMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordQueryMsg, ok := msg.(*types.ChordQuerySuccessorMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
	}

	// The queried key should be within the range of chord bits, if it is, ignore the packet
	if !c.validRange(chordQueryMsg.Key) {
		return nil
	}

	c.successorLock.RLock()
	defer c.successorLock.RUnlock()
	c.fingersLock.RLock()
	defer c.fingersLock.RUnlock()

	// Check that whether we are the direct predecessor of the key, if we are, return our successor
	isPredecesor := c.isPredecessor(chordQueryMsg.Key)
	if isPredecesor {
		// If we are the predecessor, return our successor directly to the source of query using Unicast
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

// execChordReplySuccMessage is the callback function to handle ChordReplySuccessorMessage
func (c *Chord) execChordReplySuccMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordReplyMsg, ok := msg.(*types.ChordReplySuccessorMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
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

// execChordQueryPredMessage is the callback function to handle ChordQueryPredecessorMessage
func (c *Chord) execChordQueryPredMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	_, ok := msg.(*types.ChordQueryPredecessorMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
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

// execChordReplyPredMessage is the callback function to handle ChordReplyPredecessorMessage
func (c *Chord) execChordReplyPredMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordReplyMsg, ok := msg.(*types.ChordReplyPredecessorMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
	}

	c.successorLock.Lock()
	defer c.successorLock.Unlock()
	c.fingersLock.Lock()
	defer c.fingersLock.Unlock()

	if chordReplyMsg.Predecessor == "" {
		// If our successor has no predecessor set, we should directly notify our successor
	} else {

		// If our successor already has one predecessor, we should check whether our successor has
		// a new predecessor, and the new predecessor is within the range between our chordID, and
		// our successor's ID
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
			c.fingers[0] = chordReplyMsg.Predecessor
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

// execChordNotifyMessage is the callback function to handle ChordNotifyMessage
func (c *Chord) execChordNotifyMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	_, ok := msg.(*types.ChordNotifyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
	}

	c.predecessorLock.Lock()
	defer c.predecessorLock.Unlock()
	update := false

	if c.predecessor == "" {
		// If we don't have a predecessor yet
		c.predecessor = pkt.Header.Source
		update = true
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
			update = true
		}
	}

	// If we have updated our predecessor, it means the range we are responsible is changed, we should notify
	// our password cracker about the change
	if update {
		c.notifyPasswordCracker()
	}

	// If we haven't had a successor set, we should set our successor to the source
	// of the packet as well
	c.successorLock.Lock()
	defer c.successorLock.Unlock()
	c.fingersLock.Lock()
	defer c.fingersLock.Unlock()
	if c.successor == "" || c.successor == c.address {
		c.successor = pkt.Header.Source
		c.fingers[0] = pkt.Header.Source
	}

	return nil
}

// execChordRingLenMessage is the callback function to handle ChordRingLenMessage
func (c *Chord) execChordRingLenMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordRingLenMsg, ok := msg.(*types.ChordRingLenMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
	}

	if chordRingLenMsg.Source == c.address {
		// If we are the one who initiates the ring length query, we should return the results, if we are still waiting
		// for the result
		ringLenChan, ok := c.ringLenChan.Load(chordRingLenMsg.RequestID)
		if ok {
			ringLenChan.(chan uint) <- chordRingLenMsg.Length
		}
		return nil
	}
	// If we are not, we should increment the length by 1, and pass this message to our successor, if we have any
	c.successorLock.RLock()
	defer c.successorLock.RUnlock()
	if c.successor != "" && c.successor != c.address {
		chordRingLenMsg.Length++
		chordRingLenMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordRingLenMsg)
		if err != nil {
			return err
		}
		return c.message.Unicast(c.successor, chordRingLenMsgTrans)
	}
	return nil
}

// execChordClearPredMessage is the callback function to handle ChordClearPredecessorMessage
func (c *Chord) execChordClearPredMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	_, ok := msg.(*types.ChordClearPredecessorMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
	}

	// If our predecessor matches the pkt source, we are the correct receptor of the message, then
	// we should set our predecessor field to Nil, and wait for the NotifyMessage for new update
	c.predecessorLock.Lock()
	defer c.predecessorLock.Unlock()
	if c.predecessor == pkt.Header.Source {
		c.predecessor = ""
	}

	return nil
}

// execChordSkipSuccMessage is the callback function to handle ChordSkipSuccessorMessage
func (c *Chord) execChordSkipSuccMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordSkipSuccessorMsg, ok := msg.(*types.ChordSkipSuccessorMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
	}

	// If our successor matches the pkt source, we are the correct receptor of the message, then
	// we should update our successor to the new successor
	c.successorLock.Lock()
	defer c.successorLock.Unlock()
	c.fingersLock.Lock()
	defer c.fingersLock.Unlock()

	if c.successor == pkt.Header.Source {
		c.successor = chordSkipSuccessorMsg.Successor
		c.fingers[0] = chordSkipSuccessorMsg.Successor
	}

	return nil
}

// execChordPingMessage is the callback function to handle ChordPingMessage
func (c *Chord) execChordPingMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordPingMsg, ok := msg.(*types.ChordPingMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
	}

	// Prepare the new chord ping reply message
	chordPingReplyMsg := types.ChordPingReplyMessage{
		ReplyPacketID: chordPingMsg.RequestID,
	}
	chordPingReplyMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordPingReplyMsg)
	if err != nil {
		return err
	}
	return c.message.Unicast(pkt.Header.Source, chordPingReplyMsgTrans)
}

// execChordPingMessage is the callback function to handle ChordPingReplyMessage
func (c *Chord) execChordPingReplyMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chordPingReplyMsg, ok := msg.(*types.ChordPingReplyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// If we are not alive, even we receive some packets, ignore them
	if c.alive.Load() == 0 {
		return nil
	}

	pingChan, ok := c.pingChan.Load(chordPingReplyMsg.ReplyPacketID)
	if ok {
		pingChan.(chan bool) <- true
	}

	return nil
}
