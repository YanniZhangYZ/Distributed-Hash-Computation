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

// Test_Password_Cracker_Invalid tests a node submits an invalid task
func Test_Password_Cracker_Invalid(t *testing.T) {
	transp := udpFac()
	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node1.Stop()

	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithChordBytes(1),
		z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	err := node1.JoinChord(node2.GetAddr())
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	// TEST 1, node 1 submit an invalid hash str, it should trigger an error
	hashStr := "---------"
	saltStr := "3c"
	err = node1.PasswordSubmitRequest(hashStr, saltStr, 0, 0)
	require.Error(t, err)

	// TEST 2, node 2 submit an invalid salt str, it should trigger an error
	hashStr = "49c13df5ec8821b2ec6973a83e077b5ca35ed93a55dc398aa3cb614ebae33d0f"
	saltStr = "--"
	err = node2.PasswordSubmitRequest(hashStr, saltStr, 0, 0)
	require.Error(t, err)
}

// Test_Password_Cracker_Simple tests a Chord ring formed by 2 nodes, each node should have the dictionary
// corresponding to its duty range
func Test_Password_Cracker_Simple(t *testing.T) {
	twoNode := func(t *testing.T) {
		transp := udpFac()
		node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithChordBytes(1),
			z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
		defer node1.Stop()

		node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0", z.WithChordBytes(1),
			z.WithChordStabilizeInterval(time.Millisecond*200), z.WithChordFixFingerInterval(time.Millisecond*200))
		defer node2.Stop()

		node1.AddPeer(node2.GetAddr())
		node2.AddPeer(node1.GetAddr())

		err := node1.JoinChord(node2.GetAddr())
		require.NoError(t, err)

		time.Sleep(time.Second * 2)

		totEntries := uint(math.Pow(2, float64(8)))
		chordID1 := node1.GetChordID()
		chordID2 := node2.GetChordID()
		if chordID1 == chordID2 {
			// Unlikely to happen, but if it happens, there is no correctness guarantee
			return
		}

		if chordID1 > chordID2 {
			require.Equal(t, int(chordID1-chordID2), node1.GetStorage().GetDictionaryStore().Len())
			require.Equal(t, int(totEntries-(chordID1-chordID2)), node2.GetStorage().GetDictionaryStore().Len())
		} else {
			require.Equal(t, int(totEntries-(chordID2-chordID1)), node1.GetStorage().GetDictionaryStore().Len())
			require.Equal(t, int(chordID2-chordID1), node2.GetStorage().GetDictionaryStore().Len())
		}

		// TEST 1, node 1 submit a password request, it has a corresponding password inside the dictionary
		hashStr := "1cfcd196cf51b7a1d44159875452ba2dca8898d675f3d33d610ab9cb0031d7b2"
		saltStr := "3c"
		err = node1.PasswordSubmitRequest(hashStr, saltStr, 0, 0)
		require.NoError(t, err)
		time.Sleep(time.Second)
		require.Equal(t, "apple", node1.PasswordReceiveResult(hashStr, saltStr))

		// TEST 2, node 2 submit a password request, it has a corresponding password inside the dictionary
		hashStr = "49c13df5ec8821b2ec6973a83e077b5ca35ed93a55dc398aa3cb614ebae33d0f"
		saltStr = "dd"
		err = node2.PasswordSubmitRequest(hashStr, saltStr, 0, 0)
		require.NoError(t, err)
		time.Sleep(time.Second)
		require.Equal(t, "egg", node2.PasswordReceiveResult(hashStr, saltStr))

		// TEST 3, node 2 submit a password request, it is not inside the dictionary
		hashStr = "349e4662785588a6cd0ebbd9dbb6cea0bbdbc71159d78901e33f758fafaf6a88"
		saltStr = "7f"
		err = node2.PasswordSubmitRequest(hashStr, saltStr, 0, 0)
		require.NoError(t, err)
		time.Sleep(time.Second)
		require.Equal(t, "", node2.PasswordReceiveResult(hashStr, saltStr))
	}

	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("Iteration %d", i), twoNode)
	}
}

// Test_Password_Cracker_Multiple_Without_Leave tests a Chord ring formed by multiple nodes, each node should have
// the dictionary corresponding to its duty range. No node leaves.
func Test_Password_Cracker_Multiple_Without_Leave(t *testing.T) {
	numNodes := 16
	transp := channelFac()

	nodes := make([]z.TestNode, numNodes)
	for i := range nodes {
		node := z.NewTestNode(t, peerFac, transp, fmt.Sprintf("127.0.0.1:%d", i+1), z.WithChordBytes(1),
			z.WithChordStabilizeInterval(time.Millisecond*500), z.WithChordFixFingerInterval(time.Millisecond*500))
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
	}

	time.Sleep(time.Second * 30)

	// First, we sort the nodes based on ChordID
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetChordID() < nodes[j].GetChordID()
	})

	totEntries := uint(math.Pow(2, float64(8)))
	for i := 1; i < numNodes; i++ {
		require.Equal(t, int(nodes[i].GetChordID()-nodes[i-1].GetChordID()),
			nodes[i].GetStorage().GetDictionaryStore().Len())
		totEntries -= nodes[i].GetChordID() - nodes[i-1].GetChordID()
	}
	require.Equal(t, int(totEntries), nodes[0].GetStorage().GetDictionaryStore().Len())

	// All following hash and salt pairs corresponds to "apple" in the dictionary
	hashStrs := []string{
		"1bd15226960ce500e8dbaabbd523b9356ec69ff1bdf2aeef6c5dbe272971986a",
		"2ea455a6c36bf264a0d933c2e9fa75e9962ceec80a55e118340d62c5b92cd930",
		"2ae367b9a8585e19e96f301bb3cb020941cd29a3925bb05c9dd12ac47c5757d6",
		"24a77708057aab975813079ab86b9b84ae300fd738f13fd6b425df0f8895b907",
		"a4ac87eb7080ed009b324a931235b27439f6e8f4ba51e2ed5c3c47f6962064d4",
		"4b446b30d876ca954a6d0a24f9d96db7b6a652465b28599d974f15033e87893f",
		"10d962ff38d51f366f7a0ffc2ab2f6497898230fb4b1c85bf1e03ff3226bfffa",
		"25b8005ff00096894f3d0dd7287efebef4f3d4ec45a19cc04159197368fa6ce1",
		"96e3f8a2bd177aa8d72f4bee847421b9555567c88bd940ee2b5cb7e25e022340",
		"865e185881d538d3e6f8158a5d7c2c58ca2a6523fa78e8916727682b83e9b13e",
		"86254408237a295b6aaa7b61b253689a50b635ab8b59a363e98ea2507d3d8e48",
		"67d5158c080528d5b50ad13426c6d6b56815fc2ade087f05bd39e4d0539bae34",
		"713da404459008f10a661b458fcc811f166f15579dfde8fd3c0d08de50fe8f03",
		"e7eabf2627703a7386b9433bf47734e7c6497dfb30712153e4f452074b0fa7ce",
		"38d9a74454af41571b5b20d9b26b7943dc191b3daadff8336dfe691e93b7a204",
		"a9bed160d86d2570e494cc39c095649d4816e76c1d31a183d3b63c205a25230c",
	}
	saltStrs := []string{"0f", "1f", "2f", "3f", "4f", "5f", "6f", "7f",
		"8f", "9f", "af", "bf", "cf", "df", "ef", "ff"}

	for i := 0; i < len(hashStrs); i++ {
		randomIdx := rand.Intn(len(nodes))
		err := nodes[randomIdx].PasswordSubmitRequest(hashStrs[i], saltStrs[i], 0, 0)
		require.NoError(t, err)
		time.Sleep(time.Second)
		require.Equal(t, "apple", nodes[randomIdx].PasswordReceiveResult(hashStrs[i], saltStrs[i]))
	}
}

// Test_Password_Cracker_Multiple_With_Leave tests a Chord ring formed by multiple nodes, and later some
// nodes leave. The password cracker should still be able to function
func Test_Password_Cracker_Multiple_With_Leave(t *testing.T) {
	numNodes := 16
	numLeave := 8
	transp := channelFac()

	nodes := make([]z.TestNode, numNodes)
	for i := range nodes {
		node := z.NewTestNode(t, peerFac, transp, fmt.Sprintf("127.0.0.1:%d", i+1), z.WithChordBytes(1),
			z.WithChordStabilizeInterval(time.Millisecond*500), z.WithChordFixFingerInterval(time.Millisecond*500))
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
	}

	for i := 0; i < numLeave; i++ {
		randomIdx := rand.Intn(len(nodes))
		nodes[randomIdx].LeaveChord()
		nodes = append(nodes[:randomIdx], nodes[randomIdx+1:]...)
		numNodes--
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second * 30)

	// First, we sort the nodes based on ChordID
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetChordID() < nodes[j].GetChordID()
	})

	totEntries := uint(math.Pow(2, float64(8)))
	for i := 1; i < numNodes; i++ {
		require.Equal(t, int(nodes[i].GetChordID()-nodes[i-1].GetChordID()),
			nodes[i].GetStorage().GetDictionaryStore().Len())
		totEntries -= nodes[i].GetChordID() - nodes[i-1].GetChordID()
	}
	require.Equal(t, int(totEntries), nodes[0].GetStorage().GetDictionaryStore().Len())

	// All following hash and salt pairs corresponds to "apple" in the dictionary
	hashStrs := []string{
		"1bd15226960ce500e8dbaabbd523b9356ec69ff1bdf2aeef6c5dbe272971986a",
		"2ea455a6c36bf264a0d933c2e9fa75e9962ceec80a55e118340d62c5b92cd930",
		"2ae367b9a8585e19e96f301bb3cb020941cd29a3925bb05c9dd12ac47c5757d6",
		"24a77708057aab975813079ab86b9b84ae300fd738f13fd6b425df0f8895b907",
		"a4ac87eb7080ed009b324a931235b27439f6e8f4ba51e2ed5c3c47f6962064d4",
		"4b446b30d876ca954a6d0a24f9d96db7b6a652465b28599d974f15033e87893f",
		"10d962ff38d51f366f7a0ffc2ab2f6497898230fb4b1c85bf1e03ff3226bfffa",
		"25b8005ff00096894f3d0dd7287efebef4f3d4ec45a19cc04159197368fa6ce1",
		"96e3f8a2bd177aa8d72f4bee847421b9555567c88bd940ee2b5cb7e25e022340",
		"865e185881d538d3e6f8158a5d7c2c58ca2a6523fa78e8916727682b83e9b13e",
		"86254408237a295b6aaa7b61b253689a50b635ab8b59a363e98ea2507d3d8e48",
		"67d5158c080528d5b50ad13426c6d6b56815fc2ade087f05bd39e4d0539bae34",
		"713da404459008f10a661b458fcc811f166f15579dfde8fd3c0d08de50fe8f03",
		"e7eabf2627703a7386b9433bf47734e7c6497dfb30712153e4f452074b0fa7ce",
		"38d9a74454af41571b5b20d9b26b7943dc191b3daadff8336dfe691e93b7a204",
		"a9bed160d86d2570e494cc39c095649d4816e76c1d31a183d3b63c205a25230c",
	}
	saltStrs := []string{"0f", "1f", "2f", "3f", "4f", "5f", "6f", "7f",
		"8f", "9f", "af", "bf", "cf", "df", "ef", "ff"}

	for i := 0; i < len(hashStrs); i++ {
		randomIdx := rand.Intn(len(nodes))
		err := nodes[randomIdx].PasswordSubmitRequest(hashStrs[i], saltStrs[i], 0, 0)
		require.NoError(t, err)
		time.Sleep(time.Second)
		require.Equal(t, "apple", nodes[randomIdx].PasswordReceiveResult(hashStrs[i], saltStrs[i]))
	}
}
