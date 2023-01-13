package impl

import (
	"strconv"

	"github.com/disiqueira/gotree" // lib for print tree structure in terminal
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

// This file is used for recording the state tree of AST
// by doing this we can keep track of the execution state of contract
// Note that this project only consider Assumption, Condition, Ifclause, Action

type StateNode struct {
	NodeID           int
	NodeType         string
	HashValidGrammar bool // for Assumption, Condition, Ifclause
	IsExecuted       bool // for Action
	Children         []*StateNode
}

func (s *StateNode) addChild(n *StateNode) {
	s.Children = append(s.Children, n)
}

func (s *StateNode) setExecuted() {
	s.IsExecuted = true
}

func (s *StateNode) getExecuted() bool {
	return s.IsExecuted
}

func (s *StateNode) setValid() {
	s.HashValidGrammar = true
}

func (s *StateNode) getValid() bool {
	return s.HashValidGrammar
}

func (s *StateNode) getNodeID() int {
	return s.NodeID
}

// Construct corresponding state tree from the given code AST
// Due to the grammar we specified in this project,
// The code AST is very structed( and to some extent known in advance).
// Therefore We don't need to traverse the AST recursively.
// Each state node is assigned with an ID
func BuildStateTree(ast *parser.Code) *StateNode {
	id := 0
	root := StateNode{id, "Code", false, false, []*StateNode{}}
	id++

	// Process assumptions state
	for i := 0; i < len(ast.Assumptions); i++ {
		assumptionNode := StateNode{id, "Assumption", false, false, []*StateNode{}}
		id++
		conditionNode := StateNode{id, "Condition", false, false, []*StateNode{}}
		id++
		assumptionNode.addChild(&conditionNode)
		root.addChild(&assumptionNode)
	}

	// Process if clauses state
	for _, ifclause := range ast.IfClauses {
		ifNode := StateNode{id, "If", false, false, []*StateNode{}}
		id++
		conditionObjObjNode := StateNode{id, "ConditionObjObj", false, false, []*StateNode{}}
		id++
		ifNode.addChild(&conditionObjObjNode)

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
func PrintStateAST(ast parser.Code, stateAST *StateNode) string {
	root := gotree.New("📝📝📝 Contract Execution State")
	boolHelper := map[bool]string{true: "✅", false: "❎"}

	assumeNode := root.Add("Assumption")

	for i, a := range ast.Assumptions {
		id := strconv.Itoa(stateAST.Children[i].NodeID)
		c := a.Condition.ToString()
		exeState := boolHelper[stateAST.Children[i].IsExecuted]
		validState := boolHelper[stateAST.Children[0].HashValidGrammar]
		assumeNode.Add(id + ": " + c + " [Executed" + exeState + "]" + " [Valid" + validState + "]")
	}

	for i, ifclause := range ast.IfClauses {
		ifNode := root.Add("If Clause")
		ifStateNode := stateAST.Children[len(ast.Assumptions)+i]

		id := strconv.Itoa(ifStateNode.Children[0].NodeID)
		c := ifclause.Condition.ToString()
		exeState := boolHelper[ifStateNode.Children[0].IsExecuted]
		validState := boolHelper[ifStateNode.Children[0].HashValidGrammar]
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

// DisplayAST displays the code AST, convenient for debug
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
