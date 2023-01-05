package types

import "fmt"

// -----------------------------------------------------------------------------
// ChordQuerySuccessorMessage

// NewEmpty implements types.Message.
func (c ChordQuerySuccessorMessage) NewEmpty() Message {
	return &ChordQuerySuccessorMessage{}
}

// Name implements types.Message.
func (c ChordQuerySuccessorMessage) Name() string {
	return "chordquerysucc"
}

// String implements types.Message.
func (c ChordQuerySuccessorMessage) String() string {
	return fmt.Sprintf("{chordquerysuccessor %d from %s}", c.Key, c.Source)
}

// HTML implements types.Message.
func (c ChordQuerySuccessorMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// ChordReplySuccessorMessage

// NewEmpty implements types.Message.
func (c ChordReplySuccessorMessage) NewEmpty() Message {
	return &ChordReplySuccessorMessage{}
}

// Name implements types.Message.
func (c ChordReplySuccessorMessage) Name() string {
	return "chordreplysucc"
}

// String implements types.Message.
func (c ChordReplySuccessorMessage) String() string {
	return fmt.Sprintf("{chordreply for packet: %s}", c.ReplyPacketID)
}

// HTML implements types.Message.
func (c ChordReplySuccessorMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// ChordQueryPredecessorMessage

// NewEmpty implements types.Message.
func (c ChordQueryPredecessorMessage) NewEmpty() Message {
	return &ChordQueryPredecessorMessage{}
}

// Name implements types.Message.
func (c ChordQueryPredecessorMessage) Name() string {
	return "chordquerypred"
}

// String implements types.Message.
func (c ChordQueryPredecessorMessage) String() string {
	return "{chordquerypred}"
}

// HTML implements types.Message.
func (c ChordQueryPredecessorMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// ChordReplyPredecessorMessage

// NewEmpty implements types.Message.
func (c ChordReplyPredecessorMessage) NewEmpty() Message {
	return &ChordReplyPredecessorMessage{}
}

// Name implements types.Message.
func (c ChordReplyPredecessorMessage) Name() string {
	return "chordreplypred"
}

// String implements types.Message.
func (c ChordReplyPredecessorMessage) String() string {
	return "{chordreplypred}"
}

// HTML implements types.Message.
func (c ChordReplyPredecessorMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// ChordNotifyMessage

// NewEmpty implements types.Message.
func (c ChordNotifyMessage) NewEmpty() Message {
	return &ChordNotifyMessage{}
}

// Name implements types.Message.
func (c ChordNotifyMessage) Name() string {
	return "chordnotify"
}

// String implements types.Message.
func (c ChordNotifyMessage) String() string {
	return "{chordnotify}"
}

// HTML implements types.Message.
func (c ChordNotifyMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// ChordRingLenMessage

// NewEmpty implements types.Message.
func (c ChordRingLenMessage) NewEmpty() Message {
	return &ChordRingLenMessage{}
}

// Name implements types.Message.
func (c ChordRingLenMessage) Name() string {
	return "chordringlen"
}

// String implements types.Message.
func (c ChordRingLenMessage) String() string {
	return fmt.Sprintf("{chordringlen from %s with length %d}", c.Source, c.Length)
}

// HTML implements types.Message.
func (c ChordRingLenMessage) HTML() string {
	return c.String()
}
