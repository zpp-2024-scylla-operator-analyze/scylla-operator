package front

type Diagnosis struct {
	resources map[string]any
}

func NewDiagnosis(r map[string]any) Diagnosis {
	return Diagnosis{
		resources: r,
	}
}
