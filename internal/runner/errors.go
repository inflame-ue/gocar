package runner

import "fmt"

type TaskError struct {
	Task string
	Err  error
}

func (te *TaskError) Error() string {
	return fmt.Sprintf("task %s: failed execution with os reporting %q", te.Task, te.Err.Error())
}

func (te *TaskError) Unwrap() error {
	return te.Err
}
