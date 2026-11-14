package engine

import (
	"container/heap"
	"context"
	"sync"

	"github.com/inflame-ue/gocar/internal/graph"
	"github.com/inflame-ue/gocar/internal/scheduler"
	"github.com/inflame-ue/gocar/internal/task"
)

type pool struct {
	g       *graph.Graph
	ready   *scheduler.ReadyQueue
	indeg   map[string]int
	keys    map[string]string
	running int
	err     error

	mu   sync.Mutex
	cond *sync.Cond
}

func (e *Engine) run(ctx context.Context, spec task.Tasks, g *graph.Graph) error {
	p := &pool{
		g:     g,
		ready: scheduler.NewReadyQueue(scheduler.Level(g)),
		indeg: make(map[string]int, len(g.Nodes())),
		keys:  make(map[string]string, len(g.Nodes())),
	}

	for _, n := range g.Nodes() {
		p.indeg[n] = len(g.Deps(n))
		if p.indeg[n] == 0 {
			heap.Push(p.ready, n)
		}
	}
	p.cond = sync.NewCond(&p.mu)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// signal context cancellation to all workers
	go func() {
		<-ctx.Done()
		p.mu.Lock()
		if p.err == nil {
			p.err = ctx.Err()
		}
		p.cond.Broadcast()
		p.mu.Unlock()
	}()

	return p.err
}
