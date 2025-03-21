package symptoms

type Issue struct {
	Symptom   *Symptom
	Resources map[string]any
}

func NewIssue(symptom *Symptom, resources map[string]any) Issue {
	return Issue{
		Symptom:   symptom,
		Resources: resources,
	}
}
