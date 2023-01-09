package password_cracker

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/chord"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"math/big"
	"sync"
)

// defaultDict is the default dictionary used for the dictionary attack, it can be replaced by reading a word
// list from a file, for example
var defaultDict = [...]string{
	"apple", "ball", "cat", "doll", "egg"}

func NewPasswordCracker(conf *peer.Configuration, message *message.Message, chord *chord.Chord) *PasswordCracker {
	var tasks sync.Map
	passwordCracker := PasswordCracker{
		address:  conf.Socket.GetAddress(),
		conf:     conf,
		message:  message,
		chord:    chord,
		hashAlgo: conf.PasswordHashAlgorithm,
		tasks:    &tasks,
	}

	/* Password Cracker callbacks */
	conf.MessageRegistry.RegisterMessageCallback(
		types.PasswordCrackerRequestMessage{}, passwordCracker.execPasswordCrackerRequestMessage)
	conf.MessageRegistry.RegisterMessageCallback(
		types.PasswordCrackerReplyMessage{}, passwordCracker.execPasswordCrackerReplyMessage)
	conf.MessageRegistry.RegisterMessageCallback(
		types.PasswordCrackerUpdDictRangeMessage{}, passwordCracker.execPasswordCrackerUpdDictRangeMessage)
	return &passwordCracker
}

type PasswordCracker struct {
	address  string
	conf     *peer.Configuration // The configuration contains Socket and MessageRegistry
	message  *message.Message    // Messaging used to communicate among nodes
	chord    *chord.Chord        // chord used for find the correct receptor
	hashAlgo crypto.Hash         // The algorithm that is used to compute from the password to hash
	tasks    *sync.Map           // The tasks that this node have published
}

// SubmitRequest submits the password cracking request to another peer using DHT
func (p *PasswordCracker) SubmitRequest(hashStr string, saltStr string) error {
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		return err
	}
	salt, err := hex.DecodeString(saltStr)
	if err != nil {
		return err
	}

	// Query Chord using the salt value as the key
	saltInt := uint(big.NewInt(0).SetBytes(salt).Uint64())
	receptor, err := p.chord.QuerySuccessor(p.address, saltInt)
	if err != nil {
		return err
	}

	// TODO:
	//  Blockchain & Broadcast transaction, and wait for the transaction gets committed
	//  the transaction probably should include in the message sent to the receptor

	// Prepare a password cracking request to the receptor
	passwordCrackerReqMsg := types.PasswordCrackerRequestMessage{
		Hash: hash,
		Salt: salt,
	}
	passwordCrackerReqMsgTrans, err := p.conf.MessageRegistry.MarshalMessage(passwordCrackerReqMsg)
	if err != nil {
		return err
	}

	// Store this task into the tasks pool
	task := map[string]string{"password": ""}
	taskKey := hex.EncodeToString(append(hash, salt...))
	p.tasks.Store(taskKey, task)

	// Unicast to the receptor and return
	return p.message.Unicast(receptor, passwordCrackerReqMsgTrans)
}

// ReceiveResult receives the results for tasks that we have already submitted
func (p *PasswordCracker) ReceiveResult(hashStr string, saltStr string) string {
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("Unable decode hash string [%s]", hashStr))
		return ""
	}
	salt, err := hex.DecodeString(saltStr)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("Unable decode salt string [%s]", saltStr))
		return ""
	}

	taskKey := hex.EncodeToString(append(hash, salt...))
	taskResult, ok := p.tasks.Load(taskKey)
	if !ok {
		log.Error().Err(xerrors.Errorf("PasswordCracker Wrong Key")).Msg(
			fmt.Sprintf("Unable to locate the task with hash [%x], salt [%x]", hash, salt))
		return ""
	}
	password := taskResult.(map[string]string)["password"]
	p.tasks.Delete(taskKey)

	// TODO: If the password is empty, we should reclaim our money back, using TXN

	return password
}
