package BayesianNetwork

import (
	"fmt"
	// "math"
	"testing"
)

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

	fmt.Println(bn.JointProbability())
}

// use example from slides instead
// because that has known probabilities
func TestJointProb(t *testing.T) {

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

	fmt.Printf("%f\n", dag.JointProbability())
}

func TestAncestralSampling(t *testing.T) {

	// return

	distRoot := 0.3

	dist1 := map[string]float64{
		"T": 0.4,
		"F": 0.6,
	}

	dist2 := map[string]float64{
		"TT": 0.1,
		"FF": 0.4,
		"TF": 0.3,
		"FT": 0.2,
	}

	dist3 := map[string]float64{
		"TTT": 0.1,
		"TFF": 0.4,
		"TTF": 0.3,
		"TFT": 0.2,
		"FTT": 0.6,
		"FFF": 0.5,
		"FTF": 0.5,
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
	dag, err := NewBayesianNetwork(p1, p2, p3, c1, c2, c3, c4)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(dag)

	sampleCount := 5
	fmt.Printf("X4 Before Ancestral Sampling:\n%v\n", dag.GetNode("X4"))
	for i := 0; i < sampleCount; i++ {
		dag.AncestralSampling()
		fmt.Printf("X4(s=%d):\n%v\n", i, dag.GetNode("X4"))
	}
}

// func TestMarkovBlanket(t *testing.T) {

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
// 	fmt.Println(dag.PrintNetwork())

// }

// func TestMarkovFig8_2(t *testing.T) {
// 	distRoot := 0.3

// 	dist1 := map[string]float64{
// 		"T": 0.4,
// 		"F": 0.6,
// 	}

// 	dist2 := map[string]float64{
// 		"TT": 0.6,
// 		"FF": 0.9,
// 		"TF": 0.3,
// 		"FT": 0.2,
// 	}

// 	dist3 := map[string]float64{
// 		"TTT": 0.6,
// 		"TFF": 0.4,
// 		"TTF": 0.4,
// 		"TFT": 0.7,
// 		"FTT": 0.5,
// 		"FFF": 0.5,
// 		"FTF": 0.3,
// 		"FFT": 0.9,
// 	}

// 	// network from BISHOP p. 362, Fig. 8.2
// 	p1 := NewRootNode("X1", distRoot)
// 	p2 := NewRootNode("X2", distRoot)
// 	p3 := NewRootNode("X3", distRoot)

// 	c1 := NewNode("X4", []string{"X1", "X2", "X3"}, dist3)
// 	c2 := NewNode("X5", []string{"X1", "X3"}, dist2)
// 	c3 := NewNode("X6", []string{"X4"}, dist1)
// 	c4 := NewNode("X7", []string{"X4", "X5"}, dist2)

// 	// construct the network
// 	dag, err := NewBayesianNetwork(p1, p2, p3, c1, c2, c3, c4)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	mapping := map[string]string{
// 		"X1": "F",
// 		"X3": "F",
// 		"X4": "F",
// 		"X7": "T",
// 	}
// 	fmt.Printf("markovSampling: %v\n", dag.MarkovBlanketSampling("X5", mapping))
// }
