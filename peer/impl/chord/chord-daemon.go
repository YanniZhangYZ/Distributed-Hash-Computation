package chord

import (
	"go.dedis.ch/cs438/peer"
	"time"
)

type daemonModule struct {
	address           string              // The node's address
	conf              *peer.Configuration // The configuration contains Socket and MessageRegistry
	message           *peer.Messaging     // Messaging used to communicate among nodes
	stopStabilizeChan chan bool           // Communication channel about whether we should stop the node
}

func (d *daemonModule) start() error {
	return nil
}

func (d *daemonModule) stop() error {
	d.stopStabilizeChan <- true
	return nil
}

func (d *daemonModule) stabilizeDaemon() {
	if d.conf.ChordStabilizeInterval == 0 {
		/* Anti-entropy mechanism is disabled */
		return
	}

	ticker := time.NewTicker(d.conf.ChordStabilizeInterval)
	for {
		select {
		case <-d.stopStabilizeChan:
			/* The node receives the stop message from the Stop() function,
			exit from the goroutine */
			ticker.Stop()
			return
		case <-ticker.C:
			// Update the successor and predecessor

		}
	}

}
