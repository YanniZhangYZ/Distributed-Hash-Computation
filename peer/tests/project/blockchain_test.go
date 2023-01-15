package project

import (
	"fmt"
	"github.com/stretchr/testify/require"
	z "go.dedis.ch/cs438/internal/testing"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Test_Blockchain_Create tests if a node could initiate the blockchain with just one genesis block
func Test_Blockchain_Create(t *testing.T) {
	transp := channelFac()

	node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0")
	defer node.Stop()

	time.Sleep(time.Millisecond * 10)

	// There should be just one genesis block
	chainLen := node.GetChain().GetBlockCount()
	require.Equal(t, chainLen, 1)

	// The only block should be the genesis block with prevHash == "0..0"
	lastBlock := node.GetChain().GetLastBlock()
	require.Equal(t, lastBlock.PrevHash, strings.Repeat("0", 64))
}

// Test_Blockchain_Initial_Balance tests if a node could initiate the blockchain with correct initial balance
// world state has just one account with address == "1" and balance == 10
func Test_Blockchain_Initial_Balance(t *testing.T) {
	transp := channelFac()

	worldState := common.QuickWorldState(1, 10)

	node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
		z.WithBlockchainAccountAddress("1"),
		z.WithBlockchainInitialState(worldState.GetSimpleMap()))

	defer node.Stop()

	time.Sleep(time.Millisecond * 10)

	// There should be just one account in the world state
	numAccount := node.GetChain().GetLastBlock().State.Len()
	require.Equal(t, numAccount, 1)

	// The balance should be 10
	balance := node.GetBalance()
	require.EqualValues(t, balance, 10)
}

// Test_Blockchain_No_Enough_Balance tests if a node could raise an error
// when it tries to transfer $10 when it only has $5
func Test_Blockchain_No_Enough_Balance(t *testing.T) {
	transp := channelFac()

	worldState := common.QuickWorldState(2, 5)

	node := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
		z.WithBlockchainAccountAddress("1"),
		z.WithBlockchainInitialState(worldState.GetSimpleMap()))

	defer node.Stop()

	time.Sleep(time.Millisecond * 10)

	// There should be two accounts in the world state
	numAccount := node.GetChain().GetLastBlock().State.Len()
	require.Equal(t, numAccount, 2)

	// The balance should be 5
	balance := node.GetBalance()
	require.EqualValues(t, balance, 5)

	err := node.TransferMoney(common.Address{HexString: "2"}, 10, time.Second)
	require.Error(t, err)
}

// Test_Blockchain_Success_Transfer tests if a node could successfully transfer some money when it has enough balance
// It also tests if a block containing this tx can be built
// The blockchain has three accounts with initial balance being $10,
// account-1 transfer $3 to account-2
func Test_Blockchain_Success_Transfer(t *testing.T) {
	transp := channelFac()

	worldState := common.QuickWorldState(3, 10)

	node1 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
		z.WithBlockchainAccountAddress("1"),
		z.WithBlockchainInitialState(worldState.GetSimpleMap()),
		z.WithBlockchainBlockTimeout(time.Second*3),
		z.WithTotalPeers(3))
	node2 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
		z.WithBlockchainAccountAddress("2"),
		z.WithBlockchainInitialState(worldState.GetSimpleMap()),
		z.WithBlockchainBlockTimeout(time.Second*3),
		z.WithTotalPeers(3))
	node3 := z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
		z.WithBlockchainAccountAddress("3"),
		z.WithBlockchainInitialState(worldState.GetSimpleMap()),
		z.WithBlockchainBlockTimeout(time.Second*3),
		z.WithTotalPeers(3))

	defer node1.Stop()
	defer node2.Stop()
	defer node3.Stop()

	node1.AddPeer(node2.GetAddr())
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node1.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())

	time.Sleep(time.Millisecond * 10)

	// Money transfer should be successful
	err := node1.TransferMoney(common.Address{HexString: "2"}, 3, time.Second*600)
	require.NoError(t, err)

	time.Sleep(time.Second * 5)

	// Check balance after transfer
	balance1 := node1.GetBalance()
	balance2 := node2.GetBalance()
	balance3 := node3.GetBalance()
	require.EqualValues(t, balance1, 7)
	require.EqualValues(t, balance2, 13)
	require.EqualValues(t, balance3, 10)

	// Check blockchain after transfer
	require.Equal(t, node1.GetChain().GetBlockCount(), 2)
	require.Equal(t, node2.GetChain().GetBlockCount(), 2)
	require.Equal(t, node3.GetChain().GetBlockCount(), 2)
	lastBlockHash := node1.GetChain().GetLastBlock().BlockHash
	require.Equal(t, node2.GetChain().GetLastBlock().BlockHash, lastBlockHash)
	require.Equal(t, node3.GetChain().GetLastBlock().BlockHash, lastBlockHash)
}

// Test_Blockchain_Multiple_Transfers tests if several nodes could successfully transfer some money when it has enough balance
// It also tests if blocks containing these txs can be built
// The blockchain has three accounts with initial balance being $10,
// Create 6 transactions
func Test_Blockchain_Multiple_Transfers(t *testing.T) {
	transp := channelFac()

	worldState := common.QuickWorldState(3, 10)

	newNode := func(address string) z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainAccountAddress(address),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(3),
			z.WithBlockchainBlockSize(3),
			z.WithTotalPeers(3))
	}

	node1 := newNode("1")
	node2 := newNode("2")
	node3 := newNode("3")

	defer node1.Stop()
	defer node2.Stop()
	defer node3.Stop()

	node1.AddPeer(node2.GetAddr())
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node1.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())

	done1 := make(chan struct{})
	done2 := make(chan struct{})
	done3 := make(chan struct{})

	// 1 -> 2 : $3
	// 1 -> 2 : $5
	go func() {
		err := node1.TransferMoney(common.Address{HexString: "2"}, 3, time.Second*600)
		require.NoError(t, err)
		err = node1.TransferMoney(common.Address{HexString: "2"}, 5, time.Second*600)
		require.NoError(t, err)
		close(done1)
	}()

	// 2 -> 3 : $6
	// 2 -> 1 : $1
	go func() {
		err := node2.TransferMoney(common.Address{HexString: "3"}, 6, time.Second*600)
		require.NoError(t, err)
		err = node2.TransferMoney(common.Address{HexString: "1"}, 1, time.Second*600)
		require.NoError(t, err)
		close(done2)
	}()

	// 3 -> 1 : $4
	// 3 -> 2 : $2
	go func() {
		err := node3.TransferMoney(common.Address{HexString: "1"}, 4, time.Second*600)
		require.NoError(t, err)
		err = node3.TransferMoney(common.Address{HexString: "2"}, 2, time.Second*600)
		require.NoError(t, err)
		close(done3)
	}()

	// Wait for all money transfers to be done
	<-done1
	<-done2
	<-done3

	// Print the blockchain of each account
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node3.GetChain().PrintChain())

	// Check balance after transfer
	balance1 := node1.GetBalance()
	balance2 := node2.GetBalance()
	balance3 := node3.GetBalance()
	require.EqualValues(t, balance1, 7)
	require.EqualValues(t, balance2, 13)
	require.EqualValues(t, balance3, 10)

	// 6 transactions in total
	// Check blockchain after transfer
	blockCnt := node1.GetChain().GetBlockCount()
	require.Equal(t, node1.GetChain().GetBlockCount(), blockCnt)
	require.Equal(t, node2.GetChain().GetBlockCount(), blockCnt)
	require.Equal(t, node3.GetChain().GetBlockCount(), blockCnt)

	require.Equal(t, node1.GetChain().GetTransactionCount(), 6)
	require.Equal(t, node2.GetChain().GetTransactionCount(), 6)
	require.Equal(t, node3.GetChain().GetTransactionCount(), 6)

	lastBlockHash := node1.GetChain().GetLastBlock().BlockHash
	require.Equal(t, node2.GetChain().GetLastBlock().BlockHash, lastBlockHash)
	require.Equal(t, node3.GetChain().GetLastBlock().BlockHash, lastBlockHash)

	require.NoError(t, node1.GetChain().ValidateChain())
	require.NoError(t, node2.GetChain().ValidateChain())
	require.NoError(t, node3.GetChain().ValidateChain())
}

// Test_Blockchain_Random_Transfers_1 creates several nodes and a lot of random transactions among them.
// It tests, after all these transactions, if the blockchain of each account is the same
// and if the total balance is the same as before.
// Transactions are organized in rounds. In each round, each node submits its transaction.
// This test may take very long time. Be patient :)
func Test_Blockchain_Random_Transfers(t *testing.T) {
	transp := channelFac()

	numNode := 3
	numTxPerNode := 5
	initBalance := 100
	txVerifyTimeout := time.Second * 600

	worldState := common.QuickWorldState(numNode, int64(initBalance))

	newNode := func(i int) z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainAccountAddress(fmt.Sprintf("%d", i)),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainBlockTimeout(time.Second*5),
			z.WithBlockchainDifficulty(3),
			z.WithBlockchainBlockSize(3),
			z.WithTotalPeers(uint(numNode)))
	}

	// Create nodes
	nodes := make([]z.TestNode, 0)
	for i := 0; i < numNode; i++ {
		nodes = append(nodes, newNode(i+1))
		defer nodes[len(nodes)-1].Stop()
	}

	// Add each other as peers
	for i := 0; i < numNode; i++ {
		for j := 0; j < numNode; j++ {
			if j == i {
				continue
			}
			nodes[i].AddPeer(nodes[j].GetAddr())
		}
	}

	// Generate random money transfer
	for i := 0; i < numTxPerNode; i++ {
		dones := make([]chan struct{}, 0)

		for n := 0; n < numNode; n++ {
			dst := rand.Intn(numNode) + 1
			for dst == n+1 {
				dst = rand.Intn(numNode) + 1
			}
			dstAddress := common.Address{HexString: fmt.Sprintf("%d", dst)}

			currNode := n
			currBalance := nodes[currNode].GetBalance()
			amount := rand.Int63n(currBalance)

			done := make(chan struct{})
			dones = append(dones, done)

			go func() {
				err := nodes[currNode].TransferMoney(dstAddress, amount, txVerifyTimeout)
				require.NoError(t, err)
				close(done)
			}()
		}
		// Wait until all nodes are done this round
		for i := 0; i < numNode; i++ {
			<-dones[i]
		}
	}

	// Print each node's blockchain
	for i := 0; i < numNode; i++ {
		fmt.Fprint(os.Stdout, nodes[i].GetChain().PrintChain())
	}

	// Check sum of balances
	balanceSum := int64(0)
	for i := 0; i < numNode; i++ {
		balanceSum += nodes[i].GetBalance()
	}
	require.EqualValues(t, initBalance*numNode, balanceSum)

	// Check blockchain
	blockCnt := nodes[0].GetChain().GetBlockCount()
	txCnt := nodes[0].GetChain().GetTransactionCount()
	lastBlockHash := nodes[0].GetChain().GetLastBlock().BlockHash
	for n := 0; n < numNode; n++ {
		require.Equal(t, nodes[n].GetChain().GetBlockCount(), blockCnt)
		require.Equal(t, nodes[n].GetChain().GetTransactionCount(), txCnt)
		require.Equal(t, nodes[n].GetChain().GetLastBlock().BlockHash, lastBlockHash)
	}

	// Full validation each node's blockchain
	for i := 0; i < numNode; i++ {
		err := nodes[i].GetChain().ValidateChain()
		require.NoError(t, err)
	}
}

// Test_Blockchain_Stress_Test creates many nodes and a lot of random transactions among them.
// The difficulty of POW is very low to produce frequent block mining conflicts.
// It tests, after all these transactions, if the blockchain of each account is the same
// and if the total balance is the same as before.
// Each node is executing in its own thread and submits its transactions independently.
// This test may take VERY long time. Be patient :)
func Test_Blockchain_Stress_Test(t *testing.T) {
	transp := channelFac()

	numNode := 20
	numTxPerNode := 10
	initBalance := 100
	txVerifyTimeout := time.Second * 600

	worldState := common.QuickWorldState(numNode, int64(initBalance))

	newNode := func(i int) z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainAccountAddress(fmt.Sprintf("%d", i)),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainBlockTimeout(time.Second*5),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(10),
			z.WithTotalPeers(uint(numNode)))
	}

	// Create nodes
	nodes := make([]z.TestNode, 0)
	for i := 0; i < numNode; i++ {
		nodes = append(nodes, newNode(i+1))
		defer nodes[len(nodes)-1].Stop()
	}

	// Add each other as peers
	for i := 0; i < numNode; i++ {
		for j := 0; j < numNode; j++ {
			if j == i {
				continue
			}
			nodes[i].AddPeer(nodes[j].GetAddr())
		}
	}

	// Create random transactions
	dones := make([]chan struct{}, 0)
	for n := 0; n < numNode; n++ {
		done := make(chan struct{})
		dones = append(dones, done)
		currNode := n
		go func() {
			for i := 0; i < numTxPerNode; i++ {
				dst := rand.Intn(numNode) + 1
				for dst == currNode+1 {
					dst = rand.Intn(numNode) + 1
				}
				dstAddress := common.Address{HexString: fmt.Sprintf("%d", dst)}

				currBalance := nodes[currNode].GetBalance()
				amount := rand.Int63n(currBalance)

				err := nodes[currNode].TransferMoney(dstAddress, amount, txVerifyTimeout)
				require.NoError(t, err)
			}
			close(done)
		}()

	}

	// Wait until all nodes are done this round
	for i := 0; i < numNode; i++ {
		<-dones[i]
	}

	// Print each node's blockchain
	for i := 0; i < numNode; i++ {
		fmt.Fprint(os.Stdout, nodes[i].GetChain().PrintChain())
	}

	// Check sum of balances
	balanceSum := int64(0)
	for i := 0; i < numNode; i++ {
		balanceSum += nodes[i].GetBalance()
	}
	require.EqualValues(t, initBalance*numNode, balanceSum)

	// Check blockchain
	blockCnt := nodes[0].GetChain().GetBlockCount()
	txCnt := nodes[0].GetChain().GetTransactionCount()
	lastBlockHash := nodes[0].GetChain().GetLastBlock().BlockHash
	for n := 0; n < numNode; n++ {
		require.Equal(t, nodes[n].GetChain().GetBlockCount(), blockCnt)
		require.Equal(t, nodes[n].GetChain().GetTransactionCount(), txCnt)
		require.Equal(t, nodes[n].GetChain().GetLastBlock().BlockHash, lastBlockHash)
	}

	// Full validation each node's blockchain
	for i := 0; i < numNode; i++ {
		err := nodes[i].GetChain().ValidateChain()
		require.NoError(t, err)
	}
}

// Test_Blockchain_Late_Start_Catch_Up tests if a node with its state already included in the world state
// can start late and then catch up
// Note that catch up is only possible if nodes enable anti-entropy and heartbeats.
func Test_Blockchain_Late_Start_Catch_Up(t *testing.T) {
	transp := channelFac()

	worldState := common.QuickWorldState(3, 10)

	newNode := func(address string) z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainAccountAddress(address),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(3),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1))
	}

	node1 := newNode("1")
	node2 := newNode("2")

	defer node1.Stop()
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	done1 := make(chan struct{})
	done2 := make(chan struct{})

	// 1 -> 2 : $3
	// 1 -> 2 : $6
	go func() {
		err := node1.TransferMoney(common.Address{HexString: "2"}, 3, time.Second*600)
		require.NoError(t, err)
		err = node1.TransferMoney(common.Address{HexString: "2"}, 6, time.Second*600)
		require.NoError(t, err)
		close(done1)
	}()

	// 2 -> 1 : $5
	// 2 -> 1 : $1
	go func() {
		err := node2.TransferMoney(common.Address{HexString: "1"}, 5, time.Second*600)
		require.NoError(t, err)
		err = node2.TransferMoney(common.Address{HexString: "1"}, 1, time.Second*600)
		require.NoError(t, err)
		close(done2)
	}()

	// Wait for all money transfers to be done
	<-done1
	<-done2

	// Node3 starts late
	node3 := newNode("3")
	defer node3.Stop()
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())

	time.Sleep(time.Second * 5)

	// Print the blockchain of each account
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node3.GetChain().PrintChain())

	// Check balance after transfer
	balance1 := node1.GetBalance()
	balance2 := node2.GetBalance()
	balance3 := node3.GetBalance()
	require.EqualValues(t, balance1, 7)
	require.EqualValues(t, balance2, 13)
	require.EqualValues(t, balance3, 10)

	// 4 transactions in total
	// Check blockchain after transfer
	blockCnt := node1.GetChain().GetBlockCount()
	require.Equal(t, node1.GetChain().GetBlockCount(), blockCnt)
	require.Equal(t, node2.GetChain().GetBlockCount(), blockCnt)
	require.Equal(t, node3.GetChain().GetBlockCount(), blockCnt)

	require.Equal(t, node1.GetChain().GetTransactionCount(), 4)
	require.Equal(t, node2.GetChain().GetTransactionCount(), 4)
	require.Equal(t, node3.GetChain().GetTransactionCount(), 4)

	lastBlockHash := node1.GetChain().GetLastBlock().BlockHash
	require.Equal(t, node2.GetChain().GetLastBlock().BlockHash, lastBlockHash)
	require.Equal(t, node3.GetChain().GetLastBlock().BlockHash, lastBlockHash)

	require.NoError(t, node1.GetChain().ValidateChain())
	require.NoError(t, node2.GetChain().ValidateChain())
	require.NoError(t, node3.GetChain().ValidateChain())
}

// Test_Blockchain_Join tests if a node could join the blockchain network by declaring its account when
// it is not included in the initial world state.
func Test_Blockchain_Join(t *testing.T) {
	transp := channelFac()

	worldState := common.QuickWorldState(2, 10)

	newNode := func(address string) z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainAccountAddress(address),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(3),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1))
	}

	node1 := newNode("1")
	node2 := newNode("2")

	defer node1.Stop()
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	done1 := make(chan struct{})
	done2 := make(chan struct{})

	// 1 -> 2 : $3
	// 1 -> 2 : $6
	go func() {
		err := node1.TransferMoney(common.Address{HexString: "2"}, 3, time.Second*600)
		require.NoError(t, err)
		err = node1.TransferMoney(common.Address{HexString: "2"}, 6, time.Second*600)
		require.NoError(t, err)
		close(done1)
	}()

	// 2 -> 1 : $1
	// 2 -> 1 : $5
	go func() {
		err := node2.TransferMoney(common.Address{HexString: "1"}, 1, time.Second*600)
		require.NoError(t, err)
		err = node2.TransferMoney(common.Address{HexString: "1"}, 5, time.Second*600)
		require.NoError(t, err)
		close(done2)
	}()

	// Wait for all money transfers to be done
	<-done1
	<-done2

	// Node3 starts late
	node3 := newNode("3")
	defer node3.Stop()
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())

	time.Sleep(time.Second * 5)

	// Node3 declare its blockchain account with initial balance of 100
	err := node3.JoinBlockchain(100, time.Second*600)
	//err := node3.TransferMoney(common.StringToAddress("3"), 100, time.Second*600)
	require.NoError(t, err)

	err = node3.TransferMoney(common.StringToAddress("1"), 10, time.Second*600)
	require.NoError(t, err)

	err = node3.TransferMoney(common.StringToAddress("2"), 10, time.Second*600)
	require.NoError(t, err)

	// Print the blockchain of each account
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node3.GetChain().PrintChain())

	// Check balance after transfer
	balance1 := node1.GetBalance()
	balance2 := node2.GetBalance()
	balance3 := node3.GetBalance()
	require.EqualValues(t, balance1, 17)
	require.EqualValues(t, balance2, 23)
	require.EqualValues(t, balance3, 80)

	// 6 transactions in total
	// Check blockchain after transfer
	blockCnt := node1.GetChain().GetBlockCount()
	require.Equal(t, node1.GetChain().GetBlockCount(), blockCnt)
	require.Equal(t, node2.GetChain().GetBlockCount(), blockCnt)
	require.Equal(t, node3.GetChain().GetBlockCount(), blockCnt)

	require.Equal(t, node1.GetChain().GetTransactionCount(), 7)
	require.Equal(t, node2.GetChain().GetTransactionCount(), 7)
	require.Equal(t, node3.GetChain().GetTransactionCount(), 7)

	lastBlockHash := node1.GetChain().GetLastBlock().BlockHash
	require.Equal(t, node2.GetChain().GetLastBlock().BlockHash, lastBlockHash)
	require.Equal(t, node3.GetChain().GetLastBlock().BlockHash, lastBlockHash)

	require.NoError(t, node1.GetChain().ValidateChain())
	require.NoError(t, node2.GetChain().ValidateChain())
	require.NoError(t, node3.GetChain().ValidateChain())
}

// Test_Blockchain_All_Join tests if the initial world state is empty and all nodes declare itself by joining
func Test_Blockchain_All_Join(t *testing.T) {
	transp := channelFac()

	newNode := func() z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainInitialState(nil),
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1))
	}

	// Create each node and let them join the blockchain
	node1 := newNode()
	defer node1.Stop()
	err1 := node1.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err1)

	node2 := newNode()
	defer node2.Stop()
	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())
	err2 := node2.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err2)

	node3 := newNode()
	defer node3.Stop()
	node1.AddPeer(node3.GetAddr())
	node2.AddPeer(node3.GetAddr())
	node3.AddPeer(node1.GetAddr())
	node3.AddPeer(node2.GetAddr())
	err3 := node3.JoinBlockchain(10, time.Second*600)
	require.NoError(t, err3)

	// Print the blockchain of each miner
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node3.GetChain().PrintChain())

	// Check the consistence of blockchain after joining
	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node2.GetChain().GetLastBlock().BlockHash)
	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node3.GetChain().GetLastBlock().BlockHash)
	require.Equal(t, node1.GetChain().GetBlockCount(), node2.GetChain().GetBlockCount())
	require.Equal(t, node1.GetChain().GetBlockCount(), node3.GetChain().GetBlockCount())

	require.NoError(t, node1.GetChain().ValidateChain())
	require.NoError(t, node2.GetChain().ValidateChain())
	require.NoError(t, node3.GetChain().ValidateChain())

	// There should be three txs in the blockchain and three accounts in the world state
	require.Equal(t, node1.GetChain().GetTransactionCount(), 3)
	require.Equal(t, node1.GetChain().GetLastBlock().State.Len(), 3)

	require.EqualValues(t, node1.GetBalance(), 10)
	require.EqualValues(t, node2.GetBalance(), 10)
	require.EqualValues(t, node3.GetBalance(), 10)

	// Try to do some transactions

	done1 := make(chan struct{})
	done2 := make(chan struct{})
	done3 := make(chan struct{})

	// 1 -> 2 : $3
	// 1 -> 2 : $5
	go func() {
		err := node1.TransferMoney(common.Address{HexString: "2"}, 3, time.Second*600)
		require.NoError(t, err)
		err = node1.TransferMoney(common.Address{HexString: "2"}, 5, time.Second*600)
		require.NoError(t, err)
		close(done1)
	}()

	// 2 -> 3 : $6
	// 2 -> 1 : $1
	go func() {
		err := node2.TransferMoney(common.Address{HexString: "3"}, 6, time.Second*600)
		require.NoError(t, err)
		err = node2.TransferMoney(common.Address{HexString: "1"}, 1, time.Second*600)
		require.NoError(t, err)
		close(done2)
	}()

	// 3 -> 1 : $4
	// 3 -> 2 : $2
	go func() {
		err := node3.TransferMoney(common.Address{HexString: "1"}, 4, time.Second*600)
		require.NoError(t, err)
		err = node3.TransferMoney(common.Address{HexString: "2"}, 2, time.Second*600)
		require.NoError(t, err)
		close(done3)
	}()

	// Wait for all money transfers to be done
	<-done1
	<-done2
	<-done3

	// Print the blockchain of each account
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node3.GetChain().PrintChain())

	// Check balance after transfer
	balance1 := node1.GetBalance()
	balance2 := node2.GetBalance()
	balance3 := node3.GetBalance()
	require.EqualValues(t, balance1, 7)
	require.EqualValues(t, balance2, 13)
	require.EqualValues(t, balance3, 10)

	// 6 transactions in total
	// Check blockchain after transfer
	blockCnt := node1.GetChain().GetBlockCount()
	require.Equal(t, node1.GetChain().GetBlockCount(), blockCnt)
	require.Equal(t, node2.GetChain().GetBlockCount(), blockCnt)
	require.Equal(t, node3.GetChain().GetBlockCount(), blockCnt)

	require.Equal(t, node1.GetChain().GetTransactionCount(), 9)
	require.Equal(t, node2.GetChain().GetTransactionCount(), 9)
	require.Equal(t, node3.GetChain().GetTransactionCount(), 9)

	lastBlockHash := node1.GetChain().GetLastBlock().BlockHash
	require.Equal(t, node2.GetChain().GetLastBlock().BlockHash, lastBlockHash)
	require.Equal(t, node3.GetChain().GetLastBlock().BlockHash, lastBlockHash)

	require.NoError(t, node1.GetChain().ValidateChain())
	require.NoError(t, node2.GetChain().ValidateChain())
	require.NoError(t, node3.GetChain().ValidateChain())
}

// Test_Blockchain_Stress_Test creates many nodes with an empty initial state.
// Each node joins by calling JoinBlockchain
// The difficulty of POW is very low to produce frequent block mining conflicts.
// It tests, after all these joining, if the blockchain of each account is the same
// and if the total balance is the same as before.
// This test may take VERY long time. Be patient :)
func Test_Blockchain_Join_Stress_Test(t *testing.T) {
	transp := channelFac()

	numNode := 5
	initBalance := 100
	txVerifyTimeout := time.Second * 600

	newNode := func(address string) z.TestNode {
		fullAddr := fmt.Sprintf("127.0.0.1:%s", address)
		return z.NewTestNode(t, peerFac, transp, fullAddr,
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1),
			z.WithBlockchainAccountAddress(address))
	}

	// Create nodes
	nodes := make([]z.TestNode, 0)
	for i := 0; i < numNode; i++ {
		nodes = append(nodes, newNode(strconv.Itoa(i+1)))
		defer nodes[len(nodes)-1].Stop()

		for j := 0; j < i; j++ {
			nodes[i].AddPeer(nodes[j].GetAddr())
			nodes[j].AddPeer(nodes[i].GetAddr())
		}

		err := nodes[i].JoinBlockchain(int64(initBalance), txVerifyTimeout)
		require.NoError(t, err)
	}

	// Add each other as peers
	for i := 0; i < numNode; i++ {
		for j := 0; j < numNode; j++ {
			if j == i {
				continue
			}
			//nodes[i].AddPeer(nodes[j].GetAddr())
		}
	}
	//
	//for n := 0; n < numNode; n++ {
	//	err := nodes[n].JoinBlockchain(int64(initBalance), txVerifyTimeout)
	//	require.NoError(t, err)
	//}

	// Print each node's blockchain
	for i := 0; i < numNode; i++ {
		fmt.Fprint(os.Stdout, nodes[i].GetChain().PrintChain())
	}

	// Check sum of balances
	balanceSum := int64(0)
	for i := 0; i < numNode; i++ {
		balanceSum += nodes[i].GetBalance()
	}
	require.EqualValues(t, initBalance*numNode, balanceSum)

	// Check blockchain
	blockCnt := nodes[0].GetChain().GetBlockCount()
	txCnt := nodes[0].GetChain().GetTransactionCount()
	lastBlockHash := nodes[0].GetChain().GetLastBlock().BlockHash
	for n := 0; n < numNode; n++ {
		require.Equal(t, nodes[n].GetChain().GetBlockCount(), blockCnt)
		require.Equal(t, nodes[n].GetChain().GetTransactionCount(), txCnt)
		require.Equal(t, nodes[n].GetChain().GetLastBlock().BlockHash, lastBlockHash)
	}

	// Check state
	numAccount := nodes[0].GetChain().GetLastBlock().State.Len()
	for n := 1; n < numAccount; n++ {
		require.Equal(t, numAccount,
			nodes[n].GetChain().GetLastBlock().State.Len())
	}

	// Full validation each node's blockchain
	for i := 0; i < numNode; i++ {
		err := nodes[i].GetChain().ValidateChain()
		require.NoError(t, err)
	}
}

// Test_Blockchain_Deploy_Contract tests if a node could publish a smart contract to the blockchain
// A new contract account should be added to the world state and the publisher should pay the deposit
func Test_Blockchain_Deploy_Contract(t *testing.T) {
	transp := channelFac()

	worldState := common.QuickWorldState(2, 10)

	newNode := func(address string) z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainAccountAddress(address),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(3),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1))
	}

	node1 := newNode("1")
	node2 := newNode("2")

	defer node1.Stop()
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	time.Sleep(time.Millisecond * 10)

	err := node1.ProposeContract("abcdefg", "xxxx", 3, "2", time.Second*600)
	require.NoError(t, err)

	// Print the blockchain of each miner
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())

	// There should be one transaction, two blocks, and three accounts in the blockchain
	require.Equal(t, node1.GetChain().GetTransactionCount(), 1)
	require.Equal(t, node2.GetChain().GetTransactionCount(), 1)
	require.Equal(t, node1.GetChain().GetBlockCount(), 2)
	require.Equal(t, node2.GetChain().GetBlockCount(), 2)
	require.Equal(t, node1.GetChain().GetLastBlock().State.Len(), 3)
	require.Equal(t, node2.GetChain().GetLastBlock().State.Len(), 3)
	require.Equal(t, node1.GetChain().GetLastBlock().BlockHash, node2.GetChain().GetLastBlock().BlockHash)

	// Node1 should first pay the deposit to the contract account
	// Check the balance
	require.EqualValues(t, node1.GetBalance(), 7)
	require.EqualValues(t, node2.GetBalance(), 10)

	contractState, _ := node1.GetChain().GetLastBlock().State.Get("1_1")
	require.EqualValues(t, contractState.Balance, 3)

}

// Test_Blockchain_Execute_Contract tests if a node could execute a smart contract and earn its reward
func Test_Blockchain_Execute_Contract(t *testing.T) {
	transp := channelFac()

	worldState := common.QuickWorldState(2, 10)

	newNode := func(address string) z.TestNode {
		return z.NewTestNode(t, peerFac, transp, "127.0.0.1:0",
			z.WithBlockchainAccountAddress(address),
			z.WithBlockchainInitialState(worldState.GetSimpleMap()),
			z.WithBlockchainBlockTimeout(time.Second*3),
			z.WithBlockchainDifficulty(2),
			z.WithBlockchainBlockSize(2),
			z.WithHeartbeat(time.Second*1),
			z.WithAntiEntropy(time.Second*1))
	}

	node1 := newNode("1")
	node2 := newNode("2")

	defer node1.Stop()
	defer node2.Stop()

	node1.AddPeer(node2.GetAddr())
	node2.AddPeer(node1.GetAddr())

	time.Sleep(time.Millisecond * 10)

	// Node1 publishes the contract
	err := node1.ProposeContract("c612f289f5324c73d96a20ca14cf834e95a359a2b28101401e1bd7daa3bac4e2",
		"002e", 3, "2", time.Second*600)
	require.NoError(t, err)

	// Node2 executes the contract
	err = node2.ExecuteContract("banana",
		"c612f289f5324c73d96a20ca14cf834e95a359a2b28101401e1bd7daa3bac4e2", "002e", "1_1", time.Second*600)
	require.NoError(t, err)

	// Print the blockchain of each miner
	fmt.Fprint(os.Stdout, node1.GetChain().PrintChain())
	fmt.Fprint(os.Stdout, node2.GetChain().PrintChain())

	// There should be two transactions, two/three blocks, and three accounts in the blockchain
	require.Equal(t, node1.GetChain().GetTransactionCount(), 2)
	require.Equal(t, node2.GetChain().GetTransactionCount(), 2)
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
