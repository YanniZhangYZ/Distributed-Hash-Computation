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

// Test_Full_Three_Nodes_One_Task_1B_Salt tests a simple scenario
// where one node submit a password cracking request
// and another node executes the request to earn the reward
// use 1 byte salt
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

// Test_Full_Three_Nodes_Two_Tasks_2B_Salt tests a scenario
// where there are three node, two of the ndoes submit password cracking request
// there should be two nodes execute the request to earn the reward
// use 2 byte salt
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

// Test_Full_Three_Nodes_Two_Tasks_2B_Salt_No_Enough_Balance tests a scenario
// where there are three node, one of the ndoes submit two password cracking requests
// when submiting the second task, the node has no enough balance
// therefore only the first task is successfully published
// there should be one node execute the the request to earn the reward
// use 2 byte salt
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

// Test_Full_Many_Nodes_One_Task_2B_Salt tests a scenario
// where there are eight nodes, one of the ndoes submit a password cracking request
// there should be one node execute the the request to earn the reward
// use 2 byte salt
func Test_Full_Many_Nodes_One_Task_2B_Salt(t *testing.T) {
	transp := channelFac()
	nodeNum := 8

	worldState := common.QuickWorldState(nodeNum, 10)

	newNode := func(address string) z.TestNode {
		fullAddr := fmt.Sprintf("127.0.0.1:%s", address)
		return z.NewTestNode(t, peerFac, transp, fullAddr,
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(1),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithChordBytes(2), // correspond to salt length
			z.WithChordStabilizeInterval(time.Millisecond*200),
			z.WithChordFixFingerInterval(time.Millisecond*200),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainAccountAddress(address))

	}

	testNode := make([]z.TestNode, nodeNum)

	// creating 8 nodes
	for i := 0; i < nodeNum; i++ {
		testNode[i] = newNode(strconv.Itoa(i + 1))
		defer testNode[i].Stop()
	}

	// fmt.Println(" ")
	// fmt.Println("Finish creating node")

	for i := 0; i < nodeNum; i++ {
		for j := 0; j < nodeNum; j++ {
			if i == j {
				continue
			}
			testNode[i].AddPeer(testNode[j].GetAddr())
		}
	}
	// fmt.Println("Finish adding peer")

	// add them all to chord
	for i := 1; i < nodeNum; i++ {
		_ = testNode[i].JoinChord(testNode[i-1].GetAddr())
		time.Sleep(time.Second)
	}
	// fmt.Println("Finish joining Chord")

	// Wait for dictionary construction
	time.Sleep(time.Second * 10)

	// submit one task. The passward is apple
	hashStr := "14ffb81ab8f435a96400880c8bf34dba05a7ef8b63710f136e87297e601d7881"
	saltStr := "0000"
	err := testNode[0].PasswordSubmitRequest(hashStr, saltStr, 1, time.Second*600)
	require.NoError(t, err)

	// fmt.Println(" submit the task")

	// Wait for the node to crack the password and earn the reward
	password := ""
	for {
		password = testNode[0].PasswordReceiveResult(hashStr, saltStr)
		if password != "" {
			break
		}
		// fmt.Println("receive nothing")
		time.Sleep(time.Second * 5)
	}
	require.Equal(t, "apple", password)

	var totalBalance int64

	for i := 0; i < nodeNum; i++ {
		// there should be two transactions:
		// publisher -> smartaccount 1
		// smart account -> finisher 1
		require.Equal(t, 2, testNode[i].GetChain().GetTransactionCount())

		// one smart account is built
		require.Equal(t, nodeNum+1, testNode[i].GetChain().GetLastBlock().State.Len())
		require.NoError(t, testNode[i].GetChain().ValidateChain())
		if i != 0 {
			require.Equal(t, testNode[0].GetChain().GetBlockCount(), testNode[i].GetChain().GetBlockCount())
			require.Equal(t, testNode[0].GetChain().GetLastBlock().BlockHash, testNode[i].GetChain().GetLastBlock().BlockHash)
		}
		totalBalance += testNode[i].GetBalance()
	}

	// Check the balance
	require.EqualValues(t, nodeNum*10, totalBalance)
	// The cracking task is correct,
	// smartAccount should transfer money to finisher
	contractState, _ := testNode[0].GetChain().GetLastBlock().State.Get("1_1")
	require.EqualValues(t, 0, contractState.Balance)

}

// Test_Full_Many_Nodes_Many_Task_2B_Salt tests a scenario
// where there are sixteen nodes, each of the ndoes submit a password cracking request
// there should be sixteen nodes execute the the request to earn the reward
// use 2 byte salt
func Test_Full_Many_Nodes_Many_Task_2B_Salt(t *testing.T) {
	transp := channelFac()
	nodeNum := 8

	worldState := common.QuickWorldState(nodeNum, 20)

	newNode := func(address string) z.TestNode {
		fullAddr := fmt.Sprintf("127.0.0.1:%s", address)
		return z.NewTestNode(t, peerFac, transp, fullAddr,
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second),
			z.WithAntiEntropy(time.Second),
			z.WithChordBytes(2), // correspond to salt length
			z.WithChordStabilizeInterval(time.Second),
			z.WithChordFixFingerInterval(time.Second),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainAccountAddress(address))
	}

	testNode := make([]z.TestNode, nodeNum)

	// create 16 nodes
	for i := 0; i < nodeNum; i++ {
		testNode[i] = newNode(strconv.Itoa(i + 1))
		defer testNode[i].Stop()
	}

	// fmt.Println("Finish creating node")

	for i := 0; i < nodeNum; i++ {
		for j := 0; j < nodeNum; j++ {
			testNode[i].AddPeer(testNode[j].GetAddr())
		}
	}
	// fmt.Println("Finish adding peer")

	// add them all to chord
	for i := 1; i < nodeNum; i++ {
		err := testNode[i].JoinChord(testNode[i-1].GetAddr())
		require.NoError(t, err)
		time.Sleep(time.Second)
	}
	// fmt.Println("Finish joining Chord")

	// Wait for dictionary construction
	time.Sleep(time.Second * 60)

	// the password is apple
	hashStrs := []string{
		"62f789df3b04f99ae8e43f3933005148a17b20d44e1758341cee12ac67ce4f6d",
		"4dd05f0d43d885d43d329722cb447f004b63f3ec001de7625b79ce1865f320b8",
		"b499e429dc73357f28e988aceb0a854d6e5cbed570b941dd6d4361ccdc4966a5",
		"c3f0fa0c8d30f06e39c9d888f9f330c9e03acf6115837f126ce9c5772dd38bad",
		"df0e048529c42bdd48eefa5296faac2b588769c9e6c5438e64cfa0ea74557b45",
		"893bced2d172dbf6e8a364b94d2e6a422cf88e0dfca8524b05dcef375cf22208",
		"856e4e597a5df93747b47afbcb0a20a0e92b276ef7dcc38667b5654ebfc2c546",
		"0736dacb4e60ee62ca58572a4001647dde67810d3a6b30ca940c58444282bf3f",
		"61c933f9a03a9ba4fe84c0b5ae0f7791f5a44060d63468cb59469abf60e1961b",
		"cfbba6a3825444d55214a2057e89c3ba883ce522769f2b272dfb49935195b2db",
		"9d5d006ed7b3300fb7a6c8fd637edbee13f714459debf22083a6f0776656e462",
		"667484efcecd8e7bead842a2030ce06a87a213e630e6acc5ce164f5661567517",
		"a8bc6350efb7ec657eeda8bbc45398cac73dd55a0048abc846526ed6fc645b35",
		"9f31a16b4ca3ff13de98be2260e5ba02740bab13d243fa4126ee85ea350602f7",
		"7fad138c9ee75a04f54cdd94f3b248291179c71c08a26486981e2c918e6906ca",
		"c4ac8d28581033995bec81fa243f527591b6a157a5c200ade13c74cd95697038",
	}
	saltStrs := []string{"0fff", "1fff", "2fff", "3fff", "4fff", "5fff", "6fff", "7fff",
		"8fff", "9fff", "afff", "bfff", "cfff", "dfff", "efff", "ffff"}

	// submit the task request
	for i := 0; i < nodeNum; i++ {
		err := testNode[i].PasswordSubmitRequest(hashStrs[i], saltStrs[i], 1, time.Second*600)
		require.NoError(t, err)

		// fmt.Println(" submit the task")

		// Wait for the node to crack the password and earn the reward
		password := ""
		for {
			password = testNode[i].PasswordReceiveResult(hashStrs[i], saltStrs[i])
			if password != "" {
				break
			}
			time.Sleep(time.Second * 3)
		}
		require.Equal(t, "apple", password)
		// msg := "successfully receive one: " + hashStrs[i]
		// fmt.Println(msg)

	}

	var totalBalance int64

	for i := 0; i < nodeNum; i++ {
		// require.Equal(t, 16, testNode[i].GetChain().GetTransactionCount())
		// require.Equal(t, nodeNum*2, testNode[i].GetChain().GetLastBlock().State.Len())
		require.NoError(t, testNode[i].GetChain().ValidateChain())
		if i != 0 {
			require.Equal(t, testNode[0].GetChain().GetBlockCount(), testNode[i].GetChain().GetBlockCount())
			require.Equal(t, testNode[0].GetChain().GetLastBlock().BlockHash, testNode[i].GetChain().GetLastBlock().BlockHash)
		}
		totalBalance += testNode[i].GetBalance()

		// The cracking task is correct,
		// smartAccount should transfer money to finisher
		blockNum := fmt.Sprintf("%d_1", i)
		contractState, _ := testNode[i].GetChain().GetLastBlock().State.Get(blockNum)
		require.EqualValues(t, 0, contractState.Balance)
	}

	// Check the balance
	require.EqualValues(t, nodeNum*20, totalBalance)

}
