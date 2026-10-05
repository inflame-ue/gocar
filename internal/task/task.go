package task

// Task is the parsed representation of a single build task
// that composes four different YAML fields from gocar.yaml.
// Deps, Inputs, Outputs can be nil, if the corresponding YAML fields were omitted.
type Task struct {
	// Cmd is the shell command that will be executed.
	// It must have a value, otherwise a Task is not valid.
	Cmd string `yaml:"cmd"`

	// Deps is the slice of names of task on which this Task depends.
	// They can be specified in any order.
	// Non-nil empty when the task has no dependencies.
	Deps []string `yaml:"deps"`

	// Inputs is the slice of input filenames necessary to execute Task.Cmd.
	// Non-nil empty when Task.Cmd does not require input files.
	Inputs []string `yaml:"inputs"`

	// Outputs is the slice of output filenames produced by execution of Task.Cmd.
	// Non-nil empty when Task.Cmd does not produce artifacts.
	Outputs []string `yaml:"outputs"`
}
