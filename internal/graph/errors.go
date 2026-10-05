package graph

import "fmt"

type MissingDepError struct {
	Task, Dep string
}

func (mde *MissingDepError) Error() string {
	return fmt.Sprintf("task %s depends on non-existent %s", mde.Task, mde.Dep)
}
