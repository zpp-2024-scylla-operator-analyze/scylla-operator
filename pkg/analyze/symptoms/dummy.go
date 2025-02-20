package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"k8s.io/klog/v2"
)

var DummySymptoms = NewSymptomSet("dummy", []*SymptomSet{
	buildBasicDummySymptoms(),
})

func buildBasicDummySymptoms() *SymptomSet {
	basicSet := NewSymptomSet("basic", []*SymptomSet{})

	emptyCluster := NewSymptom("cluster", "diag", "sug",
		selectors.
			Select("cluster", selectors.Type[scyllav1.ScyllaCluster]()).
			Collect([]string{"cluster"}, func(cluster *scyllav1.ScyllaCluster) bool {
				klog.Infof("found %v", cluster)
				return false
			}))
	basicSet.Add(&emptyCluster)

	return &basicSet
}
