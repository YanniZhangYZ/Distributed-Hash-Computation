package project

import (
	"fmt"
	"github.com/stretchr/testify/require"
	z "go.dedis.ch/cs438/internal/testing"
	"os"
	"testing"
	"time"
)

// Test_Simple_Submit_Execute tests a simple scenario where one node submit a password cracking request
// and another node executes the request to earn the reward
func Test_Simple_Submit_Execute(t *testing.T) {
	transp := channelFac()

	newNode := func() z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(3),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithChordBytes(1),
			z.WithChordStabilizeInterval(time.Millisecond*200),
			z.WithChordFixFingerInterval(time.Millisecond*200))
	}

	// Create two nodes and let them join the blockchain
	node1 := newNode()
	defer node1.Stop()
	err := node1.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	node2 := newNode()
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	err = node2.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	err = node1.JoinChord(node2.GetAddr())
	require.NoError(t, err)

	// Node1 submits a request
	// Password is apple
	hashStr := "1cfcd196cf51b7a1d44159875452ba2dca8898d675f3d33d610ab9cb0031d7b2"
	saltStr := "3c"

	err = node1.PasswordSubmitRequest(hashStr, saltStr, 3, time.Second*600)
	require.NoError(t, err)

	// Wait for the node to crack the password and earn the reward
	time.Sleep(time.Second * 120)

	// Print the blockchain of each miner
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())

	// Check the blockchain
	require.Equal(t, node1.GetChain().GetTransactionCount(), 4)
	require.Equal(t, node2.GetChain().GetTransactionCount(), 4)
	require.Equal(t, node1.GetChain().GetBlockCount(), node2.GetChain().GetBlockCount())
	require.Equal(t, node1.GetChain().GetLastBlock().State.Len(), 3)
	require.Equal(t, node2.GetChain().GetLastBlock().State.Len(), 3)
	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node2.GetChain().GetLastBlock().BlockHash)

	// The contract account should transfer node1's deposit to node2
	// Check the balance
	require.EqualValues(t, node1.GetBalance(), 7)
	require.EqualValues(t, node2.GetBalance(), 13)
	contractState, _ := node1.GetChain().GetLastBlock().State.Get("1_1")
	require.EqualValues(t, contractState.Balance, 0)

}
