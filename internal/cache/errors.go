package cache

import "fmt"

type PathError struct {
	Path, Msg string
}

func (pe *PathError) Error() string {
	return fmt.Sprintf("path %q: %s", pe.Path, pe.Msg)
}

type IOError struct {
	Path  string
	Cause error
}

func (e *IOError) Error() string { 
	return fmt.Sprintf("path %q: %v", e.Path, e.Cause) 
}

func (e *IOError) Unwrap() error { 
	return e.Cause 
}