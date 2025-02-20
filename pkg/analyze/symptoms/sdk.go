package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
)

type Symptom interface {
	Name() string
	Diagnoses() []string
	Suggestions() []string
	Match(*sources.DataSource) ([]front.Diagnosis, error)
}

type symptom struct {
	name        string
	diagnoses   []string
	suggestions []string
	selector    func(*sources.DataSource) (bool, error)
}

func NewSymptom(name string, diag string, suggestions string, selector func(*sources.DataSource) (bool, error)) Symptom {
	return &symptom{
		name:        name,
		diagnoses:   []string{diag},
		suggestions: []string{suggestions},
		selector:    selector,
	}
}

func (s *symptom) Name() string {
	return s.name
}

func (s *symptom) Diagnoses() []string {
	return s.diagnoses
}

func (s *symptom) Suggestions() []string {
	return s.suggestions
}

func (s *symptom) Match(ds *sources.DataSource) ([]front.Diagnosis, error) {
	match, err := s.selector(ds)
	if err != nil {
		return nil, err
	}
	if match {
		// TODO: construct diagnosis
		return make([]front.Diagnosis, 0), nil
	}
	return nil, nil
}

type MultiSymptom interface {
	Symptom
	SubSymptoms() []*Symptom
}

type multiSymptom struct {
	name     string
	symptoms []*Symptom
	selector func(*sources.DataSource) (bool, error)
}

func NewMultiSymptom(name string, symptoms []*Symptom, glueSelector string) MultiSymptom {
	return &multiSymptom{
		name:     name,
		symptoms: symptoms,
		selector: func(_ *sources.DataSource) (bool, error) { panic("not implemented :(") },
	}
}

func (m *multiSymptom) Name() string {
	return m.name
}

func (m *multiSymptom) Diagnoses() []string {
	diagnoses := make([]string, 0)
	for _, sym := range m.symptoms {
		diagnoses = append(diagnoses, (*sym).Diagnoses()...)
	}
	return diagnoses
}

func (m *multiSymptom) Suggestions() []string {
	suggestions := make([]string, 0)
	for _, sym := range m.symptoms {
		suggestions = append(suggestions, (*sym).Suggestions()...)
	}
	return suggestions
}

func (m *multiSymptom) Match(ds *sources.DataSource) ([]front.Diagnosis, error) {
	match, err := m.selector(ds)
	if err != nil {
		return nil, err
	}
	if match {
		// TODO: construct diagnosis
		return make([]front.Diagnosis, 0), nil
	}
	return nil, nil
}

func (m *multiSymptom) SubSymptoms() []*Symptom {
	return m.symptoms
}

type SymptomSet interface {
	Name() string
	Symptoms() []*Symptom

	Add(*Symptom)
	AddChild(*SymptomSet)
	Match(*sources.DataSource) ([]front.Diagnosis, error)
}

type symptomSet struct {
	name     string
	symptoms []*Symptom
	children map[string]*SymptomSet
}

func NewSymptomSet(name string) SymptomSet {
	return &symptomSet{
		name:     name,
		symptoms: make([]*Symptom, 0),
	}
}

func (s *symptomSet) Name() string {
	return s.name
}

func (s *symptomSet) Symptoms() []*Symptom {
	return s.symptoms
}

func (s *symptomSet) Add(ss *Symptom) {
	s.symptoms = append(s.symptoms, ss)
}

func (s *symptomSet) AddChild(ss *SymptomSet) {
	s.children[(*ss).Name()] = ss
}

func (s *symptomSet) Match(ds *sources.DataSource) ([]front.Diagnosis, error) {
	for _, sym := range s.symptoms {
		diag, err := (*sym).Match(ds)
		if err != nil {
			return nil, err
		}
		if diag != nil || len(diag) > 0 {
			return diag, nil
		}
	}
	return nil, nil
}
