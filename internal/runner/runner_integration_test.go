package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
	"github.com/inflame-ue/gocar/internal/testutil"
)

func TestRunTask_Integration(t *testing.T) {
	tests := map[string]struct {
		name        string
		task        task.Task
		outputFile  string
		wantContent string
		wantErr     *TaskError
	}{
		"valid task": {
			name: "alpha",
			task: task.Task{
				Cmd:     "printf 'alpha' >> a.txt",
				Outputs: []string{"a.txt"},
			},
			outputFile:  "a.txt",
			wantContent: "alpha",
		},
		"failling task": {
			name: "alpha",
			task: task.Task{
				Cmd:    "cat a.txt",
				Inputs: []string{"a.txt"},
			},
			wantErr: &TaskError{
				Task: "alpha",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, dir := context.Background(), t.TempDir()
			err := Run(ctx, tc.name, tc.task, dir)

			tErr := testutil.CheckErrAs(t, err, tc.wantErr)
			if tc.wantErr != nil && tErr.Task != tc.wantErr.Task {
				t.Errorf("expected to fail on task %s, failed on %s instead", tc.wantErr.Task, tErr.Task)
			}

			if tc.outputFile != "" {
				data, err := os.ReadFile(filepath.Join(dir, tc.outputFile))
				if err != nil {
					t.Fatal(err)
				}

				if string(data) != tc.wantContent {
					t.Errorf("expected output file to contain %s, got %s instead", tc.wantContent, string(data))
				}
			}
		})
	}
}
