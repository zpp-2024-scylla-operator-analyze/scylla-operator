package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

//import (
//	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
//	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
//)

var CsiDriverSymptoms = NewSymptomSet("csi-driver", []*SymptomSet{
	buildLocalCsiDriverMissingSymptoms(),
})

func buildLocalCsiDriverMissingSymptoms() *SymptomSet {
	nodriverbasic := NewSymptom("abc", "diaguwu", "sug", func(source *sources.DataSource) (bool, error) {
		selectors.
			Select("scylla-cluster", selectors.Type[*scyllav1.ScyllaCluster]()).
			Select("scylla-pod", selectors.Type[*v1.Pod]()).
			Select("cluster-storage-class", selectors.Type[*storagev1.StorageClass]()).
			Select("csi-driver", selectors.Type[*storagev1.CSIDriver]()).
			Filter("scylla-cluster", func(c *scyllav1.ScyllaCluster) bool {
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
			Relate("scylla-cluster", "scylla-pod", "").
			Relate("scylla-cluster", "storage-class", "").
			Relate("storage-class", "csi-driver", "StorageClass").
			Collect([]string{"scylla-cluster", "storage-class"}, func() {})
		return true, nil
	})

	csiDriverSymptomSet := NewEmptySymptomSet("local-csi-driver-missing")
	csiDriverSymptomSet.Add(&nodriverbasic)
	return &csiDriverSymptomSet
}
