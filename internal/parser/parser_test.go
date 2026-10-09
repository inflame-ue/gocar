package parser

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
)

type wantError struct {
	is          error
	as          error
	task, field string
}

func stringSliceEqual(a, b []string) bool {
	s1, s2 := slices.Clone(a), slices.Clone(b)

	if len(s1) != len(s2) {
		return false
	}

	slices.Sort(s1)
	slices.Sort(s2)

	return slices.Equal(s1, s2)
}

func checkParserErr(t *testing.T, got error, want *wantError) {
	t.Helper()

	if got == nil && want == nil {
		return
	}

	if got != nil && want == nil {
		t.Fatalf("expected no err, got %v", got)
	}

	if got == nil && want != nil {
		t.Fatal("expected err, got no err instead")
	}

	if want.is != nil && !errors.Is(got, want.is) {
		t.Errorf("expected err type %T, got %T instead", want.is, got)
	}

	if want.task != "" || want.field != "" {
		var ve *ValidationError

		if !errors.As(got, &ve) {
			t.Fatalf("expected ValidationError, got %T", got)
		}

		if ve.Task != want.task {
			t.Errorf("expected task %s, got %s instead", want.task, ve.Task)
		}

		if ve.Field != want.field {
			t.Errorf("expected field %s, got %s instead", want.field, ve.Field)
		}
	}
}

func TestParse(t *testing.T) {
	tests := map[string]struct {
		filename string
		wantErr  *wantError
		want     Tasks
	}{
		"valid YAML": {
			filename: "gocar.yaml",
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
			wantErr: &wantError{
				is: EmptyDocumentError,
			},
		},
		"whitespace-only": {
			filename: "whitespace.yaml",
			wantErr: &wantError{
				is: EmptyDocumentError,
			},
		},
		"comment-only": {
			filename: "comment.yaml",
			wantErr: &wantError{
				is: EmptyDocumentError,
			},
		},
		"empty document": {
			filename: "emptydoc.yaml",
			wantErr: &wantError{
				is: EmptyDocumentError,
			},
		},
		"bad document": {
			filename: "baddoc.yaml",
			wantErr: &wantError{
				as: &DocumentError{},
			},
		},
		"missing cmd": {
			filename: "nocmd.yaml",
			wantErr: &wantError{
				as:    &ValidationError{},
				task:  "build",
				field: "cmd",
			},
		},
		"unknown field": {
			filename: "unknown.yaml",
			wantErr: &wantError{
				as:   &ValidationError{},
				task: "build",
			},
		},
		"duplicate key": {
			filename: "duplicate.yaml",
			wantErr: &wantError{
				as:   &ValidationError{},
				task: "build",
			},
		},
		"empty task": {
			filename: "emptytask.yaml",
			wantErr: &wantError{
				as:   &ValidationError{},
				task: "build",
			},
		},
		"task not a map": {
			filename: "notamap.yaml",
			wantErr: &wantError{
				as:   &ValidationError{},
				task: "build",
			},
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

			checkParserErr(t, err, tc.wantErr)

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
