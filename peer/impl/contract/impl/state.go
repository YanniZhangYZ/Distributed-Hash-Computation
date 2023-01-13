package impl

import (
	"strconv"

	"github.com/disiqueira/gotree" // lib for print tree structure in terminal
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

// This file is used for recording the state tree of AST
// by doing this we can keep track of the execution state of contract
// The execution state will be displayed everytime the contract is executed
// Note that this project only consider Assumption, Condition, Ifclause, Action

type StateNode struct {
	NodeID           int
	NodeType         string
	HashValidGrammar bool // indicate whether the condition of this node is validate
	IsExecuted       bool // indicate whether this node is executed
	Children         []*StateNode
}

// add child to a given node
func (s *StateNode) addChild(child *StateNode) {
	s.Children = append(s.Children, child)
}

// mark the node as executed
// this function will be called
// when everytime assumption is checked and actions are gathered
func (s *StateNode) setExecuted() {
	s.IsExecuted = true
}

// get the execution state of this node
// this function will be used in when displaying the contract state tree
func (s *StateNode) getExecuted() bool {
	return s.IsExecuted
}

// mark the node as executed
// this function will be called
// when everytime the condition of assumption and if-then clause are checked
// for those contions are true, the node will be set as valid
func (s *StateNode) setValid() {
	s.HashValidGrammar = true
}

// get the validity of this node
// this function will be used in when displaying the contract state tree
func (s *StateNode) getValid() bool {
	return s.HashValidGrammar
}

// get the node ID
func (s *StateNode) getNodeID() int {
	return s.NodeID
}

// Construct corresponding state tree from the given code AST
// Due to the grammar we specified in this project,
// The code AST is very structed( and to some extent known in advance).
// Therefore We don't need to recursively traverse the AST.
// Each state node is assigned with an ID
func BuildStateTree(ast *parser.Code) *StateNode {
	id := 0
	root := StateNode{id, "Code", false, false, []*StateNode{}}
	id++

	// We first process the Assumptions of the contract
	for i := 0; i < len(ast.Assumptions); i++ {
		assumptionNode := StateNode{id, "Assumption", false, false, []*StateNode{}}
		id++
		// in assumption, we accept variables with only one attribute
		conditionNode := StateNode{id, "ConditionOneAttribute", false, false, []*StateNode{}}
		id++
		assumptionNode.addChild(&conditionNode)
		root.addChild(&assumptionNode)
	}

	// Then we process the if-then
	for _, ifclause := range ast.IfClauses {
		ifNode := StateNode{id, "If", false, false, []*StateNode{}}
		id++
		// in if-then, we accept variables with exactly two attributes
		conditionNode := StateNode{id, "ConditionTwoAttribute", false, false, []*StateNode{}}
		id++
		ifNode.addChild(&conditionNode)

		// Process the actions
		for i := 0; i < len(ifclause.Actions); i++ {
			actionNode := StateNode{id, "Action", false, false, []*StateNode{}}
			id++
			ifNode.addChild(&actionNode)
		}
		root.addChild(&ifNode)
	}

	return &root
}

// show the execution state of AST
// this function will be called everytime
// when assumptions is checked and actions are gathered
func PrintStateAST(ast parser.Code, stateAST *StateNode) string {
	root := gotree.New("📝📝📝 Contract Execution State")
	boolHelper := map[bool]string{true: "✅", false: "❎"}

	assumeNode := root.Add("Assumption")

	// We first process Assumptions in the contract
	for i, a := range ast.Assumptions {
		id := strconv.Itoa(stateAST.Children[i].NodeID)
		c := a.Condition.ToString()
		exeState := boolHelper[stateAST.Children[i].IsExecuted]
		validState := boolHelper[stateAST.Children[0].HashValidGrammar]
		// here are two possibilities
		// 1: [Executed✅][Valid✅] This means the node is executed and the condition is valid
		// 2: [Executed✅][Valid❎] This means the node is executed but the condition is wrong
		assumeNode.Add(id + ": " + c + " [Executed" + exeState + "]" + " [Valid" + validState + "]")
	}

	// Then we process the if-then
	for i, ifclause := range ast.IfClauses {
		ifNode := root.Add("If Clause")
		ifStateNode := stateAST.Children[len(ast.Assumptions)+i]

		id := strconv.Itoa(ifStateNode.Children[0].NodeID)
		c := ifclause.Condition.ToString()
		exeState := boolHelper[ifStateNode.Children[0].IsExecuted]
		validState := boolHelper[ifStateNode.Children[0].HashValidGrammar]
		// here are two possibilities
		// 1: [Executed✅][Valid✅] This means the node is executed and the condition is valid
		// 2: [Executed✅][Valid❎] This means the node is executed but the condition is wrong
		ifNode.Add(id + ": " + c + " [Executed" + exeState + "]" + " [Valid" + validState + "]")

		for j, action := range ifclause.Actions {
			id = strconv.Itoa(ifStateNode.Children[j+1].NodeID)
			act := action.ToString()
			exeState = boolHelper[ifStateNode.Children[j+1].IsExecuted]
			ifNode.Add(
				id + ": " + act + " [Executed" + exeState + "]")
		}
	}

	return root.Print()
}

// show the code AST
func PrintCodeAST(ast parser.Code) string {
	root := gotree.New("CodeAST")

	assumptionNode := root.Add("Assumption")
	for _, a := range ast.Assumptions {
		assumptionNode.Add(a.Condition.ToString())
	}

	for _, ifClause := range ast.IfClauses {
		ifNode := root.Add("If Clause")
		ifNode.Add(ifClause.Condition.ToString())
		for _, action := range ifClause.Actions {
			ifNode.Add(action.ToString())
		}
	}
	return root.Print()
}
