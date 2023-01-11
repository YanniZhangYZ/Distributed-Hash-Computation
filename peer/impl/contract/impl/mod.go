package impl

import (
	// "fmt"
	"encoding/json"
	"strings"

	"go.dedis.ch/cs438/peer/impl/contract"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
	"golang.org/x/xerrors"
)

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

	codeAST, err := parser.Parse(plainCode)
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

// Unmarshal unmarshals the data into the current Contract instance.
func Unmarshal(data []byte, contract *Contract) error {
	return json.Unmarshal(data, contract)
}

func (c *Contract) GetPublisherAccount() string {
	return c.publisher
}

func (c *Contract) GetFinisherAccount() string {
	return c.finisher
}

func (c *Contract) GetCodeAST() parser.Code {
	return c.codeAST
}

func (c *Contract) GetStateAST() *StateNode {
	return c.stateTree
}

// // Check the condition from the world state of the underlying blockchain
// // func (c *Contract) CheckConditionObjObj(condition parser.Condition, worldState storage.KV) (bool, error) {
// func (c *Contract) CheckConditionObjObj(condition parser.ConditionObjObj) (bool, error) {
// 	role1 := condition.Object1.Role
// 	fields1 := condition.Object1.Fields
// 	operator := condition.Operator
// 	role2 := condition.Object2.Role
// 	fields2 := condition.Object2.Fields

// 	// check role and account condition
// 	account1State, err := c.checkRoleAccount(role1)
// 	if err != nil {
// 		return false, err
// 	}
// 	account2State, err := c.checkRoleAccount(role2)
// 	if err != nil {
// 		return false, err
// 	}

// 	temp := fmt.Sprint(account2State) + " " + fmt.Sprint(account1State)
// 	fmt.Println("use variable to pass golint " + temp)

// 	left_val, err := c.checkObj(fields1, 2)
// 	if err != nil {
// 		return false, err
// 	}
// 	right_val, err := c.checkObj(fields2, 1)
// 	if err != nil {
// 		return false, err
// 	}

// 	// type assertion
// 	if reflect.TypeOf(left_val).String() == "string" {
// 		var left_string = left_val.(string)
// 		var right_string = right_val.(string)
// 		return CompareString(left_string, right_string, operator)
// 	} else {
// 		return false, xerrors.Errorf("type not supported: %v", reflect.TypeOf(left_val))
// 	}
// }

// // func (c *Contract) checkRoleAccount(role string, worldState storage.KV) (*account.State, error) {
// func (c *Contract) checkRoleAccount(role string) (int, error) {
// 	var account string
// 	if role == "publisher" {
// 		account = c.publisher
// 	} else if role == "finisher" {
// 		account = c.finisher
// 	}

// 	// retrieve value corresponding to role.fields from the world state
// 	// value, err := worldState.Get(account)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// account_state, ok := value.(*account.State)
// 	// if !ok {
// 	// 	return false, xerrors.Errorf("account state is corrupted: %v", account_state)
// 	// }
// 	fmt.Println("just to use account to pass golint " + account)
// 	account_state := 0
// 	return account_state, nil

// }

// // func (c *Contract) checkObj(fields []*parser.Field, accountState *account.State, fieldLength int) (bool, error) {
// func (c *Contract) checkObj(fields []*parser.Field, fieldLength int) (interface{}, error) {

// 	var value interface{}
// 	if len(fields) != fieldLength {
// 		return false, xerrors.Errorf("Condition field unknown, only support specific number of attribute.")
// 	}
// 	attribute1 := fields[0].Name
// 	attribute2 := ""

// 	fmt.Println("just to use account to pass golint " + attribute1 + " " + attribute2)

// 	if fieldLength == 2 {
// 		attribute2 = fields[1].Name
// 	}

// 	if fieldLength == 1 { // finisher.hash0
// 		value = string("account_state.hash0")
// 	} else { //finisher.key0.hash
// 		valueTemp := string("account_state.StorageRoot.Get(attribute1)") // get key0
// 		// valueTemp, err = account_state.StorageRoot.Get(attribute)
// 		// if reflect.TypeOf(value).String() == "string" {
// 		// 	value = float64(value.(int))
// 		// }
// 		// if err != nil {
// 		// 	return false, xerrors.Errorf("key not exist in storage: %v", attribute)
// 		// }

// 		// compute the hash of valueTemp
// 		fmt.Println("just to use account to pass golint " + valueTemp)

// 	}

// 	return value, nil

// }

func CompareString(left_val string, right_val string, operator string) (bool, error) {
	switch operator {
	case "==":
		return (left_val == right_val), nil
	case "!=":
		return (left_val != right_val), nil
	}
	return false, xerrors.Errorf("comparator not supported on string: %v", operator)
}

func CompareNumber(left_val float64, right_val float64, operator string) (bool, error) {
	switch operator {
	case ">":
		return (left_val > right_val), nil
	case "<":
		return (left_val < right_val), nil
	case ">=":
		return (left_val >= right_val), nil
	case "<=":
		return (left_val <= right_val), nil
	case "==":
		return (left_val == right_val), nil
	case "!=":
		return (left_val != right_val), nil
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
