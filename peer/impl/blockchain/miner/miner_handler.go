package miner

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

func (m *Miner) execTransactionMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	txMsg, ok := msg.(*types.TransactionMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	// Verify the signature upon receiving the signed TransactionMessage

	// TODO: verify the signature

	m.mu.Lock()
	defer m.mu.Unlock()

	m.txPending.Enqueue(&txMsg.SignedTX)

	return nil
}

func (m *Miner) execBlockMessage(msg types.Message, pkt transport.Packet) error {
	/* cast the message to its actual type. You assume it is the right type. */
	blockMsg, ok := msg.(*types.BlockMessage)
	if !ok {
		return xerrors.Errorf("wrong type: %T", msg)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.processBlock(blockMsg)

	return nil
}
