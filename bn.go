// A DAG (Directed Acyclic Graph) implementation of a Bayesian Network enabling Ancestral Sampling and Gibbs Sampling on Binary Discrete Variables
package BayesianNetwork

import (
	"bytes"
	"fmt"
	// "math"
	"math/rand"
	// "sort"
	// "time"
)

type BayesianNetwork struct {
	// nodesName -> node-pointer map
	nodes map[string]*Node
	// connections between nodes
	edges map[string][]string
	// index of nodes sorted on id
	// - root-ids > child nodes etc..
	// [1;len(nodes)]
	nodeIndex BayNodes
}

// Creates a directed bayesian network from each node
func NewBayesianNetwork(nodes ...*Node) *BayesianNetwork {
	bn := &BayesianNetwork{
		nodes:     make(map[string]*Node, len(nodes)),
		nodeIndex: make([]*Node, 0, len(nodes)),
		edges:     make(map[string][]string),
	}

	// add nodes to network
	for _, node := range nodes {
		if err := bn.addNode(node); err != nil {
			panic(err)
		}
	}

	// generate connections
	for _, node := range nodes {
		bn.addConnections(node.GetParentNames(), node.Name())
	}

	// validate that CPT has the correct dimensions
	// wrt. number of parents
	if err := bn.validateCPTs(); err != nil {
		panic(err)
	}

	// index nodes in a breath first fashion
	bn.indexNetwork()

	return bn
}

// takes the node argument of interest (X5) and the truth-value
// mapping for the surrounding markov blanket
// it returns the map containing the frequencies of each of the inferred values:

// if X5 sample == false:

func (bn *BayesianNetwork) GibbsSampling(observations map[string]string, n, m int) StatMap {

	// only sample from the variables that
	// are not defined
	nodes_of_interest := make(BayNodes, 0, len(bn.nodeIndex)-len(observations))
	// just to be sure
	bn.Reset()
	// make sure no node is not assigned
	// either using the mapping or using a default of "F"
	if len(observations) == 0 {
		// no observations => sample from entire network
		nodes_of_interest = bn.nodeIndex[0:len(bn.nodeIndex)]
		// reset entire graph to false
		bn.ResetWithAssignment("F")
	} else {
		// update the graph with all the observed values
		err := bn.UpdateGraphValues(observations)
		if err != nil {
			panic(err)
		}

		// gather all the nodes of interest
		// and initialize them
		for _, node := range bn.nodeIndex {
			if node.GetAssignment() == "" {
				node.SetAssignment("F")
				nodes_of_interest = append(nodes_of_interest, node)
			}
		}
	}

	// initialize stat gathering
	ns := NewNetworkStat(bn)

	// run n times before we start registering statistics
	for i := 0; i < n; i++ {
		for _, xi := range nodes_of_interest {
			sample := bn.MarkovBlanketSample(xi)
			xi.SetAssignment(sample)
		}
	}

	// run m times while gathering stats
	for i := 0; i < m; i++ {
		for _, xi := range nodes_of_interest {
			sample := bn.MarkovBlanketSample(xi)
			xi.SetAssignment(sample)
		}
		// update stats
		ns.Update()
	}

	// clean up
	bn.Reset()

	return ns.GetStats()
}

func (bn *BayesianNetwork) MarkovBlanketSample(node *Node) string {
	// ******* numerator *******
	numerator := node.P()
	// now sample the children given the sampled node of interest
	for _, childNode := range node.GetChildren() {
		numerator *= childNode.P()
	}

	// ******* normalization *******
	Z := 0.0
	// set the value of node of interest to each value
	for _, cond := range []string{"T", "F"} {
		// assign truth value
		node.SetAssignment(cond)
		// sample the probability given the assignment
		sampleProb := node.SampleOnCondition(cond)

		// now sample the children given the sampled node of interest
		for _, childNode := range node.GetChildren() {
			sampleProb *= childNode.P()
		}

		Z += sampleProb
	}
	markovProb := numerator / Z

	random := rand.Float64()
	if random > markovProb {
		return "F"
	}
	return "T"

}

// Given a truth-assignment for a markov blanket,
// this method updates the nodes to reflect those values.
// mapping example: map[string]string{ "X1":"F", X3:"T"}
// - reports an error if just one of the nodes does not exist
func (bn *BayesianNetwork) UpdateGraphValues(mapping map[string]string) error {
	for nodeName, value := range mapping {
		node := bn.nodes[nodeName]
		if node == nil {
			return fmt.Errorf("Node '%s' does not exist in network\n\tmapping: %v\n\tnetwork: %v\n", nodeName, mapping, bn.nodeIndex)
		}
		node.SetAssignment(value)
	}
	return nil
}

// does a complete ancestral sampling of the network
func (bn *BayesianNetwork) AncestralSampling(n int) StatMap {
	// initialize stats gathering
	stat := NewNetworkStat(bn)
	for i := 0; i < n; i++ {
		for _, node := range bn.nodeIndex {
			// fmt.Printf("%s = %s\n", node.Name(), node.AssignmentValue())
			node.SetAssignment(node.Sample())
		}
		// upate stats
		stat.Update()
	}
	// cleanup
	bn.Reset()
	return stat.GetStats()
}

// Reset network after running a destructive method
func (bn *BayesianNetwork) Reset() {
	for _, node := range bn.nodeIndex {
		node.Reset()
	}
}

func (bn *BayesianNetwork) ResetWithAssignment(assignment string) {

	if assignment != "F" && assignment != "T" {
		panic(fmt.Sprintf("Invalid assignment: '%s' should be T or F", assignment))
	}

	for _, node := range bn.nodeIndex {
		node.SetAssignment(assignment)
	}
}

// prints every node in the system seperately
// - also, print assignment value, if it has been
//   set. meant for DEABUG
func (bn *BayesianNetwork) PrintNetwork() string {
	if len(bn.nodeIndex) == 0 {
		return "[]"
	}

	var buffer bytes.Buffer

	buffer.WriteString(" ")
	for _, node := range bn.nodeIndex {
		buffer.WriteString(node.String())
		buffer.WriteString(" ")
	}

	return fmt.Sprintf("[%v]", buffer.String())
}

// joint probability of the network
func (bn *BayesianNetwork) JointProbability() float64 {
	p := 1.0
	bn.ResetWithAssignment("T")
	for _, node := range bn.nodeIndex {
		p *= node.P()
	}
	return p
}

// validate every node in the system for invalid
// conditional probability tables
func (bn *BayesianNetwork) validateCPTs() error {
	for _, node := range bn.nodes {
		if err := node.ValidateCPT(); err != nil {
			return err
		}
	}

	return nil
}

// index the graph in a breath-first fashion
// - guarantees that every parent has an index
//   that is larger than every one of their children
func (bn *BayesianNetwork) indexNetwork() {

	roots := make(BayNodes, 0, 5)
	for _, node := range bn.nodes {
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
			node.setId(id)
			bn.nodeIndex = append(bn.nodeIndex, node)
			id++
			ch := node.GetChildren()
			children = append(children, ch...)

		}
		// swap childnodes for parentNodes
		roots = children
	}
}

func (bn *BayesianNetwork) addNode(node *Node) error {
	if _, ok := bn.nodes[node.Name()]; ok == true {
		return fmt.Errorf("Duplicate nodeName: %s", node.Name())
	}

	// add parent to child and visa versa
	bn.nodes[node.Name()] = node
	bn.edges[node.Name()] = make([]string, 0, 5)

	return nil
}

// returns the BayesianNode probided a valid name
func (bn *BayesianNetwork) GetNode(name string) *Node {
	return bn.nodes[name]
}

func (bn *BayesianNetwork) addConnections(parentNames []string, childName string) error {

	child, ok := bn.nodes[childName]
	if !ok {
		return fmt.Errorf("Child '%s' does not exist \n", childName)
	}

	for _, parentName := range parentNames {
		parent, ok := bn.nodes[parentName]
		if !ok {
			return fmt.Errorf("Parent '%s' does not exist \n", parentName)
		}
		parent.AddChild(child)
		child.AddParent(parent)
		for _, v := range bn.edges[parentName] {
			if child.Name() == v {
				return nil
			}
		}

		bn.edges[parent.Name()] = append(bn.edges[parent.Name()],
			child.Name())
	}

	return nil
}

func (bn *BayesianNetwork) NodeCount() int {
	return len(bn.nodeIndex)
}

// returns all the nodes in the graph
func (bn *BayesianNetwork) GetNodes() BayNodes {
	return bn.nodeIndex
}

func (bn *BayesianNetwork) String() string {
	var buffer bytes.Buffer
	// buffer.WriteString(fmt.Sprintf("nodes: %d\n", bn.nodeCount))
	// buffer.WriteString(fmt.Sprintf("nodeIndex: %d\n", bn.nodeIndex))
	for _, node := range bn.nodeIndex {
		buffer.WriteString(fmt.Sprintf("%s ", node.AssignmentString()))
	}

	buffer.WriteString("\n")

	return buffer.String()
}
