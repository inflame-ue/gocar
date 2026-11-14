package scheduler

import (
	"container/heap"
	"maps"
	"slices"
	"testing"

	"github.com/inflame-ue/gocar/internal/graph"
	"github.com/inflame-ue/gocar/internal/task"
	"github.com/inflame-ue/gocar/internal/testutil"
)

func levelOf(t *testing.T, spec task.Tasks) map[string]int {
	t.Helper()

	g, err := graph.Build(spec)
	testutil.CheckNoErr(t, err)

	return Level(g)
}

func TestLevel_Chain(t *testing.T) {
	t.Parallel()

	spec := task.Tasks{
		"format": {Cmd: "go fmt ."},
		"vet":    {Cmd: "go vet .", Deps: []string{"format"}},
		"build":  {Cmd: "go build .", Deps: []string{"vet"}},
	}
	want := map[string]int{"format": 3, "vet": 2, "build": 1}

	if got := levelOf(t, spec); !maps.Equal(got, want) {
		t.Errorf("Level() = %v, want %v", got, want)
	}
}

func TestLevel_Diamond(t *testing.T) {
	t.Parallel()

	spec := task.Tasks{
		"format": {Cmd: "go fmt ."},
		"vet":    {Cmd: "go vet .", Deps: []string{"format"}},
		"lint":   {Cmd: "golangci-lint run", Deps: []string{"format"}},
		"build":  {Cmd: "go build .", Deps: []string{"vet", "lint"}},
	}
	want := map[string]int{"format": 3, "vet": 2, "lint": 2, "build": 1}

	if got := levelOf(t, spec); !maps.Equal(got, want) {
		t.Errorf("Level() = %v, want %v", got, want)
	}
}

func TestLevel_Disconnected(t *testing.T) {
	t.Parallel()

	spec := task.Tasks{
		"format": {Cmd: "go fmt ."},
		"vet":    {Cmd: "go vet .", Deps: []string{"format"}},
		"build":  {Cmd: "go build .", Deps: []string{"vet"}},
		"lint":   {Cmd: "golangci-lint run"},
		"run":    {Cmd: "go run .", Deps: []string{"lint"}},
	}
	want := map[string]int{"format": 3, "vet": 2, "build": 1, "lint": 2, "run": 1}

	if got := levelOf(t, spec); !maps.Equal(got, want) {
		t.Errorf("Level() = %v, want %v", got, want)
	}
}

func TestReadyQueue_Order(t *testing.T) {
	t.Parallel()

	level := map[string]int{
		"A": 1,
		"B": 3,
		"C": 2,
		"D": 2,
		"E": 3,
	}

	q := NewReadyQueue(level)
	for _, name := range []string{"C", "A", "E", "B", "D"} {
		heap.Push(q, name)
	}

	var got []string
	for q.Len() > 0 {
		got = append(got, heap.Pop(q).(string))
	}

	want := []string{"B", "E", "C", "D", "A"}
	if !slices.Equal(got, want) {
		t.Errorf("pop order = %v, want %v", got, want)
	}
}
