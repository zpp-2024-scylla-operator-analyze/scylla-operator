package symptoms

import (
	"errors"
	"fmt"
	"github.com/scylladb/scylla-operator/pkg/analyze/selector"
	"github.com/scylladb/scylla-operator/pkg/analyze/snapshot"
	"k8s.io/klog/v2"
)

const DefaultLimit = 4

type Symptom interface {
	Name() string
	Diagnoses() []string
	Suggestions() []string
	Match(snapshot.Snapshot) ([]Issue, error)
}

type symptom struct {
	name        string
	diagnoses   []string
	suggestions []string
	selector    *selector.Selector
}

func NewSymptom(name string, diagnoses []string, suggestions []string, selector *selector.Selector) Symptom {
	return &symptom{
		name:        name,
		diagnoses:   diagnoses,
		suggestions: suggestions,
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

func (s *symptom) Match(sn snapshot.Snapshot) ([]Issue, error) {
	it, err := s.selector.IteratorFromSnapshot(sn)
	if err != nil {
		return nil, err
	}

	res, err := it.Take(DefaultLimit)
	if err != nil {
		return nil, err
	}

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

type childrenMatcher func(map[string]SymptomTreeNode, snapshot.Snapshot) ([]Issue, bool, error)

type SymptomTreeNode interface {
	Name() string
	Symptom() Symptom
	SetSymptom(Symptom) error
	Parent() SymptomTreeNode
	SetParent(SymptomTreeNode)
	IsLeaf() bool
	MatchChildren(snapshot.Snapshot) ([]Issue, bool, error)
	getCallback() childrenMatcher

	Children() map[string]SymptomTreeNode
	AddChild(SymptomTreeNode) error
}

type symptomTreeNode struct {
	name             string
	parent           SymptomTreeNode
	symptom          Symptom
	leaf             bool
	children         map[string]SymptomTreeNode
	childrenCallback childrenMatcher
}

func NewEmptySymptomNode(name string) SymptomTreeNode {
	return &symptomTreeNode{
		name:     name,
		children: make(map[string]SymptomTreeNode),
	}
}

func NewSymptomTreeLeaf(name string, symptom Symptom) SymptomTreeNode {
	return &symptomTreeNode{
		name:             name,
		symptom:          symptom,
		parent:           nil,
		children:         nil,
		childrenCallback: nil,
		leaf:             true,
	}
}

func NewSymptomTreeNode(name string, symptom Symptom, callback childrenMatcher) SymptomTreeNode {
	return &symptomTreeNode{
		name:             name,
		symptom:          symptom,
		parent:           nil,
		children:         make(map[string]SymptomTreeNode),
		childrenCallback: callback,
		leaf:             false,
	}
}

func NewSymptomTreeNodeGroup(name string, callback childrenMatcher, children ...SymptomTreeNode) SymptomTreeNode {
	node := symptomTreeNode{
		name:             name,
		symptom:          NewSymptom(name, []string{"Node group"}, []string{"Node group"}, selector.New()),
		parent:           nil,
		children:         make(map[string]SymptomTreeNode),
		childrenCallback: callback,
		leaf:             false,
	}

	for _, c := range children {
		err := node.AddChild(c)
		if err != nil {
			klog.Warningf("can't add child symptoms for set %s: %v", name, err)
			return nil
		}
	}
	return &node
}

func NewSymptomTreeNodeWithChildren(name string, symptom Symptom, callback childrenMatcher, children ...SymptomTreeNode) SymptomTreeNode {
	node := symptomTreeNode{
		name:             name,
		symptom:          symptom,
		parent:           nil,
		children:         make(map[string]SymptomTreeNode),
		childrenCallback: callback,
		leaf:             false,
	}

	for _, c := range children {
		err := node.AddChild(c)
		if err != nil {
			klog.Warningf("can't add child symptoms for set %s: %v", name, err)
			return nil
		}
	}
	return &node
}

func (s *symptomTreeNode) Name() string {
	return s.name
}

func (s *symptomTreeNode) Symptom() Symptom {
	return s.symptom
}

func (s *symptomTreeNode) SetSymptom(symptom Symptom) error {
	if symptom == nil {
		return errors.New("Can't set nil symtptom")
	}
	s.symptom = symptom
	return nil
}

func (s *symptomTreeNode) Children() map[string]SymptomTreeNode {
	return s.children
}

func (s *symptomTreeNode) Parent() SymptomTreeNode {
	return s.parent
}

func (s *symptomTreeNode) SetParent(parent SymptomTreeNode) {
	s.parent = parent
}

func (s *symptomTreeNode) IsLeaf() bool {
	return s.leaf
}

func (s *symptomTreeNode) getCallback() childrenMatcher {
	return s.childrenCallback
}

func (s *symptomTreeNode) AddChild(c SymptomTreeNode) error {
	if c == nil {
		return errors.New("SymptomTreeNode is nil")
	}
	_, isIn := s.children[c.Name()]
	if isIn {
		return errors.New(fmt.Sprintf("symptom already exists: %v", c))
	}
	s.children[c.Name()] = c

	c.SetParent(s)
	return nil
}

func (s *symptomTreeNode) MatchChildren(ds snapshot.Snapshot) ([]Issue, bool, error) {
	return s.childrenCallback(s.Children(), ds)
}

func OrCondition(children map[string]SymptomTreeNode, ds snapshot.Snapshot) ([]Issue, bool, error) {
	for _, child := range children {
		diag, matched, err := MatchTree(child, ds)
		if err != nil {
			return nil, false, err
		}
		if matched {
			return diag, true, nil
		}
	}
	return nil, false, nil
}

func AndCondition(children map[string]SymptomTreeNode, ds snapshot.Snapshot) ([]Issue, bool, error) {
	diags := make([]Issue, 0)
	for _, child := range children {
		diag, matched, err := MatchTree(child, ds)
		if err != nil {
			return nil, false, err
		}
		if !matched {
			return nil, false, nil
		}
		diags = append(diags, diag...)
	}
	return diags, true, nil
}
