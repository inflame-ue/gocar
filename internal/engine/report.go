package engine

import (
	"sync"
	"time"
)

type outcome uint8

const (
	outcomeCached outcome = iota
	outcomeUpToDate
	outcomeExecuted
)

func (o outcome) String() string {
	switch o {
	case outcomeCached:
		return "cached"
	case outcomeUpToDate:
		return "up-to-date"
	default:
		return "run"
	}
}

// reportStart records the state for a single task.
// Guadrded by a mutex because each individual worker
// records the progress concurrently.
type reportState struct {
	mu        sync.Mutex
	start     time.Time
	total     int
	completed int
	cached    int
	upToDate  int
	executed  int
	failed    int
	cancelled int
}

func (r *reportState) reset(total int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.start = time.Now()
	r.total = total
	r.completed = 0
	r.cached = 0
	r.upToDate = 0
	r.executed = 0
	r.failed = 0
	r.cancelled = 0
}
