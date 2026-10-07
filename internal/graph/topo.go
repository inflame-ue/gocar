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

func (g *Graph) visit(node string, colors map[string]color, order *[]string) error {
	if col, ok := colors[node]; ok && col == black {
		return nil
	}

	if col, ok := colors[node]; ok && col == grey {
		cycle, _ := g.Cycle()
		return &CycleError{
			Path: cycle,
		}
	}

	colors[node] = grey

	for _, dep := range g.deps[node] {
		g.visit(dep, colors, order)
	}

	colors[node] = black
	*order = append(*order, node)

	return nil
}

func (g *Graph) DFSOrder() ([]string, error) {
	var order []string
	colors := map[string]color{}

	nodes := slices.Sorted(maps.Keys(g.deps))
	for _, node := range nodes {
		if _, ok := colors[node]; ok {
			continue
		}
		
		if err := g.visit(node, colors, &order); err != nil {
			return nil, err
		}
	}

	return order, nil
}