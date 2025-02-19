package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"k8s.io/klog/v2"
)

type Symptom interface {
	Name() string
	RawDiagnosis() string
	Match(*sources.DataSource) []front.Diagnosis
	AddSign(string, func(*sources.DataSource) (bool, error))
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

// TODO: symptom trees with joins
type SymptomSet struct {
	name         string
	diagnosisFmt string
	selectors    []func(*sources.DataSource) (bool, error)
}

func NewSymptomSet(name, diagnosisFmt string) *SymptomSet {
	return &SymptomSet{
		name:         name,
		diagnosisFmt: diagnosisFmt,
		selectors:    make([]func(*sources.DataSource) (bool, error), 0),
	}
}

func (s *SymptomSet) Name() string {
	return s.name
}

func (s *SymptomSet) RawDiagnosis() string {
	return s.diagnosisFmt
}

func (s *SymptomSet) Match(source *sources.DataSource) []front.Diagnosis {
	for i, selector := range s.selectors {
		match, err := selector(source)
		if err != nil {
			klog.Fatalf("error while matching selector no. %d of symptom %s %s", i, s.name, err)
		}
		if match {
			// TODO: construct appropriate diagnosis and return
			return []front.Diagnosis{}
		}
	}
	return nil
}

// TODO: builder
func (s *SymptomSet) AddSign(_ string, f func(*sources.DataSource) (bool, error)) *SymptomSet {
	s.selectors = append(s.selectors, f)
	return s
}
