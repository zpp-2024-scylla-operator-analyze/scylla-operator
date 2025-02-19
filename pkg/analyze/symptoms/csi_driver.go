package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
)

var CsiDriverSymptoms = []*Symptom{
	buildLocalCsiDriverMissingSymptom(),
}

func buildLocalCsiDriverMissingSymptom() *Symptom {
	symptomSet := NewSymptomSet(
		"local-csi-driver-missing",
		"local-csi-driver CSIDriver, referenced by <NAME> StorageClass, is missing")

	symptomSet.AddSign("", selectors.
		New().
		Select("scylla-cluster", "ScyllaCluster").
		Select("scylla-pod", "Pod").
		Select("cluster-storage-class", "StorageClass").
		Select("csi-driver", "CsiDriver").
		Where("scylla-cluster", func(c *scyllav1.ScyllaCluster) bool {
			storageClassXfs := false
			conditionControllerProgressing := false
			conditionProgressing := false
			for _, rack := range c.Spec.Datacenter.Racks {
				if *rack.Storage.StorageClassName == "scylladb-local-xfs" {
					storageClassXfs = true
				}
			}
			for _, cond := range c.Status.Conditions {
				if cond.Type == "StatefulSetControllerProgressing" {
					conditionControllerProgressing = true
				} else if cond.Type == "Progressing" {
					conditionProgressing = true
				}
			}
			return storageClassXfs && conditionProgressing && conditionControllerProgressing
		}).
		Join("scylla-cluster", "scylla-pod", "").
		Join("scylla-cluster", "storage-class", "").
		Join("storage-class", "csi-driver", "StorageClass").
		Any())
	return symptomSet
}
