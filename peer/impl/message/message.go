package message

import (
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"log"
	"sync"
)

type originInfoEntry struct {
	nextPeer string
	seq      uint
	rumors   []types.Rumor
}

func NewMessage(conf *peer.Configuration) *Message {
	var originTable, async, seenRequest sync.Map
	originTable.Store(conf.Socket.GetAddress(), originInfoEntry{
		conf.Socket.GetAddress(),
		uint(0),
		[]types.Rumor{}})

	message := Message{
		address:     conf.Socket.GetAddress(),
		conf:        conf,
		originTable: &originTable,
		Async:       &async,
		SeenRequest: &seenRequest,
	}

	/* Register different message callbacks */
	conf.MessageRegistry.RegisterMessageCallback(types.ChatMessage{}, message.execChatMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.RumorsMessage{}, message.execRumorsMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.StatusMessage{}, message.execStatusMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.AckMessage{}, message.execAckMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.EmptyMessage{}, message.execEmptyMessage)
	conf.MessageRegistry.RegisterMessageCallback(types.PrivateMessage{}, message.execPrivateMessage)

	return &message
}

type Message struct {
	address               string
	conf                  *peer.Configuration // The configuration contains Socket and MessageRegistry
	originTable           *sync.Map           // The table contains routing info, sequence number for each origin
	originTableUpdateLock sync.Mutex          // The mutex to protect (read from + write to) to the originTable
	Async                 *sync.Map           // The asynchronous notification channel for ACK / Chunks
	SeenRequest           *sync.Map           // File search duplicates
}

func (m *Message) AddPeer(addr ...string) {
	// Iterate and add all addresses
	m.originTableUpdateLock.Lock()
	defer m.originTableUpdateLock.Unlock()
	for _, a := range addr {
		m.originTable.Store(a, originInfoEntry{a, 0, []types.Rumor{}})
	}

	// Update the conf.TotalPeers according to the number of entries in originTable
	var i uint
	m.originTable.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	m.conf.TotalPeers = i
}

func (m *Message) GetRoutingTable() peer.RoutingTable {
	// Make a copy of the routing table
	var copyRoutingTable = make(peer.RoutingTable)

	m.originTable.Range(func(origin, originInfo interface{}) bool {
		copyRoutingTable[origin.(string)] = originInfo.(originInfoEntry).nextPeer
		return true
	})

	return copyRoutingTable
}

func (m *Message) SetRoutingEntry(origin, relayAddr string) {
	m.originTableUpdateLock.Lock()
	defer m.originTableUpdateLock.Unlock()
	if relayAddr == "" {
		// Remove the entry from the map
		m.originTable.Delete(origin)
	} else {
		// Update or create the entry
		originInfo, ok := m.originTable.Load(origin)
		if !ok {
			m.originTable.Store(origin, originInfoEntry{relayAddr, 0, []types.Rumor{}})
		} else {
			m.originTable.Store(origin, originInfoEntry{
				relayAddr,
				originInfo.(originInfoEntry).seq,
				originInfo.(originInfoEntry).rumors})
		}
	}
}

func (m *Message) Unicast(dest string, msg transport.Message) error {
	originInfo, ok := m.originTable.Load(dest)
	if !ok {
		return xerrors.Errorf("Unicast unknown address: %v %v", m.address, dest)
	}

	nextPeer := originInfo.(originInfoEntry).nextPeer

	header := transport.NewHeader(
		m.address, // source
		m.address, // relay
		dest,      // destination
		0,         // TTL
	)
	pkt := transport.Packet{
		Header: &header,
		Msg:    &msg,
	}

	// Send the packet to the next peer instead of the final destination, in case of relaying
	return m.conf.Socket.Send(nextPeer, pkt, 0)
}

func (m *Message) Broadcast(msg transport.Message) error {
	m.originTableUpdateLock.Lock()
	originInfo, _ := m.originTable.Load(m.address)
	seq := originInfo.(originInfoEntry).seq + 1

	rumorMsg := types.RumorsMessage{
		Rumors: []types.Rumor{{
			Origin:   m.address,
			Sequence: seq,
			Msg:      &msg,
		}}}
	rumorMsgTrans, err := m.conf.MessageRegistry.MarshalMessage(rumorMsg)
	if err != nil {
		m.originTableUpdateLock.Unlock()
		return err
	}

	/* Update the local info entry */
	m.originTable.Store(m.address,
		originInfoEntry{
			originInfo.(originInfoEntry).nextPeer,
			seq,
			append(originInfo.(originInfoEntry).rumors, rumorMsg.Rumors[0])})
	m.originTableUpdateLock.Unlock()

	/* Select a random node to send the rumor Message, and wait for the ACK (non-blocking) */
	m.tryNewNeighbor(map[string]struct{}{}, m.address, rumorMsgTrans)

	/* Process the message locally */
	header := transport.NewHeader(m.address, m.address, m.address, 0)
	pkt := transport.Packet{Header: &header, Msg: &msg}
	go func() {
		err = m.conf.MessageRegistry.ProcessPacket(pkt)
		if err != nil {
			log.Panicln("ListenDaemon: ", err)
		}
	}()
	return nil
}

func (m *Message) GetConf() *peer.Configuration {
	return m.conf
}
