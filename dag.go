package BayesianNetwork

import (
	"bytes"
	"fmt"
	// "sort"
	// "math"
	// "math/rand"
	// "time"
)

const (
	epsilon = 0.0001
)

type BayesianNetwork struct {
	nodes map[string]BayesianNode
	// indexOfNodes map[int]BayesianNode
	edges map[string][]string
	// list of nodes sorted by their placement
	// in their tree (breath first)
	// [1;len(nodes)]
	nodeIndex []BayesianNode
}

func NewBayesianNetwork(nodes ...BayesianNode) (*BayesianNetwork, error) {
	bn := &BayesianNetwork{
		nodes:     make(map[string]BayesianNode, len(nodes)),
		nodeIndex: make([]BayesianNode, 0, len(nodes)),
		edges:     make(map[string][]string),
	}

	// add nodes to network
	for _, node := range nodes {
		if err := bn.addNode(node); err != nil {
			return nil, err
		}
	}

	// generate connections
	for _, node := range nodes {
		bn.addConnections(node.GetParentNames(), node.Name())
	}

	if err := bn.validateCPTs(); err != nil {
		return nil, err
	}

	bn.indexNetwork()

	return bn, nil
}

// func (dag *BayesianNetwork) MarkovBlanket(nodeName string) float64 {

// }

func (dag *BayesianNetwork) AncestralSampling() {
	// to prevent recurring calls to ancestralsampling from reporting the 
	// same result
	dag.Reset()

	for _, node := range dag.nodeIndex {
		node.Sample()
	}
}

func (dag *BayesianNetwork) Reset() {
	for _, node := range dag.nodeIndex {
		node.Reset()
	}
}

func (dag *BayesianNetwork) JointProbability() float64 {
	// JointProbability(NodeNames ...string) float64 {

	// nodes := make(BayNodes, 0, len(NodeNames))
	// for _, name := range NodeNames {
	// 	nodes = append(nodes, dag.nodes[name])
	// }
	// // fmt.Println(nodes)
	// sort.Sort(nodes)
	// fmt.Println(nodes)

	jointProb := 1.0
	for _, node := range dag.nodeIndex {
		jointProb *= node.ProbTruth()
	}

	return jointProb
}

func (dag *BayesianNetwork) validateCPTs() error {
	for _, node := range dag.nodes {
		if err := node.ValidateCPT(); err != nil {
			return err
		}
	}

	return nil
}

func (dag *BayesianNetwork) indexNetwork() {

	roots := make(BayNodes, 0, 5)
	for _, node := range dag.nodes {
		if node.NumParents() != 0 {
			continue
			// children := append(children, node.GetChildren()...)
		}
		roots = append(roots, node)
	}

	id := 1
	for len(roots) > 0 {
		children := make(BayNodes, 0, 10)
		for _, node := range roots {
			if node.Id() != 0 {
				continue
			}
			node.SetId(id)
			dag.nodeIndex = append(dag.nodeIndex, node)
			id++
			ch := node.GetChildren()
			children = append(children, ch...)

		}
		// swap childnodes for parentNodes
		roots = children
	}
	// dag.nodeCount = id
}

func (dag *BayesianNetwork) addNode(node BayesianNode) error {
	// node := NewNode(name, cpt)
	if _, ok := dag.nodes[node.Name()]; ok == true {
		return fmt.Errorf("Duplicate nodeName: %s", node.Name())
	}

	dag.nodes[node.Name()] = node
	dag.edges[node.Name()] = make([]string, 0, 5)

	return nil
}

func (dag *BayesianNetwork) GetNode(name string) BayesianNode {
	return dag.nodes[name]
}

func (dag *BayesianNetwork) addConnections(parentNames []string, childName string) error {

	child, ok := dag.nodes[childName]
	if !ok {
		return fmt.Errorf("Child '%s' does not exist \n", childName)
	}

	for _, parentName := range parentNames {
		parent, ok := dag.nodes[parentName]
		if !ok {
			return fmt.Errorf("Parent '%s' does not exist \n", parentName)
		}
		parent.AddChild(child)
		child.AddParent(parent)
		for _, v := range dag.edges[parentName] {
			if child.Name() == v {
				return nil
			}
		}

		dag.edges[parent.Name()] = append(dag.edges[parent.Name()],
			child.Name())
	}

	return nil
}

func (dag *BayesianNetwork) GetNodes() BayNodes {
	nodes := make([]BayesianNode, 0, len(dag.nodes))

	for _, v := range dag.nodes {
		nodes = append(nodes, v)
	}
	return nodes
}

func (dag *BayesianNetwork) String() string {
	var buffer bytes.Buffer
	// buffer.WriteString(fmt.Sprintf("nodes: %d\n", dag.nodeCount))
	buffer.WriteString(fmt.Sprintf("nodeIndex: %d\n", dag.nodeIndex))
	for k, v := range dag.edges {
		buffer.WriteString(fmt.Sprintf("%s -> %v\n", k, v))
	}

	return buffer.String()
}
