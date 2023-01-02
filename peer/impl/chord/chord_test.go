package chord

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer"
	"math"
	"testing"
)

func TestName2ID(t *testing.T) {
	c := ChordModule{}
	c.conf = &peer.Configuration{}
	c.conf.ChordBytes = 2

	// The upper bound of the hash value should be 2^(ChordBytes * 8)
	// If ChordBytes = 1, upperBound = 256
	// If ChordBytes = 2, upperBound = 65536
	upperBound := int(math.Pow(2, float64(c.conf.ChordBytes)*8))

	for i := 0; i < upperBound; i++ {
		// The address is used for the chordId computation. If two nodes have different addresses,
		// it is likely that they also have two different chordIds. This feature is powered by
		// the collision resistance of the crypto-hash function
		
		c.address = fmt.Sprintf("127.0.0.1:{%d}", i)
		c.name2ID()

		// All hash values should be within the range 0 <= hash value < upperBound
		require.GreaterOrEqual(t, c.chordId, uint(0))
		require.Less(t, c.chordId, uint(upperBound))
	}
}
