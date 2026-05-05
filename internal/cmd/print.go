package cmd

import (
	"fmt"
	
	"github.com/inflam-ue/gocar/internal/tasks"
	"github.com/spf13/cobra"
)

var printCmd = &cobra.Command{
	Use:     "print",
	Aliases: []string{"prettyprint", "pprint"},
	Short:   "Print the task struct from YAML",
	Long: `This is a command intended for debugging purposes throughout the development lifecycle.
It prints the Task struct created from the YAML file and displays it for expection command-bycommand.
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filepath := args[0]
		
		taskSpec, err := tasks.ParseYAML(filepath)
		if err != nil {
			return err
		}

		task, err := tasks.ParseTask(taskSpec)
		if err != nil {
			return err
		}

		for _, command := range task.Commands {
			fmt.Printf("%#v\n", command)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(printCmd)
}