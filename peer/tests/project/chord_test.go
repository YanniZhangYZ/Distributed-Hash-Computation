package project

import (
	"github.com/stretchr/testify/require"
	z "go.dedis.ch/cs438/internal/testing"
	"go.dedis.ch/cs438/transport/channel"
	"testing"
	"time"
)

// Test_Chord_Create tests a chord node is correctly initiated
func Test_Chord_Create(t *testing.T) {
	transp := channel.NewTransport()

	node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithTotalPeers(1), z.WithChordBytes(4),
		z.WithChordStabilizeInterval(0), z.WithChordFixFingerInterval(0))
	defer node.Stop()

	predecessor := node.GetPredecessor()
	require.Equal(t, "", predecessor)

	successor := node.GetSuccessor()
	require.Equal(t, "", successor)

	// > The length of the finger tables should be the number of chord bits
	fingers := node.GetFingerTable()
	require.Equal(t, 4*8, len(fingers))
}

// Test_Chord_Join_Simple tests a chord node joins the Chord ring of another node
func Test_Chord_Join_Simple(t *testing.T) {
	transp := channel.NewTransport()

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithTotalPeers(1), z.WithChordBytes(4),
		z.WithChordStabilizeInterval(0), z.WithChordFixFingerInterval(0))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithTotalPeers(1), z.WithChordBytes(4),
		z.WithChordStabilizeInterval(0), z.WithChordFixFingerInterval(0))
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	err := node1.JoinChord(node2.GetAddr())
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	n1Ins := node1.GetIns()
	n2Ins := node2.GetIns()

	//n1Outs := node1.GetOuts()
	//n2Outs := node2.GetOuts()

	// > n2 should have received a Query from n1
	require.Len(t, n2Ins, 1)
	pkt := n2Ins[0]
	require.Equal(t, "chordquery", pkt.Msg.Type)

	// > n1 should have received a Reply from n2
	require.Len(t, n1Ins, 1)
	pkt = n1Ins[0]
	require.Equal(t, "chordreply", pkt.Msg.Type)

	// After the join, node1 must have node2 as its successor, the predecessor should remain
	// empty for both nodes. It will be set by the stabilization daemon
	predecessor1 := node1.GetPredecessor()
	require.Equal(t, "", predecessor1)

	predecessor2 := node1.GetPredecessor()
	require.Equal(t, "", predecessor2)

	successor1 := node1.GetSuccessor()
	require.Equal(t, node2.GetAddr(), successor1)

	successor2 := node2.GetSuccessor()
	require.Equal(t, "", successor2)
}
