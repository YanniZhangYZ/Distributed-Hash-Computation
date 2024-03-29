package cmd

import (
	"github.com/fatih/color"
	"go.dedis.ch/cs438/peer"
	"log"
	"time"
)

// SimuUserInterface provides a command line interface of the program, it exposes only one peer, but there are nbNodes
// of peers running behind
func SimuUserInterface(nbNodes int) {
	configs := make([]peer.Configuration, nbNodes)
	nodes := make([]peer.Peer, nbNodes)
	for i := 0; i < nbNodes; i++ {
		configs[i] = nodeDefaultConf(udpFac())
		node := nodeCreateWithConf(peerFac, configs[i])
		err := node.Start()
		if err != nil {
			log.Fatalf("failed to start node: %v", err)
		}
		defer func() {
			err = node.Stop()
			if err != nil {
				log.Fatalf("failed to stop node: %v", err)
			}
		}()
		nodes[i] = node
	}

	for _, n1 := range nodes {
		for _, n2 := range nodes {
			n1.AddPeer(n2.GetAddr())
		}
	}

	for _, node := range nodes {
		err := node.JoinBlockchain(100, time.Second*600)
		if err != nil {
			log.Fatalf("failed to join blockchain: %v", err)
		}
	}

	for i := 1; i < nbNodes; i++ {
		err := nodes[i].JoinChord(nodes[i-1].GetAddr())
		if err != nil {
			log.Fatalf("failed to join chord: %v", err)
		}
		time.Sleep(time.Second)
	}

	node := nodes[0]
	color.HiYellow("=======================================================\n"+
		"=======             Node started!               =======\n"+
		"=======  UDP Address        := %s  =======\n"+
		"=======  Chord ID           := %3d              =======\n"+
		"=======  Balance            := %3d              =======\n"+
		"=======  Total No. of Peers := %3d              =======\n"+
		"=======================================================\n",
		node.GetAddr(), node.GetChordID(), node.GetBalance(), nbNodes)
	postJoin(node)
}
