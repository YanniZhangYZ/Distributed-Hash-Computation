package chord

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer"
	"math"
	"testing"
)

func Test_Valid_Range(t *testing.T) {
	c := Chord{}
	c.conf = &peer.Configuration{}

	for chordBytes := 1; chordBytes < 3; chordBytes++ {
		c.conf.ChordBytes = chordBytes

		// The upper bound of the hash value should be 2^(ChordBytes * 8)
		// If ChordBytes = 1, upperBound = 256
		// If ChordBytes = 2, upperBound = 65536
		upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))

		for i := uint(0); i < upperBound; i++ {
			// All values within the range 0 <= hash value < upperBound, should be evaluated to true
			require.Equal(t, true, c.validRange(uint(i)))
		}

		for i := upperBound; i < 2*upperBound; i++ {
			// All values exceed the range, should be evaluated to false
			require.Equal(t, false, c.validRange(uint(i)))
		}
	}
}

func Test_Name2ID(t *testing.T) {
	c := Chord{}
	c.conf = &peer.Configuration{}

	for chordBytes := 1; chordBytes < 3; chordBytes++ {
		c.conf.ChordBytes = chordBytes

		// The upper bound of the hash value should be 2^(ChordBytes * 8)
		// If ChordBytes = 1, upperBound = 256
		// If ChordBytes = 2, upperBound = 65536
		upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))

		for i := uint(0); i < upperBound; i++ {
			// The address is used for the chordID computation. If two nodes have different addresses,
			// it is likely that they also have two different chordIds. This feature is powered by
			// the collision resistance of the crypto-hash function

			c.address = fmt.Sprintf("127.0.0.1:{%d}", i)
			chordID := c.name2ID(c.address)

			// All hash values should be within the valid range: 0 <= hash value < upperBound
			require.Equal(t, true, c.validRange(chordID))
		}
	}
}

func Test_Is_Predecessor(t *testing.T) {
	withoutSuccessor := func(t *testing.T) {
		// withoutSuccessor tests the case that only one node inside the Chord ring
		c := Chord{}
		c.conf = &peer.Configuration{}
		c.conf.ChordBytes = 2
		upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))

		for i := uint(0); i < upperBound; i++ {
			// Without any successor, we are the only node inside the ring, therefore,
			// all isPredecessor evaluates to true
			require.Equal(t, true, c.isPredecessor(uint(i)))
		}
	}

	withSuccessor := func(t *testing.T) {
		// withSuccessor tests the case that the node has a successor, and the successorID is
		// larger than its own ID
		c := Chord{}
		c.address = "127.0.0.0:1"
		c.successor = "127.0.0.4"
		c.conf = &peer.Configuration{}
		c.conf.ChordBytes = 2
		c.chordID = c.name2ID(c.address)
		upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))

		addressID := c.name2ID(c.address)
		successorID := c.name2ID(c.successor)

		for i := uint(0); i < upperBound; i++ {
			if addressID < i && i <= successorID {
				// If the value is within range (addressID, successorID], then we are the
				// predecessor of the given key
				require.Equal(t, true, c.isPredecessor(uint(i)))
			} else {
				// Else, we are not
				require.Equal(t, false, c.isPredecessor(uint(i)))
			}
		}
	}

	withSuccessorCrossBoundary := func(t *testing.T) {
		// withSuccessorCrossBoundary tests the case that the node has a successor, but the
		// successorID is smaller than its own ID, which means the covered range cross the
		// upperBound of the ring range
		c := Chord{}
		c.address = "127.0.0.0:1"
		c.successor = "127.0.0.2"
		c.conf = &peer.Configuration{}
		c.conf.ChordBytes = 2
		c.chordID = c.name2ID(c.address)
		upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))

		addressID := c.name2ID(c.address)
		successorID := c.name2ID(c.successor)

		for i := uint(0); i < upperBound; i++ {
			if i > addressID || i <= successorID {
				// If the value is within range (addressID, successorID], then we are the
				// predecessor of the given key
				require.Equal(t, true, c.isPredecessor(uint(i)))
			} else {
				// Else, we are not
				require.Equal(t, false, c.isPredecessor(uint(i)))
			}
		}
	}

	t.Run("Without successor", withoutSuccessor)
	t.Run("With successor", withSuccessor)
	t.Run("With successor and range crossing boundary", withSuccessorCrossBoundary)
}

func Test_FingerStartEnd(t *testing.T) {
	c := Chord{}
	c.address = "127.0.0.1:1"
	c.conf = &peer.Configuration{}
	c.conf.ChordBytes = 1
	c.chordID = c.name2ID(c.address) // chordID = 97 for ChordBytes = 1

	fingerStart, fingerEnd := c.fingerStartEnd(0)
	require.Equal(t, fingerStart, uint(98))
	require.Equal(t, fingerEnd, uint(99))

	fingerStart, fingerEnd = c.fingerStartEnd(1)
	require.Equal(t, fingerStart, uint(99))
	require.Equal(t, fingerEnd, uint(101))

	fingerStart, fingerEnd = c.fingerStartEnd(3)
	require.Equal(t, fingerStart, uint(105))
	require.Equal(t, fingerEnd, uint(113))

	fingerStart, fingerEnd = c.fingerStartEnd(7)
	require.Equal(t, fingerStart, uint(225))
	require.Equal(t, fingerEnd, uint(97))
}
