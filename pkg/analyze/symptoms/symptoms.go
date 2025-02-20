package symptoms

var Symptoms = NewSymptomSet()

func BuildSymptoms() []*Symptom {
	symptoms := make([]*Symptom, 0)
	symptoms = append(symptoms, CsiDriverSymptoms...)
	return symptoms
}
