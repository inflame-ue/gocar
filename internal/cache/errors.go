package cache

import "fmt"

type PathError struct {
	Path, Msg string
}

func (pe *PathError) Error() string {
	return fmt.Sprintf("path %q: %s", pe.Path, pe.Msg)
}
