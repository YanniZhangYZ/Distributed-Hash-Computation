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
