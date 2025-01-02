package diagnoses

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"k8s.io/klog/v2"
)

func Diagnose(ds *sources.DataSource) ([]Diagnosis, error) {
	klog.Infof("***")

	diagnoses := make([]Diagnosis, 0)

	result, err := noOperator(ds)
	if err != nil {
		return nil, err
	}

	diagnoses = append(diagnoses, result...)

	klog.Infof("***")

	return diagnoses, nil
}
