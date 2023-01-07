package password_cracker

import (
	"crypto"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/message"
)

// defaultDict is the default dictionary used for the dictionary attack, it can be replaced by reading a word
// list from a file, for example
var defaultDict = [...]string{
	"apple", "ball", "cat", "doll", "egg", "frog", "glass", "hat", "igloo", "jam",
	"kite", "lamb", "man", "net", "onion", "pen", "queen", "ring", "star", "train",
	"umbrella", "van", "watch", "xylophone", "yacht", "zebra"}

func NewPasswordCracker(conf *peer.Configuration, message *message.Message) *PasswordCracker {
	passwordCracker := PasswordCracker{
		address:  conf.Socket.GetAddress(),
		conf:     conf,
		message:  message,
		hashAlgo: conf.PasswordHashAlgorithm,
	}
	return &passwordCracker
}

type PasswordCracker struct {
	address  string
	conf     *peer.Configuration // The configuration contains Socket and MessageRegistry
	message  *message.Message    // Messaging used to communicate among nodes
	hashAlgo crypto.Hash         // The algorithm that is used to compute from the password to hash
}

func (p *PasswordCracker) SubmitRequest(salt []byte, hash []byte) {
	// Query Chord using salt value
	
	// TODO: Blockchain & Broadcast transaction

	// Unicast reception
}

func (p *PasswordCracker) ReceiveResult(salt []byte, hash []byte) string {
	return ""
}
