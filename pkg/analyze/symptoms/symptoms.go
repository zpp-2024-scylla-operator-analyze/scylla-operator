package symptoms

import (
	"errors"
	"fmt"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"k8s.io/klog/v2"
)

const DefaultLimit = 4

type Symptom interface {
	Name() string
	Diagnoses() []string
	Suggestions() []string
	Match(*sources.DataSource2) ([]Issue, error)
}

type symptom struct {
	name        string
	diagnoses   []string
	suggestions []string
	selector    func(*sources.DataSource2) []map[string]any
}

func NewSymptom(name string, diag string, suggestions string, selector func(*sources.DataSource2) []map[string]any) Symptom {
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

func (s *symptom) Match(ds *sources.DataSource2) ([]Issue, error) {
	res := s.selector(ds)
	if res != nil && len(res) > 0 {
		issues := make([]Issue, len(res))

		var sym Symptom = s
		for i, r := range res {
			issues[i] = NewIssue(&sym, r)
		}

		return issues, nil
	}
	return nil, nil
}

//type AndSymptom interface {
//	Symptom
//	SubSymptoms() []*Symptom
//}
//
//type multiSymptom struct {
//	name     string
//	symptoms []*Symptom
//	selector func(*sources.DataSource) (bool, error)
//}
//
//func NewMultiSymptom(name string, symptoms []*Symptom) AndSymptom {
//	return &multiSymptom{
//		name:     name,
//		symptoms: symptoms,
//		selector: func(_ *sources.DataSource) (bool, error) { panic("not implemented :(") },
//	}
//}
//
//func (m *multiSymptom) Name() string {
//	return m.name
//}
//
//func (m *multiSymptom) Diagnoses() []string {
//	diagnoses := make([]string, 0)
//	for _, sym := range m.symptoms {
//		diagnoses = append(diagnoses, (*sym).Diagnoses()...)
//	}
//	return diagnoses
//}
//
//func (m *multiSymptom) Suggestions() []string {
//	suggestions := make([]string, 0)
//	for _, sym := range m.symptoms {
//		suggestions = append(suggestions, (*sym).Suggestions()...)
//	}
//	return suggestions
//}
//
//func (m *multiSymptom) Match(ds *sources.DataSource) ([]front.Diagnosis, error) {
//	match, err := m.selector(ds)
//	if err != nil {
//		return nil, err
//	}
//	if match {
//		// TODO: construct diagnosis
//		return make([]front.Diagnosis, 0), nil
//	}
//	return nil, nil
//}
//
//func (m *multiSymptom) SubSymptoms() []*Symptom {
//	return m.symptoms
//}

type SymptomSet interface {
	Name() string
	Symptoms() map[string]*Symptom
	DerivedSets() map[string]*SymptomSet
	Parent() *SymptomSet
	SetParent(*SymptomSet)

	Add(*Symptom) error
	AddChild(*SymptomSet) error
}

type symptomSet struct {
	name     string
	parent   *SymptomSet
	symptoms map[string]*Symptom
	children map[string]*SymptomSet
}

func NewEmptySymptomSet(name string) SymptomSet {
	return &symptomSet{
		name:     name,
		parent:   nil,
		symptoms: make(map[string]*Symptom),
		children: make(map[string]*SymptomSet),
	}
}

func NewSymptomSet(name string, children []*SymptomSet) SymptomSet {
	ss := NewEmptySymptomSet(name)
	for _, subset := range children {
		err := ss.AddChild(subset)
		if err != nil {
			klog.Warningf("can't add child symptoms for set %s: %v", name, err)
			return nil
		}
	}
	return ss
}

func (s *symptomSet) Name() string {
	return s.name
}

func (s *symptomSet) Symptoms() map[string]*Symptom {
	return s.symptoms
}

func (s *symptomSet) DerivedSets() map[string]*SymptomSet {
	return s.children
}

func (s *symptomSet) Parent() *SymptomSet {
	return s.parent
}

func (s *symptomSet) SetParent(parent *SymptomSet) {
	s.parent = parent
}

func (s *symptomSet) Add(ss *Symptom) error {
	_, isIn := s.symptoms[(*ss).Name()]
	if isIn {
		return errors.New("symptom already exists")
	}
	s.symptoms[(*ss).Name()] = ss
	return nil
}

func (s *symptomSet) AddChild(ss *SymptomSet) error {
	_, isIn := s.children[(*ss).Name()]
	if isIn {
		return errors.New(fmt.Sprintf("symptom already exists: %v", ss))
	}
	s.children[(*ss).Name()] = ss

	var thisAsInterface SymptomSet = s
	(*ss).SetParent(&thisAsInterface)
	return nil
}
