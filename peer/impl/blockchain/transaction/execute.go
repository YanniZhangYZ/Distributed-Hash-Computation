package transaction

import (
	"fmt"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
)

// VerifyAndExecuteTransaction verify and execute a transaction on a given world state
func VerifyAndExecuteTransaction(tx *SignedTransaction, worldState *common.WorldState) error {
	var err error
	switch tx.TX.Type {
	case TRANSFER_TX:
		err = executeTransferTx(tx, worldState)
	case CONTRACT_DEPLOYMENT_TX:
		{
			//TODO implement me
			panic("implement me")
		}
	case CONTRACT_EXECUTION_TX:
		{
			//TODO implement me
			panic("implement me")
		}
	default:
		panic("Unknown transaction type")
	}

	return err

}

func executeTransferTx(tx *SignedTransaction, worldState *common.WorldState) error {
	// Check if the transaction is a new account declaration transaction
	if tx.TX.Src.String() == tx.TX.Dst.String() {

		_, ok := (*worldState).Get(tx.TX.Src.String())
		if ok {
			return fmt.Errorf("invalid account declaration transaction, account already exists")
		}

		(*worldState).Set(tx.TX.Src.String(), common.State{
			Nonce:       0,
			Balance:     tx.TX.Value,
			CodeHash:    "",
			StorageRoot: "",
		})

		return nil
	}

	// Prepare states
	srcState, ok1 := (*worldState).Get(tx.TX.Src.String())
	dstState, ok2 := (*worldState).Get(tx.TX.Dst.String())
	if !ok1 {
		return fmt.Errorf("TX src not found in the world state")
	}
	if !ok2 {
		return fmt.Errorf("TX dst not found in the world state")
	}

	// 1. Verify nonce
	// TODO : do not verify transaction nonce for now
	//if tx.TX.Nonce != srcState.Nonce+1 {
	//	return fmt.Errorf("invalid nonce %d, which should be %d", tx.TX.Nonce, srcState.Nonce+1)
	//}
	//
	//// Nonce verified, update the state
	//srcState.Nonce++
	//m.tmpWorldState.Set(tx.TX.Src.String(), srcState)

	// 2. Check balance
	if srcState.Balance < tx.TX.Value {
		return fmt.Errorf("insufficient balance, src has %d but tries to debit %d", srcState.Balance, tx.TX.Value)
	}

	// Balance checked, update the state
	srcState.Balance -= tx.TX.Value
	dstState.Balance += tx.TX.Value
	(*worldState).Set(tx.TX.Src.String(), srcState)
	(*worldState).Set(tx.TX.Dst.String(), dstState)

	return nil
}
