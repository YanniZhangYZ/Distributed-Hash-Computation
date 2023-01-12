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
	nodeID           int
	nodeType         string
	hashValidGrammar bool // for Assumption, Condition, Ifclause
	isExecuted       bool // for Action
	children         []*StateNode
}

func (s *StateNode) addChild(n *StateNode) {
	s.children = append(s.children, n)
}

func (s *StateNode) setExecuted() {
	s.isExecuted = true
}

func (s *StateNode) getExecuted() bool {
	return s.isExecuted
}

func (s *StateNode) setValid() {
	s.hashValidGrammar = true
}

func (s *StateNode) getValid() bool {
	return s.hashValidGrammar
}

func (s *StateNode) getNodeID() int {
	return s.nodeID
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
func GetStateAST(ast parser.Code, stateAST *StateNode) string {
	root := gotree.New("State")
	boolHelper := map[bool]string{true: "T", false: "F"}

	assumeNode := root.Add("Assumption")

	for i, a := range ast.Assumptions {
		id := strconv.Itoa(stateAST.children[i].nodeID)
		c := a.Condition.ToString()
		exeState := boolHelper[stateAST.children[i].isExecuted]
		assumeNode.Add(id + ": " + c + " [" + exeState + "]")
	}

	for i, ifclause := range ast.IfClauses {
		ifNode := root.Add("If Clause")
		ifStateNode := stateAST.children[len(ast.Assumptions)+i]

		id := strconv.Itoa(ifStateNode.children[0].nodeID)
		c := ifclause.Condition.ToString()
		exeState := boolHelper[ifStateNode.children[0].isExecuted]
		ifNode.Add(id + ": " + c + " [" + exeState + "]")

		for j, action := range ifclause.Actions {
			id = strconv.Itoa(ifStateNode.children[j+1].nodeID)
			act := action.ToString()
			exeState = boolHelper[ifStateNode.children[j+1].isExecuted]
			ifNode.Add(
				id + ": " + act + " [" + exeState + "]")
		}
	}

	return root.Print()
}
