package BayesianNetwork

import (
	"bytes"
	"fmt"
	// "math"
	"sort"
	// "math/rand"
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
func NewBayesianNetwork(nodes ...*Node) (*BayesianNetwork, error) {
	bn := &BayesianNetwork{
		nodes:     make(map[string]*Node, len(nodes)),
		nodeIndex: make([]*Node, 0, len(nodes)),
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

// takes the node argument of interest (X5) and the truth-value
// mapping for the surrounding markov blanket
// it returns the map containing the frequencies of each of the inferred values:
func (bn *BayesianNetwork) MarkovBlanketSampling(nodeName string, mapping map[string]string, n int) (map[float64]float64, error) {

	node := bn.GetNode(nodeName)
	if node == nil {
		return nil, fmt.Errorf("Invalid nodeName: %s", nodeName)
	}
	// make sure graph is reset
	bn.Reset()

	// update the graph with all the observed values
	// from the mapping
	bn.UpdateGraphValues(mapping)

	if node.AssignmentValue() != "" {
		return nil, fmt.Errorf("target node '%s' cannot be defined in the mapping", node.Name())
	}

	results := map[float64]float64{}

	for i := 0; i < n; i++ {

		// ******* numerator *******
		// reset node of interest
		node.SetAssignmentValue("")
		// sample and set node of interest
		_, numerator, err := node.Sample()
		if err != nil {
			return nil, err
		}
		// now sample the children given the sampled node of interest
		for _, childNode := range node.GetChildren() {
			_, prob, err := childNode.Sample()
			if err != nil {
				return nil, err
			}
			numerator *= prob
		}

		// ******** denominator *******
		denominator := 0.0
		for _, assign := range []string{"T", "F"} {
			// set the value of node of interest to each value
			node.SetAssignmentValue(assign)
			_, prob, err := node.Sample()
			if err != nil {
				return nil, err
			}
			// now sample the children given the sampled node of interest
			for _, childNode := range node.GetChildren() {
				_, prob, err := childNode.Sample()
				if err != nil {
					return nil, err
				}
				prob *= prob
			}
			denominator += prob
		}

		stat := numerator / denominator

		results[stat] += 1.0
	}

	for key, val := range results {
		results[key] = val / float64(n)
	}

	return results, nil
}

// Given a truth-assignment for a markov blanket,
// this method updates the nodes to reflect those values.
// mapping example: map[string]string{ "X1":"F", X3:"T"}
// - reports an error if just one of the nodes does not exist
func (bn *BayesianNetwork) UpdateGraphValues(mapping map[string]string) error {

	for nodeName, value := range mapping {
		node := bn.nodes[nodeName]
		if node == nil {
			return fmt.Errorf("Node '%s' does not exist\n", nodeName)
		}
		node.SetAssignmentValue(value)
	}

	return nil
}

// does a complete ancestral sampling of the network
func (bn *BayesianNetwork) AncestralSampling(mapping map[string]string, n int) (map[string][]float64, error) {
	// reset graph before running
	bn.Reset()
	// update graph from mapping
	err := bn.UpdateGraphValues(mapping)
	if err != nil {
		return nil, err
	}
	// backup the state so we can reset it afterwards
	backup := make([]string, len(bn.nodeIndex))
	for i, node := range bn.nodeIndex {
		backup[i] = node.AssignmentValue()
	}
	// initialize the stats gathering
	stat := NewStat(bn)

	for i := 0; i < n; i++ {

		for _, node := range bn.nodeIndex {
			node.Sample()
		}
		// update stats
		stat.Update()
		// reset graph with assigned values
		for j, node := range bn.nodeIndex {
			node.SetAssignmentValue(backup[j])
		}
	}
	// cleanup
	bn.Reset()
	return stat.GetStats(), nil
}

func (bn *BayesianNetwork) GatherNodes(nodeNames []string) (BayNodes, error) {

	nodes := make(BayNodes, 0, len(nodeNames))
	for _, nodeName := range nodeNames {
		node := bn.GetNode(nodeName)
		if node == nil {
			return nil, fmt.Errorf("Node %s not in network", nodeName)
		}
		nodes = append(nodes, node)
	}

	// sort nodes by index value
	sort.Sort(nodes)

	return nodes, nil
}

// Reset network after runing a destructive method
func (bn *BayesianNetwork) Reset() {
	for _, node := range bn.nodeIndex {
		node.Reset()
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
func (bn *BayesianNetwork) JointProbability() (float64, error) {

	bn.Reset()

	p := 1.0
	for _, node := range bn.nodeIndex {
		prob, err := node.SampleWithAssignment("T")
		if err != nil {
			return 0.0, err
		}
		p *= prob
	}
	return p, nil
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
	buffer.WriteString(fmt.Sprintf("nodeIndex: %d\n", bn.nodeIndex))
	for k, v := range bn.edges {
		buffer.WriteString(fmt.Sprintf("%s -> %v\n", k, v))
	}

	return buffer.String()
}
