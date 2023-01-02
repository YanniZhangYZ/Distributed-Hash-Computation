package impl

import (
	"errors"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"log"
	"time"
)

type DaemonModule struct {
	address  string              // The node's address
	conf     *peer.Configuration // The configuration contains Socket and MessageRegistry
	message  *MessageModule
	stopChan chan bool // Communication channel about whether we should stop the node
}

func (d *DaemonModule) start() error {
	/* Start listening to the socket */
	go d.listenDaemon()
	/* Start the anti-entropy daemon*/
	go d.antiEntropyDaemon()
	/* Start the heartbeat daemon */
	go d.heartbeatDaemon()
	return nil
}

func (d *DaemonModule) stop() error {
	d.stopChan <- true
	return nil
}

func (d *DaemonModule) listenDaemon() {
	for {
		select {
		case <-d.stopChan:
			/* The node receives the stop message from the Stop() function,
			exit from the goroutine */
			return
		default:
			pkt, err := d.conf.Socket.Recv(time.Second * 1)
			if errors.Is(err, transport.TimeoutError(0)) {
				/* The socket is unable to receive a message from the specified duration with a
				timeout error. It should continue listening, i.e., ignoring the error. */
				continue
			}

			/* If the packet's destination is the node */
			if pkt.Header.Destination == d.address {
				go func() {
					err = d.conf.MessageRegistry.ProcessPacket(pkt)
					if err != nil {
						log.Panicln("ListenDaemon: ", err)
					}
				}()
			} else {
				pkt.Header.RelayedBy = d.address
				err = d.conf.Socket.Send(pkt.Header.Destination, pkt, 0)
				if err != nil {
					log.Panicln("ListenDaemon: ", err)
				}
			}
		}
	}
}

func (d *DaemonModule) antiEntropyDaemon() {
	if d.conf.AntiEntropyInterval == 0 {
		/* Anti-entropy mechanism is disabled */
		return
	}

	ticker := time.NewTicker(d.conf.AntiEntropyInterval)
	for {
		select {
		case <-d.stopChan:
			/* The node receives the stop message from the Stop() function,
			exit from the goroutine */
			ticker.Stop()
			return
		case <-ticker.C:
			/* Send the status information to a random neighbor */
			statusMsgTrans, err := d.message.createStatusMessageTrans()
			if err != nil {
				log.Panicln("AntiEntropyDaemon: ", err)
			}

			/* Select a random node to send the rumor Message */
			directNeighborSet := d.message.directNeighbor(map[string]struct{}{})
			if len(directNeighborSet) > 0 {
				rumorNeighbor := d.message.selectRandomNeighbor(directNeighborSet)
				err = d.message.sendDirectMsg(rumorNeighbor, rumorNeighbor, statusMsgTrans)
				if err != nil {
					log.Panicln("AntiEntropyDaemon: ", err)
				}
			}
		}
	}

}

func (d *DaemonModule) heartbeatDaemon() {
	if d.conf.HeartbeatInterval == 0 {
		/* Heartbeat mechanism is disabled */
		return
	}

	/* Send a heartbeat at bootstrap */
	emptyMsg := types.EmptyMessage{}
	emptyMsgTrans, err := d.conf.MessageRegistry.MarshalMessage(emptyMsg)
	if err != nil {
		log.Panicln("HeartbeatDaemon: ", err)
	}

	err = d.message.broadcast(emptyMsgTrans)
	if err != nil {
		log.Panicln("HeartbeatDaemon: ", err)
	}

	/* Send at a regular interval */
	ticker := time.NewTicker(d.conf.HeartbeatInterval)
	for {
		select {
		case <-d.stopChan:
			/* The node receives the stop message from the Stop() function,
			exit from the goroutine */
			ticker.Stop()
			return
		case <-ticker.C:
			/* Send the rumor with empty embedded message to a random neighbor */
			err = d.message.broadcast(emptyMsgTrans)
			if err != nil {
				log.Panicln("AntiEntropyDaemon: ", err)
			}
		}
	}

}
