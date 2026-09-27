package parser

import (
	"errors"
	"fmt"
)

var EmptyDocumentError = errors.New("task specification is empty")

// Indicates a validation problem with task specification during parsing.
type ValidationError struct {
	Task, Field, Msg string
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("task %q, field %q: %s", ve.Task, ve.Field, ve.Msg)
}
