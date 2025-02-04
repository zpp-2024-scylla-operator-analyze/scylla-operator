package symptoms

type Symptom interface {
	// TODO
}

func Export(name string, symptom Symptom) map[string]Symptom {
	return map[string]Symptom{name: symptom}
}

func Reexport(name string, m []map[string]Symptom) map[string]Symptom {
	result := make(map[string]Symptom)

	for _, i := range m {
		for key, val := range i {
			result[name+"."+key] = val
		}
	}

	return result
}
