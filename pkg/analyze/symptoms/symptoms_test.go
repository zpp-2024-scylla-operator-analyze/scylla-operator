package symptoms

import (
	"fmt"
	"github.com/scylladb/scylla-operator/pkg/analyze/snapshot"
	"reflect"
	"testing"
)

type dummySymptom struct {
	name          string
	diagnoses     []string
	suggestions   []string
	matchCallback func(snapshot.Snapshot) ([]Issue, error)
}

func newEmptyFakeSymptom(name string) Symptom {
	return &dummySymptom{
		name:        name,
		diagnoses:   []string{"diagnoses"},
		suggestions: []string{"suggestions"},
		matchCallback: func(s snapshot.Snapshot) ([]Issue, error) {
			return nil, nil
		},
	}
}

func newFakeSymptom(name string, match func(snapshot.Snapshot) ([]Issue, error)) Symptom {
	return &dummySymptom{
		name:          name,
		diagnoses:     []string{"diagnoses"},
		suggestions:   []string{"suggestions"},
		matchCallback: match,
	}
}

func (d *dummySymptom) Name() string {
	return d.name
}

func (d *dummySymptom) Diagnoses() []string {
	return d.diagnoses
}

func (d *dummySymptom) Suggestions() []string {
	return d.suggestions
}

func (d *dummySymptom) Match(snapshot snapshot.Snapshot) ([]Issue, error) {
	return d.matchCallback(snapshot)
}

func trueSelector(ss snapshot.Snapshot) []map[string]any {
	s := make([]map[string]any, 0)
	m := make(map[string]any)
	m["true"] = "true"
	s = append(s, m)
	return s
}

func falseSelector(ss snapshot.Snapshot) []map[string]any {
	return make([]map[string]any, 0)
}

type testSymptom struct {
	willMatch bool
}

func (t *testSymptom) Name() string {
	return fmt.Sprintf("Test %v symptom", t.willMatch)
}

func (t *testSymptom) Diagnoses() []string {
	return nil
}

func (t *testSymptom) Suggestions() []string {
	return nil
}

func (t *testSymptom) Match(snapshot.Snapshot) ([]Issue, error) {
	if t.willMatch {
		return []Issue{NewIssue(nil, nil)}, nil
	} else {
		return nil, nil
	}
}

var trueSymptom = testSymptom{willMatch: true}
var falseSymptom = testSymptom{willMatch: false}

func proxySelector(pairing map[string]string) func(snapshot.Snapshot) []map[string]any {
	return func(s snapshot.Snapshot) []map[string]any {
		objects := make(map[string]any)
		for k, v := range pairing {
			vals := s.All()

			var found any = nil
			for _, objs := range vals {
				for _, obj := range objs {
					val := reflect.ValueOf(obj)
					if val.IsValid() {
						if val.Kind() == reflect.Ptr {
							val = val.Elem()
						}
						name := val.FieldByName("Name")
						if name.IsValid() && name.String() == v {
							found = obj
							break
						}
					}
				}
				if found != nil {
					break
				}
			}

			if found != nil {
				objects[k] = found
			}
		}
		if len(objects) == 0 {
			return nil
		} else {
			return []map[string]any{objects}
		}
	}
}

type fakeSnapshot struct {
	objects map[reflect.Type][]any
}

func (m *fakeSnapshot) Add(obj interface{}) {
	m.objects[reflect.TypeOf(obj)] = append(m.objects[reflect.TypeOf(obj)], obj)
}

func (m *fakeSnapshot) List(objType reflect.Type) []interface{} {
	return m.objects[objType]
}

func (m *fakeSnapshot) All() map[reflect.Type][]interface{} {
	return m.objects
}

func makeNicer(m map[string]any) map[string]string {
	nicer := make(map[string]string)
	for k, v := range m {
		val := reflect.ValueOf(v)
		if val.IsValid() {
			id := ""
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			fieldNamespace := val.FieldByName("Namespace")
			if fieldNamespace.IsValid() && fieldNamespace.String() == v {
				id = fieldNamespace.String()
			}
			fieldName := val.FieldByName("Name")
			if fieldName.IsValid() && fieldName.String() == v {
				id = fieldNamespace.String() + "." + fieldName.String()
			}
			nicer[k] = id
		}
	}
	return nicer
}

func TestSymptoms(t *testing.T) {
	tt := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "Empty symptom set",
			testFunc: func(t *testing.T) {
				// given
				expectedName := "dummySet"

				// when
				set := NewEmptySymptomNode(expectedName)

				// then
				if set.Name() != expectedName {
					t.Errorf("name differs - got %s, wwant %s", set.Name(), expectedName)
				}
				if set.Symptom() != nil {
					t.Errorf("Symptom is not nil - got %v", set.Symptom())
				}
				if len(set.Children()) > 0 {
					t.Errorf("Children() is not empty - got %v", set.Children())
				}
				if set.Parent() != nil {
					t.Errorf("Parent() is not nil - got %v", set.Parent())
				}
			},
		},
		{
			name: "Non-empty symptom set",
			testFunc: func(t *testing.T) {
				// given
				expectedName := "dummySet"
				child1Name := "child1"
				child2Name := "child2"
				child1 := NewEmptySymptomNode(child1Name)
				child2 := NewEmptySymptomNode(child2Name)
				children := []SymptomTreeNode{child1, child2}

				// when
				set := NewSymptomTreeNodeWithChildren(expectedName, nil, nil, children...)
				//set := NewSymptomSet(expectedName, children)

				// then
				if set.Name() != expectedName {
					t.Errorf("name differs - got %s, wwant %s", set.Name(), expectedName)
				}
				if set.Symptom() != nil {
					t.Errorf("Symptom() is not empty - got %v", set.Symptom())
				}
				if set.Parent() != nil {
					t.Fatalf("Parent() is not nil - got %v", set.Parent())
				}

				for _, child := range children {
					found := false
					for k, ds := range set.Children() {
						if child.Name() == k {
							found = true
							if ds.Parent() == nil || ds.Parent().Name() != set.Name() {
								t.Errorf("wrong parent for %s - got %v, want %v", k, ds.Parent(), &set)
							}
							if ds.Symptom() != nil {
								t.Errorf("Symptom() is not empty for %s - got %v", k, set.Symptom())
							}
							if len(ds.Children()) > 0 {
								t.Errorf("Children() is not empty for %s - got %v", k, set.Children())
							}
						}
					}
					if !found {
						t.Errorf("child missing: %s", child.Name())
					}
				}
			},
		},
		{
			name: "SetSymptom with valid symptom",
			testFunc: func(t *testing.T) {
				// given
				ss := NewEmptySymptomNode("symptomSet")
				s := newFakeSymptom("symptom", func(snapshot.Snapshot) ([]Issue, error) { return nil, nil })

				// when
				err := ss.SetSymptom(s)

				// then
				if err != nil {
					t.Fatalf("add shouldn't return an error %v", err)
				}
				if ss.Symptom() == nil {
					t.Fatalf("symptom shouldn't be nil, got nil, want %v", s)
				}
				if ss.Symptom().Name() != "symptom" {
					t.Errorf("symptom name invalid, got %s want symptom", ss.Symptom().Name())
				}
			},
		},
		{
			name: "SetSymptom with nil",
			testFunc: func(t *testing.T) {
				// given
				ss := NewEmptySymptomNode("symptomSet")

				// when
				err := ss.SetSymptom(nil)

				// then
				if err == nil {
					t.Fatalf("add should return an error %v", err)
				}
			},
		},
		{
			name: "AddChild with valid child",
			testFunc: func(t *testing.T) {
				// given
				ss := NewEmptySymptomNode("symptomSet1")
				ss2 := NewEmptySymptomNode("symptomSet2")

				// when
				err := ss.AddChild(ss2)

				// then
				if err != nil {
					t.Fatalf("AddChild shouldn't return an error %v", err)
				}
				if len(ss.Children()) != 1 {
					t.Fatalf("Children length mismatch, got %d want 1", len(ss.Children()))
				}
				if _, ok := ss.Children()["symptomSet2"]; !ok {
					t.Fatalf("Children should contain symptomSet2")
				}
				if (ss.Children()["symptomSet2"]).Name() != "symptomSet2" {
					t.Errorf("Children name invalid, got %s want symptom", (ss.Children()["symptomSet2"]).Name())
				}
			},
		},
		{
			name: "AddChild with nil",
			testFunc: func(t *testing.T) {
				// given
				ss := NewEmptySymptomNode("symptomSet")

				// when
				err := ss.AddChild(nil)

				// then
				if err == nil {
					t.Fatalf("add should return an error %v", err)
				}
			},
		},
		{
			name: "Match tree errors",
			testFunc: func(t *testing.T) {
				trueNode := NewSymptomTreeLeaf("true", &trueSymptom)
				root := NewSymptomTreeNode("", &trueSymptom, OrCondition)

				_, _, err := MatchTree(root, nil)
				if err == nil {
					t.Errorf("Matching non-leaf node with no children should return an error")
				}
				root = NewSymptomTreeNode("", nil, OrCondition)
				root.AddChild(trueNode)
				_, _, err = MatchTree(root, nil)
				if err == nil {
					t.Errorf("Matching non-leaf symptom node with nil symptom should return an error")
				}
				root = NewSymptomTreeNode("", &trueSymptom, nil)
				root.AddChild(trueNode)
				_, _, err = MatchTree(root, nil)
				if err == nil {
					t.Errorf("Matching non-leaf node with no callback function should return an error")
				}
			},
		},
		{
			name: "Or condition",
			testFunc: func(t *testing.T) {
				root := NewSymptomTreeNode("or", &trueSymptom, OrCondition)
				falseNode := NewSymptomTreeLeaf("false", &falseSymptom)
				root.AddChild(falseNode)
				_, matched, err := MatchTree(root, nil)
				if err != nil {
					t.Errorf("MatchTree with false child shouldn't return an error %v", err)
				}
				if matched {
					t.Errorf("Tree with or condition and one false Child shouldn't match")
				}
				trueNode := NewSymptomTreeLeaf("true", &trueSymptom)
				root.AddChild(trueNode)
				_, matched, err = MatchTree(root, nil)
				if err != nil {
					t.Errorf("MatchTree with false and true child shouldn't return an error %v", err)
				}
				if !matched {
					t.Errorf("Tree with or condition and false and true child should match")
				}
			},
		},
		{
			name: "And condition",
			testFunc: func(t *testing.T) {
				root := NewSymptomTreeNode("and", &trueSymptom, AndCondition)
				trueNode := NewSymptomTreeLeaf("true", &trueSymptom)
				root.AddChild(trueNode)
				_, matched, err := MatchTree(root, nil)
				if err != nil {
					t.Errorf("MatchTree with false child shouldn't return and error %v", err)
				}
				if !matched {
					t.Errorf("Tree with and condition and true child should match %v", err)
				}
				falseNode := NewSymptomTreeLeaf("false", &falseSymptom)
				root.AddChild(falseNode)
				_, matched, err = MatchTree(root, nil)
				if err != nil {
					t.Errorf("MatchTree with false and true child shouldn't return and error %v", err)
				}
				if matched {
					t.Errorf("Tree with and condition and false and true child shouldn't match")
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.testFunc(t)
		})
	}
}
