package BayesianNetwork

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

const (
	epsilon = 0.05
)

func validateInterval(stats StatMap, nodeName string, exp float64, t *testing.T) {

	act := stats[nodeName]

	if math.Abs(act[0]-exp) > epsilon {
		t.Errorf("%s: Exp (%.4f,%.4f) != (%.4f,%.4f) Act\n", nodeName, exp, 1-exp, act[0], act[1])
	}
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

func TestMarkovFig8_2(t *testing.T) {

	rand.Seed(100)

	distRoot := 0.7

	dist1 := map[string]float64{
		"T": 0.1,
		"F": 0.7,
	}

	dist2 := map[string]float64{
		"TT": 0.9,
		"FF": 0.2,
		"TF": 0.2,
		"FT": 0.09,
	}

	dist3 := map[string]float64{
		"TTT": 0.2,
		"TFF": 0.3,
		"TTF": 0.6,
		"TFT": 0.7,
		"FTT": 0.1,
		"FFF": 0.8,
		"FTF": 0.3,
		"FFT": 0.6,
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
	dag := NewBayesianNetwork(p1, p2, p3, c1, c2, c3, c4)

	observations := map[string]string{
		"X1": "T",
		"X2": "F",
		"X3": "F",
		"X4": "F",
		"X6": "F",
		// // "X5": "T",
		"X7": "F",
	}

	mp := dag.GibbsSampling(observations, 10000, 10000)

	fmt.Printf("markovSampling: %v\n", mp)
}

func BuildStudentNetwork() *BayesianNetwork {
	// STUDENT NETWORK
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

	return NewBayesianNetwork(e, i, d, p, r, j, u)

}

func TestAncestralSamplingExample(t *testing.T) {
	sn := BuildStudentNetwork()
	stats := sn.AncestralSampling(10000)
	fmt.Println(stats)
}

// run both gibbsampling and ancestral sampling without evidence
// both resulting distributions should be comparable
// within some acceptable error epsilon ≈ 0.5 (was 0.01)
func TestAncestralVSGibbsSampling(t *testing.T) {
	bn := BuildStudentNetwork()

	ancestral := bn.AncestralSampling(10000)

	fmt.Printf("\tAncestral: %v\n", ancestral)

	obs := map[string]string{}

	gibbs := bn.GibbsSampling(obs, 1000, 10000)

	fmt.Printf("\tGibbs:     %v\n", gibbs)

	compareStatMaps(ancestral, gibbs, t)
}

func compareStatMaps(stat1, stat2 StatMap, t *testing.T) {

	if len(stat1) != len(stat2) {
		t.Errorf("len(dist1) != len(dist2)\n\t%v\n\t%v\n", stat1, stat2)
	}

	for key, a := range stat1 {
		var b []float64
		var ok bool
		if b, ok = stat2[key]; ok == false {
			t.Errorf("%s from stat1 does not exist in stat2", key)
		}

		if math.Abs(a[0]-b[0]) > epsilon {
			t.Errorf("%s: Exp (%.2f,%.2f) !≈ (%.2f,%.2f) Act\n", key, a[0], a[1], b[0], b[1])
		}
	}
}

func TestGibbSampling(t *testing.T) {

	// seed with value for repeatable results
	rand.Seed(42)

	observations := map[string]string{
		"J": "T",
		"E": "T",
		"I": "F",
		"D": "F",
		"R": "F",
		// "P": "F",
		"U": "T",
	}
	sn := BuildStudentNetwork()
	stats := sn.GibbsSampling(observations, 1000, 10000)

	validateInterval(stats, "P", 0.2, t)

	fmt.Printf("Stats: %v\n", stats)
}

// func BuildExampleGraph() (*BayesianNetwork, error) {
// 	x1 := NewRootNode("X1", 0.5)

// 	X2Dist := map[string]float64{
// 		"T": 0.1,
// 		"F": 0.5,
// 	}
// 	x2 := NewNode("X2", []string{"X1"}, X2Dist)

// 	X3Dist := map[string]float64{
// 		"T": 0.8,
// 		"F": 0.2,
// 	}
// 	x3 := NewNode("X3", []string{"X1"}, X3Dist)

// 	X4Dist := map[string]float64{
// 		"TT": 0.99,
// 		"TF": 0.9,
// 		"FT": 0.9,
// 		"FF": 0.0,
// 	}
// 	x4 := NewNode("X4", []string{"X2", "X3"}, X4Dist)

// 	bn := NewBayesianNetwork(x1, x2, x3, x4)
// 	return bn, nil
// }
