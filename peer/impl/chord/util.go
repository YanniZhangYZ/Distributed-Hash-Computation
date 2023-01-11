package chord

import (
	"crypto"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"math"
	"math/big"
)

// validRange checks that a given key is within a valid range, the value of the key is valid if
// it is greater than or equal to 0, and it is lower than the upperBound, the upperBound is
// defined by the ChordBytes inside the configuration
func (c *Chord) validRange(key uint) bool {
	// The upper bound of the hash value should be 2^(ChordBytes * 8)
	// If ChordBytes = 1, upperBound = 256
	// If ChordBytes = 2, upperBound = 65536
	upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))
	return key < upperBound
}

// Name2ID computes from the address to the chordID, with the given ChordBits limit
func (c *Chord) Name2ID(name string) uint {
	h := crypto.SHA256.New()
	h.Write([]byte(name))
	hashSlice := h.Sum(nil)

	// Crop the hashSlice to only the specified chord bits, which is the size of the salt value, i.e.,
	// if the salt is 16 bits, then conf.ChordBytes = 2
	hashSlice = hashSlice[:c.conf.ChordBytes]
	return uint(big.NewInt(0).SetBytes(hashSlice).Uint64())
}

// isPredecessor checks whether we are the predecessor of the given key, if we are, return true,
// otherwise, return false
func (c *Chord) isPredecessor(key uint) bool {
	// This is the initial state of the Chord ring, when we create it. We are the only node inside
	// the ring. Our successor is either set to empty or our own address, depending on the execution
	// of fix finger daemon. In this case, we will be both the predecessor and the successor of the given key.
	if c.successor == "" || c.successor == c.address {
		return true
	}

	successorID := c.Name2ID(c.successor)
	if successorID <= c.chordID {
		// If the successorID is smaller than our chordID, it means we are crossing the boundary of the
		// ring. For example, the successorID = 2, and c.chordID = 15, and the ring has length = 16. Since
		// we have checked the validity of the key before calling isPredecessor, therefore, we only need to
		// check the key is either larger than our chordID, or smaller than or equal to the successor ID.
		return c.chordID < key || key <= successorID
	}
	// This is the normal case, we only need to check the key is within the range (c.chordID, successorID]
	return c.chordID < key && key <= successorID
}

// fingerStartEnd computes the interval of a finger, it returns two uint, indicates the start and end. The
// finger interval is [start, end)
func (c *Chord) fingerStartEnd(idx int) (uint, uint) {
	upperBound := uint(math.Pow(2, float64(c.conf.ChordBytes)*8))
	fingerStart := (c.chordID + uint(math.Pow(2, float64(idx)))) % upperBound
	fingerEnd := (c.chordID + uint(math.Pow(2, float64(idx+1)))) % upperBound
	return fingerStart, fingerEnd
}

// closestPrecedingFinger returns the closest finger preceding ID
func (c *Chord) closestPrecedingFinger(key uint) string {
	for i := c.conf.ChordBytes*8 - 1; i >= 0; i-- {
		// If we already have this finger, check whether its start is in the range (c.chordID, key)
		// If c.chordID == key, then we should return the first non-empty entry we encountered during
		// the lookup.
		if c.fingers[i] != "" {
			fingerID := c.Name2ID(c.fingers[i])
			within := false

			if key < c.chordID {
				within = c.chordID < fingerID || fingerID < key
			} else {
				within = c.chordID < fingerID && fingerID < key
			}

			if within || c.chordID == key {
				return c.fingers[i]
			}
		}
	}
	return ""
}

// notifyPasswordCracker notifies the password cracker the change of predecessor of the node, i.e., the password
// cracker should change the pre-compute dictionary they are storing
func (c *Chord) notifyPasswordCracker() {
	newPredecessor := c.predecessor
	start := c.Name2ID(newPredecessor)
	updatePasswordCracker := func() {
		passwordCrackerUpdDictRangeMsg := types.PasswordCrackerUpdDictRangeMessage{
			Start: start,
			End:   c.chordID,
		}
		passwordCrackerUpdDictRangeMsgTrans, err :=
			c.conf.MessageRegistry.MarshalMessage(passwordCrackerUpdDictRangeMsg)
		if err != nil {
			log.Error().Err(err).Msg("updatePasswordCracker Marshal")
		}

		// Process message locally, it will update the password_cracker module
		header := transport.NewHeader(c.address, c.address, c.address, 0)
		localPkt := transport.Packet{Header: &header, Msg: &passwordCrackerUpdDictRangeMsgTrans}
		err = c.conf.MessageRegistry.ProcessPacket(localPkt)
		if err != nil {
			log.Error().Err(err).Msg("updatePasswordCracker ProcessPacket")
		}
	}
	go updatePasswordCracker()
}
