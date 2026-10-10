package fingerprint

import (
	"errors"
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
)

func checkInputErr(t *testing.T, got error, wantErr *InputError) {
	t.Helper()

	if got == nil && wantErr != nil {
		t.Fatal("expected an err, got no err instead")
	}

	if got == nil && wantErr != nil {
		t.Fatalf("expected no err, got %v instead", got)
	}

	var ie *InputError

	if ok := errors.As(got, &ie); !ok {
		t.Errorf("expected err of type %T, got %T instead", wantErr, got)
	}

	if ie.Task != wantErr.Task {
		t.Errorf("expected task %s, got %s instead", wantErr.Task, ie.Task)
	}

	if ie.Path != wantErr.Path {
		t.Errorf("expected path %s, got %s instead", wantErr.Path, ie.Path)
	}
}

func TestKeys(t *testing.T) {
	tests := map[string]struct {
		spec    map[string]task.Task
		order   []string
		fp      *Fingerprinter
		wantErr *InputError
	}{}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			keys, err := tc.fp.Keys(tc.spec, tc.order)

			checkInputErr(t, err, tc.wantErr)

			
		})
	}
}
