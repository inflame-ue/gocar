package graph

import (
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
)

func checkWantErr(t *testing.T, got error, want *MissingDepError) {
	t.Helper()

	if got == nil && want == nil {
		return
	}

	if got != nil && want == nil {
		t.Fatalf("expected no err, got %v", got)
	}

	if got == nil && want != nil {
		t.Fatal("expected an err, got no err instead")
	}

	var mde *MissingDepError
	if !errors.As(got, &mde) {
		t.Errorf("expected err type MissingDepError, got %T instead", got)
	}

	if mde.Task != want.Task {
		t.Errorf("expected task %s, got %s instead", want.Task, mde.Task)
	}

	if mde.Dep != want.Dep {
		t.Errorf("expected dep %s, got %s instead", want.Dep, mde.Dep)
	}
}

func TestBuild(t *testing.T) {
	tests := map[string]struct {
		spec    map[string]task.Task
		want    *Graph
		wantErr *MissingDepError
	}{
		"single node graph": {
			spec: map[string]task.Task{
				"build": task.Task{
					Cmd: "go build .",
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"build": []string{},
				},
				rdeps: map[string][]string{
					"build": []string{},
				},
				indeg: map[string]int{
					"build": 0,
				},
			},
			wantErr: nil,
		},
		"valid linked list graph": {
			spec: map[string]task.Task{
				"format": task.Task{
					Cmd: "go fmt ./...",
				},
				"build": task.Task{
					Cmd:  "go build ./...",
					Deps: []string{"format"},
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"build":  []string{"format"},
				},
				rdeps: map[string][]string{
					"format": []string{"build"},
					"build":  []string{},
				},
				indeg: map[string]int{
					"format": 0,
					"build":  1,
				},
			},
			wantErr: nil,
		},
		"valid diamond graph": {
			spec: map[string]task.Task{
				"format": task.Task{
					Cmd: "go fmt ./...",
				},
				"vet": task.Task{
					Cmd:  "go vet ./...",
					Deps: []string{"format"},
				},
				"lint": task.Task{
					Cmd:  "golangci-lint run",
					Deps: []string{"format"},
				},
				"build": task.Task{
					Cmd:  "go build ./...",
					Deps: []string{"vet", "lint"},
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"vet":    []string{"format"},
					"lint":   []string{"format"},
					"build":  []string{"lint", "vet"},
				},
				rdeps: map[string][]string{
					"format": []string{"lint", "vet"},
					"vet":    []string{"build"},
					"lint":   []string{"build"},
					"build":  []string{},
				},
				indeg: map[string]int{
					"format": 0,
					"vet":    1,
					"lint":   1,
					"build":  2,
				},
			},
		},
		"missing dep graph": {
			spec: map[string]task.Task{
				"build": task.Task{
					Cmd:  "go build ./...",
					Deps: []string{"format"},
				},
			},
			want: nil,
			wantErr: &MissingDepError{
				Task: "build",
				Dep:  "format",
			},
		},
		"duplicate deps graph": {
			spec: map[string]task.Task{
				"format": task.Task{
					Cmd: "go fmt ./...",
				},
				"build": task.Task{
					Cmd:  "go build ./...",
					Deps: []string{"format", "format"},
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"build":  []string{"format"},
				},
				rdeps: map[string][]string{
					"format": []string{"build"},
					"build":  []string{},
				},
				indeg: map[string]int{
					"format": 0,
					"build":  1,
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := Build(tc.spec)

			checkWantErr(t, err, tc.wantErr)

			if got == nil || tc.want == nil {
				return
			}

			if !maps.Equal(tc.want.indeg, got.indeg) {
				t.Errorf("graph build: in-degree counts for each node are incorrect")
			}

			if len(tc.want.deps) != len(got.deps) {
				t.Errorf("graph build: expected %d tasks, got %d instead", len(tc.want.deps), len(got.deps))
			}

			for task := range tc.want.deps {
				if !slices.Equal(tc.want.deps[task], got.deps[task]) {
					t.Error("graph build: deps for each node are not equal")
				}

				if !slices.Equal(tc.want.rdeps[task], got.rdeps[task]) {
					t.Error("graph build: rdeps for each node are not equal")
				}
			}
		})
	}
}

func TestCycle(t *testing.T) {
	tests := map[string]struct {
		graph     *Graph
		wantCycle []string
		wantFound bool
	}{
		"valid linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"build":  []string{"format"},
				},
			},
			wantFound: false,
			wantCycle: nil,
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
			wantFound: false,
			wantCycle: nil,
		},
		"tree-style branching graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"vet": []string{"format"},
					"lint": []string{"format"},
					"build": []string{"vet", "lint"},
					"run": []string{"lint"},
				},	
			},
			wantFound: false,
			wantCycle: nil,
		},
		"cyclic linked list graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{"build"},
					"build":  []string{"format"},
				},
			},
			wantFound: true,
			wantCycle: []string{"build", "format", "build"},
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
			wantFound: true,
			wantCycle: []string{"build", "lint", "format", "build"},
		},
		"self-loop graph": {
			graph: &Graph{
				deps: map[string][]string{
					"build": []string{"build"},
				},
			},
			wantFound: true,
			wantCycle: []string{"build", "build"},
		},
		"disconnected component graph": {
			graph: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"build":  []string{"format"},
					"lint":   []string{"run"},
					"vet":    []string{"lint"},
					"run":    []string{"vet"},
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
