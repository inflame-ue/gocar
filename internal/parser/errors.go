package parser

import (
	"errors"
	"fmt"
)

var ErrEmptyDocument = errors.New("task specification is empty")

// DocumentError indicates a decoding/loading problem with a particular YAML
// configuration. 
type DocumentError struct {
	Msg   string
	Cause error
}

func (de *DocumentError) Error() string {
	if de.Cause != nil {
		return de.Cause.Error()
	}

	return de.Msg
}

func (de *DocumentError) Unwrap() error {
	return de.Cause
}

// ValidationError indicates a validation problem with a particular task
// specification during parsing. 
type ValidationError struct {
	Task, Field, Msg string
	Cause            error
}

func (ve *ValidationError) Error() string {
	prefix := fmt.Sprintf("task %q", ve.Task)
	if ve.Field != "" {
		prefix += fmt.Sprintf(", field %q", ve.Field)
	}

	if ve.Cause != nil {
		return fmt.Sprintf("%s: %v", prefix, ve.Cause)
	}

	return fmt.Sprintf("%s: %s", prefix, ve.Msg)
}

func (ve *ValidationError) Unwrap() error {
	return ve.Cause
}

func errTask(parseErr error) string {
	err, ok := errors.AsType[*ValidationError](parseErr)

	if !ok {
		return ""
	}

	return err.Task
}
