package contract

import (
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

// SmartContract describes functions to manipulate the code section in smart contract
type SmartContract interface {

	//This function the outputs the info of contract including plain contract code
	ToString() string

	// This function marshals the Contract instance into a byte representation.
	// we need to use marshal and unmarshal to enable contract instance transition in packet
	// It should be noted that for those data structures we put in Marshal
	// It is necessary to name the inner variables with names that is Capitalized
	// Or else the json.Marshal cannot resolve the corresponding value
	Marshal() ([]byte, error)

	// This function checks the validity of Assumptions made in the contract
	// In this project we specify that the left part of condition we used in Assumption
	// should be a variable and can have only one attribute.
	// The right part of the condition should be a string or a float
	// e.g. ASSUME smartAccount.balance > 0
	CheckAssumptions(*common.WorldState) (bool, error)

	// This function deals with the if-then clause.
	// In this project we specify that the left part of condition we used in if-then
	// should be a variable and have two attributes.
	// The right part of the condition should be a string or a float
	// e.g. IF finisher.crackedPwd.hash == "someHash" THEN
	GatherActions(*common.WorldState) (bool, []parser.Action, error)

	PrintContractExecutionState()

	// This function gets the publisher of this contract
	GetPublisherAccount() string

	// This function gets the finisher of this contract
	GetFinisherAccount() string
}
