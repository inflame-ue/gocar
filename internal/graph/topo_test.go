package graph

import (
	"errors"
	"slices"
	"testing"
)

func TestKahn(t *testing.T) {
	tests := map[string]struct {
		graph     *Graph
		wantOrder []string
		wantErr   *CycleError
	}{
		"valid linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"build":  []string{"format"},
				},
			},
			wantOrder: []string{"format", "build"},
			wantErr:   nil,
		},
		"valid diamond graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"vet":    []string{"format"},
					"lint":   []string{"format"},
					"build":  []string{"lint", "vet"},
				},
			},
			wantOrder: []string{"format", "vet", "lint", "build"},
			wantErr:   nil,
		},
		"tree-style branching graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"vet":    []string{"format"},
					"lint":   []string{"format"},
					"build":  []string{"vet", "lint"},
					"run":    []string{"lint"},
				},
			},
			wantOrder: []string{"format", "vet", "lint", "run", "build"},
			wantErr:   nil,
		},
		"disconnected component graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"build":  []string{"format"},
					"lint":   []string{},
					"vet":    []string{},
					"run":    []string{"vet"},
				},
			},
			wantOrder: []string{"format", "build", "lint", "vet", "run"},
			wantErr:   nil,
		},
		"cyclic linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{"build"},
					"build":  []string{"format"},
				},
			},
			wantOrder: nil,
			wantErr: &CycleError{
				Path: []string{"build", "format", "build"},
			},
		},
		"cyclic diamond graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{"build"},
					"vet":    []string{"format"},
					"lint":   []string{"format"},
					"build":  []string{"lint", "vet"},
				},
			},
			wantOrder: nil,
			wantErr: &CycleError{
				Path: []string{"build", "lint", "format", "build"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			order, err := tc.graph.Kahn()

			if err != nil && tc.wantErr == nil {
				t.Fatalf("expected no err, got %v", err)
			}

			if err == nil && tc.wantErr != nil {
				t.Fatalf("expected an err %v, got no err instead", tc.wantErr)
			}

			if err != nil && tc.wantErr != nil {
				var ce *CycleError
				if ok := errors.As(err, &ce); !ok {
					t.Errorf("expected err type %T, got %T instead", tc.wantErr, err)
				}

				if !slices.Equal(ce.Path, tc.wantErr.Path) {
					t.Errorf("expected cycle %v, got %v instead", tc.wantErr.Path, ce.Path)
				}
			}

			if !slices.Equal(order, tc.wantOrder) {
				t.Errorf("expected order %v, got %v instead", tc.wantOrder, order)
			}
		})
	}
}
