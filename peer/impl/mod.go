package impl

import (
	"github.com/rs/zerolog"
	"io"
	"regexp"
	"time"

	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/blockchain/block"
	"go.dedis.ch/cs438/peer/impl/blockchain/blockchain"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/chord"
	"go.dedis.ch/cs438/peer/impl/consensus"
	"go.dedis.ch/cs438/peer/impl/daemon"
	"go.dedis.ch/cs438/peer/impl/fileshare"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/peer/impl/passwordcracker"
	"go.dedis.ch/cs438/transport"
)

// node implements a peer to build a Peerster system
//
// - implements peer.Peer
type node struct {
	peer.Peer                                        // The node implements peer.Peer
	address         string                           // The node's address
	conf            *peer.Configuration              // The configuration contains Socket and MessageRegistry
	message         *message.Message                 // message module, handles packet sending
	daemon          *daemon.Daemon                   // daemon module, runs all daemons
	file            *fileshare.File                  // file module, handles file upload download
	consensus       *consensus.Consensus             // The node's consensus component
	chord           *chord.Chord                     // The node's chord component (DHT)
	Blockchain      *blockchain.Blockchain           // The node's blockchain component (currently exposed for testing)
	passwordCracker *passwordcracker.PasswordCracker // The node's password cracker
}

// NewPeer creates a new peer. You can change the content and location of this
// function, but you MUST NOT change its signature and package location.
func NewPeer(conf peer.Configuration) peer.Peer {
	messageMod := message.NewMessage(&conf)
	daemonMod := daemon.NewDaemon(&conf, messageMod)
	fileMod := fileshare.NewFile(&conf, messageMod)
	consensusMod := consensus.NewConsensus(&conf, messageMod)
	chordMod := chord.NewChord(&conf, messageMod)
	blockchainMod := blockchain.NewBlockchain(&conf, messageMod, consensusMod, conf.Storage)
	passwordCracker := passwordcracker.NewPasswordCracker(&conf, messageMod, chordMod, blockchainMod)

	n := node{
		address:         conf.Socket.GetAddress(),
		conf:            &conf,
		message:         messageMod,
		daemon:          daemonMod,
		file:            fileMod,
		consensus:       consensusMod,
		chord:           chordMod,
		Blockchain:      blockchainMod,
		passwordCracker: passwordCracker,
	}

	return &n
}

// Start implements peer.Service
func (n *node) Start() error {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	n.chord.StartDaemon()
	n.Blockchain.Start()
	return n.daemon.Start()
}

// Stop implements peer.Service
func (n *node) Stop() error {
	n.chord.StopDaemon()
	n.Blockchain.Stop()
	return n.daemon.Stop()
}

// AddPeer implements peer.Service
func (n *node) AddPeer(addr ...string) {
	n.message.AddPeer(addr...)
}

// GetRoutingTable implements peer.Service
func (n *node) GetRoutingTable() peer.RoutingTable {
	return n.message.GetRoutingTable()
}

// SetRoutingEntry implements peer.Service
func (n *node) SetRoutingEntry(origin, relayAddr string) {
	n.message.SetRoutingEntry(origin, relayAddr)
}

// GetAddr implements peer.Messaging
func (n *node) GetAddr() string {
	return n.address
}

// Unicast implements peer.Messaging
func (n *node) Unicast(dest string, msg transport.Message) error {
	return n.message.Unicast(dest, msg)
}

// Broadcast implements peer.Messaging
func (n *node) Broadcast(msg transport.Message) error {
	return n.message.Broadcast(msg)
}

// Upload implements peer.DataSharing
func (n *node) Upload(data io.Reader) (string, error) {
	return n.file.Upload(data)
}

// Download implements peer.DataSharing
func (n *node) Download(metahash string) ([]byte, error) {
	return n.file.Download(metahash)
}

// Tag implements peer.DataSharing
func (n *node) Tag(name string, mh string) error {
	if n.conf.TotalPeers > 1 {
		/* Use consensus to tag name to metahash */
		return n.consensus.Tag(name, mh)
	}
	return n.file.Tag(name, mh)
}

// Resolve implements peer.DataSharing
func (n *node) Resolve(name string) string {
	return n.file.Resolve(name)
}

// GetCatalog implements peer.DataSharing
func (n *node) GetCatalog() peer.Catalog {
	return n.file.GetCatalog()
}

// UpdateCatalog implements peer.DataSharing
func (n *node) UpdateCatalog(key string, peer string) {
	n.file.UpdateCatalog(key, peer)
}

// SearchAll implements peer.DataSharing
func (n *node) SearchAll(reg regexp.Regexp, budget uint, timeout time.Duration) (names []string, err error) {
	return n.file.SearchAll(reg, budget, timeout)
}

// SearchFirst implements peer.DataSharing
func (n *node) SearchFirst(pattern regexp.Regexp, conf peer.ExpandingRing) (string, error) {
	return n.file.SearchFirst(pattern, conf)
}

// GetChordID implements peer.Chord
func (n *node) GetChordID() uint {
	return n.chord.GetChordID()
}

// QueryChordID implements peer.Chord
func (n *node) QueryChordID(addr string) uint {
	return n.chord.Name2ID(addr)
}

// GetPredecessor implements peer.Chord
func (n *node) GetPredecessor() string {
	return n.chord.GetPredecessor()
}

// GetSuccessor implements peer.Chord
func (n *node) GetSuccessor() string {
	return n.chord.GetSuccessor()
}

// GetFingerTable implements peer.Chord
func (n *node) GetFingerTable() []string {
	return n.chord.GetFingerTable()
}

// JoinChord implements peer.Chord
func (n *node) JoinChord(remoteNode string) error {
	return n.chord.Join(remoteNode)
}

// LeaveChord implements peer.Chord
func (n *node) LeaveChord() error {
	return n.chord.Leave()
}

// RingLen implements peer.Chord
func (n *node) RingLen() uint {
	return n.chord.RingLen()
}

// JoinBlockchain implements peer.IBlockchain
func (n *node) JoinBlockchain(balance int64, timeout time.Duration) error {
	return n.Blockchain.JoinBlockchain(balance, timeout)
}

// LeaveBlockchain informs the blockchain network of the leave of the account
func (n *node) LeaveBlockchain() error {
	return n.Blockchain.LeaveBlockchain()
}

// TransferMoney implements peer.IBlockchain
func (n *node) TransferMoney(dst common.Address, amount int64, timeout time.Duration) error {
	return n.Blockchain.TransferMoney(dst, amount, timeout)
}

// ProposeContract implements peer.IBlockchain
func (n *node) ProposeContract(hash string, salt string, reward int64, recipient string, timeout time.Duration) error {
	_, err := n.Blockchain.ProposeContract(hash, salt, reward, recipient, timeout)
	return err
}

// ExecuteContract implements peer.IBlockchain
func (n *node) ExecuteContract(password string, hash string, salt string, contractAddr string, timeout time.Duration) error {
	return n.Blockchain.ExecuteContract(password, hash, salt, contractAddr, timeout)
}

// GetAccountAddress implements peer.IBlockchain
func (n *node) GetAccountAddress() string {
	return n.Blockchain.GetAccountAddress()
}

// GetBalance implements peer.IBlockchain
func (n *node) GetBalance() int64 {
	return n.Blockchain.GetBalance()
}

// GetChain implements peer.IBlockchain
func (n *node) GetChain() *block.Chain {
	return n.Blockchain.GetChain()
}

// PasswordSubmitRequest implements peer.PasswordCracker
func (n *node) PasswordSubmitRequest(hashStr string, saltStr string, reward int, timeout time.Duration) error {
	return n.passwordCracker.SubmitRequest(hashStr, saltStr, reward, timeout)
}

// PasswordReceiveResult implements peer.PasswordCracker
func (n *node) PasswordReceiveResult(hashStr string, saltStr string) string {
	return n.passwordCracker.ReceiveResult(hashStr, saltStr)
}
