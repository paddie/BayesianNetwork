GoBayesianNetwork
=================

BayesianNetwork

## Example

Building an example graph:
```Go
package main

import (
  . "github.com/paddie/BayesianNetwork"
  "fmt"
)

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

bn := NewBayesianNetwork(e, i, d, p, r, j, u)

stats := bn.AncestralSampling(10000)
fmt.Println(stats)
```
## Output:
```Bash
map[E:[0.3007 0.6993] I:[0.7044 0.2956] D:[0.2054 0.7946] P:[0.5013 0.4987] R:[0.5601 0.4399] J:[0.4988 0.5012] U:[0.6611 0.3389]]
```