package front

import "github.com/scylladb/scylla-operator/pkg/analyze/symptoms"

type Diagnosis struct {
	symptom   *symptoms.Symptom
	resources map[string]any
}

func NewDiagnosis(symptom *symptoms.Symptom, r map[string]any) Diagnosis {
	return Diagnosis{
		symptom:   symptom,
		resources: r,
	}
}
