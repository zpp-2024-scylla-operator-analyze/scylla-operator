package front

import "fmt"

func Print(diagnoses []Diagnosis) error {
	for _, diag := range diagnoses {
		symptom := "<UNKNOWN>"
		if diag.symptom != nil {
			symptom = (*diag.symptom).Name()
		}
		fmt.Printf("SYMPTOM %s RESOURCES: %v\n", symptom, diag.resources)
	}
	return nil
}
