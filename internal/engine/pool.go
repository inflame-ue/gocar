package engine

import (
	"container/heap"
	"context"
	"sync"
	"time"

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
	cancel  context.CancelFunc

	mu   sync.Mutex
	cond *sync.Cond
}

// run orchestrates the worker pool of task execution. It constructs a new
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
	p.cancel = cancel

	// signal context cancellation to all workers
	go func() {
		<-ctx.Done()
		p.mu.Lock()
		p.cond.Broadcast()
		p.mu.Unlock()
	}()

	fp := fingerprint.NewFingerprinter(e.Dir)

	var wg sync.WaitGroup
	jobs := e.jobs()
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.worker(ctx, spec, fp, p)
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

		start := time.Now()
		key, out, err := e.runOne(ctx, name, t, depKeys, spec, fp)
		e.report.record(ctx, e.Log, e.Verbose, name, out, time.Since(start), err)

		p.mu.Lock()
		p.running--
		if err != nil {
			if p.err == nil {
				p.err = err
				// stop the tasks that are already in flight
				// do not wait for them to finish on their own.
				p.cancel()
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
