package graph

import (
	"fmt"
	"strings"
)

type MissingDepError struct {
	Task, Dep string
}

func (mde *MissingDepError) Error() string {
	return fmt.Sprintf("task %s depends on non-existent %s", mde.Task, mde.Dep)
}

type CycleError struct {
	Path []string
}

func (ce *CycleError) Error() string {
	return strings.Join(ce.Path, " -> ")
}
