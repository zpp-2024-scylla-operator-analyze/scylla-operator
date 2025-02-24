package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/selectors"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
)

var DummySymptoms = NewSymptomSet("dummy", []*SymptomSet{
	buildBasicDummySymptoms(),
})

func buildBasicDummySymptoms() *SymptomSet {
	basicSet := NewSymptomSet("basic", []*SymptomSet{})

	emptyCluster := NewSymptom("cluster", "cluster diagnosis", "cluster suggestion",
		selectors.Select("cluster", selectors.Type[scyllav1.ScyllaCluster]()))
	basicSet.Add(&emptyCluster)

	return &basicSet
}
