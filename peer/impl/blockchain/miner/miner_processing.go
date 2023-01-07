package miner

import (
	"fmt"
	"go.dedis.ch/cs438/peer/impl/blockchain/transaction"
)

func (m *Miner) executeTransaction(tx *transaction.SignedTransaction) error {
	var err error
	switch tx.TX.Type {
	case transaction.TRANSFER_TX:
		{
			err = m.executeTransferTx(tx)
		}
	case transaction.CONTRACT_DEPLOYMENT_TX:
		{
			//TODO implement me
			panic("implement me")
		}
	case transaction.CONTRACT_EXECUTION_TX:
		{
			//TODO implement me
			panic("implement me")
		}
	default:
		panic("Unknown transaction type")
	}

	if err != nil {
		m.txProcessed.Enqueue(tx)
		return nil
	} else {
		m.txInvalid.Enqueue(tx)
		return err
	}
}

func (m *Miner) executeTransferTx(tx *transaction.SignedTransaction) error {
	// 1. Verify balance
	srcState, ok := m.tmpWorldState.Get(tx.TX.Src.String())
	if !ok {
		return fmt.Errorf("TX src not found in the world state")
	}

	dstState, ok := m.tmpWorldState.Get(tx.TX.Dst.String())
	if !ok {
		return fmt.Errorf("TX src not found in the world state")
	}

	if srcState.Balance < tx.TX.Value {
		return fmt.Errorf("insufficient balance, transaction invalid")
	}

	// 2. Verify nonce
	if tx.TX.Nonce != srcState.Nonce+1 {
		return fmt.Errorf("invalid nonce, transaction invalid")
	}

	srcState.Balance -= tx.TX.Value
	dstState.Balance += tx.TX.Value

	m.tmpWorldState.Set(tx.TX.Src.String(), srcState)
	m.tmpWorldState.Set(tx.TX.Dst.String(), dstState)

	return nil
}
