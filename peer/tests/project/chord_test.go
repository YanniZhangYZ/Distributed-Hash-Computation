package project

import (
	"fmt"
	"github.com/stretchr/testify/require"
	z "go.dedis.ch/cs438/internal/testing"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

// Test_Chord_Create tests a chord node is correctly initiated
func Test_Chord_Create(t *testing.T) {
	transp := channelFac()

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
// entries inside the finger tables to itself, except for the successor
func Test_Chord_Create_With_Daemon(t *testing.T) {
	transp := channelFac()

	node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(0), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node.Stop()

	// With ChordBytes = 1, we have 8 entries in finger table, we should wait at least 8 * FixFingerInterval,
	// to see all finger table entries get updated. The successor should also be updated accordingly.
	time.Sleep(time.Second * 2)

	predecessor := node.GetPredecessor()
	require.Equal(t, "", predecessor)

	successor := node.GetSuccessor()
	require.Equal(t, "", successor)

	// > The length of the finger tables should be the number of chord bits
	fingers := node.GetFingerTable()
	require.Equal(t, 8, len(fingers))
	for i := 1; i < 8; i++ {
		require.Equal(t, node.GetAddr(), fingers[i])
	}
}

// Test_Chord_Join_Simple tests a chord node joins the Chord ring of another node, with all daemons
// disabled
func Test_Chord_Join_Simple(t *testing.T) {
	transp := channelFac()

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

	time.Sleep(time.Second * 2)

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
	transp := channelFac()

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

	ringLen1 := node1.RingLen()
	require.Equal(t, uint(2), ringLen1)

	ringLen2 := node2.RingLen()
	require.Equal(t, uint(2), ringLen2)
}

// Test_Chord_Join_With_Stabilization_Fix_Finger tests a chord node joins the Chord ring of another
// node, and both the stabilization daemon and fix finger daemon are turned on. Both nodes should
// have the correct successor and predecessor set, after the stabilization interval. Also, it should
// have all finger table entry set after (ChordBytes * 8 * FixFingerInterval)
func Test_Chord_Join_With_Stabilization_Fix_Finger(t *testing.T) {
	transp := channelFac()

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

	ringLen1 := node1.RingLen()
	require.Equal(t, uint(2), ringLen1)

	ringLen2 := node2.RingLen()
	require.Equal(t, uint(2), ringLen2)
}

// Test_Chord_Join_Three_Node tests a chord ring consisting of 3 nodes
// For chordBytes = 2, Node 1 has chord ID 24963, Node 2 has chord ID 25694, Node 3 has chord ID 14865
// Therefore, the Chord ring should look like as follows
// Topology:
//
//	Node3 (14865)  ---->  Node1 (24963)
//	  ↖--- Node2 (25694) ---↙
func Test_Chord_Join_Three_Node(t *testing.T) {
	transp := channelFac()

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:1", z.WithChordBytes(2),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:2", z.WithChordBytes(2),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node2.Stop()

	node3 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:3", z.WithChordBytes(2),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
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

	time.Sleep(time.Second * 5)

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

	// The finger table of each node should be as follows
	// Node 1:
	//     finger 0 - 9: Node 2
	//     finger 10 - 15: Node 3
	// Node 2:
	//     finger 0 - 15: Node 3
	// Node 3:
	//     finger 0 - 13: Node 1
	//     finger 14 - 15: Node 3
	finger1 := node1.GetFingerTable()
	for i := 0; i < 10; i++ {
		require.Equal(t, node2.GetAddr(), finger1[i])
	}
	for i := 10; i < 16; i++ {
		require.Equal(t, node3.GetAddr(), finger1[i])
	}

	finger2 := node2.GetFingerTable()
	for i := 0; i < 16; i++ {
		require.Equal(t, node3.GetAddr(), finger2[i])
	}

	finger3 := node3.GetFingerTable()
	for i := 0; i < 14; i++ {
		require.Equal(t, node1.GetAddr(), finger3[i])
	}
	for i := 14; i < 16; i++ {
		require.Equal(t, node3.GetAddr(), finger3[15])
	}

	ringLen1 := node1.RingLen()
	require.Equal(t, uint(3), ringLen1)

	ringLen2 := node2.RingLen()
	require.Equal(t, uint(3), ringLen2)

	ringLen3 := node3.RingLen()
	require.Equal(t, uint(3), ringLen3)
}

// Test_Chord_Join_Multiple_Node tests the case of multiple nodes join a Chord ring. The correctness verification
// is done through sorting the nodes by its ChordID, and checks the predecessor and successor follows the sorted
// order. The finger table could also be computed based on the sorted ChordID. Besides, the ring length of all
// nodes should be equal to the total number of nodes
func Test_Chord_Join_Multiple_Node(t *testing.T) {
	numNodes := 16
	chordBytes := 1
	transp := channelFac()

	nodes := make([]z.TestNode, numNodes)
	for i := range nodes {
		node := z.NewTestNode(t, peerFac, transp, fmt.Sprintf("127.0.0.1:%d", i+1), z.WithChordBytes(chordBytes),
			z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
		defer node.Stop()
		nodes[i] = node
	}

	for _, n1 := range nodes {
		for _, n2 := range nodes {
			n1.AddPeer(n2.GetAddr())
		}
	}

	for i := 1; i < numNodes; i++ {
		err := nodes[i].JoinChord(nodes[i-1].GetAddr())
		require.NoError(t, err)

		time.Sleep(time.Second)

		// Already joined Chord nodes should have ring length = i + 1
		for j := 0; j <= i; j++ {
			require.Equal(t, uint(i+1), nodes[j].RingLen())
		}
	}

	time.Sleep(time.Second * 60)

	// After every node gets stabilized, we check, for every node, its predecessor, successor, and finger table.
	// First, we sort the nodes based on ChordID
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetChordID() < nodes[j].GetChordID()
	})

	for i := 0; i < numNodes; i++ {
		require.Equal(t, nodes[i].GetPredecessor(), nodes[(i-1+numNodes)%numNodes].GetAddr())
		require.Equal(t, nodes[i].GetSuccessor(), nodes[(i+1)%numNodes].GetAddr())
		fingers := nodes[i].GetFingerTable()
		for j := 0; j < chordBytes*8; j++ {
			fingerStart := nodes[i].GetChordID() + uint(math.Pow(2, float64(j)))

			// Try to find the node that has a ChordID larger than the fingerStart, i.e., it should
			// be the node that goes into the finger table
			fingerIdx := -1
			for k := i + 1; k < numNodes && fingerIdx == -1; k++ {
				if nodes[k].GetChordID() >= fingerStart {
					fingerIdx = k
				}
			}

			for k := 0; k <= i && fingerIdx == -1; k++ {
				if nodes[k].GetChordID()+uint(math.Pow(2, float64(chordBytes)*8)) >= fingerStart {
					fingerIdx = k
				}
			}

			require.Equal(t, nodes[fingerIdx].GetAddr(), fingers[j])
		}
	}
}

// Test_Chord_Ring_Len tests the RingLen function. It should return the correct number of nodes
// inside a Chord ring
func Test_Chord_Ring_Len(t *testing.T) {
	transp := channelFac()

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:1", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:2", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node2.Stop()

	node3 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:3", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node3.Stop()

	node1.AddPeer(node2.GetAddr())
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node1.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())

	// Before node2 join node1, every node should have a ring length = 1
	require.Equal(t, uint(1), node1.RingLen())
	require.Equal(t, uint(1), node2.RingLen())
	require.Equal(t, uint(1), node3.RingLen())

	err := node2.JoinChord(node1.GetAddr())
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	// After the topology is stabilized, the ring length should be 2 for node1 and node2
	require.Equal(t, uint(2), node1.RingLen())
	require.Equal(t, uint(2), node2.RingLen())
	require.Equal(t, uint(1), node3.RingLen())

	err = node3.JoinChord(node2.GetAddr())
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	// After the topology is stabilized, the ring length should be 3 for all nodes
	require.Equal(t, uint(3), node1.RingLen())
	require.Equal(t, uint(3), node2.RingLen())
	require.Equal(t, uint(3), node3.RingLen())
}

// Test_Chord_Leave_Simple tests a node volunteers to leave, the other node should have the
// correct information after the leave is done.
func Test_Chord_Leave_Simple(t *testing.T) {
	transp := channelFac()

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:1", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:2", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node2.Stop()

	node3 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:3", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
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

	time.Sleep(time.Second * 5)

	// After 3 nodes forms a Chord ring, we should have a Chord ring length = 3
	ringLen1 := node1.RingLen()
	require.Equal(t, uint(3), ringLen1)

	ringLen2 := node2.RingLen()
	require.Equal(t, uint(3), ringLen2)

	ringLen3 := node3.RingLen()
	require.Equal(t, uint(3), ringLen3)

	err = node3.LeaveChord()
	require.NoError(t, err)

	time.Sleep(time.Second * 5)

	// After node3 leaves, node1 and node2 should have each other as its successor and predecessor. The
	// finger table entries should also be updated as well.
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

	ringLen1 = node1.RingLen()
	require.Equal(t, uint(2), ringLen1)

	ringLen2 = node2.RingLen()
	require.Equal(t, uint(2), ringLen2)
}

// Test_Chord_Leave_Multiple_Node tests a node volunteers to leave, the other node should have the
// correct information after the leave is done. Now, there would be multiple nodes inside the system.
func Test_Chord_Leave_Multiple_Node(t *testing.T) {
	numNodes := 16
	chordBytes := 1
	transp := channelFac()

	nodes := make([]z.TestNode, numNodes)
	for i := range nodes {
		node := z.NewTestNode(t, peerFac, transp, fmt.Sprintf("127.0.0.1:%d", i+1), z.WithChordBytes(chordBytes),
			z.WithChordStabilizeInterval(time.Millisecond*500), z.WithChordFixFingerInterval(time.Millisecond*500),
			z.WithChordPingInterval(time.Second*5))
		defer node.Stop()
		nodes[i] = node
	}

	for _, n1 := range nodes {
		for _, n2 := range nodes {
			n1.AddPeer(n2.GetAddr())
		}
	}

	for i := 1; i < numNodes; i++ {
		err := nodes[i].JoinChord(nodes[i-1].GetAddr())
		require.NoError(t, err)
	}

	time.Sleep(time.Second * 30)

	// After every node gets stabilized, every node should have a ring length = numNodes.
	// First, we sort the nodes based on ChordID
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetChordID() < nodes[j].GetChordID()
	})

	for i := 0; i < numNodes; i++ {
		require.Equal(t, uint(numNodes), nodes[i].RingLen())
	}

	// Nodes leave in sequence, the remaining nodes should still have the correct information to continue
	for leaveIdx, sumLeave := numNodes-1, 1; leaveIdx >=
		(sumLeave+numNodes-1)/2; leaveIdx, sumLeave = leaveIdx-1, sumLeave+1 {

		leaveTest := func(t *testing.T) {
			nodes[leaveIdx].LeaveChord()
			numNodes--
			time.Sleep(time.Second * 30)

			for i := 0; i < numNodes; i++ {
				require.Equal(t, nodes[i].GetPredecessor(), nodes[(i-1+numNodes)%numNodes].GetAddr())
				require.Equal(t, nodes[i].GetSuccessor(), nodes[(i+1)%numNodes].GetAddr())
				fingers := nodes[i].GetFingerTable()
				for j := 0; j < chordBytes*8; j++ {
					fingerStart := nodes[i].GetChordID() + uint(math.Pow(2, float64(j)))

					// Try to find the node that has a ChordID larger than the fingerStart, i.e., it should
					// be the node that goes into the finger table
					fingerIdx := -1
					for k := i + 1; k < numNodes && fingerIdx == -1; k++ {
						if nodes[k].GetChordID() >= fingerStart {
							fingerIdx = k
						}
					}

					for k := 0; k <= i && fingerIdx == -1; k++ {
						if nodes[k].GetChordID()+uint(math.Pow(2, float64(chordBytes)*8)) >= fingerStart {
							fingerIdx = k
						}
					}

					require.Equal(t, nodes[fingerIdx].GetAddr(), fingers[j])
				}
			}
		}
		t.Run(fmt.Sprintf("%d node(s) leave from %d nodes", sumLeave, sumLeave+numNodes-1), leaveTest)
	}
}

// Test_Chord_Stress tests the Chord functions under stressful environment
func Test_Chord_Stress(t *testing.T) {
	numNodes := 32
	chordBytes := 2
	transp := channelFac()

	nodes := make([]z.TestNode, numNodes)
	for i := range nodes {
		node := z.NewTestNode(t, peerFac, transp, fmt.Sprintf("127.0.0.1:%d", i+1), z.WithChordBytes(chordBytes),
			z.WithChordStabilizeInterval(time.Second),
			z.WithChordFixFingerInterval(time.Second),
			z.WithHeartbeat(0),
			z.WithAntiEntropy(0))
		defer node.Stop()
		nodes[i] = node
	}

	for i := 1; i < numNodes; i++ {
		err := nodes[i].JoinChord(nodes[i-1].GetAddr())
		require.NoError(t, err)
		time.Sleep(time.Millisecond * time.Duration(1000+rand.Intn(1000)))
	}

	time.Sleep(time.Second * 60)

	// After every node gets stabilized, we check, for every node, its predecessor, successor, and finger table.
	// First, we sort the nodes based on ChordID
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetChordID() < nodes[j].GetChordID()
	})

	for i := 0; i < numNodes; i++ {
		require.Equal(t, nodes[i].GetPredecessor(), nodes[(i-1+numNodes)%numNodes].GetAddr())
		require.Equal(t, nodes[i].GetSuccessor(), nodes[(i+1)%numNodes].GetAddr())
		fingers := nodes[i].GetFingerTable()
		for j := 0; j < chordBytes*8; j++ {
			fingerStart := nodes[i].GetChordID() + uint(math.Pow(2, float64(j)))

			// Try to find the node that has a ChordID larger than the fingerStart, i.e., it should
			// be the node that goes into the finger table
			fingerIdx := -1
			for k := i + 1; k < numNodes && fingerIdx == -1; k++ {
				if nodes[k].GetChordID() >= fingerStart {
					fingerIdx = k
				}
			}

			for k := 0; k <= i && fingerIdx == -1; k++ {
				if nodes[k].GetChordID()+uint(math.Pow(2, float64(chordBytes)*8)) >= fingerStart {
					fingerIdx = k
				}
			}

			require.Equal(t, nodes[fingerIdx].GetAddr(), fingers[j])
		}
	}
}
