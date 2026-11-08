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

	flagDir := flag.String("dir", "", "The project root, location where task config lives.")
	flag.Parse()

	if *flagDir != "" {
		dir = *flagDir
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return engine.Build(ctx, spec)
}
