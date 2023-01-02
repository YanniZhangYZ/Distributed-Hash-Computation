package impl

import (
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/chord"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"io"
	"regexp"
	"sync"
	"time"
)

type originInfoEntry struct {
	nextPeer string
	seq      uint
	rumors   []types.Rumor
}

// node implements a peer to build a Peerster system
//
// - implements peer.Peer
type node struct {
	peer.Peer                     // The node implements peer.Peer
	address   string              // The node's address
	conf      *peer.Configuration // The configuration contains Socket and MessageRegistry
	message   *MessageModule      // message module, handles packet sending
	daemon    *DaemonModule       // daemon module, runs all daemons
	file      *FileModule         // file module, handles file upload download
	consensus *ConsensusModule    // The node's consensus component
	chord     *chord.Module       // TODO
}

// NewPeer creates a new peer. You can change the content and location of this
// function but you MUST NOT change its signature and package location.
func NewPeer(conf peer.Configuration) peer.Peer {
	var stopChan = make(chan bool, 1)
	var originTable, async, catalog, fullKnown, seenRequest sync.Map

	/* The routing table should have the peer's own address */
	originTable.Store(conf.Socket.GetAddress(), originInfoEntry{
		conf.Socket.GetAddress(),
		uint(0),
		[]types.Rumor{}})

	message := MessageModule{
		address:     conf.Socket.GetAddress(),
		conf:        &conf,
		originTable: &originTable,
		async:       &async,
		seenRequest: &seenRequest,
	}

	daemon := DaemonModule{
		address:  conf.Socket.GetAddress(),
		conf:     &conf,
		message:  &message,
		stopChan: stopChan,
	}

	file := FileModule{
		address:   conf.Socket.GetAddress(),
		conf:      &conf,
		message:   &message,
		catalog:   &catalog,
		fullKnown: &fullKnown,
	}

	consensus := ConsensusModule{
		address:       conf.Socket.GetAddress(),
		conf:          &conf,
		message:       &message,
		threshold:     conf.PaxosThreshold(conf.TotalPeers),
		totalPeers:    conf.TotalPeers,
		tlcCnt:        make(map[uint]int),
		tlcValue:      make(map[uint]*types.BlockchainBlock),
		tlcChangeChan: make(chan *types.BlockchainBlock, 1000),
	}
	consensus.cond = sync.NewCond(&consensus.RWMutex)
	consensus.createNewPaxos()

	chordModule := chord.Module{} // TODO

	n := node{
		address:   conf.Socket.GetAddress(),
		conf:      &conf,
		message:   &message,
		daemon:    &daemon,
		file:      &file,
		consensus: &consensus,
		chord:     &chordModule,
	}

	/* Register different message callbacks */
	conf.MessageRegistry.RegisterMessageCallback(types.ChatMessage{}, message.execChatMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.RumorsMessage{}, message.execRumorsMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.StatusMessage{}, message.execStatusMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.AckMessage{}, message.execAckMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.EmptyMessage{}, message.execEmptyMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.PrivateMessage{}, message.execPrivateMessage)

	/* File sharing callbacks */
	conf.MessageRegistry.RegisterMessageCallback(types.DataRequestMessage{}, file.execDataRequestMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.DataReplyMessage{}, file.execDataReplyMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.SearchRequestMessage{}, file.execSearchRequestMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.SearchReplyMessage{}, file.execSearchReplyMessage)

	/* Consensus callbacks */
	conf.MessageRegistry.RegisterMessageCallback(types.PaxosPrepareMessage{}, consensus.execPaxosPrepareMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.PaxosProposeMessage{}, consensus.execPaxosProposeMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.PaxosPromiseMessage{}, consensus.execPaxosPromiseMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.PaxosAcceptMessage{}, consensus.execPaxosAcceptMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.TLCMessage{}, consensus.execTLCMessage)
	return &n
}

// Start implements peer.Service
func (n *node) Start() error {
	return n.daemon.start()
}

// Stop implements peer.Service
func (n *node) Stop() error {
	return n.daemon.stop()
}

// AddPeer implements peer.Service
func (n *node) AddPeer(addr ...string) {
	n.message.addPeer(addr...)
}

// GetRoutingTable implements peer.Service
func (n *node) GetRoutingTable() peer.RoutingTable {
	return n.message.getRoutingTable()
}

// SetRoutingEntry implements peer.Service
func (n *node) SetRoutingEntry(origin, relayAddr string) {
	n.message.setRoutingEntry(origin, relayAddr)
}

// Unicast implements peer.Messaging
func (n *node) Unicast(dest string, msg transport.Message) error {
	return n.message.unicast(dest, msg)
}

// Broadcast implements peer.Messaging
func (n *node) Broadcast(msg transport.Message) error {
	return n.message.broadcast(msg)
}

// Upload implements peer.DataSharing
func (n *node) Upload(data io.Reader) (string, error) {
	return n.file.upload(data)
}

// Download implements peer.DataSharing
func (n *node) Download(metahash string) ([]byte, error) {
	return n.file.download(metahash)
}

// Tag implements peer.DataSharing
func (n *node) Tag(name string, mh string) error {
	if n.conf.TotalPeers > 1 {
		/* Use consensus to tag name to metahash */
		return n.consensus.tag(name, mh)
	}
	return n.file.tag(name, mh)
}

// Resolve implements peer.DataSharing
func (n *node) Resolve(name string) string {
	return n.file.resolve(name)
}

// GetCatalog implements peer.DataSharing
func (n *node) GetCatalog() peer.Catalog {
	return n.file.getCatalog()
}

// UpdateCatalog implements peer.DataSharing
func (n *node) UpdateCatalog(key string, peer string) {
	n.file.updateCatalog(key, peer)
}

// SearchAll implements peer.DataSharing
func (n *node) SearchAll(reg regexp.Regexp, budget uint, timeout time.Duration) (names []string, err error) {
	return n.file.searchAll(reg, budget, timeout)
}

// SearchFirst implements peer.DataSharing
func (n *node) SearchFirst(pattern regexp.Regexp, conf peer.ExpandingRing) (string, error) {
	return n.file.searchFirst(pattern, conf)
}
