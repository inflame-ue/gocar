package graph

import (
	"slices"
	"testing"

	"github.com/inflame-ue/gocar/internal/testutil"
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
					"format": {},
					"build":  {"format"},
				},
				rdeps: map[string][]string{
					"format": {"build"},
					"build":  {},
				},
				indeg: map[string]int{
					"format": 0,
					"build":  1,
				},
			},
			wantOrder: []string{"format", "build"},
			wantErr:   nil,
		},
		"valid diamond graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"vet":    {"format"},
					"lint":   {"format"},
					"build":  {"lint", "vet"},
				},
				rdeps: map[string][]string{
					"format": {"lint", "vet"},
					"vet":    {"build"},
					"lint":   {"build"},
					"build":  {},
				},
				indeg: map[string]int{
					"format": 0,
					"vet":    1,
					"lint":   1,
					"build":  2,
				},
			},
			wantOrder: []string{"format", "lint", "vet", "build"},
			wantErr:   nil,
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
				rdeps: map[string][]string{
					"format": {"vet", "lint"},
					"vet":    {"build"},
					"lint":   {"build", "run"},
					"build":  {},
					"run":    {},
				},
				indeg: map[string]int{
					"format": 0,
					"vet":    1,
					"lint":   1,
					"build":  2,
					"run":    1,
				},
			},
			wantOrder: []string{"format", "lint", "run", "vet", "build"},
			wantErr:   nil,
		},
		"disconnected component graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"build":  {"format"},
					"lint":   {},
					"vet":    {"lint"},
					"run":    {"vet"},
				},
				rdeps: map[string][]string{
					"format": {"build"},
					"build":  {},
					"lint":   {"vet"},
					"vet":    {"run"},
					"run":    {},
				},
				indeg: map[string]int{
					"format": 0,
					"build":  1,
					"lint":   0,
					"vet":    1,
					"run":    1,
				},
			},
			wantOrder: []string{"format", "build", "lint", "vet", "run"},
			wantErr:   nil,
		},
		"cyclic linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {"build"},
					"build":  {"format"},
				},
				rdeps: map[string][]string{
					"format": {"build"},
					"build":  {"format"},
				},
				indeg: map[string]int{
					"format": 1,
					"build":  1,
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
					"format": {"build"},
					"vet":    {"format"},
					"lint":   {"format"},
					"build":  {"lint", "vet"},
				},
				rdeps: map[string][]string{
					"format": {"vet", "lint"},
					"vet":    {"build"},
					"lint":   {"build"},
					"build":  {"format"},
				},
				indeg: map[string]int{
					"format": 1,
					"vet":    1,
					"lint":   1,
					"build":  2,
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

			tErr := testutil.CheckErrAs(t, err, tc.wantErr)
			if tc.wantErr != nil && !slices.Equal(tErr.Path, tc.wantErr.Path) {
				t.Errorf("expected cycle %v in error, got %v instead", tc.wantErr.Path, tErr.Path)
			}

			if !slices.Equal(order, tc.wantOrder) {
				t.Errorf("expected order %v, got %v instead", tc.wantOrder, order)
			}
		})
	}
}

func TestDFSOrder(t *testing.T) {
	tests := map[string]struct {
		graph     *Graph
		wantOrder []string
		wantErr   *CycleError
	}{
		"valid linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"build":  {"format"},
				},
				rdeps: map[string][]string{
					"format": {"build"},
					"build":  {},
				},
				indeg: map[string]int{
					"format": 0,
					"build":  1,
				},
			},
			wantOrder: []string{"format", "build"},
			wantErr:   nil,
		},
		"valid diamond graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"vet":    {"format"},
					"lint":   {"format"},
					"build":  {"lint", "vet"},
				},
				rdeps: map[string][]string{
					"format": {"lint", "vet"},
					"vet":    {"build"},
					"lint":   {"build"},
					"build":  {},
				},
				indeg: map[string]int{
					"format": 0,
					"vet":    1,
					"lint":   1,
					"build":  2,
				},
			},
			wantOrder: []string{"format", "lint", "vet", "build"},
			wantErr:   nil,
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
				rdeps: map[string][]string{
					"format": {"vet", "lint"},
					"vet":    {"build"},
					"lint":   {"build", "run"},
					"build":  {},
					"run":    {},
				},
				indeg: map[string]int{
					"format": 0,
					"vet":    1,
					"lint":   1,
					"build":  2,
					"run":    1,
				},
			},
			wantOrder: []string{"format", "vet", "lint", "build", "run"},
			wantErr:   nil,
		},
		"disconnected component graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {},
					"build":  {"format"},
					"lint":   {},
					"vet":    {"lint"},
					"run":    {"vet"},
				},
				rdeps: map[string][]string{
					"format": {"build"},
					"build":  {},
					"lint":   {"vet"},
					"vet":    {"run"},
					"run":    {},
				},
				indeg: map[string]int{
					"format": 0,
					"build":  1,
					"lint":   0,
					"vet":    1,
					"run":    1,
				},
			},
			wantOrder: []string{"format", "build", "lint", "vet", "run"},
			wantErr:   nil,
		},
		"cyclic linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": {"build"},
					"build":  {"format"},
				},
				rdeps: map[string][]string{
					"format": {"build"},
					"build":  {"format"},
				},
				indeg: map[string]int{
					"format": 1,
					"build":  1,
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
					"format": {"build"},
					"vet":    {"format"},
					"lint":   {"format"},
					"build":  {"lint", "vet"},
				},
				rdeps: map[string][]string{
					"format": {"vet", "lint"},
					"vet":    {"build"},
					"lint":   {"build"},
					"build":  {"format"},
				},
				indeg: map[string]int{
					"format": 1,
					"vet":    1,
					"lint":   1,
					"build":  2,
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

			order, err := tc.graph.DFSOrder()

			tErr := testutil.CheckErrAs(t, err, tc.wantErr)
			if tc.wantErr != nil && !slices.Equal(tErr.Path, tc.wantErr.Path) {
				t.Errorf("expected cycle %v in error, got %v instead", tc.wantErr.Path, tErr.Path)
			}

			if !slices.Equal(order, tc.wantOrder) {
				t.Errorf("expected order %v, got %v instead", tc.wantOrder, order)
			}
		})
	}
}
