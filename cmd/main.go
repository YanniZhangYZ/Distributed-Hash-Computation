package cmd

import (
	"github.com/fatih/color"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/transport/channel"
	"go.dedis.ch/cs438/transport/udp"
)

var peerFac peer.Factory = impl.NewPeer
var channelFac transport.Factory = channel.NewTransport
var udpFac transport.Factory = udp.NewUDP

// UserInterface provides a command line interface of the program
func UserInterface() {
	config := nodeDefaultConf(udpFac(), "127.0.0.1:0")
	node := nodeCreateWithConf(peerFac, config)
	node.Start()
	defer node.Stop()

	color.HiYellow("================================================\n"+
		"=======  Node started!                   =======\n"+
		"=======  UDP Address := %s  =======\n"+
		"=======  Chord ID    := %03d              =======\n"+
		"================================================\n",
		config.Socket.GetAddress(), node.GetChordID())

	leave := true

	for leave {
		join := preJoin(node)
		if join {
			leave = postJoin(node)
		}
	}
}
