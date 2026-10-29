package graph

import (
	"maps"
	"slices"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
	"github.com/inflame-ue/gocar/internal/testutil"
)

func TestBuild(t *testing.T) {
	tests := map[string]struct {
		spec    map[string]task.Task
		want    *Graph
		wantErr *MissingDepError
	}{
		"single node graph": {
			spec: map[string]task.Task{
				"build": {
					Cmd: "go build .",
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"build": {},
				},
				rdeps: map[string][]string{
					"build": {},
				},
				indeg: map[string]int{
					"build": 0,
				},
			},
			wantErr: nil,
		},
		"valid linked list graph": {
			spec: map[string]task.Task{
				"format": {
					Cmd: "go fmt ./...",
				},
				"build": {
					Cmd:  "go build ./...",
					Deps: []string{"format"},
				},
			},
			want: &Graph{
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
			wantErr: nil,
		},
		"valid diamond graph": {
			spec: map[string]task.Task{
				"format": {
					Cmd: "go fmt ./...",
				},
				"vet": {
					Cmd:  "go vet ./...",
					Deps: []string{"format"},
				},
				"lint": {
					Cmd:  "golangci-lint run",
					Deps: []string{"format"},
				},
				"build": {
					Cmd:  "go build ./...",
					Deps: []string{"vet", "lint"},
				},
			},
			want: &Graph{
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
		},
		"missing dep graph": {
			spec: map[string]task.Task{
				"build": {
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
				"format": {
					Cmd: "go fmt ./...",
				},
				"build": {
					Cmd:  "go build ./...",
					Deps: []string{"format", "format"},
				},
			},
			want: &Graph{
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
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := Build(tc.spec)

			tErr := testutil.CheckErrAs(t, err, tc.wantErr)
			if tc.wantErr != nil && tErr.Task != tc.wantErr.Task {
				t.Errorf("expected task %s for error, got %s instead", tc.wantErr.Task, tErr.Task)
			}

			if tc.wantErr != nil && tErr.Dep != tc.wantErr.Dep {
				t.Errorf("expected dep %s, got %s instead", tc.wantErr.Dep, tErr.Dep)
			}

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
