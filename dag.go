package BayesianNetwork

import (
	"bytes"
	"fmt"
	// "math"
	"sort"
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

	// validate that CPT has the correct dimensions
	// wrt. number of parents 
	if err := bn.validateCPTs(); err != nil {
		return nil, err
	}

	// index nodes in a breath first fashion
	bn.indexNetwork()

	return bn, nil
}

func (dag *BayesianNetwork) MarkovBlanketSampling(nodeName string, mapping map[string]string) map[float64]int {

	node := dag.GetNode(nodeName)
	if node == nil {
		return nil
	}
	// make sure graph is reset
	dag.Reset()

	// update the graph with all the observed values
	// from the mapping
	dag.UpdateGraph(mapping)

	results := map[float64]int{}

	for i := 0; i < 100; i++ {
		// fmt.Println(node)

		// ******* numerator
		fmt.Printf("Resetting node %v\n", node)
		node.SetSample("")
		// sample and set node of interest
		numerator := node.MarkovSample()
		fmt.Printf("Fresh Sample: %v\n", node)

		// fmt.Println(node)

		// fmt.Printf("  Numerator: %f\n", numerator)
		// fmt.Printf("1)Numerator: %f\n", numerator)
		// now sample the children given the sampled node of interest
		for _, childNode := range node.GetChildren() {
			numerator *= childNode.MarkovSample()
		}

		fmt.Printf("2)Numerator: %f\n", numerator)

		// ******** denominator
		denominator := 0.0
		for _, value := range []string{"T", "F"} {
			// set the value of node of interest to each value
			node.SetSample(value)
			prob := node.MarkovSample()
			// now sample the children given the sampled node of interest
			for _, childNode := range node.GetChildren() {
				prob *= childNode.MarkovSample()
			}
			denominator += prob
		}
		fmt.Printf("Denominator: %f\n", denominator)

		stat := numerator / denominator

		results[stat] += 1
	}
	return results
}

// func (dag *BayesianNetwork) GetMarkovBlanket(node BayesianNode) BayNodes {

// 	blanket := make(BayNodes, 0, 10)

// 	// add parent nodes
// 	blanket = append(blanket, node.GetParents()...)
// 	// add child nodes
// 	childNodes := node.GetChildren()
// 	blanket = append(blanket, childNodes...)

// 	// co-parents
// 	// extremely inefficient
// 	mm := map[int]BayesianNode{}
// 	for _, childNode := range childNodes {
// 		for _, parent := range childNode.GetParents() {
// 			if parent == node {
// 				continue
// 			}
// 			mm[parent.Id()] = parent
// 		}
// 	}
// 	for _, childNode := range mm {
// 		blanket = append(blanket, childNode)
// 	}

// 	return blanket
// }

// Given a truth-assignment for a number of nodes,
// update the nodes to reflect those values
// mapping example: map[string]string{ "X1":"F", X3:"T"}
// - reports an error if just one of the nodes does not exist
func (dag *BayesianNetwork) UpdateGraph(mapping map[string]string) error {

	for nodeName, value := range mapping {
		node := dag.nodes[nodeName]
		if node == nil {
			return fmt.Errorf("Node '%s' does not exist\n", nodeName)
		}

		if err := node.SetSample(value); err != nil {
			return err
		}
	}
	// sort.Sort(markovBlanket)
	return nil
}

// does a complete ancestral sampling of the network
func (dag *BayesianNetwork) AncestralSampling(nodeNames ...string) {
	// reset graph before running
	dag.Reset()

	// generate list of all the nodes 
	nodes, err := dag.GatherNodes(nodeNames)
	if err != nil {
		fmt.Println(err)
		return
	}

	sort.Sort(nodes)

	// for i := 0; i < max; {
	for _, node := range dag.nodeIndex {
		node.Sample()
	}
	// }
}

func (dag *BayesianNetwork) GatherNodes(nodeNames []string) (BayNodes, error) {

	nodes := make(BayNodes, 0, len(nodeNames))
	for _, nodeName := range nodeNames {
		node := dag.GetNode(nodeName)
		if node == nil {
			return nil, fmt.Errorf("Node %s not in network", nodeName)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (dag *BayesianNetwork) Reset() {
	for _, node := range dag.nodeIndex {
		node.Reset()
	}
}

// prints every node in the system seperately
// - also, print assignment value, if it has been
//   set.
func (dag *BayesianNetwork) PrintNetwork() string {
	if len(dag.nodeIndex) == 0 {
		return "[]"
	}

	var buffer bytes.Buffer

	buffer.WriteString(" ")
	for _, node := range dag.nodeIndex {
		buffer.WriteString(node.String())
		buffer.WriteString(" ")
	}

	return fmt.Sprintf("[%v]", buffer.String())
}

// joint probability of the network
func (dag *BayesianNetwork) JointProbability() float64 {
	// JointProbability(NodeNames ...string) float64 {

	// nodes := make(BayNodes, 0, len(NodeNames))
	// for _, name := range NodeNames {
	// 	nodes = append(nodes, dag.nodes[name])
	// }
	// // fmt.Println(nodes)
	// sort.Sort(nodes)
	// fmt.Println(nodes)

	p := 1.0
	for _, node := range dag.nodeIndex {
		p *= node.TruthProb()
	}

	return p
}

// validate every node in the system for invalid 
// conditional probability tables
func (dag *BayesianNetwork) validateCPTs() error {
	for _, node := range dag.nodes {
		if err := node.ValidateCPT(); err != nil {
			return err
		}
	}

	return nil
}

// index the graph in a breath-first fashion
// - guarantees that every parent has an index
//   that is larger than every one of their children
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
