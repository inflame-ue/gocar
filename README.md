# GoCar

A parallel, incremental build system with content-addressed caching.

## Overview

GoCar operates on the basis of separating a scheduler and a rebuilder, in accordance with the framework introduced in the paper "Build Systems à la Carte" by Mokhov, Mitchell, and Jones. It parses a `gocar.yaml` from the project directory that defines a collection of executable tasks, uses a critical-path-ready queue to execute the independent ones in parallel(i.e. the scheduler), and fingerprints tasks with a Merkle cascade over command, outputs, inputs, and dependency keys (i.e. the rebuilder). Unchanged tasks are restored from a local cache. 

## Installation

Before using the build system, you have to install it. Make sure that you have a Go toolchain available on your system, otherwise please see [this page](https://go.dev/doc/install) first. 

```bash
go install github.com/inflame-ue/gocar/cmd/gocar@v1.0.0
```

**Important:** This requires Go 1.27.

Alternatively, you can build the project from source by cloning

```bash
git clone https://github.com/inflame-ue/gocar && cd gocar
go build -o gocar ./cmd/gocar
```

## Quick Start

Here's a quick example of the build system pipeline. You can follow along if you have a POSIX-compliant shel.

First, create the project files:

```bash
mkdir -p src
printf 'hello gocar\n' > src/message.txt
printf 'done\n' > src/suffix.txt
```

Then define `gocar.yaml` in the same directory:

```yaml
prepare:
  cmd: mkdir -p out

side:
  cmd: printf 'not affected\n' > out/side.txt
  deps: [prepare]
  outputs: [out/side.txt]

upper:
  cmd: tr '[:lower:]' '[:upper:]' < src/message.txt > out/upper.txt
  deps: [prepare]
  inputs: [src/message.txt]
  outputs: [out/upper.txt]

bundle:
  cmd: cat out/upper.txt src/suffix.txt > out/bundle.txt
  deps: [upper]
  inputs: [out/upper.txt, src/suffix.txt]
  outputs: [out/bundle.txt]

report:
  cmd: cat out/bundle.txt
  deps: [bundle]
  inputs: [out/bundle.txt]
```

Run the build with per-task progress (`./gocar` if you built it locally):

```bash
gocar -v
```

```text
[1/5] run        prepare (2.0ms)
[2/5] run        side (1.7ms)
[3/5] run        upper (2.1ms)
[4/5] run        bundle (1.9ms)
HELLO GOCAR
done
[5/5] run        report (1.6ms)
built 5/5 tasks in 8ms: 0 cached, 0 up-to-date, 5 executed, 0 failed
```

If you run it again, then `side`, `upper`, and `bundle` are restored from the cache;
`prepare` and `report` have no outputs, so they always run. 


## Configuration Reference

The top-level name, for example `prepare`, corresponds to the task name inside the system. 

- `cmd`: the required shell command.
- `deps`: task names on which the current task depends.
- `inputs`: literal paths or globs like `src/**/*.c`.
- `outputs`: literal paths.

Note that all paths specified should be relative to the project directory and that tasks with no outputs always run.

## CLI Reference

| Flag | Default | Description |
|---|---|---|
| `-f`, `-file` | `gocar.yaml` | Task configuration file, resolved relative to `-dir`. |
| `-dir` | current directory | Project root; task paths and the default cache location are relative to it. |
| `-j`, `-jobs` | `0` (number of CPUs) | Maximum number of tasks executed in parallel. |
| `-rebuilder` | `hash` | Rebuild strategy: `hash` (content-addressed cache) or `mtime` (timestamp baseline). |
| `-cache-root` | `<dir>/.gocar/cache` | Local cache directory. Used only by the hash rebuilder. |
| `-no-cache` | `false` | Disable the cache. With `hash`, every task executes; with `mtime`, it has no effect. |
| `-v`, `-verbose` | `false` | Print one progress line per completed task to stderr. |
| `-h`, `-help` | | Print usage. |

GoCar accepts both `-flag` and `--flag` forms.

### Examples

```bash
gocar -v                                     # build ./gocar.yaml with per-task progress
gocar -dir ./project -j 8                    # project root, eight workers
gocar -f tasks.yaml -rebuilder mtime         # timestamp-based rebuild
gocar --rebuilder hash -cache-root /tmp/gc   # custom cache location
gocar -no-cache -j 1                         # serial cold build
```

