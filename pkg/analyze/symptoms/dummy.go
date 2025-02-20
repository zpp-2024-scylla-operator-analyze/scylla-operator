package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
)

var DummySymptoms = NewSymptomSet("dummy", []*OrSymptom{
	buildBasicDummySymptoms(),
})

func buildBasicDummySymptoms() *OrSymptom {
	basicSet := NewSymptomSet("basic", []*OrSymptom{})

	emptyCluster := NewSymptom("cluster", "cluster diagnosis", "cluster suggestion",
		selectors.Select("cluster", selectors.Type[scyllav1.ScyllaCluster]()))
	basicSet.Add(&emptyCluster)

	return &basicSet
}
