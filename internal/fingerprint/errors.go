package fingerprint

import "fmt"

type InputError struct {
	Task, Path string
	Cause      error
}

func (ie *InputError) Error() string {
	return fmt.Sprintf("task %s, input %s: %v", ie.Task, ie.Path, ie.Cause)
}

func (ie *InputError) Unwrap() error {
	return ie.Cause
}
