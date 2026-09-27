package parser

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/inflame-ue/gocar/internal/task"
	"go.yaml.in/yaml/v4"
)

type Tasks map[string]task.Task
type rawTasks map[string]yaml.Node

var EmptyTaskError = errors.New("task specification is empty")

func ParseRawTasks(data []byte) (rawTasks, error) {
	var raw map[string]yaml.Node

	r := bytes.NewReader(data)
	d := yaml.NewDecoder(r)
	d.KnownFields(true)

	if err := d.Decode(&raw); err != nil {
		return nil, err
	}

	return raw, nil
}

func ParseTasks(raw rawTasks) (Tasks, error) {
	var tasks Tasks

	for name, node := range raw {
		var task task.Task

		if node.IsZero() {
			return nil, EmptyTaskError
		}

		if err := node.Decode(&task); err != nil {
			return nil, fmt.Errorf("task %q: %w", name, err)
		}

		tasks[name] = task
	}

	return tasks, nil
}
