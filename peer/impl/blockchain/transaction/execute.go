package transaction

import (
	"fmt"
	"strings"

	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/contract/impl"
)

// VerifyAndExecuteTransaction verify and execute a transaction on a given world state
func VerifyAndExecuteTransaction(tx *SignedTransaction, worldState *common.WorldState, print bool) error {
	var err error
	switch tx.TX.Type {
	case TRANSFER_TX:
		err = executeTransferTx(tx, worldState)
	case CONTRACT_DEPLOYMENT_TX:
		err = executeContractDeploymentTx(tx, worldState)
	case CONTRACT_EXECUTION_TX:
		err = executeContractExecutionTx(tx, worldState, print)
	default:
		panic("Unknown transaction type")
	}

	return err
}

func executeTransferTx(tx *SignedTransaction, worldState *common.WorldState) error {
	// Check if the transaction is an account join declaration transaction (i.e. Src == Dst && Value >= 0)
	if tx.TX.Src.String() == tx.TX.Dst.String() && tx.TX.Value >= 0 {

		_, ok := (*worldState).Get(tx.TX.Src.String())
		if ok {
			return fmt.Errorf("invalid account join declaration transaction, account already exists")
		}

		(*worldState).Set(tx.TX.Src.String(), common.State{
			Nonce:   0,
			Balance: tx.TX.Value,
		})

		return nil
	}

	// Check if the transaction is an account leave declaration transaction (i.e. Src == Dst && Value == -1)
	if tx.TX.Src.String() == tx.TX.Dst.String() && tx.TX.Value == -1 {

		_, ok := (*worldState).Get(tx.TX.Src.String())
		if !ok {
			return fmt.Errorf("invalid account leave declaration transaction, account doesn't exists")
		}

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

func executeContractDeploymentTx(tx *SignedTransaction, worldState *common.WorldState) error {
	if len(tx.TX.Contract) == 0 {
		return fmt.Errorf("invalid contract deployment transaction")
	}

	// Check if the publisher has enough balance to pay the deposit
	srcState, ok := (*worldState).Get(tx.TX.Src.String())
	if !ok {
		return fmt.Errorf("TX src not found in the world state")

	}
	if srcState.Balance < tx.TX.Value {
		return fmt.Errorf("insufficient balance, publisher has %d but tries to publish a contract with %d reward", srcState.Balance, tx.TX.Value)
	}

	// Create the contract instance
	var contract impl.Contract
	err := impl.Unmarshal(tx.TX.Contract, &contract)
	if err != nil {
		return err
	}

	// Create a smart contract account in the world state
	worldState.Set(tx.TX.Dst.String(), common.State{
		Nonce:       0,
		Balance:     tx.TX.Value,
		Contract:    tx.TX.Contract,
		StorageRoot: "",
		Tasks:       make(map[string][2]string),
	})

	// Publisher pay the deposit
	srcState.Balance -= tx.TX.Value
	(*worldState).Set(tx.TX.Src.String(), srcState)

	return nil
}

func executeContractExecutionTx(tx *SignedTransaction, worldState *common.WorldState, print bool) error {
	contractAddr := tx.TX.Dst.String()
	contractState, ok := worldState.Get(contractAddr)
	if !ok {
		return fmt.Errorf("contract account address invalid")
	}

	// Update the finisher's state
	finisherState, ok := worldState.Get(tx.TX.Src.String())
	if !ok {
		return fmt.Errorf("invalid finisher address")
	}

	// Parse the data to get password, hash, salt
	data := strings.Split(tx.TX.Data, ",")
	if len(data) != 3 {
		return fmt.Errorf("invalid data of a ContractExecutionTx, data : %s", tx.TX.Data)
	}
	finisherState.Tasks[data[1]] = [2]string{data[0], data[2]}
	worldState.Set(tx.TX.Src.String(), finisherState)

	// Retrieve the smart contract instance from the world state
	var contract impl.Contract
	err := impl.Unmarshal(contractState.Contract, &contract)
	if err != nil {
		return err
	}

	// Check if the contract has enough balance to pay the execution
	assumptionValid, err2 := contract.CheckAssumptions(worldState)
	if err2 != nil {
		return err2
	}
	if !assumptionValid {
		return fmt.Errorf("smart contract assumption checking failed")
	}

	// Execute the contract
	ifThenValid, actions, err3 := contract.GatherActions(worldState)
	if print {
		contract.PrintContractExecutionState()
	}

	// Update the contract instance after executing the contract
	contractState.Contract, _ = contract.Marshal()

	if err3 != nil {
		return err3
	}
	if !ifThenValid {
		return fmt.Errorf("unable to execute action, if-then condition not satisfied")
	}

	if len(actions) != 1 || actions[0].Action != "transfer" || len(actions[0].Params) != 2 {
		return fmt.Errorf("unable to execute such action, len(actions)=%d", len(actions))
	}

	var receiver string
	var reward int64
	if actions[0].Params[0].String != nil {
		receiver = *actions[0].Params[0].String
		reward = int64(*actions[0].Params[1].Number)
	} else {
		receiver = *actions[0].Params[1].String
		reward = int64(*actions[0].Params[0].Number)
	}

	// Transfer the money
	receiverState, ok2 := worldState.Get(receiver)
	if !ok2 {
		return fmt.Errorf("reward receiver account not found in the world state")
	}

	receiverState.Balance += reward
	contractState.Balance -= reward

	worldState.Set(receiver, receiverState)
	worldState.Set(contractAddr, contractState)

	return nil
}
