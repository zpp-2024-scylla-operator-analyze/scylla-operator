package symptoms

import (
	"errors"
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

var StorageSymptoms = NewSymptomSet("storage", []*SymptomSet{
	buildLocalCsiDriverMissingSymptoms(),
	buildStorageClassMissingSymptoms(),
})

func buildLocalCsiDriverMissingSymptoms() *SymptomSet {
	// Scenario #2: local-csi-driver CSIDriver, referenced by scylladb-local-xfs StorageClass, is missing
	noCsiDriver := NewSymptom("CSIDriver is missing",
		"%[csi-driver.Name]% CSIDriver, referenced by %[storage-class.Name]% StorageClass, is missing",
		"deploy %[csi-driver.Name]% provisioner (or change StorageClass)",
		selectors.
			Select("scylla-cluster", selectors.Type[scyllav1.ScyllaCluster]()).
			Select("scylla-pod", selectors.Type[v1.Pod]()).
			Select("storage-class", selectors.Type[storagev1.StorageClass]()).
			//SelectPhantom("csi-driver", selectors.Type[storagev1.CSIDriver]()).
			Relate("scylla-cluster", "scylla-pod", func(c *scyllav1.ScyllaCluster, p *v1.Pod) bool {
				return c.Name == p.Labels["scylla/cluster"]
			}).
			Relate("scylla-cluster", "storage-class", func(c *scyllav1.ScyllaCluster, sc *storagev1.StorageClass) bool {
				for _, rack := range c.Spec.Datacenter.Racks {
					if *rack.Storage.StorageClassName == sc.Name {
						return true
					}
				}
				return false
				//}).
				//Relate("storage-class", "csi-driver", func(sc *storagev1.StorageClass, d *storagev1.CSIDriver) bool {
				//	return sc.Provisioner == d.Name
			}))

	csiDriverMissingSymptoms := NewEmptySymptomSet("csi-driver-missing")
	err := csiDriverMissingSymptoms.Add(&noCsiDriver)
	if err != nil {
		panic(errors.New("failed to create csiDriverMissing symptom" + err.Error()))
	}
	return &csiDriverMissingSymptoms
}

func buildStorageClassMissingSymptoms() *SymptomSet {
	// Scenario #1: scylladb-local-xfs StorageClass used by a ScyllaCluster is missing
	notDeployedStorageClass := NewSymptom("StorageClass is missing",
		"%[cluster-storage-class.Name]% StorageClass used by a ScyllaCluster is missing",
		"deploy %[cluster-storage-class.Name]% StorageClass (or change StorageClass)",
		selectors.
			Select("scylla-cluster", selectors.Type[scyllav1.ScyllaCluster]()).
			Select("scylla-pod", selectors.Type[v1.Pod]()).
			Select("cluster-storage-class", selectors.Type[storagev1.StorageClass]()).
			Select("csi-driver", selectors.Type[storagev1.CSIDriver]()).
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
			Relate("scylla-cluster", "scylla-pod", func(c *scyllav1.ScyllaCluster, p *v1.Pod) bool {
				return c.Name == p.Labels["scylla/cluster"]
			}).
			Relate("scylla-cluster", "cluster-storage-class", func(c *scyllav1.ScyllaCluster, sc *storagev1.StorageClass) bool {
				for _, rack := range c.Spec.Datacenter.Racks {
					if *rack.Storage.StorageClassName == sc.Name {
						return true
					}
				}
				return false
				//}).
				//Relate("storage-class", "csi-driver", func(sc *storagev1.StorageClass, d *storagev1.CSIDriver) bool {
				//	return sc.Provisioner == d.Name
			}))

	storageClassMissingSymptoms := NewEmptySymptomSet("StorageClass missing")
	err := storageClassMissingSymptoms.Add(&notDeployedStorageClass)
	if err != nil {
		panic(errors.New("failed to create storageClassMissingSymptoms symptom" + err.Error()))
	}
	return &storageClassMissingSymptoms
}
