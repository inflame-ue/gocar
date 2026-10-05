package graph

import (
	"fmt"
	"maps"
	"slices"

	"github.com/inflame-ue/gocar/internal/task"
)

type Graph struct {
	deps  map[string][]string
	rdeps map[string][]string
	indeg map[string]int
}

func Build(spec map[string]task.Task) (*Graph, error) {
	if spec == nil {
		return &Graph{}, nil
	}

	names := slices.Sorted(maps.Keys(spec))
	for _, name := range names {
		for _, dep := range spec[name].Deps {
			if _, ok := spec[dep]; !ok {
				return nil, &MissingDepError{
					Task: name,
					Dep: dep,
				}
			}
		}
	}

	return nil, nil
}
