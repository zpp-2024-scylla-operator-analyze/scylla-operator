package symptoms

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestNewIssue(t *testing.T) {
	// given
	mockedSymptom := newEmptyFakeSymptom("mocked symptom")
	mockedResourceMap := map[string]any{
		"pod1": &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1",
			},
		},
		"pod2": &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod2",
			},
		},
		"serviceAccount": &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "serviceAccount1",
			},
		},
	}

	// when
	issue := NewIssue(&mockedSymptom, mockedResourceMap)

	// then
	if (*issue.Symptom).Name() != mockedSymptom.Name() {
		t.Errorf("symptom name mismatch: got %q, want %q", (*issue.Symptom).Name(), mockedSymptom.Name())
	}
	for k, v := range mockedResourceMap {
		if issue.Resources[k] != v {
			t.Errorf("resource key %s mismatch: got %v, want %v", k, issue.Resources[k], v)
		}
	}
}
