package cmd

import (
	"crypto"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/registry/standard"
	"go.dedis.ch/cs438/storage/inmemory"
	"go.dedis.ch/cs438/transport"
	"golang.org/x/xerrors"
	"log"
	"time"
)

func nodeDefaultConf(trans transport.Transport, addr string) peer.Configuration {
	socket, err := trans.CreateSocket(addr)
	if err != nil {
		panic(err)
	}
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
	config.BlockchainDifficulty = 3
	config.BlockchainBlockSize = 5
	config.BlockchainBlockTimeout = time.Second * 5
	config.BlockchainInitialState = make(map[string]common.State)
	config.PasswordHashAlgorithm = crypto.SHA256
	return config
}

func nodeCreateWithConf(f peer.Factory) peer.Peer {
	return f(config)
}

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

func leaveChord(node peer.Peer) error {
	return node.LeaveChord()
}

func showChordInfo(node peer.Peer) error {
	pred := node.GetPredecessor()
	succ := node.GetSuccessor()
	finger := node.GetFingerTable()

	color.HiYellow("\n"+
		"=======  My address      := %s with Chord ID %d\n"+
		"=======  Predecessor     := %s with Chord ID %d\n"+
		"=======  Successor       := %s with Chord ID %d\n"+
		"=======  Finger Table\n",
		config.Socket.GetAddress(), node.GetChordID(),
		pred, node.QueryChordID(pred),
		succ, node.QueryChordID(succ))

	fingerStr := ""
	for i := 0; i < config.ChordBytes*8; i++ {
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
	return node.PasswordSubmitRequest(hash, salt, reward)
}

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
