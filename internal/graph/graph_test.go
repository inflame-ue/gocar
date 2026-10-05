package graph

import (
	"testing"

	"github.com/inflame-ue/gocar/internal/task"
)

func TestBuild(t *testing.T) {
	tests := map[string]struct {
		spec    map[string]task.Task
		want    *Graph
		wantErr *MissingDepError
	}{
		"single node graph": {
			spec: map[string]task.Task{
				"build": task.Task{
					Cmd: "go build .",
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"build": []string{},
				},
				rdeps: map[string][]string{
					"build": []string{},
				},
				indeg: map[string]int{
					"build": 0,
				},
			},
			wantErr: nil,
		},
		"valid linked list graph": {
			spec: map[string]task.Task{
				"format": task.Task{
					Cmd: "go fmt ./...",
				},
				"build": task.Task{
					Cmd:  "go build ./...",
					Deps: []string{"format"},
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"build":  []string{"format"},
				},
				rdeps: map[string][]string{
					"format": []string{"build"},
					"build":  []string{},
				},
				indeg: map[string]int{
					"format": 0,
					"build":  1,
				},
			},
			wantErr: nil,
		},
		"valid diamond graph": {
			spec: map[string]task.Task{
				"format": task.Task{
					Cmd: "go fmt ./...",
				},
				"vet": task.Task{
					Cmd:  "go vet ./...",
					Deps: []string{"format"},
				},
				"lint": task.Task{
					Cmd:  "golangci-lint run",
					Deps: []string{"format"},
				},
				"build": task.Task{
					Cmd:  "go build ./...",
					Deps: []string{"vet", "lint"},
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"vet":    []string{"format"},
					"lint":   []string{"format"},
					"build":  []string{"vet", "lint"},
				},
				rdeps: map[string][]string{
					"format": []string{"vet", "lint"},
					"vet":    []string{"build"},
					"lint":   []string{"build"},
					"build":  []string{},
				},
				indeg: map[string]int{
					"format": 0,
					"vet":    1,
					"lint":   1,
					"build":  2,
				},
			},
		},
		"missing dep graph": {
			spec: map[string]task.Task{
				"build": task.Task{
					Cmd: "go build ./...",
					Deps: []string{"format"},
				},
			},
			want: nil,
			wantErr: &MissingDepError{
				Task: "build",
				Dep: "format",
			},
		},
		"duplicate deps graph": {
			spec: map[string]task.Task{
				"format": task.Task{
					Cmd: "go fmt ./...",
				},
				"build": task.Task{
					Cmd: "go build ./...",
					Deps: []string{"format", "format"},
				},
			},
			want: &Graph{
				deps: map[string][]string{
					"format": []string{},
					"build": []string{"format"},
				},
				rdeps: map[string][]string{
					"format": []string{"build"},
					"build": []string{},
				},
				indeg: map[string]int{
					"format": 0,
					"build": 1,
				},
			},
		},
	}
}
