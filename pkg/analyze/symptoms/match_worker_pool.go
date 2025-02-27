package symptoms

import (
	"context"
	"github.com/scylladb/scylla-operator/pkg/analyze/snapshot"
)

type Job struct {
	Symptom *Symptom
}

type JobStatus struct {
	Job    *Job
	Error  error
	Issues []Issue
}

type MatchWorkerPool struct {
	ds            *snapshot.Snapshot
	jobs          chan *Job
	statusChan    chan JobStatus
	numWorkers    int
	started       bool
	workerContext context.Context
	workerCancel  context.CancelFunc
}

func NewMatchWorkerPool(ctx context.Context, ds *snapshot.Snapshot, statusChan chan JobStatus, numWorkers int) *MatchWorkerPool {
	workerContext, workerCancel := context.WithCancel(ctx)
	return &MatchWorkerPool{
		ds:            ds,
		jobs:          make(chan *Job, numWorkers),
		statusChan:    statusChan,
		numWorkers:    numWorkers,
		started:       false,
		workerContext: workerContext,
		workerCancel:  workerCancel,
	}
}

func (w *MatchWorkerPool) EnqueueAll(symptoms *SymptomSet) int {
	count := len((*symptoms).Symptoms())
	for _, s := range (*symptoms).Symptoms() {
		w.Enqueue(Job{
			Symptom: s,
		})
	}

	for _, s := range (*symptoms).DerivedSets() {
		count += w.EnqueueAll(s)
	}
	return count
}

func (w *MatchWorkerPool) Enqueue(job Job) {
	w.jobs <- &job
}

// Start initializes and starts the worker pool if it has not been started yet; panics if already started.
// This method is not thread safe.
func (w *MatchWorkerPool) Start() {
	if w.started {
		panic("MatchWorkerPool already started")
	}
	w.started = true
	for i := 0; i < w.numWorkers; i++ {
		go worker(w.workerContext, w)
	}
}

func (w *MatchWorkerPool) Finish() {
	w.workerCancel()
	close(w.jobs)
}

func worker(ctx context.Context, pool *MatchWorkerPool) {
	for {
		select {
		case <-ctx.Done():
			break
		case job := <-pool.jobs:
			diag, err := (*job.Symptom).Match(pool.ds)
			pool.statusChan <- JobStatus{
				Job:    job,
				Error:  err,
				Issues: diag,
			}
		}
	}
}
