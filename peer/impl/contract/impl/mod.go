package impl

import (
	"crypto/sha256"
	"encoding/hex"
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
const smartAccountText = "smartAccount"

// implements contract.ContractCode, maintained in contract account
// It should be noted that the contract will be marshaled and unmarshaled to transmit in the channel
// Therefore it is necessary to name the variables with names that is Capitalized
// Or else the json.Marshal and json.Unmarshal cannot resolve the corresponding value
type Contract struct {
	contract.SmartContract
	ContractID   string      // The address of the smart contract account in the blockchain network
	ContractName string      // Contract Name
	CodeAST      parser.Code // The code AST
	CodePlain    string      // Thhe text version of user input contract code
	StateTree    *StateNode  // The state AST to keep track of the execution of contract
	Publisher    string      // Address of the task publisher
	Finisher     string      // Address of the task finisher
}

// Initialized a new contract
// The codeAST and stateAST are also initialized given the plain contract code
func NewContract(ContractID string,
	ContractName string,
	CodePlain string,
	Publisher string,
	Finisher string) contract.SmartContract {

	CodeAST, err := parser.BuildCodeAST(CodePlain)
	if err != nil {
		panic(err)
	}
	stateAST := BuildStateTree(&CodeAST)

	return &Contract{
		ContractID:   ContractID,
		ContractName: ContractName,
		CodeAST:      CodeAST,
		CodePlain:    CodePlain,
		StateTree:    stateAST,
		Publisher:    Publisher,
		Finisher:     Finisher,
	}
}

// This function will print the text version of contract code itself
func (c *Contract) PrintPlainContract() {
	fmt.Println(c.CodePlain)
}

// This function helps to facilitate constructing the  text version contract code
// The template is provided
func BuildPlainContract(targetHash string, finisherAddr string, reward int64) string {
	plain := fmt.Sprintf(`
	ASSUME smartAccount.balance > 0
	IF finisher.crackedPwd.hash == "%s" THEN
		smartAccount.transfer("%s", %d)
	`, targetHash, finisherAddr, reward)

	return plain
}

func (c *Contract) PrintContractExecutionState() {
	fmt.Println("📝📝📝 Contract Info.")
	fmt.Println(c.ToStringBasic())
	fmt.Println(PrintStateAST(c.GetCodeAST(), c.GetStateAST()))
}

// This function marshals the Contract instance into a byte representation.
// we need to use marshal and unmarshal to enable contract instance transition in packet
// It should be noted that for those data structures we put in Marshal
// It is necessary to name the inner variables with names that is Capitalized
// Or else the json.Marshal cannot resolve the corresponding value
func (c *Contract) Marshal() ([]byte, error) {
	tmp, err := json.Marshal(c)
	return tmp, err
}

// Unmarshal unmarshals the data into the Contract instance.
// It should be noted that for those data structures we put in Marshal before
// It is necessary to name the inner variables with names that is Capitalized
// Or else the json.Unmarshal cannot resolve the corresponding value
func Unmarshal(data []byte, contract *Contract) error {
	return json.Unmarshal(data, contract)
}

// This function gets the publisher of this contract
func (c *Contract) GetPublisherAccount() string {
	return c.Publisher
}

// This function gets the finisher of this contract
func (c *Contract) GetFinisherAccount() string {
	return c.Finisher
}

// This function gets the code AST
func (c *Contract) GetCodeAST() parser.Code {
	return c.CodeAST
}

// This function gets the state tree
func (c *Contract) GetStateAST() *StateNode {
	return c.StateTree
}

// This function checks the validity of Assumptions made in the contract
// In this project we specify that the left part of condition we used in Assumption
// should be a variable and can have only one attribute.
// The right part of the condition should be a string or a float
// e.g. ASSUME smartAccount.balance > 0
func (c *Contract) CheckAssumptions(worldState *common.WorldState) (bool, error) {
	isValid := true

	// We support multiple assumptions before the if-then clause
	for i, assumption := range c.CodeAST.Assumptions {
		condition := assumption.Condition
		conditionValid, err := c.CheckConditionOneAttribute(condition, worldState)
		if err != nil {
			return false, err
		}
		// update the execution state to the contract state tree.
		c.StateTree.Children[i].setExecuted()
		c.StateTree.Children[i].Children[0].setExecuted()

		if !conditionValid {
			isValid = false
		} else {
			// update the validity of condition to the contract state tree.
			c.StateTree.Children[i].setValid()
			c.StateTree.Children[i].Children[0].setValid()
		}

	}

	// We only show the contract info
	// when the assumption is executed but its condition is not valid
	// if isValid == false {
	// 	fmt.Println("📝📝📝 Contract Info.")
	// 	fmt.Println(c.ToStringBasic())
	// 	fmt.Println(PrintStateAST(c.GetCodeAST(), c.GetStateAST()))
	// }

	return isValid, nil
}

// This function deals with the if-then clause.
// In this project we specify that the left part of condition we used in if-then
// should be a variable and have two attributes.
// The right part of the condition should be a string or a float
// e.g. IF finisher.crackedPwd.hash == "someHash" THEN
func (c *Contract) GatherActions(worldState *common.WorldState) (bool, []parser.Action, error) {
	var actions []parser.Action
	// Here we loop all the if clause in the code AST tree
	// we assume the contion in the if clause
	// is the comparison between object and value
	condition := c.CodeAST.IfClauses[0].Condition
	conditionValid, err := c.CheckConditionTwoAttribute(condition, worldState)
	if err != nil {
		return false, []parser.Action{}, err
	}
	ifclauseState := c.StateTree.Children[len(c.CodeAST.Assumptions)]
	conditionState := ifclauseState.Children[0]
	// update the execution state to the contract state tree.
	conditionState.setExecuted()
	ifclauseState.setExecuted()

	if conditionValid {
		// update the validity to the contract state tree.
		// ifclauseState.setExecuted()
		ifclauseState.setValid()
		// conditionState.setExecuted()
		conditionState.setValid()

		for j := 1; j < len(ifclauseState.Children); j++ {
			// update the action inside the if-then as executed
			ifclauseState.Children[j].setExecuted()
		}
		for _, action := range c.CodeAST.IfClauses[0].Actions {
			actions = append(actions, *action)
		}
	}

	// print out the contract info and execution state
	// fmt.Println("📝📝📝 Contract Info.")
	// fmt.Println(c.ToStringBasic())
	// fmt.Println(PrintStateAST(c.GetCodeAST(), c.GetStateAST()))

	return conditionValid, actions, nil
}

// This function compares two strings given the operator
func CompareTwoString(leftVal string, rightVal string, operator string) (bool, error) {
	// fmt.Println(leftVal)
	// fmt.Println(rightVal)
	switch operator {
	case "==":
		return (leftVal == rightVal), nil
	case "!=":
		return (leftVal != rightVal), nil
	}
	return false, xerrors.Errorf("operator not supported on string: %v", operator)
}

// This function compares two values given the operator
func CompareTwoNumber(leftVal int64, rightVal int64, operator string) (bool, error) {
	switch operator {
	case ">":
		return (leftVal > rightVal), nil
	case ">=":
		return (leftVal >= rightVal), nil
	case "<":
		return (leftVal < rightVal), nil
	case "<=":
		return (leftVal <= rightVal), nil
	case "==":
		return (leftVal == rightVal), nil
	case "!=":
		return (leftVal != rightVal), nil
	}
	return false, xerrors.Errorf("operator not supported on number: %v", operator)
}

// Contract.String() outputs the info of contract including plain contract code
func (c Contract) ToString() string {
	out := new(strings.Builder)

	out.WriteString("================================================\n")
	out.WriteString("| Contract: " + c.ContractName + "\n")
	out.WriteString("| ID: " + c.ContractID + "\n")
	out.WriteString("| Publisher: [" + c.Publisher + "] \n")
	out.WriteString("| Finisher: [" + c.Finisher + "] \n")
	out.WriteString("| Contract code: " + "\n")
	out.WriteString(c.CodePlain + "\n")
	out.WriteString("================================================\n")

	return out.String()
}

// outputs the info of contract without plain contract code
// this function provide complement information of execution state
func (c Contract) ToStringBasic() string {
	out := new(strings.Builder)

	out.WriteString("================================================\n")
	out.WriteString("| Contract: " + c.ContractName + "\n")
	out.WriteString("| ID: " + c.ContractID + "\n")
	out.WriteString("| Publisher: [" + c.Publisher + "] \n")
	out.WriteString("| Finisher: [" + c.Finisher + "] \n")
	out.WriteString("================================================\n")

	return out.String()
}

// Here we check the condition who is doing comparison between an obj and another obj
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
		account1 = c.Publisher
	} else if role1 == finisherText {
		account1 = c.Finisher
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
		account2 = c.Publisher
	} else if role2 == finisherText {
		account2 = c.Finisher
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

// This check is used in Assumption
// for comparison between left is a variable and right is a value
// here only smartAccount is involved
// and balance is expected
// e.g. smartAccount.balance > 0
func (c *Contract) CheckConditionOneAttribute(condition parser.Condition, worldState *common.WorldState) (bool, error) {
	role := condition.Object.Role
	fields := condition.Object.Fields
	operator := condition.Operator
	value := condition.Value

	// evaluate and retrieve the compared value
	var account string
	if role == smartAccountText {
		// here we get the account from contract.ContractID
		// This is the address of samrtAccount in the blockchain
		account = c.ContractID
	} else {
		return false, xerrors.Errorf("invalid grammar. Expecting [smartAccount], get: %v", role)
	}

	// we assume the fields restricted to balance / storage key
	var leftVal interface{}
	if len(fields) != 1 {
		return false, xerrors.Errorf("Condition field unknown. Can only have one attribute")
	}
	attribute := fields[0].Name

	// retrieve value corresponding to role.fields from the world state
	state, ok := (*worldState).Get(account)
	if !ok {
		return false, fmt.Errorf("account doesn't exists or account state is corrupted")
	}

	// here we expect smartAccount.balance
	if attribute == "balance" {
		leftVal = int64(state.Balance)
	} else {
		return false, xerrors.Errorf("invalid grammar. Expecting [balance], get: %v", attribute)
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
	// and we compare the left and right value
	return c.CompareLeftRightVal(leftVal, rightVal, operator)
}

// This check is used in if-then clause
// for comparison between left is a variable and right is a value
// here only finisher is involved
// and we expect the grammar to be finisher.crackedPwd.hash == some hash
// e.g. finisher.crackedPwd.hash == "ddnisoqhfqp0unu1h"
func (c *Contract) CheckConditionTwoAttribute(condition parser.Condition, worldState *common.WorldState) (bool, error) {
	role := condition.Object.Role
	fields := condition.Object.Fields
	operator := condition.Operator
	value := condition.Value

	// evaluate and retrieve the compared value
	var account string
	if role == finisherText {
		account = c.Finisher
	} else {
		return false, xerrors.Errorf("invalid grammar. Expecting [finisher], get: %v", role)

	}

	// here we get the hash string
	var rightVal interface{}
	if value.String != nil {
		rightVal = *value.String
	} else {
		return false, xerrors.Errorf("invalid grammar. Expecting a hash string")
	}

	// we expect exactly two fields
	var leftVal interface{}
	if len(fields) != 2 {
		return false, xerrors.Errorf("Condition field unknown. Need to have two attributes")
	}
	attribute1 := fields[0].Name
	attribute2 := fields[1].Name

	// retrieve value corresponding to role.fields from the world state
	state, ok := (*worldState).Get(account)
	if !ok {
		return false, fmt.Errorf("account doesn't exists or account state is corrupted")
	}

	// we expect the grammar like finisher.crackedPwd.hash == some hash
	if attribute1 == "crackedPwd" && attribute2 == "hash" {
		// we compute the hash of cracked key + salt
		// this value is used for the verification
		crackedPwdHash, err := GetTaskHash(state.Tasks, rightVal.(string))
		if err != nil {
			return false, err
		}
		leftVal = crackedPwdHash
	} else {
		errMsg := "invalid grammar. Expecting [crackedPwd.hash], get: " + attribute1 + "." + attribute2
		return false, xerrors.Errorf(errMsg)
	}

	// Here we check whether the left and righ data have the same type
	if !c.CheckLeftRightType(leftVal, rightVal) {
		return false, xerrors.Errorf("left and right value type are not consistent.")

	}
	// from now on, we can know that the left and right value have the same data type
	return c.CompareLeftRightVal(leftVal, rightVal, operator)
}

// This function checks whether the left and right data have the same data type
func (c *Contract) CheckLeftRightType(left interface{}, right interface{}) bool {
	return reflect.TypeOf(left) == reflect.TypeOf(right)
}

// This function compares the value of left and right data
func (c *Contract) CompareLeftRightVal(left interface{}, right interface{}, operator string) (bool, error) {
	if reflect.TypeOf(left).String() == "int64" {
		var leftNum = left.(int64)
		var rightNum = right.(int64)
		return CompareTwoNumber(leftNum, rightNum, operator)

	} else if reflect.TypeOf(left).String() == "string" {
		var leftStr = left.(string)
		var rightStr = right.(string)
		// fmt.Println("hello!!!")
		// fmt.Println(leftStr)
		// fmt.Println(rightStr)

		return CompareTwoString(leftStr, rightStr, operator)

	}
	return false, xerrors.Errorf("unsupported type: %v", reflect.TypeOf(left))
}

// This function firt search for the target hash in State.Tasks
// It then retrive the cracked passward and salt, and recompute the hash
// The return value is then used for password carcking correctness verification
func GetTaskHash(tasks map[string][2]string, targetHash string) (string, error) {
	if len(tasks) == 0 {
		fmt.Printf("Task list is empty. No such hash.")
		return "", xerrors.Errorf("Task list is empty. No such hash.")
	}
	v, ok := tasks[targetHash]
	if !ok {
		return "", xerrors.Errorf("No such hash in the tasks.")
	}
	crackedPwd := v[0]
	salt := v[1]
	// fmt.Println("haha I GOT")
	// fmt.Println(crackedPwd)
	// fmt.Println(salt)
	saltBytes, _ := hex.DecodeString(salt)
	hashStr := HashCrackedPassword(crackedPwd, saltBytes)
	return hashStr, nil
}

// This function hash the given password and salt using sha256
func HashCrackedPassword(password string, salt []byte) string {
	passwordBytes := []byte(password)

	h := sha256.New()
	// Append salt to password
	passwordBytes = append(passwordBytes, salt...)
	h.Write(passwordBytes)
	hashedPasswordBytes := h.Sum(nil)
	hashStr := hex.EncodeToString(hashedPasswordBytes)
	return hashStr
}
