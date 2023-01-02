package impl

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"log"
	"math/rand"
)

func (m *MessageModule) execChatMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	chatMsg, ok := msg.(*types.ChatMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	log.Println(chatMsg)

	return nil
}

func (m *MessageModule) execEmptyMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	_, ok := msg.(*types.EmptyMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	return nil
}

func (m *MessageModule) execPrivateMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	privateMsg, ok := msg.(*types.PrivateMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	_, ok = privateMsg.Recipients[m.address]
	if ok {
		/* The node is in the recipient list, process the message locally */
		header := transport.NewHeader(m.address, m.address, m.address, 0)
		localPkt := transport.Packet{Header: &header, Msg: privateMsg.Msg}
		err := m.conf.MessageRegistry.ProcessPacket(localPkt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MessageModule) execRumorsMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	rumorsMsg, ok := msg.(*types.RumorsMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Update our routing table based on the pkt header */
	/* If no entry is found for the source, add it to our routing table */
	m.originTableUpdateLock.Lock()
	_, ok = m.originTable.Load(pkt.Header.Source)
	if !ok {
		m.originTable.Store(pkt.Header.Source,
			originInfoEntry{
				pkt.Header.RelayedBy,
				uint(0),
				[]types.Rumor{},
			})
	}
	m.originTableUpdateLock.Unlock()

	/* Check the rumor is new */
	anyNew := false
	for _, rumor := range rumorsMsg.Rumors {
		m.originTableUpdateLock.Lock()
		originInfo, ok := m.originTable.Load(rumor.Origin)
		prevSeq := uint(0)
		var rumors []types.Rumor
		if ok {
			/* If there is an entry from the node */
			prevSeq = originInfo.(originInfoEntry).seq
			rumors = originInfo.(originInfoEntry).rumors
		}

		/* If it is the new rumor that we are waiting for */
		if prevSeq+1 == rumor.Sequence {
			anyNew = true
			/* Update the status table, process the message locally */
			m.originTable.Store(rumor.Origin,
				originInfoEntry{
					pkt.Header.RelayedBy,
					rumor.Sequence,
					append(rumors, rumor)})
			m.originTableUpdateLock.Unlock()

			localPkt := transport.Packet{
				Header: pkt.Header,
				Msg:    rumor.Msg,
			}
			err := m.conf.MessageRegistry.ProcessPacket(localPkt)
			if err != nil {
				return err
			}
		} else {
			m.originTableUpdateLock.Unlock()
		}
	}

	/* If there is some new rumors inside, re-broadcast the rumor to a random neighbor */
	if anyNew {
		/* Change the destination and relayed, and send the msg out */
		msgTrans, _ := m.conf.MessageRegistry.MarshalMessage(msg)
		previousRelay := map[string]struct{}{}
		previousRelay[pkt.Header.RelayedBy] = struct{}{}
		m.tryNewNeighbor(previousRelay, pkt.Header.Source, msgTrans)
	}

	/* Send back the ACK message direct, to mark the reception of packet */
	statusMsg := m.createStatusMessage()
	ackMsg := types.AckMessage{
		AckedPacketID: pkt.Header.PacketID,
		Status:        statusMsg,
	}
	ackMsgTrans, err := m.conf.MessageRegistry.MarshalMessage(ackMsg)
	if err != nil {
		return err
	}
	return m.sendDirectMsg(pkt.Header.RelayedBy, pkt.Header.RelayedBy, ackMsgTrans)
}

func (m *MessageModule) execStatusMessage(msg types.Message, pkt transport.Packet) error {

	/* cast the message to its actual type. You assume it is the right type. */
	statusMsg, ok := msg.(*types.StatusMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	localMissing, remoteMissing, localSync, remoteMissingRumors := m.checkLocalRemoteSync(statusMsg, pkt.Header.Source)

	statusMsgTrans, err := m.createStatusMessageTrans()
	if err != nil {
		return err
	}

	/* Local is missing rumors, send our status to the remote */
	if localMissing {
		/* Send our status to the remote peer */
		err = m.sendDirectMsg(pkt.Header.Source, pkt.Header.Source, statusMsgTrans)
		if err != nil {
			return err
		}
	}

	/* Remote is missing rumors, send the missing rumors to it */
	if remoteMissing {
		rumorMsg := types.RumorsMessage{
			Rumors: remoteMissingRumors}
		rumorMsgTrans, err := m.conf.MessageRegistry.MarshalMessage(rumorMsg)
		if err != nil {
			return err
		}
		/* Send and do not wait for an ACK */
		err = m.sendDirectMsg(pkt.Header.Source, pkt.Header.Source, rumorMsgTrans)
		if err != nil {
			return err
		}
	}

	/* Both peers have the same view */
	if localSync && rand.Float64() < m.conf.ContinueMongering {
		/* Select a random node to send the rumor Message, with given probability */
		previousRelay := map[string]struct{}{}
		previousRelay[pkt.Header.RelayedBy] = struct{}{}
		directNeighborSet := m.directNeighbor(previousRelay)
		if len(directNeighborSet) == 0 {
			return nil
		}
		rumorNeighbor := m.selectRandomNeighbor(directNeighborSet)
		return m.sendDirectMsg(rumorNeighbor, rumorNeighbor, statusMsgTrans)
	}
	return nil
}

func (m *MessageModule) execAckMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	ackMsg, ok := msg.(*types.AckMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Load the channel from the map and send the finish signal, if we are waiting for it */
	finChan, ok := m.async.Load(ackMsg.AckedPacketID)
	if ok {
		finChan.(chan int) <- 1
	}

	/* Process the ackMsg locally with the embedded status message */
	ackMsgTrans, err := m.conf.MessageRegistry.MarshalMessage(ackMsg.Status)
	if err != nil {
		return err
	}
	statusPkt := transport.Packet{
		Header: pkt.Header,
		Msg:    &ackMsgTrans,
	}
	return m.conf.MessageRegistry.ProcessPacket(statusPkt)
}
