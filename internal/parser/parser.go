package parser

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/inflame-ue/gocar/internal/task"
	"go.yaml.in/yaml/v4"
)

type Tasks map[string]task.Task
type rawTasks map[string]yaml.Node

func Parse(data []byte) (Tasks, error) {
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

	if len(raw) == 0 {
		return nil, EmptyDocumentError
	}

	return raw, nil
}

func parseTasks(raw rawTasks) (Tasks, error) {
	var tasks = make(Tasks)

	for name, node := range raw {
		var task task.Task

		if err := node.Load(&task, yaml.WithKnownFields(), yaml.WithUniqueKeys()); err != nil {
			return nil, fmt.Errorf("task %q: %w", name, err)
		}

		if task.Cmd == "" {
			return nil, &ValidationError{
				Task:  name,
				Field: "cmd",
				Msg:   "cmd specification cannot be empty",
			}
		}

		if reflect.ValueOf(task).IsZero() {
			return nil, &ValidationError{
				Task:  name,
				Field: "",
				Msg:   "task specification cannot be empty",
			}
		}

		tasks[name] = task
	}

	return tasks, nil
}
