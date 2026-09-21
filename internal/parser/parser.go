package parser

import (
	"github.com/inflame-ue/gocar/internal/task"
	"go.yaml.in/yaml/v4"
)

type Tasks map[string]task.Task

func Parse(data []byte) (Tasks, error) {
	var tasks Tasks

	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}
