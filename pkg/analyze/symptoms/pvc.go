package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"k8s.io/apimachinery/pkg/labels"
)

func PersistentVolumeClaimPending(ds *sources.DataSource) (bool, error) {
	//clusters, err := ds.PersistentVolumeClaimLister.PersistentVolumeClaims("scylla").List(labels.Everything())
	clusters, err := ds.PersistentVolumeClaimLister.List(labels.Everything())
	if err != nil {
		return false, err
	}

	for _, cluster := range clusters {
		if cluster.Status.Phase == "Pending" {
			return true, nil
		}
	}

	return true, nil
}
