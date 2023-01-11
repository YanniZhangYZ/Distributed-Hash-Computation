package chord

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer"
	"math"
	"testing"
)

// Test_Valid_Range tests the validRange function
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

// Test_Name2ID tests the Name2ID function
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
			chordID := c.Name2ID(c.address)

			// All hash values should be within the valid range: 0 <= hash value < upperBound
			require.Equal(t, true, c.validRange(chordID))
		}
	}
}

// Test_Is_Predecessor tests the isPredecessor function
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
		c.conf = &peer.Configuration{}
		c.conf.ChordBytes = 2
		c.address = "127.0.0.0:1"
		c.successor = "127.0.0.4"
		c.chordID = c.Name2ID(c.address)
		upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))

		addressID := c.Name2ID(c.address)
		successorID := c.Name2ID(c.successor)

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
		c.conf = &peer.Configuration{}
		c.conf.ChordBytes = 2
		c.address = "127.0.0.0:1"
		c.successor = "127.0.0.2"
		c.chordID = c.Name2ID(c.address)
		upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))

		addressID := c.Name2ID(c.address)
		successorID := c.Name2ID(c.successor)

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

// Test_Finger_Start_End tests the fingerStartEnd function
func Test_Finger_Start_End(t *testing.T) {
	c := Chord{}
	c.conf = &peer.Configuration{}
	c.conf.ChordBytes = 1
	c.address = "127.0.0.1:1"
	c.chordID = c.Name2ID(c.address) // chordID = 97 for ChordBytes = 1

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

// Test_Closest_Preceding_Finger tests the closestPrecedingFinger function
func Test_Closest_Preceding_Finger(t *testing.T) {
	c := Chord{}
	c.conf = &peer.Configuration{}
	c.conf.ChordBytes = 1
	c.address = "10"
	c.chordID = c.Name2ID(c.address) // ChordID = 74 for address = "10"
	c.fingers = make([]string, c.conf.ChordBytes*8)

	c.fingers[0] = "4"   // ChordID = 75    (75  - 74 = 1)    for address = "4"
	c.fingers[1] = "257" // ChordID = 76    (76  - 74 = 2)    for address = "257"
	c.fingers[2] = "3"   // ChordID = 78    (78  - 74 = 4)    for address = "3"
	c.fingers[3] = "184" // ChordID = 82    (82  - 74 = 8)    for address = "184"
	c.fingers[4] = "199" // ChordID = 90    (90  - 74 = 16)   for address = "199"
	c.fingers[5] = "124" // ChordID = 106   (106 - 74 = 32)   for address = "124"
	c.fingers[6] = "238" // ChordID = 138   (138 - 74 = 64)   for address = "238"
	c.fingers[7] = "451" // ChordID = 202   (202 - 74 = 128)  for address = "451"

	require.Equal(t, c.fingers[0], c.closestPrecedingFinger(76))
	for i := uint(0); i < 2; i++ {
		require.Equal(t, c.fingers[1], c.closestPrecedingFinger(77+i))
	}
	for i := uint(0); i < 4; i++ {
		require.Equal(t, c.fingers[2], c.closestPrecedingFinger(79+i))
	}
	for i := uint(0); i < 8; i++ {
		require.Equal(t, c.fingers[3], c.closestPrecedingFinger(83+i))
	}
	for i := uint(0); i < 16; i++ {
		require.Equal(t, c.fingers[4], c.closestPrecedingFinger(91+i))
	}
	for i := uint(0); i < 32; i++ {
		require.Equal(t, c.fingers[5], c.closestPrecedingFinger(107+i))
	}
	for i := uint(0); i < 64; i++ {
		require.Equal(t, c.fingers[6], c.closestPrecedingFinger(139+i))
	}
	for i := uint(0); i < 128; i++ {
		require.Equal(t, c.fingers[7], c.closestPrecedingFinger((203+i)%256))
	}
}
