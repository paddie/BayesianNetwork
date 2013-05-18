package BayesianNetwork

import (
	"fmt"
	"math"
	"testing"
)

const (
	epsilon = 0.01
)

var bn *BayesianNetwork
var bn2 *BayesianNetwork

func BuildExampleGraph() (*BayesianNetwork, error) {
	x1 := NewRootNode("X1", 0.5)

	X2Dist := map[string]float64{
		"T": 0.1,
		"F": 0.5,
	}
	x2 := NewNode("X2", []string{"X1"}, X2Dist)

	X3Dist := map[string]float64{
		"T": 0.8,
		"F": 0.2,
	}
	x3 := NewNode("X3", []string{"X1"}, X3Dist)

	X4Dist := map[string]float64{
		"TT": 0.99,
		"TF": 0.9,
		"FT": 0.9,
		"FF": 0.0,
	}
	x4 := NewNode("X4", []string{"X2", "X3"}, X4Dist)

	bn, err := NewBayesianNetwork(x1, x2, x3, x4)
	if err != nil {
		return nil, err
	}
	return bn, nil
}

func validateInterval(stats StatMap, nodeName string, exp float64, t *testing.T) {

	act := stats[nodeName]

	if math.Abs(act[0]-exp) > epsilon {
		t.Errorf("%s: Exp (%.4f,%.4f) != (%.4f,%.4f) Act\n", nodeName, exp, 1-exp, act[0], act[1])
	}
}

func TestAncestralSamplingExample(t *testing.T) {
	bn, err := BuildExampleGraph()
	if err != nil {
		t.Error(err)
	}

	stats, err := bn.AncestralSampling(map[string]string{}, 10000)
	if err != nil {
		t.Error(err)
	}
	// x1 = <0.5, 0.5>
	validateInterval(stats, "X1", 0.5, t)

	x2Map := map[string]string{
		"X1": "T",
	}
	stats, err = bn.AncestralSampling(x2Map, 10000)
	if err != nil {
		t.Error(err)
	}
	// x2 = <0.1, 0.9>
	validateInterval(stats, "X2", 0.1, t)
	// x3 = <0.8, 0.2>
	validateInterval(stats, "X3", 0.8, t)

	x4Map := map[string]string{
		"X2": "F",
		"X3": "T",
	}
	stats, err = bn.AncestralSampling(x4Map, 10000)
	if err != nil {
		t.Error(err)
	}
	// x4 = <0.9, 0.1>
	validateInterval(stats, "X4", 0.9, t)

	bn.JointProbability()

}

func (bn *BayesianNetwork) ValidateIndex(t *testing.T) {
	for _, node := range bn.GetNodes() {
		for _, child := range node.childIds {
			if node.Id() > child.Id() {
				t.Errorf("Invalid ID on '%s': child '%s' has id ",
					node.Name(), child.Name(), child.Id())
			}
		}
	}
}

func TestValidateIndexing(t *testing.T) {
	bn.ValidateIndex(t)
}

func TestValidateCPTs(t *testing.T) {
	err := bn.validateCPTs()
	if err != nil {
		t.Error(err)
	}
}

// func TestNetworkConstruction(t *testing.T) {

// 	e := NewRootNode("E", 0.3)
// 	i := NewRootNode("I", 0.7)
// 	d := NewRootNode("D", 0.2)

// 	pDist := map[string]float64{
// 		"TTT": 0.9,
// 		"TFF": 0.2,
// 		"TTF": 0.5,
// 		"TFT": 0.7,
// 		"FTT": 0.8,
// 		"FFF": 0.07,
// 		"FTF": 0.6,
// 		"FFT": 0.7,
// 	}

// 	p := NewNode("P", []string{"E", "I", "D"}, pDist)

// 	rDist := map[string]float64{
// 		"TT": 0.9,
// 		"FF": 0.2,
// 		"TF": 0.6,
// 		"FT": 0.9,
// 	}

// 	r := NewNode("R", []string{"I", "D"}, rDist)

// 	jDist := map[string]float64{
// 		"T": 0.7,
// 		"F": 0.3,
// 	}
// 	j := NewNode("J", []string{"P"}, jDist)

// 	uDist := map[string]float64{
// 		"TT": 0.9,
// 		"FF": 0.3,
// 		"TF": 0.6,
// 		"FT": 0.8,
// 	}
// 	u := NewNode("U", []string{"P", "R"}, uDist)

// 	dag, err := NewBayesianNetwork(e, i, d, p, r, j, u)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Println(dag)
// }

func TestJointProbabilityCH3P32(t *testing.T) {

	r1 := NewRootNode("R", 0.8)
	r2 := NewRootNode("S", 0.4)

	c1 := NewNode("J", []string{"R"}, map[string]float64{
		"T": 1.0,
		"F": 0.2,
	})
	c2 := NewNode("T", []string{"R", "S"}, map[string]float64{
		"TF": 1.0,
		"TT": 1.0,
		"FT": 0.9,
		"FF": 0.0,
	})

	bn, err := NewBayesianNetwork(r1, r2, c1, c2)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(bn)

	fmt.Println(bn.JointProbability())
}

// use example from slides instead
// because that has known probabilities
func TestJointProbabilityBasic(t *testing.T) {

	dist1 := map[string]float64{
		"T": 2.0 / 3.0,
		"F": 3.0 / 4.0,
	}

	distRoot := 0.5

	p1 := NewRootNode("X1", distRoot)
	c1 := NewNode("X2", []string{"X1"}, dist1)

	dag, err := NewBayesianNetwork(p1, c1)
	if err != nil {
		t.Error(err)
		return
	}
	jp, err := dag.JointProbability()

	fmt.Printf("%f\n", jp)
}

// // func TestAncestralSampling(t *testing.T) {
// // 	fmt.Printf("All nodes before ancestral sampling:\n%s\n", dag.PrintNetwork())
// // 	dag.AncestralSampling("X1", "X2", "X3", "X4", "X5", "X6", "X7")
// // 	fmt.Printf("All nodes after ancestral sampling:\n%s\n", dag.PrintNetwork())
// // }

func TestMarkovFig8_2(t *testing.T) {
	distRoot := 0.9

	dist1 := map[string]float64{
		"T": 0.4,
		"F": 0.6,
	}

	dist2 := map[string]float64{
		"TT": 0.4,
		"FF": 0.9,
		"TF": 0.3,
		"FT": 0.9,
	}

	dist3 := map[string]float64{
		"TTT": 0.6,
		"TFF": 0.4,
		"TTF": 0.4,
		"TFT": 0.7,
		"FTT": 0.5,
		"FFF": 0.5,
		"FTF": 0.3,
		"FFT": 0.9,
	}

	// network from BISHOP p. 362, Fig. 8.2
	p1 := NewRootNode("X1", distRoot)
	p2 := NewRootNode("X2", distRoot)
	p3 := NewRootNode("X3", distRoot)

	c1 := NewNode("X4", []string{"X1", "X2", "X3"}, dist3)
	c2 := NewNode("X5", []string{"X1", "X3"}, dist2)
	c3 := NewNode("X6", []string{"X4"}, dist1)
	c4 := NewNode("X7", []string{"X4", "X5"}, dist2)

	// construct the network
	dag, err := NewBayesianNetwork(p1, p2, p3, c1, c2, c3, c4)
	if err != nil {
		t.Error(err)
	}

	// nomap := map[string]string{}

	mapping := map[string]string{
		"X1": "T",
		"X3": "T",
		"X4": "T",
		"X7": "F",
	}

	jp, _ := dag.JointProbability()
	fmt.Printf("jointProbability: %4f", jp)

	mp, err := dag.MarkovBlanketSampling("X5", mapping, 10000)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("markovSampling: %v\n", mp)
}

func init() {
	// ****** bn 1 ********
	e := NewRootNode("E", 0.3)
	i := NewRootNode("I", 0.7)
	d := NewRootNode("D", 0.2)

	pDist := map[string]float64{
		"TTT": 0.9,
		"TFF": 0.2,
		"TTF": 0.5,
		"TFT": 0.7,
		"FTT": 0.8,
		"FFF": 0.07,
		"FTF": 0.6,
		"FFT": 0.7,
	}

	p := NewNode("P", []string{"E", "I", "D"}, pDist)

	rDist := map[string]float64{
		"TT": 0.9,
		"FF": 0.2,
		"TF": 0.6,
		"FT": 0.9,
	}

	r := NewNode("R", []string{"I", "D"}, rDist)

	jDist := map[string]float64{
		"T": 0.7,
		"F": 0.3,
	}
	j := NewNode("J", []string{"P"}, jDist)

	uDist := map[string]float64{
		"TT": 0.9,
		"FF": 0.3,
		"TF": 0.6,
		"FT": 0.8,
	}
	u := NewNode("U", []string{"P", "R"}, uDist)

	var err error
	bn, err = NewBayesianNetwork(e, i, d, p, r, j, u)
	if err != nil {
		panic(err)
	}
}
