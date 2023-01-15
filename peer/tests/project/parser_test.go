package project

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
	"golang.org/x/xerrors"
)

// Test parsing Value of String type
func Test_Parse_Value_String(t *testing.T) {
	str1 := `"Qiyuan Liang"`
	str2 := `"Yanni Zhang"`
	str3 := `"Qiyuan Dong"`
	testStr := []string{str1, str2, str3}
	expectedStr := []string{`Qiyuan Liang`, `Yanni Zhang`, `Qiyuan Dong`}
	var parsedStr []string

	var ValParser = participle.MustBuild(&parser.Value{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	for _, s := range testStr {
		valAST := &parser.Value{}
		err := ValParser.ParseString("", s, valAST)
		require.NoError(t, err)
		parsedStr = append(parsedStr, *valAST.String)
	}
	require.Equal(t, expectedStr, parsedStr)

}

// Test parsing Value of Int type
func Test_Parse_Value_Int64(t *testing.T) {
	test := []string{"0", "666", "1234", "1"}
	expected := []int64{0, 666, 1234, 1}
	var parsed []int64

	var ValParser = participle.MustBuild(&parser.Value{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)
	for _, s := range test {
		valAST := &parser.Value{}
		err := ValParser.ParseString("", s, valAST)
		require.NoError(t, err)
		parsed = append(parsed, *valAST.Number)
	}
	require.Equal(t, expected, parsed)
}

// Test parsing Object
func Test_Parse_Object(t *testing.T) {

	objPlain := []string{
		`publisher.budget`,
		`finisher.key35`,
		`smartAccount.attribute`,
	}
	var parsedObjs []*parser.Object

	expectedObjs := []*parser.Object{
		{
			Role: "publisher",
			Fields: []*parser.Field{
				{Name: "budget"},
			},
		},
		{
			Role: "finisher",
			Fields: []*parser.Field{
				{Name: "key35"},
			},
		},
		{
			Role: "smartAccount",
			Fields: []*parser.Field{
				{Name: "attribute"},
			},
		},
	}

	var ObjectParser = participle.MustBuild(&parser.Object{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	for _, obj := range objPlain {
		objAST := &parser.Object{}
		err := ObjectParser.ParseString("", obj, objAST)
		require.NoError(t, err)
		parsedObjs = append(parsedObjs, objAST)
	}
	require.Equal(t, expectedObjs, parsedObjs)

}

func Test_Parse_Object_Error(t *testing.T) {

	objPlain := []string{
		`a.budget`,
		`b.key35`,
		`c.attribute`,
	}
	var ObjectParser = participle.MustBuild(&parser.Object{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	unexpectedToken := []string{`"a"`, `"b"`, `"c"`}
	for i, obj := range objPlain {
		objAST := &parser.Object{}
		err := ObjectParser.ParseString("", obj, objAST)

		expectedErrMsg := fmt.Sprintf(`1:1: unexpected token %s`, unexpectedToken[i])
		expectedErr := xerrors.Errorf(expectedErrMsg)
		require.EqualError(t, err, expectedErr.Error())
	}

}

func Test_Parse_Object_Multi_Attribute(t *testing.T) {

	objPlain := []string{
		`publisher.budget.blah`,
		`finisher.key0.hash`,
		`smartAccount.attribute.attribute.attribute`,
	}
	var parsedObjs []*parser.Object

	expectedObjs := []*parser.Object{
		{
			Role: "publisher",
			Fields: []*parser.Field{
				{Name: "budget"},
				{Name: "blah"},
			},
		},
		{
			Role: "finisher",
			Fields: []*parser.Field{
				{Name: "key0"},
				{Name: "hash"},
			},
		},
		{
			Role: "smartAccount",
			Fields: []*parser.Field{
				{Name: "attribute"},
				{Name: "attribute"},
				{Name: "attribute"},
			},
		},
	}

	var ObjectParser = participle.MustBuild(&parser.Object{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	for _, obj := range objPlain {
		objAST := &parser.Object{}
		err := ObjectParser.ParseString("", obj, objAST)
		require.NoError(t, err)
		parsedObjs = append(parsedObjs, objAST)
	}
	require.Equal(t, expectedObjs, parsedObjs)

}

// Test parsing Condition
func Test_Parse_Condition(t *testing.T) {
	conditionStrings := []string{
		`publisher.budget > 3`,
		`finisher.key24.verified > 0`,
		`smartAccount.attribute.attribute == "yeah"`,
		`smartAccount.a.b.c.d != "hahaha"`,
	}
	expectedValue1 := int64(3)
	expectedValue2 := int64(0)
	expectedValue3 := "yeah"
	expectedValue4 := "hahaha"

	var parsedConditions []*parser.Condition

	expectedConditions := []*parser.Condition{
		{
			Object: parser.Object{
				Role: "publisher",
				Fields: []*parser.Field{
					{Name: "budget"},
				},
			},
			Operator: ">",
			Value: parser.Value{
				String: nil,
				Number: &expectedValue1,
			},
		},
		{
			Object: parser.Object{
				Role: "finisher",
				Fields: []*parser.Field{
					{Name: "key24"},
					{Name: "verified"},
				},
			},
			Operator: ">",
			Value: parser.Value{
				String: nil,
				Number: &expectedValue2,
			},
		},
		{
			Object: parser.Object{
				Role: "smartAccount",
				Fields: []*parser.Field{
					{Name: "attribute"},
					{Name: "attribute"},
				},
			},
			Operator: "==",
			Value: parser.Value{
				String: &expectedValue3,
				Number: nil,
			},
		},
		{
			Object: parser.Object{
				Role: "smartAccount",
				Fields: []*parser.Field{
					{Name: "a"},
					{Name: "b"},
					{Name: "c"},
					{Name: "d"},
				},
			},
			Operator: "!=",
			Value: parser.Value{
				String: &expectedValue4,
				Number: nil,
			},
		},
	}

	var ConditionParser = participle.MustBuild(&parser.Condition{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	for _, c := range conditionStrings {
		condAST := &parser.Condition{}
		err := ConditionParser.ParseString("", c, condAST)
		require.NoError(t, err)
		parsedConditions = append(parsedConditions, condAST)
	}
	require.Equal(t, expectedConditions, parsedConditions)

}

// Test parsing ConditionObjObj
// this is the comparison between obj and obj
func Test_Parse_ConditionObjObj(t *testing.T) {

	conditionStrings := []string{
		`publisher.budget > publisher.account`,
		`finisher.key24.hash ==  finisher.hash24`,
		`publisher.attribute.attribute >= finisher.attribute.attribute`,
	}

	var parsedConditions []*parser.ConditionObjObj

	expectedConditions := []*parser.ConditionObjObj{
		{
			Object1: parser.Object{
				Role: "publisher",
				Fields: []*parser.Field{
					{Name: "budget"},
				},
			},
			Operator: ">",
			Object2: parser.Object{
				Role: "publisher",
				Fields: []*parser.Field{
					{Name: "account"},
				},
			},
		},
		{
			Object1: parser.Object{
				Role: "finisher",
				Fields: []*parser.Field{
					{Name: "key24"},
					{Name: "hash"},
				},
			},
			Operator: "==",
			Object2: parser.Object{
				Role: "finisher",
				Fields: []*parser.Field{
					{Name: "hash24"},
				},
			},
		},
		{
			Object1: parser.Object{
				Role: "publisher",
				Fields: []*parser.Field{
					{Name: "attribute"},
					{Name: "attribute"},
				},
			},
			Operator: ">=",
			Object2: parser.Object{
				Role: "finisher",
				Fields: []*parser.Field{
					{Name: "attribute"},
					{Name: "attribute"},
				},
			},
		},
	}

	var ConditionParser = participle.MustBuild(&parser.ConditionObjObj{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	for _, c := range conditionStrings {
		condAST := &parser.ConditionObjObj{}
		err := ConditionParser.ParseString("", c, condAST)
		require.NoError(t, err)
		parsedConditions = append(parsedConditions, condAST)
	}
	require.Equal(t, expectedConditions, parsedConditions)

}

// test contract code that has one assumption
func Test_Parse_Assumption(t *testing.T) {
	assumeStrings := []string{
		`ASSUME publisher.budget > 49`,
		`ASSUME publisher.attribute.attribute == "yeah"`,
		`ASSUME smartAccount.attribute.attribute != "hahaha"`,
	}
	expectedValue1 := int64(49)
	expectedValue2 := "yeah"
	expectedValue3 := "hahaha"

	var parsedAssume []*parser.Assumption

	expectedAssume := []*parser.Assumption{
		{
			Condition: parser.Condition{
				Object: parser.Object{
					Role: "publisher",
					Fields: []*parser.Field{
						{Name: "budget"},
					},
				},
				Operator: ">",
				Value: parser.Value{
					String: nil,
					Number: &expectedValue1,
				},
			},
		},
		{
			Condition: parser.Condition{
				Object: parser.Object{
					Role: "publisher",
					Fields: []*parser.Field{
						{Name: "attribute"},
						{Name: "attribute"},
					},
				},
				Operator: "==",
				Value: parser.Value{
					String: &expectedValue2,
					Number: nil,
				},
			},
		},
		{
			Condition: parser.Condition{
				Object: parser.Object{
					Role: "smartAccount",
					Fields: []*parser.Field{
						{Name: "attribute"},
						{Name: "attribute"},
					},
				},
				Operator: "!=",
				Value: parser.Value{
					String: &expectedValue3,
					Number: nil,
				},
			},
		},
	}

	var AssumeParser = participle.MustBuild(&parser.Assumption{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)
	for _, s := range assumeStrings {
		assumeAST := &parser.Assumption{}
		err := AssumeParser.ParseString("", s, assumeAST)
		require.NoError(t, err)
		parsedAssume = append(parsedAssume, assumeAST)
	}
	require.Equal(t, expectedAssume, parsedAssume)

}

func Test_Parse_Assumption_Error(t *testing.T) {
	// should use ASSUME
	assumeString := `ASSUMPTION publisher.budget > 49`
	var AssumeParser = participle.MustBuild(&parser.Assumption{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)
	assumeAST := &parser.Assumption{}
	err := AssumeParser.ParseString("", assumeString, assumeAST)
	expectedErrMsg := `1:1: unexpected token "ASSUMPTION"`
	expectedErr := xerrors.Errorf(expectedErrMsg)
	require.EqualError(t, err, expectedErr.Error())

}

// test id clause
func Test_Parse_Ifclause(t *testing.T) {
	ifStrings := []string{
		`IF finisher.key67.hash == "inowrogionjde" THEN
			smartAccount.transfer("finisher_ID", 46)
		`,
	}
	// expectedValue2 := "publisher_ID"
	expectedValue3 := "inowrogionjde"
	expectedValue4 := "finisher_ID"
	expectedValue5 := int64(46)

	var parsedIf []*parser.IfClause

	expectedIf := []*parser.IfClause{
		{
			Condition: parser.Condition{
				Object: parser.Object{
					Role: "finisher",
					Fields: []*parser.Field{
						{Name: "key67"},
						{Name: "hash"},
					},
				},
				Operator: "==",
				Value: parser.Value{
					String: &expectedValue3,
					Number: nil,
				},
			},
			Actions: []*parser.Action{
				{
					Role:   "smartAccount",
					Action: "transfer",
					Params: []*parser.Value{
						{String: &expectedValue4, Number: nil},
						{String: nil, Number: &expectedValue5},
					},
				},
			},
		},
	}

	var IfParser = participle.MustBuild(&parser.IfClause{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	for _, s := range ifStrings {
		ifAST := &parser.IfClause{}
		err := IfParser.ParseString("", s, ifAST)
		require.NoError(t, err)
		parsedIf = append(parsedIf, ifAST)
	}
	require.Equal(t, expectedIf, parsedIf)

}

func Test_Parse_Ifclause_Error(t *testing.T) {
	// lack THEN
	ifString := `IF finisher.key67.hash == "inowrogionjde"
			smartAccount.transfer("finisher_ID", 46)
		`
	var IfParser = participle.MustBuild(&parser.IfClause{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)
	ifAST := &parser.IfClause{}
	err := IfParser.ParseString("", ifString, ifAST)
	expectedErrMsg := `2:4: unexpected token "smartAccount" (expected ("THEN" Action+))`
	expectedErr := xerrors.Errorf(expectedErrMsg)
	require.EqualError(t, err, expectedErr.Error())

}

// Test parsing Action with multiple attribute
func Test_Parse_Action(t *testing.T) {
	actionStrings := []string{
		`smartAccount.transfer("finisher_ID", 46)`,
		`smartAccount.transfer("finisher_ID", "crackedKey")`,
	}
	expectedValue11 := "finisher_ID"
	expectedValue12 := int64(46)
	expectedValue21 := "finisher_ID"
	expectedValue22 := "crackedKey"

	var parsedActions []*parser.Action

	expectedActions := []*parser.Action{
		{
			Role:   "smartAccount",
			Action: "transfer",
			Params: []*parser.Value{
				{String: &expectedValue11, Number: nil},
				{String: nil, Number: &expectedValue12},
			},
		},
		{
			Role:   "smartAccount",
			Action: "transfer",
			Params: []*parser.Value{
				{String: &expectedValue21, Number: nil},
				{String: &expectedValue22, Number: nil},
			},
		},
	}

	var ActionParser = participle.MustBuild(&parser.Action{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	for _, s := range actionStrings {
		actionAST := &parser.Action{}
		err := ActionParser.ParseString("", s, actionAST)
		require.NoError(t, err)
		parsedActions = append(parsedActions, actionAST)
	}
	require.Equal(t, expectedActions, parsedActions)

}

func Test_Parse_Action_Error(t *testing.T) {
	actionStrings := []string{
		`smartAccount.abc("finisher_ID", 46)`,
		`smartAccount.def("finisher_ID", "crackedKey")`,
	}

	var ActionParser = participle.MustBuild(&parser.Action{},
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)

	unexpectedToken := []string{`"abc"`, `"def"`}
	for i, s := range actionStrings {
		actionAST := &parser.Action{}
		err := ActionParser.ParseString("", s, actionAST)
		expectedErrMsg := fmt.Sprintf(`1:14: unexpected token %s (expected "transfer")`, unexpectedToken[i])
		expectedErr := xerrors.Errorf(expectedErrMsg)
		require.EqualError(t, err, expectedErr.Error())
	}
}

// test the functionality of parsing entire contract
func Test_Parse_Contract(t *testing.T) {
	codeStrings := []string{
		`
		ASSUME publisher.budget > 0
		IF finisher.key98.hash == "inowrogionjde" THEN
			smartAccount.transfer("finisher_ID", 46)
		`,
	}
	expectedValue1 := int64(0)
	// expectedValue2 := int64(0)
	// expectedValue3 := "publisher_ID"
	expectedValue4 := "inowrogionjde"
	expectedValue5 := "finisher_ID"
	expectedValue6 := int64(46)
	var parsedCode []*parser.Code

	expectedCode := []*parser.Code{
		{
			Assumptions: []*parser.Assumption{
				{
					Condition: parser.Condition{
						Object: parser.Object{
							Role: "publisher",
							Fields: []*parser.Field{
								{Name: "budget"},
							},
						},
						Operator: ">",
						Value: parser.Value{
							String: nil,
							Number: &expectedValue1,
						},
					},
				},
			},
			IfClauses: []*parser.IfClause{
				{
					Condition: parser.Condition{
						Object: parser.Object{
							Role: "finisher",
							Fields: []*parser.Field{
								{Name: "key98"},
								{Name: "hash"},
							},
						},
						Operator: "==",
						Value: parser.Value{
							String: &expectedValue4,
							Number: nil,
						},
					},
					Actions: []*parser.Action{
						{
							Role:   "smartAccount",
							Action: "transfer",
							Params: []*parser.Value{
								{String: &expectedValue5, Number: nil},
								{String: nil, Number: &expectedValue6},
							},
						},
					},
				},
			},
		},
	}

	for _, s := range codeStrings {
		codeAST, err := parser.BuildCodeAST(s)
		require.NoError(t, err)
		parsedCode = append(parsedCode, &codeAST)
	}
	require.Equal(t, expectedCode, parsedCode)

}
