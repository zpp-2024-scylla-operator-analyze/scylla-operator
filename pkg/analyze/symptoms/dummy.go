package symptoms

type dummySymptom struct {
	name string
}

func NewDummySymptom(name string) *dummySymptom {
	return &dummySymptom{name: name}
}
