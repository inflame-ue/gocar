package engine

import (
	"container/heap"
	"context"
	"sync"

	"github.com/inflame-ue/gocar/internal/fingerprint"
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

// run orchestrates the worker pool of task execution. It construct a new
// goroutine pool, handles in-flight context cancellation, runs each worker.
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

	fp := fingerprint.NewFingerprinter(e.Dir)

	var wg sync.WaitGroup
	for i := 0; i < e.jobs(); i++ {
		wg.Add(1)
		go func() {
			e.worker(ctx, spec, fp, p)
			defer wg.Done()
		}()
	}
	wg.Wait()

	return p.err
}

func (e *Engine) worker(ctx context.Context, spec task.Tasks, fp *fingerprint.Fingerprinter, p *pool) {
	for {
		p.mu.Lock()

		for p.ready.Len() == 0 && p.running > 0 && p.err == nil && ctx.Err() == nil {
			p.cond.Wait()
		}

		if p.err != nil {
			p.mu.Unlock()
			return
		}

		if ctx.Err() != nil {
			if p.err == nil {
				p.err = ctx.Err()
			}
			p.cond.Broadcast()
			p.mu.Unlock()
			return
		}

		if p.ready.Len() == 0 {
			p.mu.Unlock()
			return
		}

		name := heap.Pop(p.ready).(string)
		t := spec[name]
		depKeys := make(map[string]string, len(t.Deps))
		for _, d := range t.Deps {
			depKeys[d] = p.keys[d]
		}
		p.running++
		p.mu.Unlock()

		key, err := e.runOne(ctx, name, t, depKeys, fp)

		p.mu.Lock()
		p.running--
		if err != nil {
			if p.err == nil {
				p.err = err
			}
			p.cond.Broadcast()
			p.mu.Unlock()
			return
		}

		p.keys[name] = key
		for _, rdep := range p.g.Rdeps(name) {
			p.indeg[rdep]--
			if p.indeg[rdep] == 0 {
				heap.Push(p.ready, rdep)
			}
		}
		p.cond.Broadcast()
		p.mu.Unlock()
	}
}
