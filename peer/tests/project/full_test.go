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

func Test_Full_Many_Nodes_One_Task_2B_Salt(t *testing.T) {
	transp := channelFac()
	nodeNum := 8

	worldState := common.QuickWorldState(nodeNum, 10)

	newNode := func(address string) z.TestNode {
		fullAddr := fmt.Sprintf("127.0.0.1:%s", address)
		return z.NewTestNode(t, peerFac, transp, fullAddr,
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithChordBytes(2), // correspond to salt length
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

	hashStr := "14ffb81ab8f435a96400880c8bf34dba05a7ef8b63710f136e87297e601d7881"
	saltStr := "0000"
	err := testNode[0].PasswordSubmitRequest(hashStr, saltStr, 1, time.Second*600)
	require.NoError(t, err)

	fmt.Println(" submit the task")

	// Wait for the node to crack the password and earn the reward
	password := ""
	for {
		password = testNode[0].PasswordReceiveResult(hashStr, saltStr)
		if password != "" {
			break
		}
		fmt.Println("receive nothing")
		time.Sleep(time.Second * 5)
	}
	require.Equal(t, "apple", password)

	var totalBalance int64

	for i := 0; i < nodeNum; i++ {
		require.Equal(t, 2, testNode[i].GetChain().GetTransactionCount())
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
	contractState, _ := testNode[0].GetChain().GetLastBlock().State.Get("1_1")
	require.EqualValues(t, 0, contractState.Balance)

}

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
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithChordBytes(2), // correspond to salt length
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
		"14ffb81ab8f435a96400880c8bf34dba05a7ef8b63710f136e87297e601d7881",
		"deb253c70e2318c3161561307094c14f0637fc9a528884125374c87d8cc9978b",
		"08cb91740f17c9e2f0dfc492031746b9c5925ec39c78b21f194381603dbf5e37",
		"c95bd1a106693dedea6570c60b1a24394fecf6f43d5c51b51819b96ccba483aa",
		"536dde0c6fdc7c5d811dd5e8cf80981c393d45fe84cf3fab7bc59cab5fac9033",
		"7389fbde2eb57ed20f942bb757854a95ccbf65508d0644e3d0353543b3316913",
		"cb045966ebe244998d4e4a24c9905ebeb8878248590231624c02ae174e83affc",
		"4b4b329d70d37f09638e8545bd708bd0e212e7a9c5c6352a3bfdc4b446f57413",
		"86f76685a10823db81f55fd81523a46cb2eb6a99c27317e0f376999cc741ec44",
		"c8d430ffe501ad5087fd31a98fbabb834beb0c82e722bd3be0991e2d399a0868",
		"e873502053f5475ec34f7be7fba48c7030a03820b2e61d2050d9db682587ca17",
		"1e6d28d2c48a2e9e0d81548a3e99852de7f9244609f6d9cf45e9f0dd35a4132c",
		"4b26e856e459ff373866707088d202ad7e745b348680fccd93e43bbd411e30c2",
		"39cef83ff1c135d71776a76439d72265b2ad99b855bbaa0e91a8004230564e7d",
		"e0371dd92ce8492f78a9be094e65d4e3ed7f8d3a819701e7afffb3922e743251",
		"83777a16726539e4f592ea8c7ec0afd9dad8e83deb6129ce39dc49f0e687f908",
		"ec21d75489c5f5a350b56a9175ceb037721a7952ba2e92bdfaf10e99b3ac05c8",
		"3a72c6e038ce875d3802582e4d436d518a1e21033caf2a84b3ca9c46bd6b20f4",
		"180e13060acd8a66e95ecdd6bd6eeb56f8fb1400c0cb9360fd9d92090e88709d",
	}
	saltStrs := []string{"0000", "0001", "0003", "0004", "0005", "0006", "0007", "0008", "0009",
		"0100", "0101", "0102", "0103", "0104", "0105", "0106", "0107", "0108", "0109"}

	for i := 1; i < nodeNum; i++ {
		err := testNode[i].PasswordSubmitRequest(hashStrs[i], saltStrs[i], 1, time.Second*600)
		require.NoError(t, err)

		fmt.Println(" submit the task")

		// Wait for the node to crack the password and earn the reward
		password := ""
		for {
			password = testNode[i].PasswordReceiveResult(hashStrs[i], saltStrs[i])
			if password != "" {
				break
			}
			fmt.Println("receive nothing")
			time.Sleep(time.Second * 5)
		}
		require.Equal(t, "apple", password)

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

		blockNum := fmt.Sprintf("%d_1", i)
		contractState, _ := testNode[i].GetChain().GetLastBlock().State.Get(blockNum)
		require.EqualValues(t, 0, contractState.Balance)
	}

	// Check the balance
	require.EqualValues(t, nodeNum*20, totalBalance)

}
