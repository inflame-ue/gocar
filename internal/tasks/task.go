package tasks

import (
	"errors"
	"log"
)

type Command struct {
	Name    string
	Cmds    []string
	Inputs  []string
	Outputs []string
	Deps    []string	
}

type Task struct {
	Commands []Command
}

func createCommand(cmdName string, args map[string][]string) (Command, error) {
	cmd := Command{
		Name: cmdName,
	}
	for arg, specs := range args {
		switch arg {
		case "cmds":
			cmd.Cmds = specs
		case "inputs":
			cmd.Inputs = specs
		case "outputs":
			cmd.Outputs = specs
	    case "deps":
			cmd.Deps = specs
	    default:
		    return Command{}, errors.New("err: argument not supported for command")
		}
	}

	return cmd, nil
}

func ParseTask(commands map[any]any) (Task, error) {
	task := Task{
		Commands: []Command{},
	}
	
	switch cmds := commands["tasks"].(type) {
		case map[string]map[string][]string:
			for cmdName, args := range cmds {
				cmd, err := createCommand(cmdName, args)
				if err != nil {
					log.Printf("err: %v...skipping the command...", err)
					continue
				}
				task.Commands = append(task.Commands, cmd)
			}
		default:
			return Task{}, errors.New("err: the commands must follow: name -> commands | inputs | outputs | deps -> slices")
	}

	return task, nil
}
