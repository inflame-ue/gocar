package parser

import (
	"errors"
	"fmt"
)

var EmptyDocumentError = errors.New("task specification is empty")

// Describes either a DocumentError or a ValidationError.
// Use Unwrap to access the underlying yamlv4 library error.
type ParserError interface {
	error
	Unwrap() error
}

// Implements the ParserError interface.
// Indicates a decoding/loading problem with a particular YAML configuration.
type DocumentError struct {
	Msg   string
	Cause error
}

func (de *DocumentError) Error() string {
	// yaml fails here -> Cause msg
	return fmt.Sprintf("cause: %s; msg: %s", de.Cause.Error(), de.Msg)
}

func (de *DocumentError) Unwrap() error {
	return de.Cause
}

// Implement the ParserError interface.
// Indicates a validation problem with a particular task specification during parsing.
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
