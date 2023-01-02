package message

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"log"
	"math"
	"math/rand"
	"time"
)

func (m *MessageModule) CreateStatusMessageTrans() (transport.Message, error) {
	statusMsg := m.CreateStatusMessage()
	statusMsgTrans, err := m.conf.MessageRegistry.MarshalMessage(statusMsg)
	return statusMsgTrans, err
}

func (m *MessageModule) CreateStatusMessage() types.StatusMessage {
	statusMsg := make(types.StatusMessage)
	m.originTable.Range(func(origin, originInfo interface{}) bool {
		if originInfo.(originInfoEntry).seq > 0 {
			statusMsg[origin.(string)] = originInfo.(originInfoEntry).seq
		}
		return true
	})
	return statusMsg
}

func (m *MessageModule) SendDirectMsg(nextPeer string, dest string, msg transport.Message) error {
	header := transport.NewHeader(m.address, m.address, dest, 0)
	pkt := transport.Packet{
		Header: &header,
		Msg:    &msg,
	}
	err := m.conf.Socket.Send(nextPeer, pkt, 0)
	return err
}

func (m *MessageModule) DirectNeighbor(except map[string]struct{}) map[string]struct{} {
	/* Select a direct neighbor set that are outside the set "except", and the neighbor is not ourselves */
	directNeighborSet := map[string]struct{}{}
	m.originTable.Range(func(origin, originInfo interface{}) bool {
		_, ok := except[originInfo.(originInfoEntry).nextPeer]
		if originInfo.(originInfoEntry).nextPeer != m.address && !ok {
			directNeighborSet[originInfo.(originInfoEntry).nextPeer] = struct{}{}
		}
		return true
	})
	return directNeighborSet
}

func (m *MessageModule) RemoteNeighbor(except map[string]struct{}) map[string]struct{} {
	/* Select a remote neighbor set that are outside the set "except", and the neighbor is not ourselves */
	remoteNeighborSet := map[string]struct{}{}
	m.originTable.Range(func(origin, originInfo interface{}) bool {
		_, ok := except[origin.(string)]
		if origin.(string) != m.address && !ok {
			remoteNeighborSet[origin.(string)] = struct{}{}
		}
		return true
	})
	return remoteNeighborSet
}

func (m *MessageModule) SelectRandomNeighbor(neighborSet map[string]struct{}) string {
	/* Read out the neighbors to an array */
	neighbors := make([]string, len(neighborSet))
	i := 0
	for k := range neighborSet {
		neighbors[i] = k
		i++
	}

	/* Select a random neighbor */
	if len(neighbors) > 0 {
		randomNeighbor := neighbors[rand.Intn(len(neighbors))]
		return randomNeighbor
	}
	return ""
}

func (m *MessageModule) SelectKNeighbors(budget uint, neighborSet map[string]struct{}) ([]string, []uint) {
	if budget > uint(len(neighborSet)) {
		/* We will send to all neighbors but with different budgets */
		budgets := make([]uint, len(neighborSet))
		for idx := range budgets {
			budgets[idx] = budget / uint(len(neighborSet))
			if uint(idx) < budget%uint(len(neighborSet)) {
				budgets[idx]++
			}
		}

		neighbors := make([]string, len(neighborSet))
		i := 0
		for k := range neighborSet {
			neighbors[i] = k
			i++
		}
		return neighbors, budgets
	}

	/* We will select a neighbor pool to send to, all neighbors get a budget 1 */
	budgets := make([]uint, budget)
	for idx := range budgets {
		budgets[idx] = 1
	}

	var neighbors []string
	neighborSetCopy := map[string]struct{}{}
	for k := range neighborSet {
		neighborSetCopy[k] = struct{}{}
	}
	for uint(len(neighbors)) < budget {
		randomNeighbor := m.SelectRandomNeighbor(neighborSetCopy)
		neighbors = append(neighbors, randomNeighbor)
		delete(neighborSetCopy, randomNeighbor)
	}
	return neighbors, budgets
}

func (m *MessageModule) checkLocalRemoteSync(statusMsg *types.StatusMessage,
	source string) (bool, bool, bool, []types.Rumor) {
	localMissing := false
	remoteMissing := false
	localSync := true
	var remoteMissingRumors []types.Rumor

	for remoteOrigin, remoteSeq := range *statusMsg {
		localOriginInfo, ok := m.originTable.Load(remoteOrigin)
		/* If there is no local entry, we are missing rumors from the remote */
		localMissing = localMissing || (!ok)
		if !ok {
			continue
		}

		/* If there is a local entry, we still have to check the sequence number */
		if localOriginInfo.(originInfoEntry).seq != remoteSeq {
			/* If there is a mismatch, then it is not synced */
			localSync = false
			if localOriginInfo.(originInfoEntry).seq > remoteSeq && remoteOrigin != source {
				/* Remote is missing rumors, append the missing rumors to the missing list */
				/* Ignore sending missing rumors to the source of the status message */
				remoteMissing = true
				remoteMissingRumors = append(remoteMissingRumors,
					localOriginInfo.(originInfoEntry).rumors[remoteSeq:]...)
			} else if localOriginInfo.(originInfoEntry).seq < remoteSeq {
				/* Local is missing messages */
				localMissing = true
			}
		}

	}

	/* Find local entries that are not inside the remote status message (remote missing some entries) */
	m.originTable.Range(func(origin, originInfo interface{}) bool {
		_, ok := (*statusMsg)[origin.(string)]
		if !ok && originInfo.(originInfoEntry).seq > 0 && origin.(string) != source {
			/* This local entry does not exist in the remote peer, send all rumors to it */
			localSync = false
			remoteMissing = true
			remoteMissingRumors = append(remoteMissingRumors, originInfo.(originInfoEntry).rumors...)
		}
		return true
	})

	return localMissing, remoteMissing, localSync, remoteMissingRumors
}

func (m *MessageModule) tryNewNeighbor(previousTargets map[string]struct{},
	source string, rumorMsgTrans transport.Message) {
	directNeighborSet := m.DirectNeighbor(previousTargets)
	if len(directNeighborSet) == 0 {
		/* We cannot find another neighbor */
		return
	}
	rumorNeighbor := m.SelectRandomNeighbor(directNeighborSet)

	header := transport.NewHeader(source, m.address, rumorNeighbor, 0)
	pkt := transport.Packet{
		Header: &header,
		Msg:    &rumorMsgTrans,
	}

	/* Make a channel for ackMessage handler to ack the reception of the packet */
	pktChan := make(chan int, 1)
	m.Async.Store(pkt.Header.PacketID, pktChan)

	err := m.conf.Socket.Send(rumorNeighbor, pkt, 0)
	if err != nil {
		log.Panicln("TryNewNeighbor: ", m.address, err)
	}

	/* If the timeout is not specified, wait forever */
	var timeoutDur time.Duration
	timeoutDur = math.MaxInt64
	if m.conf.AckTimeout > 0 {
		timeoutDur = m.conf.AckTimeout
	}

	/* Either we receive the timeout or we receive the ACK message */
	go func() {
		select {
		case <-pktChan:
			/* Delete the entry in the ack channels */
			m.Async.Delete(pkt.Header.PacketID)
			return
		case <-time.After(timeoutDur):
			m.Async.Delete(pkt.Header.PacketID)
			previousTargets[rumorNeighbor] = struct{}{}
			m.tryNewNeighbor(previousTargets, source, rumorMsgTrans)
			return
		}
	}()
}
