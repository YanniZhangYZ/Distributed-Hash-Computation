package passwordcracker

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/blockchain/blockchain"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/chord"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"math/big"
	"strings"
	"sync"
	"time"
)

// defaultDict is the default dictionary used for the dictionary attack, it can be replaced by reading a word
// list from a file, for example
var defaultDict = [...]string{
	"apple", "ball", "cat", "doll", "egg"}

func NewPasswordCracker(conf *peer.Configuration, message *message.Message, chord *chord.Chord, blockchain *blockchain.Blockchain) *PasswordCracker {
	var tasks sync.Map
	passwordCracker := PasswordCracker{
		address:    conf.Socket.GetAddress(),
		conf:       conf,
		message:    message,
		chord:      chord,
		blockchain: blockchain,
		hashAlgo:   conf.PasswordHashAlgorithm,
		tasks:      &tasks,
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
	address     string
	conf        *peer.Configuration    // The configuration contains Socket and MessageRegistry
	message     *message.Message       // Messaging used to communicate among nodes
	chord       *chord.Chord           // chord used for find the correct receptor
	blockchain  *blockchain.Blockchain // Blockchain used for submit request and execute contract
	hashAlgo    crypto.Hash            // The algorithm that is used to compute from the password to hash
	tasks       *sync.Map              // The tasks that this node have published
	dictUpdLock sync.Mutex             // The dictionary update lock
}

// SubmitRequest submits the password cracking request to another peer using DHT
func (p *PasswordCracker) SubmitRequest(hashStr string, saltStr string, reward int, timeout time.Duration) error {
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

	// Propose a password-cracking smart contract to the blockchain
	// It blocks until the ContractDeployTx has been confirmed
	contractAddr := ""
	if timeout > 0 {
		receptorBlockchainAddr := strings.Split(receptor, ":")[1]
		contractAddr, err = p.blockchain.ProposeContract(hashStr, saltStr, int64(reward), receptorBlockchainAddr, timeout)
		if err != nil {
			return err
		}
	}

	// Prepare a password cracking request to the receptor
	passwordCrackerReqMsg := types.PasswordCrackerRequestMessage{
		Hash:            hash,
		Salt:            salt,
		ContractAddress: common.StringToAddress(contractAddr),
	}
	passwordCrackerReqMsgTrans, err := p.conf.MessageRegistry.MarshalMessage(passwordCrackerReqMsg)
	if err != nil {
		return err
	}

	// Store this task into the tasks pool
	task := map[string]string{"password": ""}
	taskKey := hex.EncodeToString(append(hash, salt...))
	p.tasks.Store(taskKey, task)

	// SendDirectMsg to the receptor and return
	return p.message.SendDirectMsg(receptor, receptor, passwordCrackerReqMsgTrans)
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
	return password
}
