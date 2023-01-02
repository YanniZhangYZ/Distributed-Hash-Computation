package types

import "fmt"

// -----------------------------------------------------------------------------
// PaxosPrepareMessage

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
	return fmt.Sprintf("{chordquery %d}", c.Key)
}

// HTML implements types.Message.
func (c ChordQueryMessage) HTML() string {
	return c.String()
}
