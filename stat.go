package BayesianNetwork

import (
	// "math"
	// "testing"
	"fmt"
)

type NetworkStat struct {
	bn    *BayesianNetwork
	count []int
	total int
}

type StatMap map[string][]float64

func NewNetworkStat(bn *BayesianNetwork) *NetworkStat {
	return &NetworkStat{
		bn:    bn,
		total: 0,
		count: make([]int, len(bn.nodeIndex)),
	}
}

// run through the entire network and increment
// if the value is "T", else ignore
func (stat *NetworkStat) Update() {
	for i, node := range stat.bn.nodeIndex {
		assignment := node.GetAssignment()
		if assignment == "F" {
			continue
		}
		stat.count[i] += 1
	}
	stat.total += 1
}

// return a mapping of the normalized probablilities
func (stat *NetworkStat) GetStats() StatMap {
	stats := make(map[string][]float64, len(stat.count))

	for i, node := range stat.bn.GetNodes() {
		stats[node.Name()] = []float64{
			float64(stat.count[i]) / float64(stat.total),
			float64(stat.total-stat.count[i]) / float64(stat.total),
		}
	}
	return stats
}

type NodeStat struct {
	node         *Node
	count, total int
}

func NewNodeStat(node *Node) *NodeStat {
	return &NodeStat{
		node:  node,
		count: 0,
		total: 0,
	}
}

func (ns *NodeStat) Update(assignment string) {
	ns.total += 1
	if assignment == "F" {
		return
	}

	ns.count += 1
}

func (ns *NodeStat) GetStats() []float64 {
	return []float64{
		float64(ns.count) / float64(ns.total),
		float64(ns.total-ns.count) / float64(ns.total),
	}
}

func (ns *NodeStat) String() string {
	return fmt.Sprintf("%s: T: %.3f F: %.3f",
		ns.node.Name(),
		float64(ns.count)/float64(ns.total),
		float64(ns.total-ns.count)/float64(ns.total))
}
