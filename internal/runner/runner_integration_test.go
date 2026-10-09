package runner

import (
	"context"
	"errors"
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
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, dir := context.Background(), t.TempDir()
			err := Run(ctx, tc.spec, tc.order, dir)

			if err != nil && tc.wantErr == nil {
				t.Fatalf("expected no err, got %v instead", err)
			}

			if err == nil && tc.wantErr != nil {
				t.Fatal("expected an err, got no err instead")
			}

			if err != nil && tc.wantErr != nil {
				if err != nil && tc.wantErr != nil {
					var te *TaskError

					if ok := errors.As(err, &te); !ok {
						t.Errorf("expected error type %T, got %T instead", tc.wantErr, err)
					}

					if te.Task != tc.wantErr.Task {
						t.Errorf("expected to fail on task %s, failed on %s instead", tc.wantErr.Task, te.Task)
					}
				}
			}

			data, err := os.ReadFile(filepath.Join(dir, tc.outputFile))
			if err != nil {
				t.Fatal(err)
			}

			if !(string(data) == tc.wantContent) {
				t.Errorf("expected output file to contain %s, got %s instead", tc.wantContent, string(data))
			}

		})
	}
}
