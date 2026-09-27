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

var EmptyDocumentError = errors.New("task specification is empty")

func Parse(data []byte) (Tasks, error) {
	if len(data) == 0 {
		return nil, EmptyDocumentError
	}

	raw, err := parseRawTasks(data)
	if err != nil {
		return nil, err
	}

	return parseTasks(raw)
}

func parseRawTasks(data []byte) (rawTasks, error) {
	var raw map[string]yaml.Node

	r := bytes.NewReader(data)
	d := yaml.NewDecoder(r)

	if err := d.Decode(&raw); err != nil {
		return nil, err
	}

	return raw, nil
}

func parseTasks(raw rawTasks) (Tasks, error) {
	var tasks = make(Tasks)

	for name, node := range raw {
		var task task.Task

		if err := node.Decode(&task); err != nil {
			return nil, fmt.Errorf("task %q: %w", name, err)
		}

		tasks[name] = task
	}

	return tasks, nil
}
