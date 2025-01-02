package front

import (
	"fmt"
	"github.com/scylladb/scylla-operator/pkg/analyze/diagnoses"
)

type Front struct {
}

func NewFront( /* Some options */ ) Front {
	return Front{}
}

/* Maybe some builder methods here */

func (f *Front) Show(diagnoses []diagnoses.Diagnosis) error {
	if len(diagnoses) == 0 {
		fmt.Println("Nothing wrong")
		return nil
	}

	for _, diagnosis := range diagnoses {
		fmt.Printf("%s\n", diagnosis.Name())
	}

	return nil
}
