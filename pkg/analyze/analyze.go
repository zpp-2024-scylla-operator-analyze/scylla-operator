package analyze

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	_ "github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func Analyze(ds *sources.DataSource) ([]front.Diagnosis, error) {
	// for key, val := range symptoms.Symptoms {
	// 	klog.Infof("%s %v", key, val)
	// }
	smp := symptoms.BuildSymptoms()
	klog.Info("Available symptoms:")
	for _, val := range smp {
		klog.Infof("%s %v", (*val).Name(), val)
	}

	issues := make([]front.Diagnosis, 0)
	for _, s := range smp {
		result := (*s).Match(ds)
		err := front.Print(result)
		if err != nil {
			return nil, err
		}
		if result != nil && len(result) > 0 {
			klog.Info(result)
			issues = append(issues, result...)
		}
	}
	return issues, nil
}
