package diagnoses

type Diagnosis interface {
	Name() string
}

type diagnosis struct {
	name string
}

func New(name string) *diagnosis {
	return &diagnosis{name: name}
}

func (d *diagnosis) Name() string {
	return d.name
}
