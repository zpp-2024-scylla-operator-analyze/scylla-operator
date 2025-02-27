package rules

import "github.com/scylladb/scylla-operator/pkg/analyze/symptoms"

var Symptoms = symptoms.NewSymptomSet("root", []*symptoms.SymptomSet{
	&StorageSymptoms,
	&DummySymptoms,
})
