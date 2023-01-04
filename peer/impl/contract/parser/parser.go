package parser

// lib for building the lexer
import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"       // lib for building the parser
	"github.com/alecthomas/participle/v2/lexer" // lib for building the lexer
)

// This module parse the smart contract code
// from plain text to internal ata structure

// Smart contract Specification:
// 2 kind of roles: proposer (publisher), acceptor (finisher)
// 3 kind of primitives: Assume, If, Submit Action, Transaction Action (supported API of blockchain)
// Expression supported: number, role.attribute, role.keyID.attribute, role.action(params) & 1 line for 1 expr
// Condition supported: role.(keyID).attribute >/</=/>=/<=/!= number

// How to execute:

// Example:
// ASSUME publisher.budget > 0
// IF finisher.key0.hash == finisher.hash0  THEN
// 		finisher.submit("123", "2ab3dvjhkj1k") # publisherID, key0's corresponding cracked key
// 		publisher.transfer("456", 20) # finisherID, money. This money should be reduce from publisher's budget

// Lexer for the contract code. Rules are specified with regexp.
// Need to tokenize to the minimum unit be
// nil meaning that the lexer is simple/stateless.
var SMTLexer = lexer.MustSimple([]lexer.Rule{
	{`Keyword`, `(?i)\b(ASSUME|IF|THEN)\b`, nil}, // not case sensitive
	{`Float`, `\d+(?:\.\d+)?`, nil},
	{`String`, `"(.*?)"`, nil},           // quoted string tokens
	{`Operator`, `==|!=|>=|<=|>|<`, nil}, // only comparison operator
	{`Ident`, `[a-zA-Z][a-zA-Z0-9_]*`, nil},
	{"comment", `[#;][^\n]*`, nil},
	{"Punct", `[(),\.]`, nil},
	{"whitespace", `\s+`, nil},
})

// Specify participle grammar for contract code
type Code struct {
	Assumptions []*Assumption `@@*` // 0 or more assumptions
	IfClauses   []*IfClause   `@@*` // 0 or more If clauses
}

type Assumption struct { // each assumption specifies a condition
	Condition Condition `( "ASSUME" @@ )`
}

type IfClause struct { // one condition with one or more actions
	ConditionObjObj ConditionObjObj `"IF" @@`
	Actions         []*Action       `( "THEN" @@+ )`
}

type Condition struct {
	Object   Object `@@`
	Operator string `@Operator`
	Value    Value  `@@`
}

type ConditionObjObj struct {
	Object_1 Object `@@`
	Operator string `@Operator`
	Object_2 Object `@@`
}

type Value struct {
	String *string  `@String`
	Number *float64 `| @Float`
}

type Object struct {
	Role   string   `( @"publisher" | @"finisher" )`
	Fields []*Field `@@*`
}

type Field struct {
	Name string `"." @Ident`
}

type Action struct {
	Role   string   ` ( @"publisher" | @"finisher" )`
	Action string   ` ( "." (@"submit" | @"transfer") )` // TODO: Action to be all blockchain primitives supported
	Params []*Value `( "(" ( @@ ( "," @@ )* )? ")" )`
}

// Parsing for the contract code
func Parse(plainCode string) (Code, error) {
	ast := &Code{}
	var codeParser = participle.MustBuild(&Code{},
		participle.Lexer(SMTLexer),
		participle.Unquote("String"),
	)
	err := codeParser.ParseString("", plainCode, ast)
	return *ast, err
}

// Provides toString for each type of elements
func (v Condition) ToString() string {
	out := new(strings.Builder)
	obj := v.Object.ToString()
	operator := v.Operator
	val := v.Value.ToString()
	out.WriteString("[Condition] " + obj + " " + operator + " " + val)

	return out.String()
}

func (v ConditionObjObj) ToString() string {
	out := new(strings.Builder)
	obj1 := v.Object_1.ToString()
	operator := v.Operator
	obj2 := v.Object_2.ToString()
	out.WriteString("[Condition] " + obj1 + " " + operator + " " + obj2)

	return out.String()
}

func (v Object) ToString() string {
	out := new(strings.Builder)
	out.WriteString(v.Role)
	for _, field := range v.Fields {
		out.WriteString("." + field.ToString())
	}

	return out.String()
}

func (v Action) ToString() string {
	out := new(strings.Builder)
	out.WriteString("[Action] " + v.Role + " " + v.Action + " (")
	for _, p := range v.Params {
		out.WriteString(" " + p.ToString())
	}
	out.WriteString(" )")

	return out.String()
}

func (v Value) ToString() string {
	if v.String != nil {
		return *v.String
	} else {
		return fmt.Sprintf("%f", *v.Number)
	}
}

func (v Field) ToString() string {
	return v.Name
}
