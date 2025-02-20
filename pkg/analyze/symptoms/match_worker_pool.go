package symptoms

import (
	"context"
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"sync"
)

type Job struct {
	Symptom  *Symptom
	Callback func([]front.Diagnosis, error)
}

type MatchWorkerPool struct {
	ds            *sources.DataSource
	jobs          chan Job
	started       bool
	poolMutex     sync.Mutex
	workerContext context.Context
	workerCancel  context.CancelFunc
}

func NewMatchWorkerPool(ctx context.Context, ds *sources.DataSource, matchThrottleLimit int) *MatchWorkerPool {
	workerContext, workerCancel := context.WithCancel(ctx)
	return &MatchWorkerPool{
		workerContext: workerContext,
		workerCancel:  workerCancel,
		ds:            ds,
		jobs:          make(chan Job, matchThrottleLimit),
	}
}

func (w *MatchWorkerPool) Enqueue(job Job) {
	go func() {
		// Throttling
		w.jobs <- job
		select {
		case <-w.workerContext.Done():
			return
		case job = <-w.jobs:
			diag, err := (*job.Symptom).Match(w.ds)
			job.Callback(diag, err)
		}
	}()
}

func (w *MatchWorkerPool) Finish() {
	w.workerCancel()
}
