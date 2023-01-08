package chord

import (
	"fmt"
	"github.com/rs/xid"
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
	/* Start the pingDaemon */
	go c.pingDaemon()
}

// StopDaemon stops daemon for Chord
func (c *Chord) StopDaemon() {
	c.stopStabilizeChan <- true
	c.stopFixFingerChan <- true
	c.stopPingChan <- true
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
			if c.fingerIdx == 0 {
				// We should only update finger entries that are not the successor
				c.fingerIdx++
			}
			fingerStart, _ := c.fingerStartEnd(c.fingerIdx)
			successor, err := c.QuerySuccessor(c.address, fingerStart)
			if err != nil {
				log.Error().Err(err).Msg(
					fmt.Sprintf("[%s] fixFingerDaemon querySuccessor with error for index %d!",
						c.address, c.fingerIdx))
			}

			c.fingersLock.Lock()
			c.fingers[c.fingerIdx] = successor
			c.fingersLock.Unlock()

			c.fingerIdx = (c.fingerIdx + 1) % len(c.fingers)
		}
	}
}

// pingDaemon
func (c *Chord) pingDaemon() {
	if c.conf.ChordPingInterval == 0 {
		// Fix finger mechanism is disabled
		return
	}

	ticker := time.NewTicker(c.conf.ChordPingInterval)
	for {
		select {
		case <-c.stopPingChan:
			// The node receives the stop message from the StopDaemon() function,
			// exit from the goroutine
			ticker.Stop()
			return
		case <-ticker.C:
			// Check for liveliness of the finger entry, except for the successor
			checkLiveliness := func(fingerEntry string, fingerIdx int) {
				// Prepare the new chord ping message
				chordPingMsg := types.ChordPingMessage{
					RequestID: xid.New().String(),
				}
				chordPingMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(chordPingMsg)
				if err != nil {
					log.Error().Err(err).Msg(
						fmt.Sprintf("[%s] pingDaemon querySuccessor MarshalMessage failed!", c.address))
				}

				// Prepare a reply channel that receives the reply from the remote peer, if any response is ready
				replyChan := make(chan bool, 1)
				c.pingChan.Store(chordPingMsg.RequestID, replyChan)

				// Send the message to the remote peer
				err = c.message.Unicast(fingerEntry, chordPingMsgTrans)
				if err != nil {
					log.Error().Err(err).Msg(
						fmt.Sprintf("[%s] pingDaemon Unicast failed!", c.address))
				}

				// Either we wait until the timeout, or we receive a response from the reply channel
				select {
				case <-replyChan:
					// The entry is still alive, continue
					return
				case <-time.After(c.conf.ChordPingInterval):
					// Timeout, we should set all entries contain expired value to empty
					c.fingersLock.Lock()
					if c.fingers[fingerIdx] == fingerEntry {
						c.fingers[fingerIdx] = ""
					}
					c.fingersLock.Unlock()
				}
			}

			c.fingersLock.RLock()
			for i := 1; i < len(c.fingers); i++ {
				if c.fingers[i] != "" {
					go checkLiveliness(c.fingers[i], i)
				}
			}
			c.fingersLock.RUnlock()
		}
	}
}
