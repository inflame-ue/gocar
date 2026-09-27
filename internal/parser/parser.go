package parser

import (
	"bytes"
	"errors"
	"io"
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

	if len(data) == 0 {
		return nil, EmptyDocumentError
	}

	r := bytes.NewReader(data)
	d := yaml.NewDecoder(r)

	if err := d.Decode(&raw); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, EmptyDocumentError
		}
		return nil, &ValidationError{
			Task:  "",
			Field: "",
			Msg:   err.Error(),
		}
	}

	return raw, nil
}

func parseTasks(raw rawTasks) (Tasks, error) {
	var tasks = make(Tasks)
	var errs []error

	for name, node := range raw {
		var task task.Task

		if err := node.Load(&task, yaml.WithKnownFields(), yaml.WithUniqueKeys()); err != nil {
			errs = append(errs, &ValidationError{
				Task:  name,
				Field: "",
				Msg:   "task specification cannot be loaded",
			})
		}

		if reflect.ValueOf(task).IsZero() {
			errs = append(errs, &ValidationError{
				Task:  name,
				Field: "",
				Msg:   "task specification cannot be empty",
			})
		}

		if task.Cmd == "" {
			errs = append(errs, &ValidationError{
				Task:  name,
				Field: "cmd",
				Msg:   "cmd specification cannot be empty",
			})
		}

		tasks[name] = task
	}

	return tasks, errors.Join(errs...)
}