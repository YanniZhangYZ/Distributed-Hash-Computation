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
	fmt.Println(impl.PrintStateAST(codeAST, stateAST))
	fmt.Println(impl.PrintCodeAST(codeAST))
}

func Test_Contract_Build_Plain_Text(t *testing.T) {
	plainContract := impl.BuildPlainContract("yuvubknluykgink", "finisherAddr", 3)
	expected := `
	ASSUME smartAccount.balance > 0
	IF finisher.crackedPwd.hash == "yuvubknluykgink" THEN
		smartAccount.transfer("finisherAddr", 3)
	`
	require.Equal(t, plainContract, expected)
}

func Test_Contract_Check_Assumption_True(t *testing.T) {

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
}

func Test_Contract_Check_Assumption_False(t *testing.T) {
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

func Test_Contract_Check_Assumption_Wrong_Role(t *testing.T) {

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
}

func Test_Contract_Check_Assumption_Wrong_Number_Attribute(t *testing.T) {

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
}

func Test_Contract_Check_Assumption_No_Such_Account(t *testing.T) {

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
}

func Test_Contract_Check_Assumption_Wrong_Attribute(t *testing.T) {

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

}

func Test_Contract_Check_Assumption_Left_Right_Type_Inconsistent(t *testing.T) {
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

	salts := [][]byte{{0x0, 0x0}, {0x0, 0x1}, {0x0, 0x2}, {0x0, 0x3}, {0x0, 0x4}, {0x0, 0x5}, {0x0, 0x6}, {0x0, 0x7}, {0x0, 0x8}, {0x0, 0x9},
		{0x1, 0x0}, {0x1, 0x1}, {0x1, 0x2}, {0x1, 0x3}, {0x1, 0x4}, {0x1, 0x5}, {0x1, 0x6}, {0x1, 0x7}, {0x1, 0x8}, {0x1, 0x9}}
	hashStrs := []string{
		"14ffb81ab8f435a96400880c8bf34dba05a7ef8b63710f136e87297e601d7881",
		"deb253c70e2318c3161561307094c14f0637fc9a528884125374c87d8cc9978b",
		"5f8fdee1def61c429a911a420e26f777c605f883e496c9823df54902c3326a99",
		"08cb91740f17c9e2f0dfc492031746b9c5925ec39c78b21f194381603dbf5e37",
		"c95bd1a106693dedea6570c60b1a24394fecf6f43d5c51b51819b96ccba483aa",
		"536dde0c6fdc7c5d811dd5e8cf80981c393d45fe84cf3fab7bc59cab5fac9033",
		"7389fbde2eb57ed20f942bb757854a95ccbf65508d0644e3d0353543b3316913",
		"cb045966ebe244998d4e4a24c9905ebeb8878248590231624c02ae174e83affc",
		"4b4b329d70d37f09638e8545bd708bd0e212e7a9c5c6352a3bfdc4b446f57413",
		"86f76685a10823db81f55fd81523a46cb2eb6a99c27317e0f376999cc741ec44",
		"c8d430ffe501ad5087fd31a98fbabb834beb0c82e722bd3be0991e2d399a0868",
		"e873502053f5475ec34f7be7fba48c7030a03820b2e61d2050d9db682587ca17",
		"1e6d28d2c48a2e9e0d81548a3e99852de7f9244609f6d9cf45e9f0dd35a4132c",
		"4b26e856e459ff373866707088d202ad7e745b348680fccd93e43bbd411e30c2",
		"39cef83ff1c135d71776a76439d72265b2ad99b855bbaa0e91a8004230564e7d",
		"e0371dd92ce8492f78a9be094e65d4e3ed7f8d3a819701e7afffb3922e743251",
		"83777a16726539e4f592ea8c7ec0afd9dad8e83deb6129ce39dc49f0e687f908",
		"ec21d75489c5f5a350b56a9175ceb037721a7952ba2e92bdfaf10e99b3ac05c8",
		"3a72c6e038ce875d3802582e4d436d518a1e21033caf2a84b3ca9c46bd6b20f4",
		"180e13060acd8a66e95ecdd6bd6eeb56f8fb1400c0cb9360fd9d92090e88709d",
	}

	for i := 0; i < len(salts); i++ {
		require.Equal(t, hashStrs[i], impl.HashCrackedPassword("apple", salts[i]))
	}

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

}

func Test_Contract_Get_Task_Hash_Error(t *testing.T) {
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

	hashNotExisit := "c612f289f5324c73d96a20ca14cf834e95a359a2b28101401e1bd7daa3bac4e9"
	_, err := impl.GetTaskHash(tasks, hashNotExisit)
	expectErr := xerrors.Errorf("No such hash in the tasks.")
	require.EqualError(t, err, expectErr.Error())

	taskEmpty := make(map[string][2]string)
	_, err2 := impl.GetTaskHash(taskEmpty, hash2)
	expectErr2 := xerrors.Errorf("Task list is empty. No such hash.")
	require.EqualError(t, err2, expectErr2.Error())

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

	_, actions, err := contract.GatherActions(worldState)
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

	_, _, err := contract.GatherActions(worldState)
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

	_, _, err = contract.GatherActions(worldState)
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

	_, _, err = contract.GatherActions(worldState)
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

	_, _, err = contract.GatherActions(worldState)
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

	_, _, err = contract.GatherActions(worldState)
	expectErr = xerrors.Errorf("invalid grammar. Expecting [crackedPwd.hash], get: abc.abc")
	require.EqualError(t, err, expectErr.Error())

}

func Test_Contract_All_True(t *testing.T) {

	plainContract :=
		`
		ASSUME smartAccount.balance > 5
		IF finisher.crackedPwd.hash == "6ad18f940ffbd30454e3c2ecf6178c6492deb33cd2fa142dad3b411762a57860" THEN
		smartAccount.transfer("finisher_ID", 46)
	`

	// create a contract instance
	contract := impl.NewContract(
		"2",                  // ID
		"crack pwd contract", // name
		plainContract,        // plain_code
		"nowqnpfe",           // publisher
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

	isValid, err := contract.CheckAssumptions(worldState)
	require.NoError(t, err)
	require.Equal(t, isValid, true)

	fmt.Println("------------after assumption check----------------")
	contract.PrintContractExecutionState()

	_, actions, err := contract.GatherActions(worldState)
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

	fmt.Println("------------after assumption check----------------")
	contract.PrintContractExecutionState()

}
