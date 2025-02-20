package front

import "fmt"

func Print(diagnoses []Diagnosis) error {
	fmt.Printf("%v", diagnoses)
	return nil
}
