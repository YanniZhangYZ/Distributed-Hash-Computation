package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"go.dedis.ch/cs438/peer"
	"log"
	"os"
	"time"
)

// preJoin is the actions allowed before a node joins the Chord ring, it should be able to
// add new peers (used for broadcast), and join a Chord ring, or exit
func preJoin(node peer.Peer) bool {
	prompt := &survey.Select{
		Message: "What do you want to do ?",
		Options: []string{
			"👫 add peer, used for broadcast",
			"🗿 join blockchain",
			"🕓 join Chord, used for password cracker",
			"👋 exit"},
	}
	var action string
	for {
		err := survey.AskOne(prompt, &action)
		if err != nil {
			fmt.Println(err)
			return false
		}

		switch action {
		case "👫 add peer, used for broadcast":
			err = addPeer(node)
			if err != nil {
				log.Fatalf("failed to add peer: %v", err)
			}
		case "🗿 join blockchain":
			err = node.JoinBlockchain(100, time.Second*600)
			if err != nil {
				log.Fatalf("failed to join blockchain: %v", err)
			}
		case "🕓 join Chord, used for password cracker":
			// Check we have a successor or not, if yes, others have joined our Chord, we
			// can return true, for postJoin actions
			if node.GetSuccessor() != "" {
				return true
			}
			err = joinChord(node)
			if err != nil {
				log.Fatalf("failed to join Chord: %v", err)
			}
			// We have successfully joined Chord, we can enter postJoin actions
			return true
		case "👋 exit":
			color.HiYellow("=======  Bye 👋")
			os.Exit(0)
		}
	}
}

// postJoin is the actions allowed after a node joins the Chord ring, it should be able to
// propose new password cracking tasks
func postJoin(node peer.Peer) bool {
	prompt := &survey.Select{
		Message: "What do you want to do ?",
		Options: []string{
			"👫 add peer, used for broadcast",
			"🪐 show predecessor, successor, and finger table",
			"🔒 propose password cracking task",
			"🔐 receive password cracking result",
			"🕓 leave Chord",
			"📖 show world state",
			"👋 exit"},
	}
	var action string
	for {
		err := survey.AskOne(prompt, &action)
		if err != nil {
			fmt.Println(err)
			return false
		}

		switch action {
		case "👫 add peer, used for broadcast":
			err = addPeer(node)
			if err != nil {
				log.Fatalf("failed to add peer: %v", err)
			}
		case "🪐 show predecessor, successor, and finger table":
			err = showChordInfo(node)
			if err != nil {
				log.Fatalf("failed to show Chord info: %v", err)
			}
		case "🔒 propose password cracking task":
			err = crackPassword(node)
			if err != nil {
				log.Fatalf("failed to submit password cracking result: %v", err)
			}
		case "🔐 receive password cracking result":
			err = receivePassword(node)
			if err != nil {
				log.Fatalf("failed to receive password cracking task: %v", err)
			}
		case "🕓 leave Chord":
			err = leaveChord(node)
			if err != nil {
				log.Fatalf("failed to join Chord: %v", err)
			}
			return true
		case "📖 show world state":
			err = showWorldState(node)
			if err != nil {
				log.Fatalf("failed to receive password cracking task: %v", err)
			}
		case "👋 exit":
			color.HiYellow("=======  Bye 👋")
			os.Exit(0)
		}
	}
}
