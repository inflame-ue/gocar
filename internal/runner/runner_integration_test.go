package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
	"github.com/inflame-ue/gocar/internal/testutil"
)

func TestRunIntegration(t *testing.T) {
	tests := map[string]struct {
		spec        map[string]task.Task
		order       []string
		outputFile  string
		wantContent string
		wantErr     *TaskError
	}{
		"valid task": {
			spec: map[string]task.Task{
				"alpha": {
					Cmd: "printf 'alpha' >> a.txt",
				},
				"beta": {
					Cmd: "printf 'beta' >> a.txt",
				},
			},
			order:       []string{"alpha", "beta"},
			outputFile:  "a.txt",
			wantContent: "alphabeta",
			wantErr:     nil,
		},
		"failling task": {
			spec: map[string]task.Task{
				"alpha": {
					Cmd: "printf 'alpha' >> a.txt",
				},
				"beta": {
					Cmd: "cat b.txt",
				},
			},
			order: []string{"alpha", "beta"},
			wantErr: &TaskError{
				Task: "beta",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, dir := context.Background(), t.TempDir()
			err := Run(ctx, tc.spec, tc.order, dir)

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
