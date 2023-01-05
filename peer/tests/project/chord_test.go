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

	node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithChordBytes(4),
		z.WithChordStabilizeInterval(0), z.WithChordFixFingerInterval(0))
	defer node.Stop()

	predecessor := node.GetPredecessor()
	require.Equal(t, "", predecessor)

	successor := node.GetSuccessor()
	require.Equal(t, "", successor)

	// > The length of the finger tables should be the number of chord bits
	fingers := node.GetFingerTable()
	require.Equal(t, 4*8, len(fingers))
	require.Equal(t, "", fingers[0])
}

// Test_Chord_Create_With_Fix_Finger tests a chord node is correctly initiated, and the fixFingerDaemon updates
// entries inside the finger tables to itself.
func Test_Chord_Create_With_Fix_Finger(t *testing.T) {
	transp := channel.NewTransport()

	node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(0), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node.Stop()

	// With ChordBytes = 1, we have 8 entries in finger table, we should wait at least 8 * FixFingerInterval,
	// to see all finger table entries get updated. The successor should also be updated accordingly.
	time.Sleep(time.Second * 2)

	predecessor := node.GetPredecessor()
	require.Equal(t, "", predecessor)

	successor := node.GetSuccessor()
	require.Equal(t, node.GetAddr(), successor)

	// > The length of the finger tables should be the number of chord bits
	fingers := node.GetFingerTable()
	require.Equal(t, 8, len(fingers))
	for i := 0; i < 8; i++ {
		require.Equal(t, node.GetAddr(), fingers[i])
	}
}

// Test_Chord_Join_Simple tests a chord node joins the Chord ring of another node, with all daemons
// disabled
func Test_Chord_Join_Simple(t *testing.T) {
	transp := channel.NewTransport()

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:1", z.WithChordBytes(4),
		z.WithChordStabilizeInterval(0), z.WithChordFixFingerInterval(0))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:2", z.WithChordBytes(4),
		z.WithChordStabilizeInterval(0), z.WithChordFixFingerInterval(0))
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	err := node1.JoinChord(node2.GetAddr())
	require.NoError(t, err)

	time.Sleep(time.Second)

	n1Ins := node1.GetIns()
	n2Ins := node2.GetIns()

	// > n2 should have received a QuerySucc from n1
	require.Len(t, n2Ins, 1)
	pkt := n2Ins[0]
	require.Equal(t, "chordquerysucc", pkt.Msg.Type)

	// > n1 should have received a ReplySucc from n2
	require.Len(t, n1Ins, 1)
	pkt = n1Ins[0]
	require.Equal(t, "chordreplysucc", pkt.Msg.Type)

	// After the join, node1 must have node2 as its successor, the predecessor should remain
	// empty for both nodes. It will be set by the stabilization daemon.
	predecessor1 := node1.GetPredecessor()
	require.Equal(t, "", predecessor1)

	predecessor2 := node2.GetPredecessor()
	require.Equal(t, "", predecessor2)

	successor1 := node1.GetSuccessor()
	require.Equal(t, node2.GetAddr(), successor1)

	successor2 := node2.GetSuccessor()
	require.Equal(t, "", successor2)
}

// Test_Chord_Join_With_Stabilization tests a chord node joins the Chord ring of another node, and the
// stabilization daemon is turned on. Both nodes should have the correct successor and predecessor
// set, after the stabilization interval.
func Test_Chord_Join_With_Stabilization(t *testing.T) {
	transp := channel.NewTransport()

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:1", z.WithChordBytes(4),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(0))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:2", z.WithChordBytes(4),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(0))
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	err := node1.JoinChord(node2.GetAddr())
	require.NoError(t, err)

	time.Sleep(time.Second)

	// After the join and stabilization, each node should have the other node as
	// its predecessor and successor
	predecessor1 := node1.GetPredecessor()
	require.Equal(t, node2.GetAddr(), predecessor1)

	predecessor2 := node2.GetPredecessor()
	require.Equal(t, node1.GetAddr(), predecessor2)

	successor1 := node1.GetSuccessor()
	require.Equal(t, node2.GetAddr(), successor1)

	successor2 := node2.GetSuccessor()
	require.Equal(t, node1.GetAddr(), successor2)
}

// Test_Chord_Join_With_Stabilization_Fix_Finger tests a chord node joins the Chord ring of another
// node, and both the stabilization daemon and fix finger daemon are turned on. Both nodes should
// have the correct successor and predecessor set, after the stabilization interval. Also, it should
// have all finger table entry set after (ChordBytes * 8 * FixFingerInterval)
func Test_Chord_Join_With_Stabilization_Fix_Finger(t *testing.T) {
	transp := channel.NewTransport()

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:1", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:2", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	err := node1.JoinChord(node2.GetAddr())
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	// After the join and stabilization, each node should have the other node as
	// its predecessor and successor
	predecessor1 := node1.GetPredecessor()
	require.Equal(t, node2.GetAddr(), predecessor1)

	predecessor2 := node2.GetPredecessor()
	require.Equal(t, node1.GetAddr(), predecessor2)

	successor1 := node1.GetSuccessor()
	require.Equal(t, node2.GetAddr(), successor1)

	successor2 := node2.GetSuccessor()
	require.Equal(t, node1.GetAddr(), successor2)

	// Node 1 has chordID = 97 and Node 2 has chordID = 100
	finger1 := node1.GetFingerTable()
	require.Equal(t, node2.GetAddr(), finger1[0])
	require.Equal(t, node2.GetAddr(), finger1[1])
	for i := 2; i < 8; i++ {
		require.Equal(t, node1.GetAddr(), finger1[i])
	}

	finger2 := node2.GetFingerTable()
	for i := 0; i < 8; i++ {
		require.Equal(t, node1.GetAddr(), finger2[i])
	}
}

// Test_Chord_Join_Three_Node tests a chord ring consisting of 3 nodes
// For chordBytes = 2, Node 1 has chord ID 24963, Node 2 has chord ID 25694, Node 3 has chord ID 14865
// Therefore, the Chord ring should look like as follows
// Topology:
//
//	Node3 (14865)  ---->  Node1 (24963)
//	  ↖--- Node2 (25694) ---↙
func Test_Chord_Join_Three_Node(t *testing.T) {
	transp := channel.NewTransport()

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:1", z.WithChordBytes(2),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(0))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:2", z.WithChordBytes(2),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(0))
	defer node2.Stop()

	node3 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:3", z.WithChordBytes(2),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(0))
	defer node3.Stop()

	node1.AddPeer(node2.GetAddr())
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node1.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())

	err := node2.JoinChord(node1.GetAddr())
	require.NoError(t, err)

	err = node3.JoinChord(node2.GetAddr())
	require.NoError(t, err)

	time.Sleep(time.Second)

	// After the join, 3 nodes should form a topology as follows
	// Topology:
	//	           Node3 (14865)  ---->  Node1 (24963)
	//	             ↖--- Node2 (25694) ---↙
	predecessor1 := node1.GetPredecessor()
	require.Equal(t, node3.GetAddr(), predecessor1)

	predecessor2 := node2.GetPredecessor()
	require.Equal(t, node1.GetAddr(), predecessor2)

	predecessor3 := node3.GetPredecessor()
	require.Equal(t, node2.GetAddr(), predecessor3)

	successor1 := node1.GetSuccessor()
	require.Equal(t, node2.GetAddr(), successor1)

	successor2 := node2.GetSuccessor()
	require.Equal(t, node3.GetAddr(), successor2)

	successor3 := node3.GetSuccessor()
	require.Equal(t, node1.GetAddr(), successor3)
}
