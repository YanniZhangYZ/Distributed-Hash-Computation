package consensus

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"log"
)

func (c *ConsensusModule) execPaxosPrepareMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	paxosPrepareMsg, ok := msg.(*types.PaxosPrepareMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Ignore messages from the previous and future TLC step, and ignore messages that with ID
	less than or equal to the max ID that we already observed */
	c.Lock()
	defer c.Unlock()
	if paxosPrepareMsg.Step != c.tlcStep || paxosPrepareMsg.ID <= c.paxos.maxID {
		if paxosPrepareMsg.Step != c.tlcStep {
			log.Printf("dropping Prepare because Step does not match %d vs %d: %s addr=%s",
				paxosPrepareMsg.Step, c.tlcStep, paxosPrepareMsg, c.address)
		} else {
			log.Printf("dropping Prepare because ID is too low %d <= %d: %s addr=%s",
				paxosPrepareMsg.ID, c.paxos.maxID, paxosPrepareMsg, c.address)
		}
		return nil
	}
	c.paxos.maxID = paxosPrepareMsg.ID
	paxosPromiseMsg := types.PaxosPromiseMessage{
		Step:          paxosPrepareMsg.Step,
		ID:            paxosPrepareMsg.ID,
		AcceptedID:    c.paxos.AcceptedID,
		AcceptedValue: c.paxos.AcceptedValue,
	}

	paxosPromiseMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(paxosPromiseMsg)
	if err != nil {
		return err
	}
	recipients := make(map[string]struct{})
	recipients[paxosPrepareMsg.Source] = struct{}{}
	privateMsg := types.PrivateMessage{
		Recipients: recipients,
		Msg:        &paxosPromiseMsgTrans,
	}
	privateMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(privateMsg)
	if err != nil {
		return err
	}
	return c.message.Broadcast(privateMsgTrans)
}

func (c *ConsensusModule) execPaxosPromiseMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	paxosPromiseMsg, ok := msg.(*types.PaxosPromiseMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Ignore messages from the previous and future TLC step, and ignore messages if we are
	in stage 2, or we are in stage 1, but the promise is from the previous round */
	c.Lock()
	defer c.Unlock()

	if paxosPromiseMsg.Step != c.tlcStep {
		log.Printf("dropping Promise because Step does not match %d vs %d: %s addr=%s",
			paxosPromiseMsg.Step, c.tlcStep, paxosPromiseMsg, c.address)
		return nil
	}

	if c.paxos.phase == 2 {
		log.Printf("dropping Promise because we are proposer but in Phase %d: %s addr=%s",
			c.paxos.phase, paxosPromiseMsg, c.address)
		return nil
	}

	fromPreviousPropose := c.paxos.phase == 1 && paxosPromiseMsg.ID < c.paxos.proposeID &&
		(c.paxos.proposeID-paxosPromiseMsg.ID)%c.totalPeers == 0
	if fromPreviousPropose {
		log.Printf("dropping Promise because we are proposer but the promise is from the "+
			"previous round with ID %d < %d: %s addr=%s",
			paxosPromiseMsg.ID, c.paxos.proposeID, paxosPromiseMsg, c.address)
		return nil
	}

	/* Set our accepted ID or value to the values contained in the promise message, with the highest acceptID */
	if paxosPromiseMsg.AcceptedID > c.paxos.AcceptedID {
		c.paxos.AcceptedID = paxosPromiseMsg.AcceptedID
		c.paxos.AcceptedValue = paxosPromiseMsg.AcceptedValue
	}

	c.paxos.promiseCnt++
	if c.paxos.promiseCnt >= c.threshold {
		/* Reset this promise stage, avoid double triggering */
		c.paxos.promiseCnt = 0
		if c.paxos.collectEnoughPromise != nil {
			/* Inform that enough promises have been collected */
			*c.paxos.collectEnoughPromise <- true
			c.paxos.collectEnoughPromise = nil
		}
	}
	return nil
}

func (c *ConsensusModule) execPaxosProposeMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	paxosProposeMsg, ok := msg.(*types.PaxosProposeMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Ignore messages from the previous and future TLC step, and ignore messages that with ID
	not equal to the max ID that we already observed */
	c.Lock()
	defer c.Unlock()
	if paxosProposeMsg.Step != c.tlcStep || paxosProposeMsg.ID != c.paxos.maxID {
		if paxosProposeMsg.Step != c.tlcStep {
			log.Printf("dropping Propose because Step does not match %d vs %d: %s addr=%s",
				paxosProposeMsg.Step, c.tlcStep, paxosProposeMsg, c.address)
		} else {
			log.Printf("dropping Propose because ID does not match %d vs %d: %s addr=%s",
				paxosProposeMsg.ID, c.paxos.maxID, paxosProposeMsg, c.address)
		}
		return nil
	}
	c.paxos.AcceptedID = paxosProposeMsg.ID
	c.paxos.AcceptedValue = &paxosProposeMsg.Value

	paxosAcceptMsg := types.PaxosAcceptMessage{
		Step:  paxosProposeMsg.Step,
		ID:    paxosProposeMsg.ID,
		Value: paxosProposeMsg.Value,
	}
	paxosAcceptMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(paxosAcceptMsg)
	if err != nil {
		return err
	}
	return c.message.Broadcast(paxosAcceptMsgTrans)
}

func (c *ConsensusModule) execPaxosAcceptMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	paxosAcceptMsg, ok := msg.(*types.PaxosAcceptMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Ignore messages from the previous and future TLC step, and ignore messages if we are proposer,
	but the accept message is from the previous round */
	c.Lock()
	defer c.Unlock()
	if paxosAcceptMsg.Step != c.tlcStep ||
		(c.paxos.phase > 0 &&
			paxosAcceptMsg.ID < c.paxos.proposeID &&
			(c.paxos.proposeID-paxosAcceptMsg.ID)%c.totalPeers == 0) {
		if paxosAcceptMsg.Step != c.tlcStep {
			log.Printf("dropping Accept because Step does not match %d vs %d: %s addr=%s",
				paxosAcceptMsg.Step, c.tlcStep, paxosAcceptMsg, c.address)
		} else {
			log.Printf("dropping Accept because we are proposer but the propose ID is too low %d <= %d: %s addr=%s",
				paxosAcceptMsg.ID, c.paxos.proposeID, paxosAcceptMsg, c.address)
		}
		return nil
	}

	c.paxos.acceptCnt[paxosAcceptMsg.Value.UniqID]++
	if c.paxos.acceptCnt[paxosAcceptMsg.Value.UniqID] >= c.threshold {
		/* Reset this accept stage, avoid double triggering */
		c.paxos.acceptCnt[paxosAcceptMsg.Value.UniqID] = 0

		/* Update our acceptID and accept*/
		c.paxos.AcceptedID = paxosAcceptMsg.ID
		c.paxos.AcceptedValue = &paxosAcceptMsg.Value

		if c.paxos.collectEnoughAccept != nil {
			/* We are the proposer, notify the phase 2 is done */
			*c.paxos.collectEnoughAccept <- &paxosAcceptMsg.Value
			c.paxos.collectEnoughAccept = nil
		}

		/* Reach a consensus */
		tlcMsg := c.buildTLCMsg()
		c.paxos.alreadyBroadcast = true

		tlcMsgTrans, err := c.conf.MessageRegistry.MarshalMessage(tlcMsg)
		if err != nil {
			return err
		}
		return c.message.Broadcast(tlcMsgTrans)
	}
	return nil
}

func (c *ConsensusModule) execTLCMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	tlcMsg, ok := msg.(*types.TLCMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	/* Ignore messages from the previous TLC step */
	c.Lock()
	defer c.Unlock()
	if tlcMsg.Step < c.tlcStep {
		log.Printf("dropping TLC Message because Step is too low %d <= %d: %s addr=%s",
			tlcMsg.Step, c.tlcStep, tlcMsg, c.address)

		return nil
	}

	c.tlcCnt[tlcMsg.Step]++
	c.tlcValue[tlcMsg.Step] = &tlcMsg.Block
	if tlcMsg.Step == c.tlcStep && c.tlcCnt[tlcMsg.Step] >= c.threshold {
		c.tlcCnt[tlcMsg.Step] = 0
		return c.advanceTLC(false)
	}
	return nil
}
