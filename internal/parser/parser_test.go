package parser

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
)

func stringSliceEqual(a, b []string) bool {
	s1, s2 := slices.Clone(a), slices.Clone(b)

	if len(s1) != len(s2) {
		return false
	}

	slices.Sort(s1)
	slices.Sort(s2)

	return slices.Equal(s1, s2)
}

func TestParse(t *testing.T) {
	tests := map[string]struct {
		filename string
		wantErr  error
		want     Tasks
	}{
		"valid YAML": {
			filename: "gocar.yaml",
			wantErr:  nil,
			want: Tasks{
				"mkout": task.Task{
					Cmd: "mkdir out",
				},
				"alpha": task.Task{
					Cmd:     `echo 'alpha\n' > out/a.txt`,
					Deps:    []string{"mkout"},
					Outputs: []string{"out/a.txt"},
				},
				"beta": task.Task{
					Cmd:     `echo 'beta\n' > out/b.txt`,
					Deps:    []string{"mkout"},
					Outputs: []string{"out/b.txt"},
				},
				"merge": task.Task{
					Cmd:     "cat out/a.txt out/b.txt > out/all.txt",
					Deps:    []string{"alpha", "beta"},
					Inputs:  []string{"out/a.txt", "out/b.txt"},
					Outputs: []string{"out/all.txt"},
				},
				"check": task.Task{
					Cmd:    "wc -l out/all.txt",
					Deps:   []string{"merge"},
					Inputs: []string{"out/all.txt"},
				},
			},
		},
		"empty": {
			filename: "empty.yaml",
			wantErr:  EmptyDocumentError,
		},
		"whitespace-only": {
			filename: "whitespace.yaml",
			wantErr:  EmptyDocumentError,
		},
		"comment-only": {
			filename: "comment.yaml",
			wantErr:  EmptyDocumentError,
		},
		"empty document": {
			filename: "emptydoc.yaml",
			wantErr:  EmptyDocumentError,
		},
		"bad document": {
			filename: "baddoc.yaml",
			wantErr:  &DocumentError{},
		},
		"missing cmd": {
			filename: "nocmd.yaml",
			wantErr:  &ValidationError{},
		},
		"unknown field": {
			filename: "unknown.yaml",
			wantErr:  &ValidationError{},
		},
		"duplicate key": {
			filename: "duplicate.yaml",
			wantErr:  &ValidationError{},
		},
		"empty task": {
			filename: "emptytask.yaml",
			wantErr:  &ValidationError{},
		},
		"task not a map": {
			filename: "notamap.yaml",
			wantErr:  &ValidationError{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			file := filepath.Join("testdata", tc.filename)
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("reading yaml file: %v", err)
			}

			tasks, err := Parse(data)
			if err != nil && tc.wantErr == nil {
				t.Fatalf("expected no err, got %v", err)
			}

			if err == nil && tc.wantErr != nil {
				t.Fatal("expected err, got no err instead")
			}

			if err != nil && tc.wantErr != nil {
				// TODO: implement the error type checking
				// return early to not compare
				return
			}

			if len(tasks) != len(tc.want) {
				t.Errorf("expected %d tasks, got %d instead", len(tc.want), len(tasks))
			}

			for name, got := range tasks {
				want := tc.want[name]

				if _, ok := tc.want[name]; !ok {
					t.Errorf("task %s: expected to not be in tasks", name)
				}

				if got.Cmd != want.Cmd {
					t.Errorf("task %s: expected command to be %s, got %s instead", name, want.Cmd, got.Cmd)
				}

				if !stringSliceEqual(got.Deps, want.Deps) {
					t.Errorf("task %s: dependency list is not the same", name)
				}

				if !stringSliceEqual(got.Inputs, want.Inputs) {
					t.Errorf("task %s: input files list is not the same", name)
				}

				if !stringSliceEqual(got.Outputs, want.Outputs) {
					t.Errorf("task %s: output files list is not the same", name)
				}
			}
		})
	}
}
