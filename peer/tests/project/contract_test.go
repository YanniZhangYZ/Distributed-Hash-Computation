package project

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer/impl/blockchain/common"
	"go.dedis.ch/cs438/peer/impl/contract/impl"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
	"golang.org/x/xerrors"
)

// test the functionality of Marshaling and Unmarshaling the contract
func Test_Contract_Marshal(t *testing.T) {

	plainContract :=
		`
		ASSUME publisher.budget > 0
		IF finisher.key98.hash == "vtiubiijk" THEN
			smartAccount.transfer("finisher_ID", 46)
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
	require.Equal(t, contract.ToString(), unmarshalContract.ToString())
	require.NoError(t, err)

}

// Test state tree
// Here we print the AST and State AST
func Test_Contract_State_Tree(t *testing.T) {

	plainContract :=
		`
		ASSUME publisher.balance > 0
		IF finisher.crackedPwd.hash == "yuvubknluykgink" THEN
			smartAccount.transfer("finisher_ID", 46)
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
	fmt.Println(impl.GetCodeAST(codeAST))
}

func Test_Contract_Plain_Text(t *testing.T) {
	plainContract := impl.BuildPlainContract("yuvubknluykgink", "finisherAddr", 3)
	expected := `
	ASSUME publisher.balance > 0
	IF finisher.crackedPwd.hash == "yuvubknluykgink" THEN
		smartAccount.transfer("finisherAddr", 3)
	`
	require.Equal(t, plainContract, expected)

}

func Test_Contract_Check_Assumption(t *testing.T) {

	plainContract :=
		`
		ASSUME smartAccount.balance > 5
		IF finisher.key98.hash == "yuvubknluykgink" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract := impl.NewContract(
		"1",                  // smartAccount Addr
		"crack pwd contract", // name
		plainContract,        // plain_code
		"1cqnopfop",          // publisher
		"f1",                 // finisher
	)

	worldState := common.QuickWorldState(5, 20)

	isValid, err := contract.CheckAssumptions(worldState)
	require.NoError(t, err)
	require.Equal(t, isValid, true)

	//----------------- second--------------------

	plainContract2 :=
		`
		ASSUME smartAccount.balance > 25
		IF finisher.key98.hash == "yuvubknluykgink" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract2 := impl.NewContract(
		"3",                  // ID
		"crack pwd contract", // name
		plainContract2,       // plain_code
		"100ixwscds",         // publisher
		"f1",                 // finisher
	)

	worldState2 := common.QuickWorldState(5, 20)

	isValid2, err := contract2.CheckAssumptions(worldState2)
	require.NoError(t, err)
	require.Equal(t, isValid2, false)

}

func Test_Contract_Check_Assumption_Error(t *testing.T) {

	plainContract :=
		`
		ASSUME finisher.budget > 5
		IF finisher.key98.hash == "yuvubknluykgink" THEN
		smartAccount.transfer("finisher_ID", 46)
	`
	// create a contract instance
	contract := impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"1",                  // publisher
		"f1",                 // finisher
	)

	worldState := common.QuickWorldState(5, 20)

	_, err := contract.CheckAssumptions(worldState)
	expectErr := xerrors.Errorf("invalid grammar. Expecting [smartAccount], get: finisher")
	require.EqualError(t, err, expectErr.Error())

	//----------------- second--------------------

	plainContract2 :=
		`
		ASSUME smartAccount.attribute.attribute > 25
		IF finisher.key98.hash == "yuvubknluykgink" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract2 := impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract2,       // plain_code
		"3",                  // publisher
		"f1",                 // finisher
	)

	worldState2 := common.QuickWorldState(5, 20)

	_, err2 := contract2.CheckAssumptions(worldState2)
	expectErr2 := xerrors.Errorf("Condition field unknown. Can only have one attribute")
	require.EqualError(t, err2, expectErr2.Error())

	//----------------- third--------------------

	plainContract3 :=
		`
		ASSUME smartAccount.balance > 25
		IF finisher.key98.hash == "yuvubknluykgink" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract3 := impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract3,       // plain_code
		"3",                  // publisher
		"f1",                 // finisher
	)

	worldState3 := common.QuickWorldState(1, 20)

	_, err3 := contract3.CheckAssumptions(worldState3)
	expectErr3 := xerrors.Errorf("account doesn't exists or account state is corrupted")
	require.EqualError(t, err3, expectErr3.Error())

	//-----------------  4th -------------------

	plainContract4 :=
		`
			ASSUME smartAccount.budget > 25
			IF finisher.key98.hash == "yuvubknluykgink" THEN
			smartAccount.transfer("finisher_ID", 46)
		`

	// create a contract instance
	contract4 := impl.NewContract(
		"3",                  // ID
		"crack pwd contract", // name
		plainContract4,       // plain_code
		"100ixwscds",         // publisher
		"f1",                 // finisher
	)

	worldState4 := common.QuickWorldState(5, 20)

	_, err4 := contract4.CheckAssumptions(worldState4)
	expectErr4 := xerrors.Errorf("invalid grammar. Expecting [balance], get: budget")
	require.EqualError(t, err4, expectErr4.Error())

	//-----------------  5th -------------------

	plainContract5 :=
		`
			ASSUME smartAccount.balance > "cbuasinfo"
			IF finisher.key98.hash == "yuvubknluykgink" THEN
			smartAccount.transfer("finisher_ID", 46)
		`

	// create a contract instance
	contract5 := impl.NewContract(
		"3",                  // ID
		"crack pwd contract", // name
		plainContract5,       // plain_code
		"100ixwscds",         // publisher
		"f1",                 // finisher
	)

	worldState5 := common.QuickWorldState(5, 20)

	_, err5 := contract5.CheckAssumptions(worldState5)
	expectErr5 := xerrors.Errorf("left and right value type are not consistent.")
	require.EqualError(t, err5, expectErr5.Error())

}

func Test_Contract_Hash_Cracked_pwd(t *testing.T) {
	password1 := "Password"
	salt1 := []byte{0x0, 0x0}
	passwordHash1 := []byte{0x48, 0x4f, 0x95, 0x73, 0x38, 0xd, 0x13, 0xc3, 0x4, 0x2d, 0x36, 0x1, 0xb2, 0x0,
		0x1b, 0x61, 0x1d, 0x2, 0xf4, 0xec, 0xc8, 0x8a, 0xf2, 0x23, 0x5e, 0xc3, 0x18, 0xd, 0xe7, 0xbd, 0x96, 0x2c}
	require.Equal(t, hex.EncodeToString(passwordHash1), impl.HashCrackedPassword(password1, salt1))

	password2 := "apple"
	salt2 := []byte{0x0, 0x3c}
	passwordHash2 := []byte{0x6a, 0xd1, 0x8f, 0x94, 0xf, 0xfb, 0xd3, 0x4, 0x54, 0xe3, 0xc2, 0xec, 0xf6, 0x17,
		0x8c, 0x64, 0x92, 0xde, 0xb3, 0x3c, 0xd2, 0xfa, 0x14, 0x2d, 0xad, 0x3b, 0x41, 0x17, 0x62, 0xa5, 0x78, 0x60}
	require.Equal(t, hex.EncodeToString(passwordHash2), impl.HashCrackedPassword(password2, salt2))

	password3 := "banana"
	salt3 := []byte{0x0, 0x2e}
	passwordHash3 := "c612f289f5324c73d96a20ca14cf834e95a359a2b28101401e1bd7daa3bac4e2"
	require.Equal(t, passwordHash3, impl.HashCrackedPassword(password3, salt3))

}

func Test_Contract_Get_Task_Hash(t *testing.T) {
	hash1 := "484f9573380d13c3042d3601b2001b611d02f4ecc88af2235ec3180de7bd962c"
	pwd1 := "Password"
	salt1 := "0000"

	hash2 := "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860"
	pwd2 := "apple"
	salt2 := "003c"

	hash3 := "c612f289f5324c73d96a20ca14cf834e95a359a2b28101401e1bd7daa3bac4e2"
	pwd3 := "banana"
	salt3 := "002e"

	tasks := map[string][2]string{
		hash1: {pwd1, salt1},
		hash2: {pwd2, salt2},
		hash3: {pwd3, salt3},
	}

	targetHash1, _ := impl.GetTaskHash(tasks, hash2)
	require.Equal(t, targetHash1, hash2)

	targetHash2, _ := impl.GetTaskHash(tasks, hash1)
	require.Equal(t, targetHash2, hash1)

	targetHash3, _ := impl.GetTaskHash(tasks, hash3)
	require.Equal(t, targetHash3, hash3)

	hashNotExisit := "c612f289f5324c73d96a20ca14cf834e95a359a2b28101401e1bd7daa3bac4e9"
	_, err3 := impl.GetTaskHash(tasks, hashNotExisit)
	expectErr3 := xerrors.Errorf("No such hash in the tasks.")
	require.EqualError(t, err3, expectErr3.Error())

}

func Test_Contract_Gather_Action_True(t *testing.T) {

	plainContract :=
		`
		ASSUME publisher.balance > 5
		IF finisher.crackedPwd.hash == "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract := impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"1",                  // finisher
	)

	hash1 := "484f9573380d13c3042d3601b2001b611d02f4ecc88af2235ec3180de7bd962c"
	pwd1 := "Password"
	salt1 := "0000"

	hash2 := "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860"
	pwd2 := "apple"
	salt2 := "003c"

	hash3 := "c612f289f5324c73d96a20ca14cf834e95a359a2b28101401e1bd7daa3bac4e2"
	pwd3 := "banana"
	salt3 := "002e"

	tasks := map[string][2]string{
		hash1: {pwd1, salt1},
		hash2: {pwd2, salt2},
		hash3: {pwd3, salt3},
	}

	worldState := common.QuickWorldState(5, 20)
	state1 := common.State{
		Nonce:       0,
		Balance:     20,
		CodeHash:    "",
		StorageRoot: "",
		Tasks:       tasks,
	}
	worldState.Set("1", state1)

	actions, err := contract.GatherActions(worldState)
	require.NoError(t, err)
	expected1 := "finisher_ID"
	expected2 := int64(46)
	expectedActions := []parser.Action{
		{
			Role:   "smartAccount",
			Action: "transfer",
			Params: []*parser.Value{
				{String: &expected1, Number: nil},
				{String: nil, Number: &expected2},
			},
		},
	}
	require.Equal(t, actions, expectedActions)

}

func Test_Contract_Gather_Action_Error(t *testing.T) {
	hash1 := "484f9573380d13c3042d3601b2001b611d02f4ecc88af2235ec3180de7bd962c"
	pwd1 := "Password"
	salt1 := "0000"

	hash2 := "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860"
	pwd2 := "apple"
	salt2 := "003c"

	hash3 := "c612f289f5324c73d96a20ca14cf834e95a359a2b28101401e1bd7daa3bac4e2"
	pwd3 := "banana"
	salt3 := "002e"

	tasks := map[string][2]string{
		hash1: {pwd1, salt1},
		hash2: {pwd2, salt2},
		hash3: {pwd3, salt3},
	}

	worldState := common.QuickWorldState(5, 20)
	state1 := common.State{
		Nonce:       0,
		Balance:     20,
		CodeHash:    "",
		StorageRoot: "",
		Tasks:       tasks,
	}
	worldState.Set("1", state1)

	//------------- 1st----------------------
	plainContract :=
		`
		ASSUME publisher.balance > 5
		IF publisher.crackedPwd.hash == "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract := impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"1",                  // finisher
	)

	_, err := contract.GatherActions(worldState)
	expectErr := xerrors.Errorf("invalid grammar. Expecting [finisher], get: publisher")
	require.EqualError(t, err, expectErr.Error())

	//------------- 2nd----------------------
	plainContract =
		`
		ASSUME publisher.balance > 5
		IF finisher.crackedPwd.hash == 36 THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract = impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"1",                  // finisher
	)

	_, err = contract.GatherActions(worldState)
	expectErr = xerrors.Errorf("invalid grammar. Expecting a hash string")
	require.EqualError(t, err, expectErr.Error())

	//------------- 3nd----------------------
	plainContract =
		`
		ASSUME publisher.balance > 5
		IF finisher.crackedPwd.hash.blah == "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract = impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"1",                  // finisher
	)

	_, err = contract.GatherActions(worldState)
	expectErr = xerrors.Errorf("Condition field unknown. Need to have two attributes")
	require.EqualError(t, err, expectErr.Error())

	//------------- 4th----------------------
	plainContract =
		`
		ASSUME publisher.balance > 5
		IF finisher.crackedPwd.hash == "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract = impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"6",                  // finisher
	)

	_, err = contract.GatherActions(worldState)
	expectErr = xerrors.Errorf("account doesn't exists or account state is corrupted")
	require.EqualError(t, err, expectErr.Error())

	//------------- 5th----------------------
	plainContract =
		`
		ASSUME publisher.balance > 5
		IF finisher.abc.abc == "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract = impl.NewContract(
		"100ixwscds",         // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"p1",                 // publisher
		"1",                  // finisher
	)

	_, err = contract.GatherActions(worldState)
	expectErr = xerrors.Errorf("invalid grammar. Expecting [crackedPwd.hash], get: abc.abc")
	require.EqualError(t, err, expectErr.Error())

}
