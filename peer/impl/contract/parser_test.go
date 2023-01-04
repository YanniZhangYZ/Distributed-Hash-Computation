package contract

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

// Unit tests of the Smart contract functionalities

func GetParser(grammar interface{}) *participle.Parser {
	return participle.MustBuild(grammar,
		participle.Lexer(parser.SMTLexer),
		participle.Unquote("String"),
	)
}

// Test parsing Value
func Test_Parser_Value_String(t *testing.T) {
	test_strings := []string{`"Qiyuan Liang"`, `"Yanni Zhang"`, `"Qiyuan Dong"`}
	expected_strings := []string{`Qiyuan Liang`, `Yanni Zhang`, `Qiyuan Dong`}
	var parsed_strings []string

	var ValParser = GetParser(&parser.Value{})
	for _, s := range test_strings {
		val_ast := &parser.Value{}
		err := ValParser.ParseString("", s, val_ast)
		require.NoError(t, err)
		parsed_strings = append(parsed_strings, *val_ast.String)
	}
	require.Equal(t, expected_strings, parsed_strings)

}

func Test_Parser_Value_Float(t *testing.T) {
	test_floats := []string{"0", "666", "1234", "1.125"}
	expected_floats := []float64{0, 666, 1234, 1.125}
	var parsed_floats []float64

	var ValParser = GetParser(&parser.Value{})
	for _, s := range test_floats {
		val_ast := &parser.Value{}
		err := ValParser.ParseString("", s, val_ast)
		require.NoError(t, err)
		parsed_floats = append(parsed_floats, *val_ast.Number)
	}
	require.Equal(t, expected_floats, parsed_floats)
}

// Test parsing Object
func Test_Parser_Object(t *testing.T) {
	fmt.Println("test obj")

	obj_plain := []string{
		`publisher.budget`,
		`finisher.key35`,
		`finisher.key0.hash`,
	}
	var parsed_objs []*parser.Object

	expected_objs := []*parser.Object{
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
			Role: "finisher",
			Fields: []*parser.Field{
				{Name: "key0"},
				{Name: "hash"},
			},
		},
	}

	var ObjectParser = GetParser(&parser.Object{})

	for _, obj := range obj_plain {
		obj_ast := &parser.Object{}
		err := ObjectParser.ParseString("", obj, obj_ast)
		require.NoError(t, err)
		parsed_objs = append(parsed_objs, obj_ast)
	}
	require.Equal(t, expected_objs, parsed_objs)

}

// Test parsing Condition
func Test_Parser_Condition(t *testing.T) {
	fmt.Println("test condition")

	condition_strings := []string{
		`publisher.budget > 3.246`,
		`finisher.key24.verified > 0`,
		`publisher.attribute.attribute == "yeah"`,
	}
	expected_value1 := float64(3.246)
	expected_value2 := float64(0)
	expected_value3 := "yeah"

	var parsed_conditions []*parser.Condition

	expected_conditions := []*parser.Condition{
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
				Number: &expected_value1,
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
				Number: &expected_value2,
			},
		},
		{
			Object: parser.Object{
				Role: "publisher",
				Fields: []*parser.Field{
					{Name: "attribute"},
					{Name: "attribute"},
				},
			},
			Operator: "==",
			Value: parser.Value{
				String: &expected_value3,
				Number: nil,
			},
		},
	}

	var ConditionParser = GetParser(&parser.Condition{})

	for _, c := range condition_strings {
		cond_ast := &parser.Condition{}
		err := ConditionParser.ParseString("", c, cond_ast)
		require.NoError(t, err)
		parsed_conditions = append(parsed_conditions, cond_ast)
	}
	require.Equal(t, expected_conditions, parsed_conditions)

}

func Test_Parser_ConditionObjObj(t *testing.T) {
	fmt.Println("test condition obj obj")

	condition_strings := []string{
		`publisher.budget > publisher.account`,
		`finisher.key24.hash ==  finisher.hash24`,
		`publisher.attribute.attribute >= finisher.attribute.attribute`,
	}

	var parsed_conditions []*parser.ConditionObjObj

	expected_conditions := []*parser.ConditionObjObj{
		{
			Object_1: parser.Object{
				Role: "publisher",
				Fields: []*parser.Field{
					{Name: "budget"},
				},
			},
			Operator: ">",
			Object_2: parser.Object{
				Role: "publisher",
				Fields: []*parser.Field{
					{Name: "account"},
				},
			},
		},
		{
			Object_1: parser.Object{
				Role: "finisher",
				Fields: []*parser.Field{
					{Name: "key24"},
					{Name: "hash"},
				},
			},
			Operator: "==",
			Object_2: parser.Object{
				Role: "finisher",
				Fields: []*parser.Field{
					{Name: "hash24"},
				},
			},
		},
		{
			Object_1: parser.Object{
				Role: "publisher",
				Fields: []*parser.Field{
					{Name: "attribute"},
					{Name: "attribute"},
				},
			},
			Operator: ">=",
			Object_2: parser.Object{
				Role: "finisher",
				Fields: []*parser.Field{
					{Name: "attribute"},
					{Name: "attribute"},
				},
			},
		},
	}

	var ConditionParser = GetParser(&parser.ConditionObjObj{})

	for _, c := range condition_strings {
		cond_ast := &parser.ConditionObjObj{}
		err := ConditionParser.ParseString("", c, cond_ast)
		require.NoError(t, err)
		parsed_conditions = append(parsed_conditions, cond_ast)
	}
	require.Equal(t, expected_conditions, parsed_conditions)

}

// Test parsing Action
func Test_Parser_Action(t *testing.T) {
	action_strings := []string{
		`publisher.transfer("finisher_ID", 46.967)`,
		`finisher.submit("publisher_ID", "crackedKey")`,
	}
	expected_value1_1 := "finisher_ID"
	expected_value1_2 := float64(46.967)
	expected_value2_1 := "publisher_ID"
	expected_value2_2 := "crackedKey"

	var parsed_actions []*parser.Action

	expected_actions := []*parser.Action{
		{
			Role:   "publisher",
			Action: "transfer",
			Params: []*parser.Value{
				{String: &expected_value1_1, Number: nil},
				{String: nil, Number: &expected_value1_2},
			},
		},
		{
			Role:   "finisher",
			Action: "submit",
			Params: []*parser.Value{
				{String: &expected_value2_1, Number: nil},
				{String: &expected_value2_2, Number: nil},
			},
		},
	}

	var ActionParser = GetParser(&parser.Action{})

	for _, s := range action_strings {
		action_ast := &parser.Action{}
		err := ActionParser.ParseString("", s, action_ast)
		require.NoError(t, err)
		parsed_actions = append(parsed_actions, action_ast)
	}
	require.Equal(t, expected_actions, parsed_actions)

}

// Parsing contract code only with one assumption
func Test_Parser_Assume(t *testing.T) {
	assume_strings := []string{
		`ASSUME publisher.budget > 49.597`,
		`ASSUME publisher.attribute.attribute == "yeah"`,
	}
	expected_value1 := float64(49.597)
	expected_value2 := "yeah"

	var parsed_assume []*parser.Assumption

	expected_assume := []*parser.Assumption{
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
					Number: &expected_value1,
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
					String: &expected_value2,
					Number: nil,
				},
			},
		},
	}

	var AssumeParser = GetParser(&parser.Assumption{})
	for _, s := range assume_strings {
		assume_ast := &parser.Assumption{}
		err := AssumeParser.ParseString("", s, assume_ast)
		require.NoError(t, err)
		parsed_assume = append(parsed_assume, assume_ast)
	}
	require.Equal(t, expected_assume, parsed_assume)

}

// Parsing contract code with one if clause
func Test_Parser_Ifclause(t *testing.T) {
	if_strings := []string{
		`IF finisher.key67.hash == finisher.hash67 THEN
			finisher.submit("publisher_ID", "crackedKey")
			publisher.transfer("finisher_ID", 46.967)
		`,
	}
	expected_value2 := "publisher_ID"
	expected_value3 := "crackedKey"
	expected_value4 := "finisher_ID"
	expected_value5 := float64(46.967)

	var parsed_if []*parser.IfClause

	expected_if := []*parser.IfClause{
		{
			ConditionObjObj: parser.ConditionObjObj{
				Object_1: parser.Object{
					Role: "finisher",
					Fields: []*parser.Field{
						{Name: "key67"},
						{Name: "hash"},
					},
				},
				Operator: "==",
				Object_2: parser.Object{
					Role: "finisher",
					Fields: []*parser.Field{
						{Name: "hash67"},
					},
				},
			},
			Actions: []*parser.Action{
				{
					Role:   "finisher",
					Action: "submit",
					Params: []*parser.Value{
						{String: &expected_value2, Number: nil},
						{String: &expected_value3, Number: nil},
					},
				},
				{
					Role:   "publisher",
					Action: "transfer",
					Params: []*parser.Value{
						{String: &expected_value4, Number: nil},
						{String: nil, Number: &expected_value5},
					},
				},
			},
		},
	}

	var IfParser = GetParser(&parser.IfClause{})

	for _, s := range if_strings {
		if_ast := &parser.IfClause{}
		err := IfParser.ParseString("", s, if_ast)
		require.NoError(t, err)
		parsed_if = append(parsed_if, if_ast)
	}
	require.Equal(t, expected_if, parsed_if)

}

// Parsing complete contract code case
func Test_Parser_Complete(t *testing.T) {
	code_strings := []string{
		`
		ASSUME publisher.budget > 0
		IF finisher.key98.hash == finisher.hash98 THEN
			finisher.submit("publisher_ID", "crackedKey")
			publisher.transfer("finisher_ID", 46.967)
		`,
	}
	expected_value1 := float64(0)
	// expected_value2 := float64(0)
	expected_value3 := "publisher_ID"
	expected_value4 := "crackedKey"
	expected_value5 := "finisher_ID"
	expected_value6 := float64(46.967)
	var parsed_code []*parser.Code

	expected_code := []*parser.Code{
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
							Number: &expected_value1,
						},
					},
				},
			},
			IfClauses: []*parser.IfClause{
				{
					ConditionObjObj: parser.ConditionObjObj{
						Object_1: parser.Object{
							Role: "finisher",
							Fields: []*parser.Field{
								{Name: "key98"},
								{Name: "hash"},
							},
						},
						Operator: "==",
						Object_2: parser.Object{
							Role: "finisher",
							Fields: []*parser.Field{
								{Name: "hash98"},
							},
						},
					},
					Actions: []*parser.Action{
						{
							Role:   "finisher",
							Action: "submit",
							Params: []*parser.Value{
								{String: &expected_value3, Number: nil},
								{String: &expected_value4, Number: nil},
							},
						},
						{
							Role:   "publisher",
							Action: "transfer",
							Params: []*parser.Value{
								{String: &expected_value5, Number: nil},
								{String: nil, Number: &expected_value6},
							},
						},
					},
				},
			},
		},
	}

	for _, s := range code_strings {
		code_ast, err := parser.Parse(s)
		require.NoError(t, err)
		parsed_code = append(parsed_code, &code_ast)
	}
	require.Equal(t, expected_code, parsed_code)

}
