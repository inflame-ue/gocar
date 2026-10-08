package runner

import (
	"context"
	"io"
	"os/exec"
)

type commander interface {
	run(ctx context.Context, cmd string, stdout, stderr io.Writer) error
}

type shellCommander struct{}

func (sc *shellCommander) run(ctx context.Context, cmd string, stdout, stderr io.Writer) error {
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	c.Stdout, c.Stderr = stdout, stderr
	return c.Run()
}
