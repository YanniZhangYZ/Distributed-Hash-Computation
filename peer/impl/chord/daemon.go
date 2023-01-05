package chord

import (
	"fmt"
	"github.com/rs/zerolog/log"
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
			// The node receives the stop message from the StopDaemon() function,
			// exit from the goroutine
			ticker.Stop()
			return
		case <-ticker.C:
			chordQueryMsg := types.ChordQueryPredecessorMessage{}
			chordQueryMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordQueryMsg)
			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("[%s] stabilizeDaemon MarshalMessage failed!", c.address))
			}

			c.successorLock.RLock()
			// If we have a successor, send a query message to it.
			if c.successor != "" && c.successor != c.address {
				err = c.message.Unicast(c.successor, chordQueryMsgTrans)
				if err != nil {
					log.Error().Err(err).Msg(fmt.Sprintf("[%s] stabilizeDaemon Unicast with error!", c.address))
				}
			}
			c.successorLock.RUnlock()
		}
	}
}

// fixFingerDaemon fix the finger table of a Chord node. After a fixed interval, it
// will ask inside the network about the newest information of a finger table, and update
// the entry accordingly.
func (c *Chord) fixFingerDaemon() {
	if c.conf.ChordFixFingerInterval == 0 {
		// Fix finger mechanism is disabled
		return
	}

	ticker := time.NewTicker(c.conf.ChordFixFingerInterval)
	for {
		select {
		case <-c.stopFixFingerChan:
			// The node receives the stop message from the StopDaemon() function,
			// exit from the goroutine
			ticker.Stop()
			return
		case <-ticker.C:
			// Update our finger table
			fingerStart, _ := c.fingerStartEnd(c.fingerIdx)
			successor, err := c.querySuccessor(c.address, fingerStart)
			if err != nil {
				log.Error().Err(err).Msg(
					fmt.Sprintf("[%s] fixFingerDaemon querySuccessor with error for index %d!",
						c.address, c.fingerIdx))
			}

			c.fingersLock.Lock()
			c.fingers[c.fingerIdx] = successor
			c.fingersLock.Unlock()
			if c.fingerIdx == 0 {
				c.successorLock.Lock()
				c.successor = successor
				c.successorLock.Unlock()
			}
			c.fingerIdx = (c.fingerIdx + 1) % len(c.fingers)
		}
	}
}
