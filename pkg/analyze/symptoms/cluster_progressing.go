package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"k8s.io/apimachinery/pkg/labels"
	_ "k8s.io/klog/v2"
)

func ScyllaClusterProgressing(ds *sources.DataSource) (bool, error) {
	clusters, err := ds.ScyllaClusterLister.List(labels.Everything())
	if err != nil {
		return false, err
	}

	for _, cluster := range clusters {
		for _, condition := range cluster.Status.Conditions {
			if condition.Type == "Progressing" {
				return true, nil
			}
		}
	}

	return false, nil
}
