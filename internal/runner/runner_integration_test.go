package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
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
				"alpha": task.Task{
					Cmd: "printf 'alpha' >> a.txt",
				},
				"beta": task.Task{
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
				"alpha": task.Task{
					Cmd: "printf 'alpha' >> a.txt",
				},
				"beta": task.Task{
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

			checkRunnerErr(t, err, tc.wantErr)

			if tc.outputFile != "" {
				data, err := os.ReadFile(filepath.Join(dir, tc.outputFile))
				if err != nil {
					t.Fatal(err)
				}

				if !(string(data) == tc.wantContent) {
					t.Errorf("expected output file to contain %s, got %s instead", tc.wantContent, string(data))
				}
			}
		})
	}
}
