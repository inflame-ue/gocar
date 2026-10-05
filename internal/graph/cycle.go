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
	found := false
	
	colors[node] = grey
	*path = append(*path, node)

	for _, dep := range g.deps[node] {
		col := colors[dep]

		if col == grey {
			temp := slices.Clone((*path)[slices.Index(*path, dep):])
			cycle, found = append(temp, dep), true

			return cycle, found
		}

		if col == white {
			return g.walk(dep, colors, path)
		}
	}

	colors[node] = black
	*path = (*path)[:len(*path)-1]

	return cycle, found
}

func (g *Graph) Cycle() ([]string, bool) {
	colors := map[string]color{}
	path := []string{}

	nodes := slices.Sorted(maps.Keys(g.deps))
	for _, node := range nodes {
		if col := colors[node]; col == white {
			if cycle, found :=  g.walk(node, colors, &path); found {
				return cycle, true
			}
		}
	}

	return nil, false
}
