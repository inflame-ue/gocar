package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/inflame-ue/gocar/internal/cache"
	"github.com/inflame-ue/gocar/internal/engine"
	"github.com/inflame-ue/gocar/internal/parser"
	"github.com/inflame-ue/gocar/internal/runner"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	var file string
	flag.StringVar(&file, "file", "gocar.yaml", "Task configuration file for GoCar.")
	flag.StringVar(&file, "f", "gocar.yaml", "Task configuration file for GoCar.")

	var cacheRoot string
	flag.StringVar(&cacheRoot, "cache-root", "", "Location to store local cache artifacts.")

	var noCache bool
	flag.BoolVar(&noCache, "no-cache", false, "No cache flag, every build will be cold.")

	var jobs int
	flag.IntVar(&jobs, "jobs", 0, "Maximum number of tasks to run in parallel (0 = number of CPUs).")
	flag.IntVar(&jobs, "j", 0, "Maximum number of tasks to run in parallel (0 = number of CPUs).")

	var rebuilder string
	flag.StringVar(&rebuilder, "rebuilder", "hash", "Rebuild strategy: hash or mtime.")

	flagDir := flag.String("dir", "", "The project root, location where task config lives.")
	flag.Parse()

	if *flagDir != "" {
		dir = *flagDir
	}

	var mode engine.Rebuilder
	switch rebuilder {
	case "hash":
		mode = engine.RebuilderHash
	case "mtime":
		mode = engine.RebuilderMTime
	default:
		return fmt.Errorf("unknown rebuilder %q", rebuilder)
	}

	if cacheRoot == "" {
		cacheRoot = filepath.Join(dir, ".gocar/cache")
	}

	data, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return err
	}

	spec, err := parser.Parse(data)
	if err != nil {
		return err
	}

	var ca engine.Cache
	if !noCache {
		ca = cache.NewL1(cacheRoot)
	}

	ex := runner.NewRunner(os.Stdout, os.Stderr, dir)
	engine := engine.NewEngine(dir, ex, ca)
	engine.Jobs = jobs
	engine.Rebuilder = mode

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return engine.Build(ctx, spec)
}
