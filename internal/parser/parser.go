package parser

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"

	"github.com/inflame-ue/gocar/internal/task"
	"go.yaml.in/yaml/v4"
)

type Tasks task.Tasks
type rawTasks map[string]yaml.Node

func Parse(data []byte) (task.Tasks, error) {
	raw, err := parseRawTasks(data)
	if err != nil {
		return nil, err
	}

	return parseTasks(raw)
}

func parseRawTasks(data []byte) (rawTasks, error) {
	var raw map[string]yaml.Node

	if len(data) == 0 {
		return nil, ErrEmptyDocument
	}

	r := bytes.NewReader(data)
	d := yaml.NewDecoder(r)

	if err := d.Decode(&raw); err != nil {

		// no valid YAML found
		if errors.Is(err, io.EOF) {
			return nil, ErrEmptyDocument
		}

		return nil, &DocumentError{
			Cause: err,
			Msg:   err.Error(),
		}
	}

	if len(raw) == 0 {
		return nil, ErrEmptyDocument
	}

	return raw, nil
}

func parseTasks(raw rawTasks) (task.Tasks, error) {
	var tasks = make(task.Tasks)
	var errs []error

	// reminder: one error per-task
nodeloop:
	for name, node := range raw {
		var task task.Task

		if err := node.Load(&task, yaml.WithKnownFields(), yaml.WithUniqueKeys()); err != nil {
			errs = append(errs, &ValidationError{
				Task:  name,
				Field: "",
				Cause: err,
				Msg:   "task specification cannot be loaded",
			})
			continue
		}

		if reflect.ValueOf(task).IsZero() {
			errs = append(errs, &ValidationError{
				Task:  name,
				Field: "",
				Msg:   "task specification cannot be empty",
			})
			continue
		}

		if task.Cmd == "" {
			errs = append(errs, &ValidationError{
				Task:  name,
				Field: "cmd",
				Msg:   "cmd specification cannot be empty",
			})
			continue
		}

		for _, output := range task.Outputs {
			if strings.ContainsAny(output, "*?[") {
				errs = append(errs, &ValidationError{
					Task:  name,
					Field: "outputs",
					Msg:   "outputs do not support glob expansion",
				})
				continue nodeloop
			}
		}

		tasks[name] = task
	}

	// the build system cannot support duplicate outputs,
	// because of the potential write races and cache corruptions
	if err := duplicateOutputs(tasks); err != nil {
		errs = append(errs, err)
	}

	// map iteration is random in Go, and we want errs to be in the same order
	// hence we sort on err.Task to ensure idempotency between runs if no changes happen
	slices.SortFunc(errs, func(errA error, errB error) int {
		return cmp.Compare(errTask(errA), errTask(errB))
	})

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return tasks, nil
}

func duplicateOutputs(spec task.Tasks) error {
	seen := make(map[string]bool)
	names := spec.Names()
	for _, name := range names {
		for _, output := range spec[name].Outputs {
			if seen[output] {
				return &ValidationError{
					Task:  name,
					Field: "outputs",
					Msg:   fmt.Sprintf("output %q is a duplicate", output),
				}
			}
			seen[output] = true
		}
	}
	return nil
}
