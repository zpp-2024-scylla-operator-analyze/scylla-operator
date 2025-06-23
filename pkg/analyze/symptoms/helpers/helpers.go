package helpers

import (
	"reflect"
	"strings"

	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func RunsGkeVersion(node *v1.Node, gkeVersion string) bool {
	runsGke := strings.HasSuffix(node.Status.NodeInfo.KernelVersion, "-gke")
	if !runsGke {
		return false
	}
	return node.Status.NodeInfo.KubeletVersion == gkeVersion
}

func NodeConfigTunesNode(nodeConfig *scyllav1alpha1.NodeConfig, node *v1.Node) bool {
	for _, nodeStatus := range nodeConfig.Status.NodeStatuses {
		if !nodeStatus.TunedNode {
			continue
		}
		if nodeStatus.Name == node.Name {
			return true
		}
	}
	return false
}
