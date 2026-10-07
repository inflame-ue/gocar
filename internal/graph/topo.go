package graph

import (
	"maps"
	"slices"
)

func (g *Graph) Kahn() ([]string, error) {
	if cycle, found := g.Cycle(); found {
		return nil, &CycleError{
			Path: cycle,
		}
	}

	indeg := maps.Clone(g.indeg)
	var ready []string
	for node := range g.deps {
		if indeg[node] != 0 {
			continue
		}
		ready = append(ready, node)
	}
	slices.Sort(ready)

	var order []string


	if len(order) != len(g.deps) {
		return nil, &CycleError{}
	}

	return order, nil
}