package symptoms

import (
	"context"
	"errors"
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"k8s.io/klog/v2"
	"sync"
)

const jobBufferSize = 128

type Executor interface {
	Enqueue(Job)
	Start() error
	Finish()
}

type Job struct {
	Symptom  *Symptom
	Callback func([]front.Diagnosis, error)
}

type workerPool struct {
	ds            *sources.DataSource
	jobs          chan Job
	nWorkers      int
	started       bool
	poolMutex     sync.Mutex
	workerContext context.Context
	workerCancel  context.CancelFunc
}

func NewMatchWorkerPool(ctx context.Context, ds *sources.DataSource, nWorkers int) Executor {
	workerContext, workerCancel := context.WithCancel(ctx)
	return &workerPool{
		workerContext: workerContext,
		workerCancel:  workerCancel,
		nWorkers:      nWorkers,
		ds:            ds,
		jobs:          make(chan Job, jobBufferSize),
	}
}

func (w *workerPool) Enqueue(job Job) {
	w.jobs <- job
}

func (w *workerPool) Start() error {
	w.poolMutex.Lock()
	defer w.poolMutex.Unlock()
	if w.started {
		return errors.New("worker pool started multiple times")
	}
	for i := 0; i < w.nWorkers; i++ {
		go worker(w.workerContext, i, w)
	}
	return nil
}

func (w *workerPool) Finish() {
	w.poolMutex.Lock()
	defer w.poolMutex.Unlock()
	w.workerCancel()
}

func worker(ctx context.Context, workerId int, pool *workerPool) {
	klog.Infof("Starting worker %d", workerId)
	for {
		select {
		case <-ctx.Done():
			klog.Infof("Worker %d finished", workerId)
			break
		case job := <-pool.jobs:
			diag, err := (*job.Symptom).Match(pool.ds)
			job.Callback(diag, err)
		}
	}
}
