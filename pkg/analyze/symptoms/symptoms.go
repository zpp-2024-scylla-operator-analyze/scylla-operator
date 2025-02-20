package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
)

var Symptoms = NewSymptomSet("root", []*OrSymptom{
	//&CsiDriverSymptoms,
	&DummySymptoms,
})

func MatchAll(symptoms *OrSymptom, executor *MatchWorkerPool, ds *sources.DataSource, callback func(*Symptom, []front.Diagnosis, error)) {
	for _, s := range (*symptoms).Symptoms() {
		(*executor).Enqueue(Job{
			Symptom:  s,
			Callback: func(diag []front.Diagnosis, err error) { callback(s, diag, err) },
		})
	}

	for _, s := range (*symptoms).DerivedSets() {
		MatchAll(s, executor, ds, callback)
	}
}
