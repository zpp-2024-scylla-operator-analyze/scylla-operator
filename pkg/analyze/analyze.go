package analyze

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	_ "github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func Analyze(ds *sources.DataSource) ([]front.Diagnosis, error) {
	// for key, val := range symptoms.Symptoms {
	// 	klog.Infof("%s %v", key, val)
	// }

	query := selectors.
		Select(selectors.Resource[scyllav1.ScyllaCluster]("cluster")).
		Select(selectors.Resource[v1.Pod]("pod")).
		Filter("pod", func(_ *v1.Pod) (bool, error) {
			return false, nil
		}).
		Relate("cluster", "pod", func(_ *scyllav1.ScyllaCluster, _ *v1.Pod) (bool, error) {
			return false, nil
		}).
		Collect()

	query := selectors.Builder().
		New("cluster", selectors.Type[scyllav1.ScyllaCluster]()).
		New("pod", selectors.Type[v1.Pod]()).
		Join(&selectors.FuncRelation[*scyllav1.ScyllaCluster, *v1.Pod]{
			Lhs: "cluster",
			Rhs: "pod",
			Lambda: func(_ *scyllav1.ScyllaCluster, _ *v1.Pod) (bool, error) {
				return false, nil
			},
		}).
		Where(&selectors.FuncConstraint[*v1.Pod]{
			Resource: "pod",
			Lambda: func(_ *v1.Pod) (bool, error) {
				return false, nil
			},
		}).
		Any()
	// TODO: Should panic if error while constructing a query?

	result, err := query(ds)
	if err != nil {
		return nil, err
	}

	klog.Info(result)

	return []front.Diagnosis{}, nil
}
