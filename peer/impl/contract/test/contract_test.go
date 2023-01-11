package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer/impl/contract/impl"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

// Marshal and Unmarshal the contract instance
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

// Test state tree by printing out AST and State AST
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

	code_ast, err := parser.Parse(plainContract)
	state_ast := impl.BuildStateTree(&code_ast)
	require.NoError(t, err)

	fmt.Println(contract.ToString())
	// fmt.Println(impl.DisplayAST(code_ast))
	fmt.Println(impl.GetStateAST(code_ast, state_ast))
}
