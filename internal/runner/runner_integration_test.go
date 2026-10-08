package runner

import (
	"context"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
)

func TestRunIntegration(t *testing.T) {
	dir := t.TempDir()
	tests := map[string]struct {
		spec    map[string]task.Task
		order   []string
		wantErr *TaskError
	}{
		"valid task": {
			spec: map[string]task.Task{
				"alpha": task.Task{
					Cmd: "echo 'alpha' > a.txt",
				},
				"beta": task.Task{
					Cmd: "echo 'beta' > b.txt",
				},
				"merge": task.Task{
					Cmd:  "cat a.txt b.txt > all.txt",
					Deps: []string{"alpha", "beta"},
				},
			},
			order:   []string{"alpha", "beta", "merge"},
			wantErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := Run(ctx, tc.spec, tc.order)

			if err != nil && tc.wantErr == nil {
				t.Fatalf("expected no err, got %v instead", err)
			}

			if err == nil && tc.wantErr != nil {
				t.Fatal("expected an err, got no err instead")
			}

			if err != nil && tc.wantErr != nil {
				
			}

			
			
		})
	}
}
