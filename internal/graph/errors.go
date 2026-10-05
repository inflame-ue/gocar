package graph

import "fmt"

type MissingDepError struct {
	Task, Dep string
}

func (mde *MissingDepError) Error() string {
	return fmt.Sprintf("task %s: depedency %s does not exist", mde.Task, mde.Dep)
}
