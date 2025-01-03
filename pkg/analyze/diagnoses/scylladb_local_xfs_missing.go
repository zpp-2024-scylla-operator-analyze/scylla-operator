package diagnoses

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
)

func storageClassMissing(ds *sources.DataSource) ([]Diagnosis, error) {
	var err error

	diagnoses := make([]Diagnosis, 0)
	result := false

	result, err = symptoms.ScyllaClusterProgressing(ds)
	if err != nil {
		return nil, err
	} else if !result {
		return diagnoses, nil
	}

	result, err = symptoms.PersistentVolumeClaimPending(ds)
	if err != nil {
		return nil, err
	} else if !result {
		return diagnoses, nil
	}

	diagnoses = append(diagnoses, New("StorageClass is missing"))

	return diagnoses, nil
}
