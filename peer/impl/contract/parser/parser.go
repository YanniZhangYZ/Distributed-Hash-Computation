package parser

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// This file parse the smart contract code
// from plain text to internal ata structure

// Here we define the smart contract grammar as follow:
// 2 kind of roles:
//	 publisher and finisher
// 3 kind of primitives:
//	 Assume, If, Submit Action, Transaction Action (supported API of blockchain)
// Expression supported:
//   number, role.attribute, role.attribute.attribute, role.action(params).
//   Note: one line for one expression
// Condition supported:
// 	role.attribute >, <, =, >=, <=, != number
// 	role.attribute >, <, =, >=, <=, != role.attribute

// How to execute:
// Example:
// ASSUME publisher.budget > 0
// IF finisher.key0.hash == finisher.hash0  THEN
//      # to submit, finisher needs publisherID and key0's corresponding cracked key
// 		finisher.submit(publisherID, finisher.key0)
// 		# to do the transaction, publisher needs finisherID and amount of money.
//      # This money should be reduce from publisher's budget
// 		smartAccount.transfer(finisherID, amount of money)

// ASSUME smartAccount.balance > 10
// IF finisher.crackedPwd.hash == "abcuusljfwpe1npo30dmks3e"  THEN
// 		smartAccount.transfer(finisherID, amount of money)

// Lexer for the contract code. Rules are specified with regexp.
// Need to tokenize to the minimum unit be
// nil meaning that the lexer is simple or stateless.

var ContractLexer = lexer.MustSimple([]lexer.Rule{
	{Name: `Keyword`, Pattern: `(?i)\b(ASSUME|IF|THEN)\b`, Action: nil}, // not case sensitive
	{Name: `Int64`, Pattern: `\d+(?:\.\d+)?`, Action: nil},
	{Name: `String`, Pattern: `"(.*?)"`, Action: nil},           // quoted string tokens
	{Name: `Operator`, Pattern: `==|!=|>=|<=|>|<`, Action: nil}, // only comparison operator
	{Name: `Ident`, Pattern: `[a-zA-Z][a-zA-Z0-9_]*`, Action: nil},
	{Name: "comment", Pattern: `[#;][^\n]*`, Action: nil},
	{Name: "Punct", Pattern: `[(),\.]`, Action: nil},
	{Name: "whitespace", Pattern: `\s+`, Action: nil},
})

//Below we specify the grammar for smart contract

// Code is constructed with assuptions and if clauses
type Code struct {
	Assumptions []*Assumption `@@*` // 0 or more assumptions
	IfClauses   []*IfClause   `@@*` // 0 or more If clauses
}

// Assuption is defined as ASSUME + condition
type Assumption struct { // each assumption specifies a condition
	Condition Condition `( "ASSUME" @@ )`
}

// If clause is defined as
// comaprison between obj and obj + actions to be executed in the clause
type IfClause struct { // one condition with one or more actions
	Condition Condition `"IF" @@`
	Actions   []*Action `( "THEN" @@+ )`
}

// this condition is for comparison between obj and value
type Condition struct {
	Object   Object `@@`
	Operator string `@Operator`
	Value    Value  `@@`
}

// this condition is for comparison between obj and obj
type ConditionObjObj struct {
	Object1  Object `@@`
	Operator string `@Operator`
	Object2  Object `@@`
}

// a value can be either a string or a float.
// The current smart contract only support these two types.
type Value struct {
	String *string `@String`
	Number *int64  `| @Int64`
}

// the object is consist of a role and fields
// e.g. publisher.budget
type Object struct {
	Role   string   `( @"publisher" | @"finisher" | @"smartAccount")`
	Fields []*Field `@@*`
}

// field corresponds to the attribute of the role
type Field struct {
	Name string `"." @Ident`
}

// Action is conducted by specific role with specific action and parameters it needs
type Action struct {
	// Role   string   ` ( @"publisher" | @"finisher" | @"smartAccount")`
	Role string ` (@"smartAccount")`

	// Action string   ` ( "." (@"submit" | @"transfer") )`
	Action string   ` ( "." (@"transfer") )`
	Params []*Value `( "(" ( @@ ( "," @@ )* )? ")" )`
}

// Parsing the contract code and return the code AST
func BuildCodeAST(plainText string) (Code, error) {
	ast := &Code{}
	var codeParser = GetCodeParser()
	err := codeParser.ParseString("", plainText, ast)
	return *ast, err
}

// return the code parser with the grammar we defined above
func GetCodeParser() *participle.Parser {
	return participle.MustBuild(&Code{},
		participle.Lexer(ContractLexer),
		participle.Unquote("String"),
	)
}

// toString for condition
func (c Condition) ToString() string {

	obj := c.Object.ToString()
	operator := c.Operator
	val := c.Value.ToString()
	str := new(strings.Builder)
	str.WriteString("[Condition] " + obj + " " + operator + " " + val)

	return str.String()
}

// toString for comparison between obj and obj
func (coo ConditionObjObj) ToString() string {
	obj1 := coo.Object1.ToString()
	operator := coo.Operator
	obj2 := coo.Object2.ToString()
	str := new(strings.Builder)
	str.WriteString("[ConditionObjObj] " + obj1 + " " + operator + " " + obj2)

	return str.String()
}

// toString for Object
func (o Object) ToString() string {
	str := new(strings.Builder)
	str.WriteString(o.Role)
	for _, f := range o.Fields {
		str.WriteString("." + f.ToString())
	}

	return str.String()
}

// toString for action
func (a Action) ToString() string {
	str := new(strings.Builder)
	str.WriteString("[Action] " + a.Role + " " + a.Action + " (")
	for _, param := range a.Params {
		str.WriteString(" " + param.ToString())
	}
	str.WriteString(" )")

	return str.String()
}

// toString for value
func (v Value) ToString() string {
	if v.String != nil {
		return *v.String
	}
	return fmt.Sprintf("%f", *v.Number)
}

// toString for role's attribut
func (f Field) ToString() string {
	return f.Name
}
