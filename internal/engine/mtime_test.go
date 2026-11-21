package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/inflame-ue/gocar/internal/task"
	"github.com/inflame-ue/gocar/internal/testutil"
)

func writeAt(t *testing.T, dir, name, content string, mtime time.Time) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}

	return name
}

func TestStale(t *testing.T) {
	t.Parallel()

	base := time.Unix(1_700_000_000, 0)
	old := base
	newer := base.Add(time.Minute)

	t.Run("newer input rebuilds", func(t *testing.T) {
		dir := t.TempDir()
		e := &Engine{Dir: dir}
		spec := task.Tasks{"A": {
			Cmd:     "x",
			Inputs:  []string{"in.txt"},
			Outputs: []string{"out.txt"},
		}}

		writeAt(t, dir, "in.txt", "v", newer)
		writeAt(t, dir, "out.txt", "v", old)

		stale, err := e.stale("A", spec["A"], spec)
		testutil.CheckNoErr(t, err)
		if !stale {
			t.Errorf("expected the task to be stale")
		}
	})

	t.Run("newer output skips", func(t *testing.T) {
		dir := t.TempDir()
		e := &Engine{Dir: dir}
		spec := task.Tasks{"A": {
			Cmd:     "x",
			Inputs:  []string{"in.txt"},
			Outputs: []string{"out.txt"},
		}}

		writeAt(t, dir, "in.txt", "v", old)
		writeAt(t, dir, "out.txt", "v", newer)

		stale, err := e.stale("A", spec["A"], spec)
		testutil.CheckNoErr(t, err)
		if stale {
			t.Errorf("expected the task to be up-to-date")
		}
	})

	t.Run("missing output rebuilds", func(t *testing.T) {
		dir := t.TempDir()
		e := &Engine{Dir: dir}
		spec := task.Tasks{"A": {
			Cmd:     "x",
			Inputs:  []string{"in.txt"},
			Outputs: []string{"out.txt"},
		}}

		writeAt(t, dir, "in.txt", "v", old)

		stale, err := e.stale("A", spec["A"], spec)
		testutil.CheckNoErr(t, err)
		if !stale {
			t.Errorf("expected a missing output to force a rebuild")
		}
	})

	t.Run("missing input errors", func(t *testing.T) {
		dir := t.TempDir()
		e := &Engine{Dir: dir}
		spec := task.Tasks{"A": {
			Cmd:     "x",
			Inputs:  []string{"missing.txt"},
			Outputs: []string{"out.txt"},
		}}

		writeAt(t, dir, "out.txt", "v", old)

		if _, err := e.stale("A", spec["A"], spec); err == nil {
			t.Errorf("expected a missing input to return an error")
		}
	})

	t.Run("newer dependency output rebuilds", func(t *testing.T) {
		dir := t.TempDir()
		e := &Engine{Dir: dir}
		spec := task.Tasks{
			"A": {Cmd: "x", Outputs: []string{"a.txt"}},
			"B": {Cmd: "x", Outputs: []string{"b.txt"}, Deps: []string{"A"}},
		}

		writeAt(t, dir, "a.txt", "v", newer)
		writeAt(t, dir, "b.txt", "v", old)

		stale, err := e.stale("B", spec["B"], spec)
		testutil.CheckNoErr(t, err)
		if !stale {
			t.Errorf("expected a newer dependency output to force a rebuild")
		}
	})

	t.Run("older dependency output skips", func(t *testing.T) {
		dir := t.TempDir()
		e := &Engine{Dir: dir}
		spec := task.Tasks{
			"A": {Cmd: "x", Outputs: []string{"a.txt"}},
			"B": {Cmd: "x", Outputs: []string{"b.txt"}, Deps: []string{"A"}},
		}

		writeAt(t, dir, "a.txt", "v", old)
		writeAt(t, dir, "b.txt", "v", newer)

		stale, err := e.stale("B", spec["B"], spec)
		testutil.CheckNoErr(t, err)
		if stale {
			t.Errorf("expected the task to be up-to-date")
		}
	})

	t.Run("missing dependency output rebuilds", func(t *testing.T) {
		dir := t.TempDir()
		e := &Engine{Dir: dir}
		spec := task.Tasks{
			"A": {Cmd: "x", Outputs: []string{"a.txt"}},
			"B": {Cmd: "x", Outputs: []string{"b.txt"}, Deps: []string{"A"}},
		}

		writeAt(t, dir, "b.txt", "v", old)

		stale, err := e.stale("B", spec["B"], spec)
		testutil.CheckNoErr(t, err)
		if !stale {
			t.Errorf("expected a missing dependency output to force a rebuild")
		}
	})
}
