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

type BayesianNode interface {
	Sample() (string, float64, error)
	Reset()
	AddChild(BayesianNode)
	AddParent(BayesianNode) error
	Name() string
	NumParents() int
	NumChildren() int
	IsSampled() bool
	GetChildren() BayNodes
	GetParents() BayNodes
	GetParentNames() []string
	Id() int
	SetId(int)
	ValidateParents(parents []string) bool
	ProbTruth() float64
	ValidateCPT() error
}

func NewRootNode(name string, dist float64) BayesianNode {
	node := &Node{
		name:     name,
		childIds: make([]BayesianNode, 0, 4),
		cpt:      map[string]float64{"T": dist, "F": 1.0 - dist},
	}
	return BayesianNode(node)
}

func NewNode(name string, parents []string, dist map[string]float64) BayesianNode {
	node := &Node{
		name:        name,
		parentNames: parents,
		parentIds:   make([]BayesianNode, 0, 4),
		childIds:    make([]BayesianNode, 0, 4),
		cpt:         dist,
	}
	return BayesianNode(node)
}

type Node struct {
	id          int
	name        string
	parentNames []string
	childIds    BayNodes
	parentIds   BayNodes
	sample      string // onle != "" after all parents are sampled
	prob        float64
	cpt         map[string]float64 // conditional probability table
}

func (self *Node) IsSampled() bool {
	if self.sample == "" {
		return false
	}
	return true
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

func (self *Node) SetId(i int) {
	self.id = i
}

func (self *Node) GetParentNames() []string {
	// names := make([]string, 0, len(self.parentIds))
	// for _, node := range self.parentIds {
	// 	names = append(names, node.Name())
	// }

	return self.parentNames
}

func (self *Node) ProbTruth() float64 {
	if len(self.parentIds) == 0 {
		return self.cpt["T"]
	}

	key := truthKey(len(self.parentIds))

	val, ok := self.cpt[key]
	if !ok {
		return 0.0
	}

	return val
}

func truthKey(length int) string {
	var buffer bytes.Buffer
	for i := 0; i < length; i++ {
		buffer.WriteString("T")
	}

	return buffer.String()
}

func (self *Node) AddChild(child BayesianNode) {

	for _, c := range self.childIds {
		if c == child {
			return
		}
	}

	self.childIds = append(self.childIds, child)
}

func (self *Node) AddParent(parent BayesianNode) error {

	for _, p := range self.parentIds {
		if p == parent {
			return nil
		}
	}
	self.parentIds = append(self.parentIds, parent)

	return nil
}

func (self *Node) String() string {

	if self.sample != "" {
		return fmt.Sprintf("%s(%d): s='%s' p=%f (%v)\n\tparents:  %v\n\tchildren: %v\n",
			self.name, self.id, self.sample, self.prob, self.cpt, self.parentIds, self.childIds)
	}

	return fmt.Sprintf("%s(%d): %v\n\tparents:  %v\n\tchildren: %v\n",
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

func (self *Node) Reset() {
	self.sample = ""
	self.prob = 0.0
}

func (self *Node) Sample() (string, float64, error) {
	// if sample has already been calculated
	// return cached value
	if self.sample != "" {
		return self.sample, self.prob, nil
	}
	// rootNode
	if len(self.parentIds) == 0 {
		self.prob = rand.Float64()
		if self.prob > self.cpt["T"] {
			self.sample = "F"
		} else {
			self.sample = "T"
		}

		return self.sample, self.prob, nil
	}
	// Not a rootnode

	// recursively sample parents
	var buffer bytes.Buffer
	for _, id := range self.parentIds {
		s, _, err := id.Sample()

		if err != nil {
			return "", 0.0, err
		}
		// cache parent samples
		// self.parentSampleCache[i] = s
		// concat the samples into a key
		// which is used to store the values
		buffer.WriteString(s)
	}

	// Sample probability given conditions
	// look up the probability in the CPT
	key := buffer.String()
	prob, ok := self.cpt[key]
	if !ok {
		return "", 0.0, fmt.Errorf("%s not a valid node key", key)
	}

	// generate random number in [0.0;1.0]
	// generate seed based on current time
	self.prob = rand.Float64()

	if self.prob < prob {
		self.sample = "T"
	} else {
		self.sample = "F"
	}
	return self.sample, self.prob, nil
}

func (self *Node) IsRoot() bool {
	if len(self.parentIds) == 0 {
		return true
	}
	return false
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

type BayNodes []BayesianNode

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
		if node.IsSampled() {
			buffer.WriteString("(")
			s, _, _ := node.Sample()
			buffer.WriteString(s)
			buffer.WriteString(")")
		}
		buffer.WriteString(" ")
	}

	return fmt.Sprintf("[%v]", buffer.String())
}
