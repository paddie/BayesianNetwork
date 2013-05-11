package BayesianNetwork

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Node struct {
	// id of a parent node must be larger than 
	// id of every one of their childnodes
	id int
	// name of the random variable
	name string
	// the list of parentNames is in the 
	// same order as one would use to
	// lookup in the CPT
	parentNames []string
	// References to child and parent nodes
	childIds  BayNodes
	parentIds BayNodes
	// truth assignment "T"/"F"
	// assignment == "" <=> unsampled
	assignment string
	// conditional probability table
	// takes a string-key consisting of truth
	// assignments with len(key) == len(parentIds)
	// such as "TTFF"
	// indicating that parent 1-2 have
	// truth assignments "T", and parents
	// 3-4 have truth assignments "F"
	// - the value returned in a CPT lookup
	//   is always the "T" value.
	cpt map[string]float64
	// after a node has been sampled
	// this will contain the 
	// key generated from the truth assignments 
	// of the parents
	// keyCache string
	// Probability cache from the lookup in the 
	// CPT given the keyCache
	// probabilityCache float64
}

// Generate a root node.
// The CPT is initialized with "T"=dist,
// and the "F" = 1-dist
func NewRootNode(name string, dist float64) *Node {
	node := &Node{
		name:     name,
		childIds: make([]*Node, 0, 4),
		cpt:      map[string]float64{"T": dist, "F": 1.0 - dist},
	}
	return node
}

// 
func NewNode(name string, parents []string, dist map[string]float64) *Node {
	node := &Node{
		name:        name,
		parentNames: parents,
		parentIds:   make([]*Node, 0, 4),
		childIds:    make([]*Node, 0, 4),
		cpt:         dist,
	}
	return node
}

func (self *Node) computeKey() (string, error) {
	// rootnode - always return the truth value
	if self.NumParents() == 0 {
		// self.keyCache = "T"
		return "T", nil
	}

	// generate key from parent assignments
	var buffer bytes.Buffer
	for _, id := range self.parentIds {
		av := id.AssignmentValue()
		// one of the parents have not been sampled
		// - error because this should never happen
		//   if we sort on the index
		if av == "" {
			return "", fmt.Errorf("%s: unset assignment value", id.Name())
		}
		buffer.WriteString(av)
	}
	// update cache
	// self.keyCache = buffer.String()
	return buffer.String(), nil
}

// Generate the CPT lookup-key from parent assignment variables
// - if this has already been generated, returned cached value
func (self *Node) GetCPTProbability() (float64, error) {
	// cache key (can be very expensive to generate)
	// if self.keyCache == "" {
	// 	if err := self.computeKey(); err != nil {
	// 		return 0.0, err
	// 	}
	// }
	key, err := self.computeKey()
	if err != nil {
		return 0.0, err
	}

	if prob, ok := self.cpt[key]; ok == true {
		return prob, nil
	}

	return 0.0, fmt.Errorf("'%s' not a valid node key for CPT: %v", key, self.cpt)
}

// Sample returns the assignment T/F and the probability
// of the node given the parent nodes
// - can return an error if the assignments of the parents
//   are invalid keys in the CPT
// - if the assignemnt of a node without a sampling 
//   as in the markov blanket =>
//   self.probabilityCache == 0.0 because it hasn't
//   been sampled.
func (self *Node) Sample() (string, float64, error) {
	// if sample has already been calculated
	// the values will have been cached
	cptProb, err := self.GetCPTProbability()
	if err != nil {
		return "", 0.0, err
	}

	// if node hasn't been defined, we generate 
	// an assignment for the node
	if self.assignment == "" {
		// generate random float64 for sampling
		random := rand.Float64()
		if random > cptProb {
			self.assignment = "F"
			return self.assignment, 1 - cptProb, nil
		}

		self.assignment = "T"
		return self.assignment, cptProb, nil
	}
	// if assignment has already been defined
	// we might be in a markov blanket
	if self.assignment == "F" {
		return self.assignment, 1 - cptProb, nil
	}

	return self.assignment, cptProb, nil
}

// Provided that the nodes parents have been sampled =>
// have truth assignment \in {T/F}
// returns the probability of either the true or false assignment
func (self *Node) SampleWithAssignment(assignment string) (float64, error) {

	if assignment != "F" && assignment != "T" {
		panic(fmt.Sprintf("SampleWithAssignment: invalid assignment: '%s'", assignment))
	}

	self.assignment = assignment

	_, prob, err := self.Sample()
	if err != nil {
		return 0.0, err
	}

	return prob, nil
}

func (self *Node) NumParents() int {
	return len(self.parentIds)
}

func (self *Node) NumChildren() int {
	return len(self.childIds)
}

func (self *Node) GetChildren() BayNodes {
	return self.childIds
}

func (self *Node) GetParents() BayNodes {
	return self.parentIds
}

func (self *Node) Name() string {
	return self.name
}

func (self *Node) Id() int {
	return self.id
}

func (self *Node) setId(i int) {
	self.id = i
}

func (self *Node) GetParentNames() []string {
	// names := make([]string, 0, len(self.parentIds))
	// for _, node := range self.parentIds {
	// 	names = append(names, node.Name())
	// }

	return self.parentNames
}

func (self *Node) AddChild(child *Node) {

	for _, c := range self.childIds {
		if c == child {
			return
		}
	}

	self.childIds = append(self.childIds, child)
}

func (self *Node) AddParent(parent *Node) error {

	for _, p := range self.parentIds {
		if p == parent {
			return nil
		}
	}
	self.parentIds = append(self.parentIds, parent)

	return nil
}

func (self *Node) String() string {

	if self.assignment != "" {
		prob, _ := self.GetCPTProbability()
		return fmt.Sprintf("%d: %s='%s' p=%f (%v)\n\tparents:  %v\n\tchildren: %v\n",
			self.id, self.name, self.assignment, prob, self.cpt, self.parentIds, self.childIds)
	}

	return fmt.Sprintf("%s(%d): (%v)\n\tparents:  %v\n\tchildren: %v\n",
		self.name, self.id, self.cpt, self.parentIds, self.childIds)
}

func (self *Node) ValidateParents(parents []string) bool {
	if len(parents) != len(self.parentIds) {
		return false
	}

	for i, parentName := range parents {
		if self.parentIds[i].Name() != parentName {
			return false
		}
	}
	return true
}

// func (self *Node) ResetKey() {
// 	self.keyCache = ""
// }

func (self *Node) Reset() {
	self.assignment = ""
}

func (self *Node) AssignmentValue() string {
	return self.assignment
}

func (self *Node) IsRoot() bool {
	if len(self.parentIds) == 0 {
		return true
	}
	return false
}

func (self *Node) SetAssignmentValue(value string) {
	self.assignment = value
	// for _, child := range self.childIds {
	// 	child.ResetKey()
	// }
}

func (self *Node) ValidateCPT() error {
	// root node
	if self.IsRoot() {
		if len(self.cpt) != 2 {
			return fmt.Errorf("(Root): %s's CPT has wrong dimension: %d != %d act (cpt: %v)",
				self.name, 2, len(self.cpt), self.cpt)
		}
		return nil
	}

	for k, _ := range self.cpt {
		if len(k) != self.NumParents() {
			return fmt.Errorf("%s's CPT has wrong key-length: exp: %d != %d act (cpt: %v)",
				self.name, self.NumParents(), len(k), self.cpt)
		}
		break
	}

	exptectedCPTSize := int(math.Pow(2, float64(self.NumParents())))
	// fmt.Printf("%s: exp: %v\n", self.name, exptectedCPTSize)
	if len(self.cpt) != exptectedCPTSize {
		return fmt.Errorf("%s's CPT has wrong dimensions: exp: %d != %d act (cpt: %v)",
			self.name, exptectedCPTSize, len(self.cpt), self.cpt)
	}

	// if math.Abs(sum-1.0) > epsilon {
	// 	return fmt.Errorf("%f != %f %v", 1.0, sum, self.cpt)
	// }

	return nil
}

type BayNodes []*Node

func (bn BayNodes) Len() int {
	fmt.Println(len(bn))
	return len(bn)
}

func (bn BayNodes) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}

func (bn BayNodes) Less(i, j int) bool {
	return bn[i].Id() < bn[j].Id()
}

func (self BayNodes) String() string {

	if len(self) == 0 {
		return "[]"
	}

	var buffer bytes.Buffer

	buffer.WriteString(" ")
	for _, node := range self {
		buffer.WriteString(node.Name())
		if node.AssignmentValue() != "" {
			buffer.WriteString("(")
			s := node.AssignmentValue()
			buffer.WriteString(s)
			buffer.WriteString(")")
		}
		buffer.WriteString(" ")
	}

	return fmt.Sprintf("[%v]", buffer.String())
}
