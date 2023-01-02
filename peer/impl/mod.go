package impl

import (
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/chord"
	"go.dedis.ch/cs438/peer/impl/consensus"
	"go.dedis.ch/cs438/peer/impl/daemon"
	"go.dedis.ch/cs438/peer/impl/fileshare"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/transport"
	"io"
	"regexp"
	"time"
)

// node implements a peer to build a Peerster system
//
// - implements peer.Peer
type node struct {
	peer.Peer                            // The node implements peer.Peer
	address   string                     // The node's address
	conf      *peer.Configuration        // The configuration contains Socket and MessageRegistry
	message   *message.MessageModule     // message module, handles packet sending
	daemon    *daemon.DaemonModule       // daemon module, runs all daemons
	file      *fileshare.FileModule      // file module, handles file upload download
	consensus *consensus.ConsensusModule // The node's consensus component
	chord     *chord.ChordModule         // TODO
}

// NewPeer creates a new peer. You can change the content and location of this
// function but you MUST NOT change its signature and package location.
func NewPeer(conf peer.Configuration) peer.Peer {
	messageMod := message.NewMessageModule(&conf)
	daemonMod := daemon.NewDaemonModule(&conf, messageMod)
	fileMod := fileshare.NewFileModule(&conf, messageMod)
	consensusMod := consensus.NewConsensusModule(&conf, messageMod)
	chordMod := chord.NewChordModule(&conf, messageMod)

	n := node{
		address:   conf.Socket.GetAddress(),
		conf:      &conf,
		message:   messageMod,
		daemon:    daemonMod,
		file:      fileMod,
		consensus: consensusMod,
		chord:     chordMod,
	}

	return &n
}

// Start implements peer.Service
func (n *node) Start() error {
	return n.daemon.Start()
}

// Stop implements peer.Service
func (n *node) Stop() error {
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
