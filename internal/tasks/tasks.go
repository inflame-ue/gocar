package tasks

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

func readTasks(filepath string) ([]byte, error) {
	// the tasks file will never be large, hence read to memory
	data, err := os.ReadFile(filepath)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func ParseTasks(filepath string) (map[interface{}]interface{}, error) {
	data, err := readTasks(filepath)
	if err != nil {
		return map[interface{}]interface{}{}, err
	}

	// the structure of the file is dynamic
	taskSpec := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(data, &taskSpec); err != nil {
		return map[interface{}]interface{}{}, fmt.Errorf("failed to unmarhsal: %v", err)
	}

	return taskSpec, nil
}
