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

func createCommand(cmdName string, args map[string]any) (Command, error) {	
	cmd := Command{
		Name: cmdName,
	}
	
	for arg, specs := range args {
		strSpecs := make([]string, len(specs.([]any)))

		for index, value := range specs.([]any) {
			if str, ok := value.(string); ok {
				strSpecs[index] = str
			} else {
				log.Printf("element %v is a valid spec", str)
				continue
			}
		}
		
		switch arg {
		case "cmds":
			cmd.Cmds = strSpecs
		case "inputs":
			cmd.Inputs = strSpecs
		case "outputs":
			cmd.Outputs = strSpecs
	    case "deps":
			cmd.Deps = strSpecs
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
		case map[string]any:
			for cmdName, args := range cmds {
				args, ok := args.(map[string]any)
				if !ok {
					log.Print("failed to establish the args type...skipping...")
					continue
				}
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
