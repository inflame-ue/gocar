package graph

import (
	"maps"
	"slices"

	"github.com/inflame-ue/gocar/internal/task"
)

type Graph struct {
	deps  map[string][]string
	rdeps map[string][]string
	indeg map[string]int
}

func missingDep(spec map[string]task.Task, task string) (string, bool) {
	for _, dep := range spec[task].Deps {
		if _, ok := spec[dep]; !ok {
			return dep, true
		}
	}

	return "", false
}

func Build(spec map[string]task.Task) (*Graph, error) {
	graph := Graph{
		deps:  map[string][]string{},
		rdeps: map[string][]string{},
		indeg: map[string]int{},
	}

	names := slices.Sorted(maps.Keys(spec))

	// pre-populate rdeps to not return nil-slice
	for _, name := range names {
		graph.rdeps[name] = []string{}
	}

	for _, name := range names {
		if dep, ok := missingDep(spec, name); ok {
			return nil, &MissingDepError{
				Task: name,
				Dep:  dep,
			}
		}

		// could potentially error here to avoid faulty tasks
		// for now remove dup dependencies for better ux
		uniqueDeps := slices.Clone(spec[name].Deps)
		slices.Sort(uniqueDeps)
		uniqueDeps = slices.Compact(uniqueDeps)

		for _, dep := range uniqueDeps {
			graph.rdeps[dep] = append(graph.rdeps[dep], name)
		}

		graph.indeg[name] = len(uniqueDeps)
		graph.deps[name] = uniqueDeps
	}

	return &graph, nil
}
