package analyze

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	"k8s.io/klog/v2"
)

func Analyze(ds *sources.DataSource) ([]front.Diagnosis, error) {
	for key, val := range symptoms.Symptoms {
		klog.Infof("%s %v", key, val)
	}

	query := selectors.New().
		Select("cluster", "ScyllaCluster").
		Select("pod", "Pod").
		Where("pod", "pod.app = scylla").
		Join("cluster", "pod", "cluster.deployed = pod.app").
		Any()
	// TODO: Maybe should panic if error while constructing a query?

	result, err := query(ds)
	if err != nil {
		return nil, err
	}

	klog.Info(result)

	return []front.Diagnosis{}, nil
}
