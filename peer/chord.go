package peer

// Chord defines the functions for the basic chord operations of a peer.
type Chord interface {
	// GetPredecessor gets the predecessor fo the current node
	GetPredecessor() string

	// GetSuccessor gets the successor of the current node
	GetSuccessor() string

	// GetFingerTable gets the finger table of the current node
	GetFingerTable() []string

	// JoinChord joins the peer to an existing Chord ring
	JoinChord(remoteNode string) error
}
