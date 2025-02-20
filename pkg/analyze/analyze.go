package analyze

import (
	"context"
	"github.com/pkg/errors"
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	_ "github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	"k8s.io/klog/v2"
	"runtime"
)

func Analyze(ctx context.Context, ds *sources.DataSource) ([]front.Diagnosis, error) {
	klog.Info("Available symptoms:")
	for _, val := range symptoms.Symptoms.Symptoms() {
		klog.Infof("%s %v", (*val).Name(), val)
	}

	matchExecutor := symptoms.NewMatchWorkerPool(ctx, ds, runtime.NumCPU())
	err := matchExecutor.Start()
	if err != nil {
		return nil, errors.Errorf("failed to start match worker pool: %v", err)
	}
	symptoms.MatchAll(&symptoms.Symptoms, &matchExecutor, ds, func(s *symptoms.Symptom, diagnoses []front.Diagnosis, err error) {
		if err != nil {
			klog.Warningf("symptom %v, error: %v", s, err)
			return
		}
		err = front.Print(diagnoses)
		if err != nil {
			klog.Warningf("can't print diagnoses for symptom %v, diagnoses: %v, error: %v", s, diagnoses, err)
		}
	})
	return nil, nil
}
