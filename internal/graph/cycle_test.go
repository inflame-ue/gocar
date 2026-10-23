package graph

import (
	"slices"
	"testing"
)

func TestCycle(t *testing.T) {
	tests := map[string]struct {
		graph     *Graph
		wantCycle []string
		wantFound bool
	}{
		"valid linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"build":  {"format"},
				},
			},
			wantFound: false,
			wantCycle: nil,
		},
		"valid diamond graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"vet":    {"format"},
					"lint":   {"format"},
					"build":  {"lint", "vet"},
				},
			},
			wantFound: false,
			wantCycle: nil,
		},
		"tree-style branching graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"vet":    {"format"},
					"lint":   {"format"},
					"build":  {"vet", "lint"},
					"run":    {"lint"},
				},
			},
			wantFound: false,
			wantCycle: nil,
		},
		"cyclic linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {"build"},
					"build":  {"format"},
				},
			},
			wantFound: true,
			wantCycle: []string{"build", "format", "build"},
		},
		"cyclic diamond graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {"build"},
					"vet":    {"format"},
					"lint":   {"format"},
					"build":  {"lint", "vet"},
				},
			},
			wantFound: true,
			wantCycle: []string{"build", "lint", "format", "build"},
		},
		"self-loop graph": {
			graph: &Graph{
				deps: map[string][]string{
					"build": {"build"},
				},
			},
			wantFound: true,
			wantCycle: []string{"build", "build"},
		},
		"disconnected component graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"build":  {"format"},
					"lint":   {"run"},
					"vet":    {"lint"},
					"run":    {"vet"},
				},
			},
			wantFound: true,
			wantCycle: []string{"lint", "run", "vet", "lint"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cycle, found := tc.graph.Cycle()

			if found && !tc.wantFound {
				t.Errorf("did not expect a cycle, but one was found")
			}

			if !found && tc.wantFound {
				t.Errorf("expected a cycle, but none were found")
			}

			if !slices.Equal(cycle, tc.wantCycle) {
				t.Errorf("expected cycle %v, got %v instead", tc.wantCycle, cycle)
			}
		})
	}
}
