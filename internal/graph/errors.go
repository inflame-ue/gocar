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
	var builder strings.Builder
	
	for _, node := range ce.Path {
		_, _ = builder.WriteString(fmt.Sprintf("%s -> ", node))
	}

	return builder.String()
}
