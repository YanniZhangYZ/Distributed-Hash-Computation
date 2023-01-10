package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/transport/channel"
	"go.dedis.ch/cs438/transport/udp"
	"log"
	"os"
)

var peerFac peer.Factory = impl.NewPeer
var channelFac transport.Factory = channel.NewTransport
var udpFac transport.Factory = udp.NewUDP
var config peer.Configuration

func main() {
	nodeDefaultConf(udpFac(), "127.0.0.1:0")
	node := nodeCreateWithConf(peerFac)
	node.Start()
	defer node.Stop()

	color.Yellow("================================================\n"+
		"=======  Node started!                   =======\n"+
		"=======  UDP Address := %s  =======\n"+
		"=======  Chord ID    := %03d              =======\n"+
		"================================================\n",
		config.Socket.GetAddress(), node.GetChordID())

	prompt := &survey.Select{
		Message: "What do you want to do ?",
		Options: []string{"👫 add peer", "🕓 join Chord", "🔒 submit password cracking task",
			"🔐 receive password cracking result", "👉 exit"},
	}
	var action string
	for {
		err := survey.AskOne(prompt, &action)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch action {
		case "👫 add peer":
			err = addPeer(node)
			if err != nil {
				log.Fatalf("failed to add peer: %v", err)
			}
		case "🕓 join Chord":
			err = joinChord(node)
			if err != nil {
				log.Fatalf("failed to join Chord: %v", err)
			}
		case "🔒 submit password cracking task":
			err = crackPassword(node)
			if err != nil {
				log.Fatalf("failed to submit password cracking result: %v", err)
			}
		case "🔐 receive password cracking result":
			err = receivePassword(node)
			if err != nil {
				log.Fatalf("failed to receive password cracking task: %v", err)
			}
		case "👉 exit":
			color.Yellow("======= Bye 👋")
			os.Exit(0)
		}
	}
}
