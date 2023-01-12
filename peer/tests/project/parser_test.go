package project

import (
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

func GetParser(grammar interface{}) *participle.Parser {
	return participle.MustBuild(grammar,
		participle.Lexer(parser.ContractLexer),
		participle.Unquote("String"),
	)
}

// Test parsing Value of String type
func Test_Parser_Value_String(t *testing.T) {
	testString := []string{`"Qiyuan Liang"`, `"Yanni Zhang"`, `"Qiyuan Dong"`}
	expectedStrings := []string{`Qiyuan Liang`, `Yanni Zhang`, `Qiyuan Dong`}
	var parsedStrings []string

	var ValParser = GetParser(&parser.Value{})
	for _, s := range testString {
		valAST := &parser.Value{}
		err := ValParser.ParseString("", s, valAST)
		require.NoError(t, err)
		parsedStrings = append(parsedStrings, *valAST.String)
	}
	require.Equal(t, expectedStrings, parsedStrings)

}

// Test parsing Value of Float type
func Test_Parser_Value_Float(t *testing.T) {
	testFloats := []string{"0", "666", "1234", "1.125"}
	expectedFloats := []float64{0, 666, 1234, 1.125}
	var parsedFloats []float64

	var ValParser = GetParser(&parser.Value{})
	for _, s := range testFloats {
		valAST := &parser.Value{}
		err := ValParser.ParseString("", s, valAST)
		require.NoError(t, err)
		parsedFloats = append(parsedFloats, *valAST.Number)
	}
	require.Equal(t, expectedFloats, parsedFloats)
}

// Test parsing Object
func Test_Parser_Object(t *testing.T) {

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

	var ObjectParser = GetParser(&parser.Object{})

	for _, obj := range objPlain {
		objAST := &parser.Object{}
		err := ObjectParser.ParseString("", obj, objAST)
		require.NoError(t, err)
		parsedObjs = append(parsedObjs, objAST)
	}
	require.Equal(t, expectedObjs, parsedObjs)

}

func Test_Parser_Object_Multi_Attribute(t *testing.T) {

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

	var ObjectParser = GetParser(&parser.Object{})

	for _, obj := range objPlain {
		objAST := &parser.Object{}
		err := ObjectParser.ParseString("", obj, objAST)
		require.NoError(t, err)
		parsedObjs = append(parsedObjs, objAST)
	}
	require.Equal(t, expectedObjs, parsedObjs)

}

// Test parsing Condition
func Test_Parser_Condition(t *testing.T) {
	conditionStrings := []string{
		`publisher.budget > 3.246`,
		`finisher.key24.verified > 0`,
		`smartAccount.attribute.attribute == "yeah"`,
	}
	expectedValue1 := float64(3.246)
	expectedValue2 := float64(0)
	expectedValue3 := "yeah"

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
	}

	var ConditionParser = GetParser(&parser.Condition{})

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
func Test_Parser_ConditionObjObj(t *testing.T) {

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

	var ConditionParser = GetParser(&parser.ConditionObjObj{})

	for _, c := range conditionStrings {
		condAST := &parser.ConditionObjObj{}
		err := ConditionParser.ParseString("", c, condAST)
		require.NoError(t, err)
		parsedConditions = append(parsedConditions, condAST)
	}
	require.Equal(t, expectedConditions, parsedConditions)

}

// Test parsing Action with multiple attribute
func Test_Parser_Action(t *testing.T) {
	actionStrings := []string{
		`smartAccount.transfer("finisher_ID", 46.967)`,
		`smartAccount.transfer("finisher_ID", "crackedKey")`,
	}
	expectedValue11 := "finisher_ID"
	expectedValue12 := float64(46.967)
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

	var ActionParser = GetParser(&parser.Action{})

	for _, s := range actionStrings {
		actionAST := &parser.Action{}
		err := ActionParser.ParseString("", s, actionAST)
		require.NoError(t, err)
		parsedActions = append(parsedActions, actionAST)
	}
	require.Equal(t, expectedActions, parsedActions)

}

// test contract code that has one assumption
func Test_Parser_Assumption(t *testing.T) {
	assumeStrings := []string{
		`ASSUME publisher.budget > 49.597`,
		`ASSUME publisher.attribute.attribute == "yeah"`,
		`ASSUME smartAccount.attribute.attribute != "hahaha"`,
	}
	expectedValue1 := float64(49.597)
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

	var AssumeParser = GetParser(&parser.Assumption{})
	for _, s := range assumeStrings {
		assumeAST := &parser.Assumption{}
		err := AssumeParser.ParseString("", s, assumeAST)
		require.NoError(t, err)
		parsedAssume = append(parsedAssume, assumeAST)
	}
	require.Equal(t, expectedAssume, parsedAssume)

}

// test id clause
func Test_Parser_Ifclause(t *testing.T) {
	ifStrings := []string{
		`IF finisher.key67.hash == "inowrogionjde" THEN
			smartAccount.transfer("finisher_ID", 46.967)
		`,
	}
	// expectedValue2 := "publisher_ID"
	expectedValue3 := "inowrogionjde"
	expectedValue4 := "finisher_ID"
	expectedValue5 := float64(46.967)

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
				// {
				// 	Role:   "finisher",
				// 	Action: "submit",
				// 	Params: []*parser.Value{
				// 		{String: &expectedValue2, Number: nil},
				// 		{String: &expectedValue3, Number: nil},
				// 	},
				// },
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

	var IfParser = GetParser(&parser.IfClause{})

	for _, s := range ifStrings {
		ifAST := &parser.IfClause{}
		err := IfParser.ParseString("", s, ifAST)
		require.NoError(t, err)
		parsedIf = append(parsedIf, ifAST)
	}
	require.Equal(t, expectedIf, parsedIf)

}

// test the functionality of parsing entire contract
func Test_Parser_Contract(t *testing.T) {
	// codeStrings := []string{
	// 	`
	// 	ASSUME publisher.budget > 0
	// 	IF finisher.key98.hash == finisher.hash98 THEN
	// 		finisher.submit("publisher_ID", "crackedKey")
	// 		publisher.transfer("finisher_ID", 46.967)
	// 	`,
	// }
	codeStrings := []string{
		`
		ASSUME publisher.budget > 0
		IF finisher.key98.hash == "inowrogionjde" THEN
			smartAccount.transfer("finisher_ID", 46.967)
		`,
	}
	expectedValue1 := float64(0)
	// expectedValue2 := float64(0)
	// expectedValue3 := "publisher_ID"
	expectedValue4 := "inowrogionjde"
	expectedValue5 := "finisher_ID"
	expectedValue6 := float64(46.967)
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
						// {
						// 	Role:   "finisher",
						// 	Action: "submit",
						// 	Params: []*parser.Value{
						// 		{String: &expectedValue3, Number: nil},
						// 		{String: &expectedValue4, Number: nil},
						// 	},
						// },
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
