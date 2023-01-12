package project

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/contract/impl"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

// test the functionality of Marshaling and Unmarshaling the contract
func Test_Contract_Marshal(t *testing.T) {

	plainContract :=
		`
		ASSUME publisher.budget > 0
		IF finisher.key98.hash == finisher.hash98 THEN
			finisher.submit("publisher_ID", "crackedKey")
			publisher.transfer("finisher_ID", 46.967)
	`

	// create a contract instance
	contract := impl.NewContract(
		"1",                  // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"f1",                 // finisher
	)

	buf, err := contract.Marshal()
	require.NoError(t, err)

	var unmarshalContract impl.Contract
	err = impl.Unmarshal(buf, &unmarshalContract)
	require.NoError(t, err)

}

// Test state tree
// Here we print the AST and State AST
func Test_Contract_State_Tree(t *testing.T) {

	plainContract :=
		`
		ASSUME publisher.budget > 0
		IF finisher.key98.hash == finisher.hash98 THEN
			finisher.submit("publisher_ID", "crackedKey")
			publisher.transfer("finisher_ID", 46.967)
	`

	// create a contract instance
	contract := impl.NewContract(
		"1",                  // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"f1",                 // finisher
	)

	codeAST, err := parser.BuildCodeAST(plainContract)
	stateAST := impl.BuildStateTree(&codeAST)
	require.NoError(t, err)

	fmt.Println(contract.ToString())
	fmt.Println(impl.GetStateAST(codeAST, stateAST))
}

func Test_Contract_Check_Assumption(t *testing.T) {

	plainContract :=
		`
		ASSUME publisher.budget > 0
		IF finisher.key98.hash == "abcdgak13eJ46" THEN
			finisher.submit("publisher_ID", "crackedKey")
			publisher.transfer("finisher_ID", 46.967)
	`

	// create a contract instance
	contract := impl.NewContract(
		"1",                  // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"f1",                 // finisher
	)

	worldState := common.QuickWorldState(1, 20)
	add := &worldState

	isValid, err := contract.CheckAssumptions(add)
	require.NoError(t, err)

	require.Equal(t, isValid, true)

}
