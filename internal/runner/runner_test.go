package runner

import (
	"context"
	"errors"
	"io"
	"slices"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
	"github.com/inflame-ue/gocar/internal/testutil"
)

type testCommander struct {
	ran    []string
	failOn string
}

func (tc *testCommander) run(ctx context.Context, cmd, dir string, stdout, stderr io.Writer) error {
	if tc.failOn == cmd {
		return errors.New("faulty command")
	}

	tc.ran = append(tc.ran, cmd)
	return nil
}

func TestRun(t *testing.T) {
	tests := map[string]struct {
		spec    map[string]task.Task
		order   []string
		failOn  string
		wantRan []string
		wantErr *TaskError
	}{
		"valid task": {
			spec: map[string]task.Task{
				"fmt": {
					Cmd: "go fmt .",
				},
				"lint": {
					Cmd: "go lint .",
				},
				"run": {
					Cmd:  "go run .",
					Deps: []string{"fmt", "lint"},
				},
			},
			order:   []string{"fmt", "lint", "run"},
			failOn:  "",
			wantRan: []string{"go fmt .", "go lint .", "go run ."},
			wantErr: nil,
		},
		"fast fail task": {
			spec: map[string]task.Task{
				"fmt": {
					Cmd: "go fmt .",
				},
				"run": {
					Cmd:  "go run .",
					Deps: []string{"fmt"},
				},
				"build": {
					Cmd:  "go build .",
					Deps: []string{"run"},
				},
			},
			order:   []string{"fmt", "run", "build"},
			failOn:  "go run .",
			wantRan: []string{"go fmt ."},
			wantErr: &TaskError{
				Task: "run",
			},
		},
		"unknown task in order": {
			spec: map[string]task.Task{
				"fmt": {
					Cmd: "go fmt .",
				},
				"run": {
					Cmd:  "go run .",
					Deps: []string{"fmt"},
				},
				"build": {
					Cmd:  "go build .",
					Deps: []string{"run"},
				},
			},
			order:   []string{"fmt", "run", "lint", "build"},
			failOn:  "",
			wantRan: []string{"go fmt .", "go run ."},
			wantErr: &TaskError{
				Task: "lint",
			},
		},
		"empty spec and order": {
			spec:    map[string]task.Task{},
			order:   []string{},
			failOn:  "",
			wantRan: []string{},
			wantErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cmder := &testCommander{
				ran:    []string{},
				failOn: tc.failOn,
			}
			r := Runner{cmder: cmder}
			ctx := context.Background()
			err := r.Run(ctx, tc.spec, tc.order)

			tErr := testutil.CheckErrAs(t, err, tc.wantErr)
			if tc.wantErr != nil && tErr.Task != tc.wantErr.Task {
				t.Errorf("expected to fail on task %s, failed on %s instead", tc.wantErr.Task, tErr.Task)
			}

			if !slices.Equal(cmder.ran, tc.wantRan) {
				t.Errorf("expected %v to run, got %v instead", tc.wantRan, cmder.ran)
			}
		})
	}
}
