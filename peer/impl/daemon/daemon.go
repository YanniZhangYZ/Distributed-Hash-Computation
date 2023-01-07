package daemon

import (
	"errors"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl/message"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"log"
	"sync"
	"time"
)

func NewDaemon(conf *peer.Configuration, message *message.Message) *Daemon {
	daemon := Daemon{
		address:             conf.Socket.GetAddress(),
		conf:                conf,
		message:             message,
		stopListenChan:      make(chan bool, 1),
		stopAntiEntropyChan: make(chan bool, 1),
		stopHeartbeatChan:   make(chan bool, 1),
	}
	return &daemon
}

type Daemon struct {
	address             string              // The node's address
	conf                *peer.Configuration // The configuration contains Socket and MessageRegistry
	message             *message.Message
	stopListenChan      chan bool
	stopAntiEntropyChan chan bool
	stopHeartbeatChan   chan bool
	blockchainWaitGroup sync.WaitGroup
}

func (d *Daemon) Start() error {
	/* Start listening to the socket */
	go d.listenDaemon()
	/* Start the anti-entropy daemon*/
	go d.antiEntropyDaemon()
	/* Start the heartbeat daemon */
	go d.heartbeatDaemon()
	return nil
}

func (d *Daemon) Stop() error {
	d.stopListenChan <- true
	d.stopAntiEntropyChan <- true
	d.stopHeartbeatChan <- true
	return nil
}

func (d *Daemon) listenDaemon() {
	for {
		select {
		case <-d.stopListenChan:
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

func (d *Daemon) antiEntropyDaemon() {
	if d.conf.AntiEntropyInterval == 0 {
		/* Anti-entropy mechanism is disabled */
		return
	}

	ticker := time.NewTicker(d.conf.AntiEntropyInterval)
	for {
		select {
		case <-d.stopAntiEntropyChan:
			/* The node receives the stop message from the Stop() function,
			exit from the goroutine */
			ticker.Stop()
			return
		case <-ticker.C:
			/* Send the status information to a random neighbor */
			statusMsgTrans, err := d.message.CreateStatusMessageTrans()
			if err != nil {
				log.Panicln("AntiEntropyDaemon: ", err)
			}

			/* Select a random node to send the rumor Message */
			directNeighborSet := d.message.DirectNeighbor(map[string]struct{}{})
			if len(directNeighborSet) > 0 {
				rumorNeighbor := d.message.SelectRandomNeighbor(directNeighborSet)
				err = d.message.SendDirectMsg(rumorNeighbor, rumorNeighbor, statusMsgTrans)
				if err != nil {
					log.Panicln("AntiEntropyDaemon: ", err)
				}
			}
		}
	}

}

func (d *Daemon) heartbeatDaemon() {
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

	err = d.message.Broadcast(emptyMsgTrans)
	if err != nil {
		log.Panicln("HeartbeatDaemon: ", err)
	}

	/* Send at a regular interval */
	ticker := time.NewTicker(d.conf.HeartbeatInterval)
	for {
		select {
		case <-d.stopHeartbeatChan:
			/* The node receives the stop message from the Stop() function,
			exit from the goroutine */
			ticker.Stop()
			return
		case <-ticker.C:
			/* Send the rumor with empty embedded message to a random neighbor */
			err = d.message.Broadcast(emptyMsgTrans)
			if err != nil {
				log.Panicln("AntiEntropyDaemon: ", err)
			}
		}
	}

}
