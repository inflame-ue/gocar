package graph

import (
	"maps"
	"slices"
)

type color uint8

const (
	white color = iota
	grey
	black
)

func (g *Graph) walk(node string, colors map[string]color, path *[]string) ([]string, bool) {
	cycle := []string{}

	colors[node] = grey
	*path = append(*path, node)

	for _, dep := range g.deps[node] {
		col := colors[dep]

		if col == grey {
			temp := slices.Clone((*path)[slices.Index(*path, dep):])
			cycle = append(temp, dep)

			return cycle, true
		}

		if col == white {
			if col, ok := g.walk(dep, colors, path); ok {
				return col, true
			}
		}
	}

	colors[node] = black
	*path = (*path)[:len(*path)-1]

	return cycle, false
}

func (g *Graph) Cycle() ([]string, bool) {
	colors := map[string]color{}
	path := []string{}

	nodes := slices.Sorted(maps.Keys(g.deps))
	for _, node := range nodes {
		// we only want to walk unvisited(white) nodes
		if col := colors[node]; col == white {
			if cycle, found := g.walk(node, colors, &path); found {
				return cycle, true
			}
		}
	}

	return nil, false
}
