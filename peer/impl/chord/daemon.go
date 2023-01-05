package chord

import (
	"go.dedis.ch/cs438/types"
	"time"
)

// StartDaemon starts daemon for Chord
func (c *Chord) StartDaemon() {
	/* Start the stabilizeDaemon */
	go c.stabilizeDaemon()
	/* Start the fixFingerDaemon */
	go c.fixFingerDaemon()
}

// StopDaemon stops daemon for Chord
func (c *Chord) StopDaemon() {
	c.stopStabilizeChan <- true
	c.stopFixFingerChan <- true
}

// stabilizeDaemon ensures the correctness of the Chord, it sends a QueryPredecessor
// message to our successor, if any. Upon receiving the reply from our successor, we will
// check in the callback that our information is up-to-date, and our successor has the
// correct predecessor as well
func (c *Chord) stabilizeDaemon() {
	if c.conf.ChordStabilizeInterval == 0 {
		// Stabilization mechanism is disabled
		return
	}

	ticker := time.NewTicker(c.conf.ChordStabilizeInterval)
	for {
		select {
		case <-c.stopStabilizeChan:
			// The node receives the stop message from the Stop() function,
			// exit from the goroutine
			ticker.Stop()
			return
		case <-ticker.C:
			chordQueryMsg := types.ChordQueryPredecessorMessage{}
			chordQueryMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordQueryMsg)
			if err != nil {
				panic(err)
			}

			c.successorLock.RLock()
			// If we have a successor, send a query message to it
			if c.successor != "" {
				err = c.message.Unicast(c.successor, chordQueryMsgTrans)
				if err != nil {
					panic(err)
				}
			}
			c.successorLock.RUnlock()
		}
	}
}

// TODO
func (c *Chord) fixFingerDaemon() {
	if c.conf.ChordFixFingerInterval == 0 {
		// Fix finger mechanism is disabled
		return
	}

	ticker := time.NewTicker(c.conf.ChordFixFingerInterval)
	for {
		select {
		case <-c.stopFixFingerChan:
			// The node receives the stop message from the Stop() function,
			// exit from the goroutine
			ticker.Stop()
			return
		case <-ticker.C:
			// Update our finger table

		}
	}
}
