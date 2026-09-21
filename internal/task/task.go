package task

type Task struct {
	Cmd     string   `yaml:"cmd"`
	Deps    []string `yaml:"deps"`
	Outputs []string `yaml:"outputs"`
	Inputs  []string `yaml:"inputs"`
}
