package cmd

import (
	"github.com/fatih/color"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/transport/udp"
	"log"
)

var peerFac peer.Factory = impl.NewPeer
var udpFac transport.Factory = udp.NewUDP

// UserInterface provides a command line interface of the program, in the normal mode
func UserInterface() {
	config := nodeDefaultConf(udpFac())
	node := nodeCreateWithConf(peerFac, config)
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

	color.HiYellow("================================================\n"+
		"=======  Node started!                   =======\n"+
		"=======  UDP Address := %s  =======\n"+
		"=======  Chord ID    := %03d              =======\n"+
		"=======  Balance     := %d                =======\n"+
		"================================================\n",
		config.Socket.GetAddress(), node.GetChordID(), node.GetBalance())

	leave := true

	for leave {
		join := preJoin(node)
		if join {
			leave = postJoin(node)
		}
	}
}
