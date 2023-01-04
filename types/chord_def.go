package types

// ChordQueryMessage describes a message sent to request a key.
//
// - implements types.Message
type ChordQueryMessage struct {
	// RequestID must be a unique identifier. Use xid.New().String() to generate
	// it.
	RequestID string

	// Source is the address of the peer that initiate the query
	Source string

	// Key is the key to query
	Key uint
}

// ChordReplyMessage describes a reply message to the ChordQueryMessage, it includes the
// indicator ReplyPacketID (which query it replies to), and the answer to the query, Successor.
//
// - implements types.Message
type ChordReplyMessage struct {
	// ReplyPacketID is the PacketID this reply is for
	ReplyPacketID string

	// Successor is the answer to the query, i.e., which successor the query key belongs to
	Successor string
}
