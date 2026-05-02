package main

import (
	"fmt"
	"log"

	"github.com/inflam-ue/gocar/internal/tasks"
)

func main() {
	taskSpec, err := tasks.ParseTasks("internal/tasks/example.yaml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v", taskSpec)
}
