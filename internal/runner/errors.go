package runner

import "fmt"

type TaskError struct {
	Task string
	Err  error
}

func (te *TaskError) Error() string {
	return fmt.Sprintf("task %s: %v", te.Task, te.Err)
}

func (te *TaskError) Unwrap() error {
	return te.Err
}
