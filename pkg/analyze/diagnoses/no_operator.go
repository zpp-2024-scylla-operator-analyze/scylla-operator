package diagnoses

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
)

func noOperator(ds *sources.DataSource) ([]Diagnosis, error) {
	result, err := symptoms.NoOperator(ds)
	if err != nil {
		return nil, err
	}

	diagnoses := make([]Diagnosis, 0)

	if result {
		diagnoses = append(diagnoses, New("No Scylla Operator"))
	}

	return diagnoses, nil
}
