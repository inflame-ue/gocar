package parser

import (
	"errors"

	"github.com/inflame-ue/gocar/internal/task"
	"go.yaml.in/yaml/v4"
)

type Tasks []task.Task

var RuntimeUnmarhsalError = errors.New("parsing of the YAML file with task specification failed")

func ParseTaskSpecification(data []byte) (Tasks, error) {
	var tasks Tasks

	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return nil, RuntimeUnmarhsalError
	}

	return tasks, nil
}
