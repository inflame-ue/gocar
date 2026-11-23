package engine

import (
	"context"
	"fmt"
	"io"
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

func (r *reportState) finish(log io.Writer, err error) {
	if log == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	status := "built"
	if err != nil {
		status = "failed"
	}

	// use _ to supress errors messages from the linter
	_, _ = fmt.Fprintf(log, "%s %d/%d tasks in %s: %d cached, %d up-to-date, %d executed, %d failed",
		status, r.completed, r.total, time.Since(r.start).Round(time.Millisecond),
		r.cached, r.upToDate, r.executed, r.failed)
	if r.cancelled > 0 {
		_, _ = fmt.Fprintf(log, ", %d cancelled", r.cancelled)
	}
	_, _ = fmt.Fprintln(log)
}

func (r *reportState) record(ctx context.Context, log io.Writer, verbose bool,
	name string, out outcome, elapsed time.Duration, err error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.completed++

	action := out.String()
	if err != nil {
		// cancelled task is distinct from a failed one
		if ctx.Err() != nil {
			r.cancelled++
			action = "cancelled"
		} else {
			r.failed++
			action = "failed"
		}
	} else {
		switch out {
		case outcomeCached:
			r.cached++
		case outcomeUpToDate:
			r.upToDate++
		case outcomeExecuted:
			r.executed++
		}
	}

	if log == nil || !verbose {
		return
	}

	_, _ = fmt.Fprintf(log, "[%d/%d] %-10s %s (%s)\n",
		r.completed, r.total, action, name, elapsed.Round(time.Microsecond))
}
