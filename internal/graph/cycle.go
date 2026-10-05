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

func (g *Graph) walk(node string, colors map[string]color, path *[]string) {
	colors[node] = grey
	*path = append(*path, node)

	for _, dep := range g.deps[node] {
		col := colors[dep]

		if col == grey {
			*path = (*path)[slices.Index(*path, dep):]
		}

		if col == white {
			g.walk(dep, colors, path)
		}
	}

	colors[node] = black
	*path = (*path)[:len(*path)-1]
}

func (g *Graph) Cycle() ([]string, bool) {
	colors := map[string]color{}
	path := []string{}

	nodes := slices.Sorted(maps.Keys(g.deps))
	for _, node := range nodes {
		g.walk(node, colors, &path)
	}

	if len(path) != 0 {
		return path, true
	}

	return nil, false
}
