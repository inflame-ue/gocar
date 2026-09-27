package parser

import (
	"errors"
	"fmt"
)

var EmptyDocumentError = errors.New("task specification is empty")

// Indicates a decoding/loading problem with a particular YAML configuration.
type DocumentError struct {
	Msg   string
	Cause error
}

func (de *DocumentError) Error() string {
	return de.Msg
}

// Indicates a validation problem with a particular task specification during parsing.
type ValidationError struct {
	Task, Field, Msg string
	Cause            error
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("task %q, field %q: %s", ve.Task, ve.Field, ve.Msg)
}
