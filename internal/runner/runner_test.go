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

func TestRunTask(t *testing.T) {
	tests := map[string]struct {
		task    task.Task
		failOn  string
		wantRan []string
		wantErr *TaskError
	}{
		"success": {
			task: task.Task{
				Cmd: "go fmt .",
			},
			failOn:  "",
			wantRan: []string{"go fmt ."},
			wantErr: nil,
		},
		"failure": {
			task: task.Task{
				Cmd: "go run .",
			},
			failOn:  "go run .",
			wantRan: []string{},
			wantErr: &TaskError{
				Task: "run",
			},
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

			err := r.RunTask(ctx, "run", tc.task)

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
