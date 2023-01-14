package project

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	z "go.dedis.ch/cs438/internal/testing"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"golang.org/x/xerrors"
)

// Test_Simple_Submit_Execute tests a simple scenario where one node submit a password cracking request
// and another node executes the request to earn the reward
func Test_Full_Three_Nodes_One_Task_1B_Salt(t *testing.T) {
	transp := channelFac()

	newNode := func() z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithChordBytes(1), // correspond to salt length
			z.WithChordStabilizeInterval(time.Millisecond*200),
			z.WithChordFixFingerInterval(time.Millisecond*200))
	}

	// Create three nodes and let them join the blockchain

	// Node 1
	node1 := newNode()
	defer node1.Stop()
	err := node1.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Node 2
	node2 := newNode()
	defer node2.Stop()
	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())
	err = node2.JoinChord(node1.GetAddr())
	require.NoError(t, err)
	err = node2.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Node 3
	node3 := newNode()
	defer node3.Stop()
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())
	err = node3.JoinChord(node2.GetAddr())
	require.NoError(t, err)
	err = node3.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Wait for dictionary construction
	time.Sleep(time.Second * 5) // neccessary to compute dictionary

	// Node1 submits a request
	// Password is apple
	hashStr := "a9bed160d86d2570e494cc39c095649d4816e76c1d31a183d3b63c205a25230c"
	saltStr := "ff"

	err = node1.PasswordSubmitRequest(hashStr, saltStr, 3, time.Second*600)
	require.NoError(t, err)

	// Wait for the node to crack the password and earn the reward
	password := ""
	for {
		password = node1.PasswordReceiveResult(hashStr, saltStr)
		if password != "" {
			break
		}
		time.Sleep(time.Second * 5)
	}
	require.Equal(t, "apple", password)

	// Print the blockchain of each miner
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node3.GetChain().PrintChain())

	// Check the blockchain
	require.Equal(t, 5, node1.GetChain().GetTransactionCount())
	require.Equal(t, 5, node2.GetChain().GetTransactionCount())
	require.Equal(t, 5, node3.GetChain().GetTransactionCount())

	require.Equal(t, node1.GetChain().GetBlockCount(), node2.GetChain().GetBlockCount())
	require.Equal(t, node1.GetChain().GetBlockCount(), node3.GetChain().GetBlockCount())

	require.Equal(t, 4, node1.GetChain().GetLastBlock().State.Len())
	require.Equal(t, 4, node2.GetChain().GetLastBlock().State.Len())
	require.Equal(t, 4, node3.GetChain().GetLastBlock().State.Len())

	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node2.GetChain().GetLastBlock().BlockHash)
	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node3.GetChain().GetLastBlock().BlockHash)

	require.NoError(t, node1.GetChain().ValidateChain())
	require.NoError(t, node2.GetChain().ValidateChain())
	require.NoError(t, node3.GetChain().ValidateChain())

	// The contract account should transfer node1's deposit to node2
	// Check the balance
	require.EqualValues(t, 30, node1.GetBalance()+node2.GetBalance()+node3.GetBalance())
	contractState, _ := node1.GetChain().GetLastBlock().State.Get("1_1")
	require.EqualValues(t, 0, contractState.Balance)

}

func Test_Full_Three_Nodes_Two_Tasks_2B_Salt(t *testing.T) {
	transp := channelFac()

	newNode := func() z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithChordBytes(2), // correspond to salt length
			z.WithChordStabilizeInterval(time.Millisecond*200),
			z.WithChordFixFingerInterval(time.Millisecond*200))
	}

	// Create three nodes and let them join the blockchain

	// Node 1
	node1 := newNode()
	defer node1.Stop()
	err := node1.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Node 2
	node2 := newNode()
	defer node2.Stop()
	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())
	err = node2.JoinChord(node1.GetAddr())
	require.NoError(t, err)
	err = node2.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Node 3
	node3 := newNode()
	defer node3.Stop()
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())
	err = node3.JoinChord(node2.GetAddr())
	require.NoError(t, err)
	err = node3.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Wait for dictionary construction
	time.Sleep(time.Second * 5) // neccessary to compute dictionary

	//------------ first task, proposed by node 1---------------------

	// Password is apple
	hashStr := "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860"
	saltStr := "003c"

	err = node1.PasswordSubmitRequest(hashStr, saltStr, 3, time.Second*600)
	require.NoError(t, err)

	// Wait for the node to crack the password and earn the reward
	password := ""
	for {
		password = node1.PasswordReceiveResult(hashStr, saltStr)
		if password != "" {
			break
		}
		time.Sleep(time.Second * 5)
	}
	require.Equal(t, "apple", password)

	//------------ second task, proposed by node 1---------------------
	// Password is egg
	hashStr2 := "f857023981c0a3e223a45d37e129c6a3ddbbfe944075895243f72e83354e1008"
	saltStr2 := "002e"

	err = node1.PasswordSubmitRequest(hashStr2, saltStr2, 5, time.Second*600)
	require.NoError(t, err)

	// Wait for the node to crack the password and earn the reward
	password2 := ""
	for {
		password2 = node1.PasswordReceiveResult(hashStr2, saltStr2)
		if password2 != "" {
			break
		}
		time.Sleep(time.Second * 5)
	}
	require.Equal(t, "egg", password2)

	// Print the blockchain of each miner
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node3.GetChain().PrintChain())

	// Check the blockchain
	// Node registration ->3
	// proposer to smartAccount -> 2
	// smartAccount to finisher ->2
	require.Equal(t, 7, node1.GetChain().GetTransactionCount())
	require.Equal(t, 7, node2.GetChain().GetTransactionCount())
	require.Equal(t, 7, node3.GetChain().GetTransactionCount())

	require.Equal(t, node1.GetChain().GetBlockCount(), node2.GetChain().GetBlockCount())
	require.Equal(t, node1.GetChain().GetBlockCount(), node3.GetChain().GetBlockCount())

	// 3 block for the node themselves
	// 1_1 and 1_2 are two blocks for smartAccount correspond to the 2 tasks
	require.Equal(t, 5, node1.GetChain().GetLastBlock().State.Len())
	require.Equal(t, 5, node2.GetChain().GetLastBlock().State.Len())
	require.Equal(t, 5, node3.GetChain().GetLastBlock().State.Len())

	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node2.GetChain().GetLastBlock().BlockHash)
	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node3.GetChain().GetLastBlock().BlockHash)

	require.NoError(t, node1.GetChain().ValidateChain())
	require.NoError(t, node2.GetChain().ValidateChain())
	require.NoError(t, node3.GetChain().ValidateChain())

	// The contract account should transfer node1's deposit to node2
	// Check the balance
	require.EqualValues(t, 30, node1.GetBalance()+node2.GetBalance()+node3.GetBalance())
	// The first cracking task is correct,
	// smartAccount should transfer money to finisher
	contractState, _ := node1.GetChain().GetLastBlock().State.Get("1_1")
	require.EqualValues(t, 0, contractState.Balance)

	// The second cracking task is correct,
	// smartAccount should transfer money to finisher
	contractState2, _ := node1.GetChain().GetLastBlock().State.Get("1_2")
	require.EqualValues(t, 0, contractState2.Balance)

}

func Test_Full_Three_Nodes_Two_Tasks_2B_Salt_No_Enough_Balance(t *testing.T) {
	transp := channelFac()

	newNode := func() z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithChordBytes(2), // correspond to salt length
			z.WithChordStabilizeInterval(time.Millisecond*200),
			z.WithChordFixFingerInterval(time.Millisecond*200))
	}

	// Create three nodes and let them join the blockchain

	// Node 1
	node1 := newNode()
	defer node1.Stop()
	err := node1.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Node 2
	node2 := newNode()
	defer node2.Stop()
	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())
	err = node2.JoinChord(node1.GetAddr())
	require.NoError(t, err)
	err = node2.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Node 3
	node3 := newNode()
	defer node3.Stop()
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())
	err = node3.JoinChord(node2.GetAddr())
	require.NoError(t, err)
	err = node3.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err)

	// Wait for dictionary construction
	time.Sleep(time.Second * 5) // neccessary to compute dictionary

	//------------ first task, node 1 propose, reward 5 --------------------------
	// Password is apple
	hashStr := "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860"
	saltStr := "003c"

	// Node1 propose a task, reward is 5
	err = node1.PasswordSubmitRequest(hashStr, saltStr, 5, time.Second*600)
	require.NoError(t, err)

	// Wait for the node to crack the password and earn the reward
	password := ""
	for {
		password = node1.PasswordReceiveResult(hashStr, saltStr)
		if password != "" {
			break
		}
		time.Sleep(time.Second * 5)
	}
	require.Equal(t, "apple", password)

	//------------ second task, node 1 propose, reward 100, no enough balance --------------------------
	// Password is egg
	hashStr2 := "f857023981c0a3e223a45d37e129c6a3ddbbfe944075895243f72e83354e1008"
	saltStr2 := "002e"

	// Node1 propose a task again, reward is 100. No engough balance
	err = node1.PasswordSubmitRequest(hashStr2, saltStr2, 100, time.Second*600)
	expectErr := xerrors.Errorf("ProposeContract failed : don't have enough balance")
	require.EqualError(t, err, expectErr.Error())

	// Print the blockchain of each miner
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node3.GetChain().PrintChain())

	// Check the blockchain
	// Node registration ->3
	// proposer to smartAccount -> 1
	// smartAccount to finisher ->1
	// The second task has no enough balance, therefore no transcation happened
	require.Equal(t, 5, node1.GetChain().GetTransactionCount())
	require.Equal(t, 5, node2.GetChain().GetTransactionCount())
	require.Equal(t, 5, node3.GetChain().GetTransactionCount())

	require.Equal(t, node1.GetChain().GetBlockCount(), node2.GetChain().GetBlockCount())
	require.Equal(t, node1.GetChain().GetBlockCount(), node3.GetChain().GetBlockCount())

	// 3 block for the node themselves
	// 1_1 is the block for smartAccount correspond to the first task
	// The second task has no enough balance, therefore no block 1_2 established
	require.Equal(t, 4, node1.GetChain().GetLastBlock().State.Len())
	require.Equal(t, 4, node2.GetChain().GetLastBlock().State.Len())
	require.Equal(t, 4, node3.GetChain().GetLastBlock().State.Len())

	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node2.GetChain().GetLastBlock().BlockHash)
	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node3.GetChain().GetLastBlock().BlockHash)

	require.NoError(t, node1.GetChain().ValidateChain())
	require.NoError(t, node2.GetChain().ValidateChain())
	require.NoError(t, node3.GetChain().ValidateChain())

	// Check the balance
	require.EqualValues(t, 30, node1.GetBalance()+node2.GetBalance()+node3.GetBalance())
	// The first cracking task is correct,
	// smartAccount should transfer money to finisher
	contractState, _ := node1.GetChain().GetLastBlock().State.Get("1_1")
	require.EqualValues(t, 0, contractState.Balance)
	// The second task has no enough balance, therefore no block 1_2 established

}

func Test_Full_Many_Nodes_One_Task(t *testing.T) {
	transp := channelFac()
	nodeNum := 15

	worldState := common.QuickWorldState(nodeNum, 10)

	newNode := func(address string) z.TestNode {
		fullAddr := fmt.Sprintf("127.0.0.1:%s", address)
		return z.NewTestNode(t, peerFac, transp, fullAddr,
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithChordBytes(1), // correspond to salt length
			z.WithChordStabilizeInterval(time.Millisecond*200),
			z.WithChordFixFingerInterval(time.Millisecond*200),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainAccountAddress(address))

	}

	testNode := make([]z.TestNode, nodeNum)

	for i := 0; i < nodeNum; i++ {
		testNode[i] = newNode(strconv.Itoa(i + 1))
		defer testNode[i].Stop()
	}

	fmt.Println(" ")
	fmt.Println("Finish creating node")

	for i := 0; i < nodeNum; i++ {
		for j := 0; j < nodeNum; j++ {
			if i == j {
				continue
			}
			testNode[i].AddPeer(testNode[j].GetAddr())
		}
	}
	fmt.Println("Finish adding peer")

	for i := 1; i < nodeNum; i++ {
		_ = testNode[i].JoinChord(testNode[i-1].GetAddr())
	}
	fmt.Println("Finish joining Chord")

	// Wait for dictionary construction
	time.Sleep(time.Second * 10)

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

	err := testNode[0].PasswordSubmitRequest(hashStrs[15], saltStrs[15], 1, time.Second*600)
	require.NoError(t, err)

	fmt.Println(" submit the task")

	// Wait for the node to crack the password and earn the reward
	password := ""
	for {
		password = testNode[0].PasswordReceiveResult(hashStrs[15], saltStrs[15])
		if password != "" {
			break
		}
		fmt.Println("receive nothing")
		time.Sleep(time.Second * 5)
	}
	require.Equal(t, "apple", password)

}
