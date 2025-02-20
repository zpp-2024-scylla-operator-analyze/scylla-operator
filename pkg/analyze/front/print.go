package front

import "fmt"

func Print(diagnoses []Diagnosis) error {
	fmt.Printf("PRINT %v\n", diagnoses)
	return nil
}
