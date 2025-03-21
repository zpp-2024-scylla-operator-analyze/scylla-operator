package symptoms

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

func fieldIsValid(f reflect.Value) bool {
	return f.IsValid() && !f.IsNil()
}

func getConditionByType(res any, condType string) *metav1.Condition {
	r := reflect.ValueOf(res)
	if !fieldIsValid(r) {
		return nil
	}
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}

	status := r.FieldByName("Status")
	if !status.IsValid() {
		return nil
	}

	conditions := status.FieldByName("Conditions")
	if !conditions.IsValid() {
		return nil
	}
	for i := 0; i < conditions.Len(); i++ {
		condEl := conditions.Index(i)
		cond := condEl.Interface().(metav1.Condition)
		if cond.Type == condType {
			return &cond
		}
	}
	return nil
}

func MeetsCondition(res any, condType string, condStatus metav1.ConditionStatus) bool {
	cond := getConditionByType(res, condType)
	return cond != nil && cond.Status == condStatus
}

func MeetsNodeSelectorPlacementRules(node *v1.Node, nodeSelector map[string]string) bool {
	for label, value := range node.Labels {
		v, exists := nodeSelector[label]
		if !exists || v != value {
			return false
		}
	}
	return true
}
