package front

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
	"testing"
)

func resourceWithGVK(res runtime.Object, gvk schema.GroupVersionKind) any {
	res.GetObjectKind().SetGroupVersionKind(gvk)
	return res
}

func TestPrint(t *testing.T) {
	tt := []struct {
		name      string
		symptom   symptoms.Symptom
		resources map[string]any
		expected  string
	}{
		{
			name:    "Simple issue",
			symptom: symptoms.NewSymptom("name", []string{"diag1", "diag2"}, []string{"sugg1", "sugg2"}, nil),
			resources: map[string]any{
				"pod": &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
				},
				"serviceAccount": &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name: "serviceAccount1",
					},
				},
			},
			expected: strings.TrimSpace(`
name
Diagnoses:
	diag1
	diag2
Suggestions:
	sugg1
	sugg2
Resources GVK:
	No GVK set
	No GVK set
---
`) + "\n",
		},
		{
			name:    "No diagnoses",
			symptom: symptoms.NewSymptom("No diag symptom", nil, []string{"sugg1", "sugg2"}, nil),
			resources: map[string]any{
				"pod": &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
				},
				"serviceAccount": &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name: "serviceAccount1",
					},
				},
			},
			expected: strings.TrimSpace(`
No diag symptom
No Diagnoses
Suggestions:
	sugg1
	sugg2
Resources GVK:
	No GVK set
	No GVK set
---
`) + "\n",
		},
		{
			name:    "No suggestions",
			symptom: symptoms.NewSymptom("name", []string{"diag1", "diag2"}, nil, nil),
			resources: map[string]any{
				"pod": &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
					},
				},
				"serviceAccount": &v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name: "serviceAccount1",
					},
				},
			},
			expected: strings.TrimSpace(`
name
Diagnoses:
	diag1
	diag2
No suggestions
Resources GVK:
	No GVK set
	No GVK set
---
`) + "\n",
		},
		{
			name:      "No resources",
			symptom:   symptoms.NewSymptom("name", []string{"diag1", "diag2"}, []string{"sugg1", "sugg2"}, nil),
			resources: nil,
			expected: strings.TrimSpace(`
name
Diagnoses:
	diag1
	diag2
Suggestions:
	sugg1
	sugg2
No resources related to this issue.
---
`) + "\n",
		},
		{
			name:      "No symptom",
			symptom:   nil,
			resources: nil,
			expected:  "",
		},
		{
			name:    "Resources with GVK",
			symptom: symptoms.NewSymptom("name", []string{"diag1", "diag2"}, []string{"sugg1", "sugg2"}, nil),
			resources: map[string]any{
				"pod": resourceWithGVK(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pod1",
						},
					},
					schema.GroupVersionKind{Group: "Pod-group", Version: "v1", Kind: "Pod"},
				),
				"serviceAccount": resourceWithGVK(
					&v1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name: "serviceAccount1",
						},
					},
					schema.GroupVersionKind{Group: "Service-account-group", Version: "v1", Kind: "Service account"},
				),
			},
			expected: strings.TrimSpace(`
name
Diagnoses:
	diag1
	diag2
Suggestions:
	sugg1
	sugg2
Resources GVK:
	Pod-group/v1.Pod, pod1
	Service-account-group/v1.Service account, serviceAccount1
---
`) + "\n",
		},
		{
			name:    "Missing resources",
			symptom: symptoms.NewSymptom("name", []string{"diag1", "diag2"}, []string{"sugg1", "sugg2"}, nil),
			resources: map[string]any{
				"csi-driver": nil,
			},
			expected: strings.TrimSpace(`
name
Diagnoses:
	diag1
	diag2
Suggestions:
	sugg1
	sugg2
No resources related to this issue.
---
`) + "\n",
		},
		{
			name:      "Symptoms with diagnoses and suggestions \"Node group\" get ignored",
			symptom:   symptoms.NewSymptomTreeNodeGroup("test", nil).Symptom(),
			resources: nil,
			expected:  "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var issue symptoms.Issue
			if tc.symptom != nil { //Allow tests with nil symptom pointer
				issue = symptoms.NewIssue(&(tc.symptom), tc.resources)
			} else {
				issue = symptoms.NewIssue(nil, tc.resources)
			}
			var buf bytes.Buffer
			err := Print(&buf, issue)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			got := buf.String()
			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("Expected and actual output differ:\n%s", diff)
			}
		})
	}

}
