package graph

type Graph struct {
	deps map[string][]string 
	rdeps map[string][]string
	indeg map[string]int
}

