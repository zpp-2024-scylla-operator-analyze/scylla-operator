package selectors

import (
	"fmt"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	"strings"
)

func ExampleCollect() {
	s := Select("cluster", Type[scyllav1.ScyllaCluster]()).
		Select("pod", Type[v1.Pod]()).
		Filter("pod", func(pod *v1.Pod) bool {
			return strings.HasPrefix(pod.Name, "scylla-operator-")
		}).
		Relate("cluster", "pod", func(cluster *scyllav1.ScyllaCluster, pod *v1.Pod) bool {
			return cluster.Name == pod.Spec.NodeName
		}).
		Collect(
			[]string{"cluster", "pod"},
			func(cluster *scyllav1.ScyllaCluster, pod *v1.Pod) bool {
				fmt.Printf("\"%s\" \"%s\"\n", cluster.Name, pod.Name)
				return true
			},
		)

	s(nil)

	// Output: "europe-central2" "scylla-operator-1"
	// "europe-central2" "scylla-operator-2"
	// "us-east1" "scylla-operator-3"
}
