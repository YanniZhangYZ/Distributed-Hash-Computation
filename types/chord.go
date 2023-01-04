package types

import "fmt"

// -----------------------------------------------------------------------------
// ChordQueryMessage

// NewEmpty implements types.Message.
func (c ChordQueryMessage) NewEmpty() Message {
	return &ChordQueryMessage{}
}

// Name implements types.Message.
func (c ChordQueryMessage) Name() string {
	return "chordquery"
}

// String implements types.Message.
func (c ChordQueryMessage) String() string {
	return fmt.Sprintf("{chordquery %d from %s}", c.Key, c.Source)
}

// HTML implements types.Message.
func (c ChordQueryMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// ChordReplyMessage

// NewEmpty implements types.Message.
func (c ChordReplyMessage) NewEmpty() Message {
	return &ChordReplyMessage{}
}

// Name implements types.Message.
func (c ChordReplyMessage) Name() string {
	return "chordreply"
}

// String implements types.Message.
func (c ChordReplyMessage) String() string {
	return fmt.Sprintf("{chordreply for packet: %s}", c.ReplyPacketID)
}

// HTML implements types.Message.
func (c ChordReplyMessage) HTML() string {
	return c.String()
}
