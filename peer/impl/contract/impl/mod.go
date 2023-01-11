package impl

import (
	// "fmt"
	"encoding/json"

	"fmt"
	"reflect"
	"strings"

	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/contract"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
	"golang.org/x/xerrors"
)

const finisherText = "finisher"
const publisherText = "publisher"

// implements contract.ContractCode, maintained in contract account
type Contract struct {
	contract.SmartContract
	contractID   string
	contractName string
	codeAST      parser.Code
	codePlain    string
	stateTree    *StateNode
	publisher    string
	finisher     string
}

// Create & initialize a new Code instance
func NewContract(contractID string,
	contractName string,
	plainCode string,
	publisher string,
	finisher string) contract.SmartContract {

	codeAST, err := parser.BuildCodeAST(plainCode)
	if err != nil {
		panic(err)
	}
	stateAST := BuildStateTree(&codeAST)

	return &Contract{
		contractID:   contractID,
		contractName: contractName,
		codeAST:      codeAST,
		codePlain:    plainCode,
		stateTree:    stateAST,
		publisher:    publisher,
		finisher:     finisher,
	}
}

// This function marshals the Contract instance into a byte representation.
// we need to use marshal and unmarshal to enable contract instance transition in packet
func (c *Contract) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

// Unmarshal unmarshals the data into the Contract instance.
func Unmarshal(data []byte, contract *Contract) error {
	return json.Unmarshal(data, contract)
}

// get the publisher of this contract
func (c *Contract) GetPublisherAccount() string {
	return c.publisher
}

// get the finisher of this contract
func (c *Contract) GetFinisherAccount() string {
	return c.finisher
}

// get the code AST
func (c *Contract) GetCodeAST() parser.Code {
	return c.codeAST
}

// get the state tree
func (c *Contract) GetStateAST() *StateNode {
	return c.stateTree
}

// Here we check the condition of comparing an obj and another obj
// finisher.key0.hash == finisher.hash0
func (c *Contract) CheckConditionObjObj(condition parser.ConditionObjObj, worldState *common.WorldState) (bool, error) {
	role1 := condition.Object1.Role
	fields1 := condition.Object1.Fields
	operator := condition.Operator
	role2 := condition.Object2.Role
	fields2 := condition.Object2.Fields

	//------------ check the first account and fields--------------------
	var account1 string
	if role1 == publisherText {
		account1 = c.publisher
	} else if role1 == finisherText {
		account1 = c.finisher
	}
	state1, ok := (*worldState).Get(account1)
	if !ok {
		return false, fmt.Errorf("account doesn't exists or account state is corrupted")
	}

	//for the obj on left, we allow at most two attribute
	// e.g. finisher.key0.hash
	if len(fields1) > 2 {
		return false, xerrors.Errorf("Condition field unknown. Can have at most two attributes")
	}
	attribute11 := fields1[0].Name
	attribute12 := ""
	if len(fields1) == 2 {
		attribute12 = fields1[1].Name
	}

	//------------ check the second account--------------------

	var account2 string
	if role2 == publisherText {
		account2 = c.publisher
	} else if role2 == finisherText {
		account2 = c.finisher
	}
	state2, ok := (*worldState).Get(account2)
	if !ok {
		return false, fmt.Errorf("account doesn't exists or account state is corrupted")
	}

	//for the obj on right we allow at most one attribute
	// e.g. finisher.hash0
	if len(fields2) > 1 {
		return false, xerrors.Errorf("Condition field unknown. Can have at most one attribute")
	}
	attribute21 := fields2[0].Name

	var leftVal interface{}
	var rightVal interface{}
	// get left value
	if len(fields1) == 2 && attribute12 == "hash" {
		leftValPlain := "leftVal, err := state1.StorageRoot.Get(attribute11)" + attribute11 + state1.CodeHash
		leftVal = leftValPlain + "hash"
	} else if len(fields1) == 1 {
		leftVal = "leftVal, err := state1.StorageRoot.Get(attribute11)" + attribute11 + state1.CodeHash
	} else {
		return false, xerrors.Errorf("invalid condition obj obj grammar.")
	}

	//get right value
	if len(fields2) == 1 {
		rightVal = "rightVal, err := state2.StorageRoot.Get(attribute21)" + attribute21 + state2.CodeHash
	} else {
		return false, xerrors.Errorf("invalid condition obj obj grammar.")
	}

	// Here we check whether the left and righ data have the same type
	if !c.CheckLeftRightType(leftVal, rightVal) {
		return false, xerrors.Errorf("left and right value type are not consistent.")

	}
	// from now on, we can know that the left and right value have the same data type
	return c.CompareLeftRightVal(leftVal, rightVal, operator)

}

func (c *Contract) CheckRoleAccount(role string, worldState *common.WorldState) (common.State, error) {
	var account string
	if role == publisherText {
		account = c.publisher
	} else if role == finisherText {
		account = c.finisher
	}

	// retrieve value corresponding to role.fields from the world state
	state, ok := (*worldState).Get(account)
	if !ok {
		return state, fmt.Errorf("account doesn't exists")
	}

	return state, nil

}

// This check is used in Assumption
// for comparison between left is a variable and right is a value
// e.g. publisher.budget > 0
func (c *Contract) CheckCondition(condition parser.Condition, worldState *common.WorldState) (bool, error) {
	role := condition.Object.Role
	fields := condition.Object.Fields
	operator := condition.Operator
	value := condition.Value

	// evaluate and retrieve the compared value
	var account string
	if role == publisherText {
		account = c.publisher
	} else if role == finisherText {
		account = c.finisher
	}

	// we assume the fields restricted to balance / storage key
	var leftVal interface{}
	if len(fields) > 1 {
		return false, xerrors.Errorf("Condition field unknown. Can only have one attribute")
	}
	attribute := fields[0].Name

	// retrieve value corresponding to role.fields from the world state
	state, ok := (*worldState).Get(account)
	if !ok {
		return false, fmt.Errorf("account doesn't exists or account state is corrupted")
	}

	if attribute == "Balance" {
		leftVal = float64(state.Balance)
	} else {
		return false, xerrors.Errorf("invalid grammar. Expecting [Balance], get: %v", attribute)
		// !!!! NEED TO CHECK with Qiyuan!!!!! how to get attribute

		// leftVal, err := state.StorageRoot.Get(attribute)
		// if reflect.TypeOf(leftVal).String() == "int" {
		// 	leftVal = float64(leftVal.(int))
		// }
		// if err != nil {
		// 	return false, xerrors.Errorf("no such key exist in storage: %v", attribute)
		// }
	}

	var rightVal interface{}
	if value.String == nil {
		rightVal = *value.Number
	} else {
		rightVal = *value.String
	}

	// Here we check whether the left and righ data have the same type
	if !c.CheckLeftRightType(leftVal, rightVal) {
		return false, xerrors.Errorf("left and right value type are not consistent.")

	}
	// from now on, we can know that the left and right value have the same data type
	return c.CompareLeftRightVal(leftVal, rightVal, operator)
}

func (c *Contract) CheckLeftRightType(left interface{}, right interface{}) bool {
	return reflect.TypeOf(left) == reflect.TypeOf(right)
}

func (c *Contract) CompareLeftRightVal(left interface{}, right interface{}, operator string) (bool, error) {
	if reflect.TypeOf(left).String() == "float64" {
		var leftNum = left.(float64)
		var rightNum = right.(float64)
		return CompareNumber(leftNum, rightNum, operator)

	} else if reflect.TypeOf(left).String() == "string" {
		var leftStr = left.(string)
		var rightStr = right.(string)
		return CompareString(leftStr, rightStr, operator)

	}
	return false, xerrors.Errorf("unsupported type: %v", reflect.TypeOf(left))
}

func (c *Contract) CheckAssumptions(worldState *common.WorldState) (bool, error) {
	isValid := true
	for i, assumption := range c.codeAST.Assumptions {
		condition := assumption.Condition
		conditionValid, err := c.CheckCondition(condition, worldState)
		if err != nil {
			return false, err
		}
		if !conditionValid {
			isValid = false
		} else { // synchronize the validity to the state tree
			c.stateTree.children[i].setValid()
			c.stateTree.children[i].children[0].setValid()
		}

		c.stateTree.children[i].setExecuted()
		c.stateTree.children[i].children[0].setExecuted()
	}

	return isValid, nil
}

func (c *Contract) GatherActions(worldState *common.WorldState) ([]parser.Action, error) {
	var actions []parser.Action

	// Here we loop all the if clause in the code AST tree
	for i, ifclause := range c.codeAST.IfClauses {
		// we assume the contion in the if clause
		// is the comparison between object and object, note object and value
		conditionObjObj := ifclause.ConditionObjObj
		conditionValid, err := c.CheckConditionObjObj(conditionObjObj, worldState)
		if err != nil {
			return []parser.Action{}, err
		}
		ifclauseState := c.stateTree.children[i+len(c.codeAST.Assumptions)]
		conditionState := ifclauseState.children[0]

		if !conditionValid {
			ifclauseState.setExecuted()
			conditionState.setExecuted()
		} else {
			ifclauseState.setExecuted()
			ifclauseState.setValid()
			conditionState.setExecuted()
			conditionState.setValid()

			for j := 1; j < len(ifclauseState.children); j++ {
				ifclauseState.children[j].setExecuted()
			}
			for _, action := range ifclause.Actions {
				actions = append(actions, *action)
			}
		}
	}

	return actions, nil
}

func CompareString(leftVal string, rightVal string, operator string) (bool, error) {
	switch operator {
	case "==":
		return (leftVal == rightVal), nil
	case "!=":
		return (leftVal != rightVal), nil
	}
	return false, xerrors.Errorf("comparator not supported on string: %v", operator)
}

func CompareNumber(leftVal float64, rightVal float64, operator string) (bool, error) {
	switch operator {
	case ">":
		return (leftVal > rightVal), nil
	case "<":
		return (leftVal < rightVal), nil
	case ">=":
		return (leftVal >= rightVal), nil
	case "<=":
		return (leftVal <= rightVal), nil
	case "==":
		return (leftVal == rightVal), nil
	case "!=":
		return (leftVal != rightVal), nil
	}
	return false, xerrors.Errorf("comparator not supported on number: %v", operator)
}

// Contract.String() outputs the contract in pretty readable format
func (c Contract) ToString() string {
	out := new(strings.Builder)

	out.WriteString("=================================================================\n")
	out.WriteString("| Contract: " + c.contractName + "\n")
	out.WriteString("| ID: " + c.contractID + "\n")
	out.WriteString("| Publisher: [" + c.publisher + "] \n")
	out.WriteString("| Finisher: [" + c.finisher + "] \n")
	out.WriteString("| Contract code: " + "\n")
	out.WriteString(c.codePlain + "\n")
	out.WriteString("=================================================================\n")

	return out.String()
}
