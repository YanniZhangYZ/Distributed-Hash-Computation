package impl

import (
	"strconv"

	"github.com/disiqueira/gotree" // lib for print tree structure in terminal
	"go.dedis.ch/cs438/peer/impl/contract/parser"
)

// Maintain the state tree of AST to keep track of the execution state of contract
// In this project we only focus on Assumption, Condition, Ifclause, Action

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

// Construct corresponding state tree, given the code AST
// The structure of AST is rather predictable, so we don't need to recursively traverse
// We assign a id to each node, so it will be easier to retrieve & manipulate with node id
func constructStateTree(ast *parser.Code) *StateNode {
	id := 0
	root := StateNode{id, "code", false, false, []*StateNode{}}
	id++

	// Process assumptions state
	for i := 0; i < len(ast.Assumptions); i++ {
		assumption_node := StateNode{id, "assumption", false, false, []*StateNode{}}
		id++
		condition_node := StateNode{id, "condition", false, false, []*StateNode{}}
		id++
		assumption_node.addChild(&condition_node)
		root.addChild(&assumption_node)
	}

	// Process if clauses state
	for _, ifclause := range ast.IfClauses {
		if_node := StateNode{id, "if", false, false, []*StateNode{}}
		id++
		conditionObjObj_node := StateNode{id, "conditionObjObj", false, false, []*StateNode{}}
		id++
		if_node.addChild(&conditionObjObj_node)
		for i := 0; i < len(ifclause.Actions); i++ {
			action_node := StateNode{id, "action", false, false, []*StateNode{}}
			id++
			if_node.addChild(&action_node)
		}
		root.addChild(&if_node)
	}

	return &root
}

// show the execution state of AST
func getStateAST(ast parser.Code, stateAST *StateNode) string {
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
		c := ifclause.ConditionObjObj.ToString()
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
