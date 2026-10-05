package parser

import (
	"errors"
	"fmt"
)

var EmptyDocumentError = errors.New("task specification is empty")

// DocumentError indicates a decoding/loading problem with a particular YAML
// configuration. Implements the ParserError interface for convenience.
type DocumentError struct {
	Msg   string
	Cause error
}

func (de *DocumentError) Error() string {
	if de.Cause == nil {
		return fmt.Sprintf("msg: %s", de.Msg)
	}

	// yaml fails here -> Cause msg
	return fmt.Sprintf("cause: %s; msg: %s", de.Cause.Error(), de.Msg)
}

func (de *DocumentError) Unwrap() error {
	return de.Cause
}

// ValidationError indicates a validation problem with a particular task
// specification during parsing. Implements the ParserError interface for convenience.
type ValidationError struct {
	Task, Field, Msg string
	Cause            error
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("task %q, field %q: %s", ve.Task, ve.Field, ve.Msg)
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
