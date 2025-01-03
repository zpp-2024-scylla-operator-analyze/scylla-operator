package diagnoses

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	_ "k8s.io/klog/v2"
)

func Diagnose(ds *sources.DataSource) ([]Diagnosis, error) {
	diagnoses := make([]Diagnosis, 0)

	result, err := noOperator(ds)
	if err != nil {
		return nil, err
	}

	diagnoses = append(diagnoses, result...)

	result, err = storageClassMissing(ds)
	if err != nil {
		return nil, err
	}

	diagnoses = append(diagnoses, result...)

	return diagnoses, nil
}
