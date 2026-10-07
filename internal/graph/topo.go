package graph

import (
	"maps"
	"slices"
	"sort"
)

func (g *Graph) Kahn() ([]string, error) {
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
	for len(ready) > 0 {
		// smallest node is first since ready is sorted
		node := ready[0]
		ready = ready[1:]
		order = append(order, node)

		for _, rdep := range g.rdeps[node] {
			indeg[rdep]--
			if indeg[rdep] == 0 {
				i := sort.Search(len(ready), func(j int) bool {
					return ready[j] >= rdep
				})
				ready = slices.Insert(ready, i, rdep)
			}
		}
	}

	if len(order) != len(g.deps) {
		cycle, _ := g.Cycle()
		return nil, &CycleError{
			Path: cycle,
		}
	}

	return order, nil
}
