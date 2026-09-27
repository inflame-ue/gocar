package task

type Task struct {
	Cmd     string   `yaml:"cmd"`
	Deps    []string `yaml:"deps"`
	Inputs  []string `yaml:"inputs"`
	Outputs []string `yaml:"outputs"`
}
