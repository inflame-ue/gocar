package scheduler

import "github.com/inflame-ue/gocar/internal/graph"

func Level(g *graph.Graph) map[string]int {
	level := make(map[string]int, len(g.Nodes()))
	var visit func(string) int
	visit = func(node string) int {
		if l, ok := level[node]; ok {
			return l
		}
		max := 0
		for _, rdep := range g.Rdeps(node) {
			if l := visit(rdep); l > max {
				max = l
			}
		}
		level[node] = max + 1
		return level[node]
	}
	for _, n := range g.Nodes() {
		visit(n)
	}
	return level
}
