package cmd

import (
	"crypto"
	"fmt"
	"log"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/registry/standard"
	"go.dedis.ch/cs438/storage/inmemory"
	"go.dedis.ch/cs438/transport"
	"golang.org/x/xerrors"
)

// nodeDefaultConf returns the default configuration of a node
func nodeDefaultConf(trans transport.Transport, addr string) peer.Configuration {
	socket, err := trans.CreateSocket(addr)
	if err != nil {
		panic(err)
	}

	var config peer.Configuration
	config.Socket = socket
	config.MessageRegistry = standard.NewRegistry()
	config.AntiEntropyInterval = 0
	config.HeartbeatInterval = 0
	config.ContinueMongering = 0.5
	config.AckTimeout = time.Second * 3
	config.Storage = inmemory.NewPersistency()
	config.ChunkSize = 8192
	config.BackoffDataRequest = peer.Backoff{
		Initial: time.Second * 2,
		Factor:  2,
		Retry:   5,
	}
	config.TotalPeers = 1
	config.PaxosThreshold = func(u uint) int {
		return int(u/2 + 1)
	}
	config.PaxosID = 0
	config.PaxosProposerRetry = time.Second * 5

	config.ChordBytes = 1
	config.ChordTimeout = time.Second * 5
	config.ChordStabilizeInterval = time.Second * 5
	config.ChordFixFingerInterval = time.Second * 5
	config.ChordPingInterval = time.Second * 60

	config.BlockchainAccountAddress = ""
	config.BlockchainDifficulty = 2
	config.BlockchainBlockSize = 5
	config.BlockchainBlockTimeout = time.Second * 5
	config.BlockchainInitialState = make(map[string]common.State)
	config.PasswordHashAlgorithm = crypto.SHA256
	return config
}

// nodeCreateWithConf creates a node with the specified config
func nodeCreateWithConf(f peer.Factory, config peer.Configuration) peer.Peer {
	return f(config)
}

// addPeer add a remote node as a peer
func addPeer(node peer.Peer) error {
	var peerAddr string
	err := survey.AskOne(
		&survey.Input{Message: "Enter peer's address: "},
		&peerAddr,
		survey.WithValidator(addressValidator))

	if err != nil {
		return xerrors.Errorf("failed to get the answer: %v", err)
	}
	node.AddPeer(peerAddr)
	return nil
}

// joinChord joins an existing Chord ring
func joinChord(node peer.Peer) error {
	var peerAddr string
	err := survey.AskOne(
		&survey.Input{Message: "Enter Chord peer's address: "},
		&peerAddr,
		survey.WithValidator(addressValidator))

	if err != nil {
		return xerrors.Errorf("failed to get the answer: %v", err)
	}
	return node.JoinChord(peerAddr)
}

// leaveChord leaves a joined Chord ring
func leaveChord(node peer.Peer) error {
	return node.LeaveChord()
}

// showChordInfo shows all fields for a Chord node
func showChordInfo(node peer.Peer) error {
	pred := node.GetPredecessor()
	succ := node.GetSuccessor()
	finger := node.GetFingerTable()

	color.HiYellow("\n"+
		"=======  My address      := %s with Chord ID %d\n"+
		"=======  Predecessor     := %s with Chord ID %d\n"+
		"=======  Successor       := %s with Chord ID %d\n"+
		"=======  Finger Table\n",
		node.GetAddr(), node.GetChordID(),
		pred, node.QueryChordID(pred),
		succ, node.QueryChordID(succ))

	fingerStr := ""
	for i := 0; i < len(finger); i++ {
		if finger[i] != "" {
			fingerStr +=
				fmt.Sprintf("           Entry %d: %s with Chord ID %d\n",
					i+1, finger[i], node.QueryChordID(finger[i]))
		} else {
			fingerStr +=
				fmt.Sprintf("           Entry %d: %s\n",
					i+1, finger[i])
		}
	}
	color.Yellow("%s\n", fingerStr)

	return nil
}

// askHashSalt asks users for hash and salt
func askHashSalt() (string, string) {
	answers := struct {
		Hash string
		Salt string
	}{}

	err := survey.Ask([]*survey.Question{
		{
			Name:     "Hash",
			Prompt:   &survey.Input{Message: "Enter the hash value in hexadecimal form: "},
			Validate: hashValidator,
		},
		{
			Name:     "Salt",
			Prompt:   &survey.Input{Message: "Enter the salt value in hexadecimal form: "},
			Validate: saltValidator,
		},
	}, &answers)

	if err != nil {
		log.Fatalf("failed to get the answers for hash and salt: %v", err)
	}

	return answers.Hash, answers.Salt
}

// crackPassword propose a new password-cracking task
func crackPassword(node peer.Peer) error {
	hash, salt := askHashSalt()
	var reward int
	err := survey.AskOne(
		&survey.Input{Message: "Enter the reward you want to spend on this task: "},
		&reward,
		survey.WithValidator(rewardValidator))
	if err != nil {
		return xerrors.Errorf("failed to get the answer: %v", err)
	}
	return node.PasswordSubmitRequest(hash, salt, reward, time.Second*600)
}

// receivePassword receives results for previous tasks
func receivePassword(node peer.Peer) error {
	hash, salt := askHashSalt()
	result := node.PasswordReceiveResult(hash, salt)
	if result == "" {
		color.Yellow("\nNo result has been received! Try another one!\n\n\n")
	} else {
		color.Yellow("\n"+
			"=======  Hash     := %s\n"+
			"=======  Salt     := %s\n", hash, salt)
		color.Red("=======  Password := %s\n\n\n", result)
	}
	return nil
}
