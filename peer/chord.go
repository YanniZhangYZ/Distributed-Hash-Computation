package peer

// Chord defines the functions for the basic chord operations of a peer.
type Chord interface {
	// GetChordID gets the chordID of the current node
	GetChordID() uint
	
	// QueryChordID queries the chordID of the given address
	QueryChordID(string) uint

	// GetPredecessor gets the predecessor of the current node
	GetPredecessor() string

	// GetSuccessor gets the successor of the current node
	GetSuccessor() string

	// GetFingerTable gets the finger table of the current node
	GetFingerTable() []string

	// JoinChord joins the peer to an existing Chord ring
	JoinChord(string) error

	// LeaveChord allows the peer to leave a joined Chord ring
	LeaveChord() error

	// RingLen returns the number of nodes inside the Chord ring
	RingLen() uint
}
