package BayesianNetwork

import (
// "math"
// "testing"
)

type NetworkStat struct {
	bn    *BayesianNetwork
	count []int
	total int
}

type StatMap map[string][]float64

func NewStat(bn *BayesianNetwork) *NetworkStat {
	return &NetworkStat{
		bn:    bn,
		total: 0,
		count: make([]int, len(bn.nodeIndex)),
	}
}

func (stat *NetworkStat) Update() {
	for i, node := range stat.bn.nodeIndex {
		assignment := node.AssignmentValue()
		if assignment == "F" {
			continue
		}
		stat.count[i] += 1
	}
	stat.total += 1
}

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
