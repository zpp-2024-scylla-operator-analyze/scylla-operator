package selectors

import (
	"fmt"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
)

func ExampleCollect() {
	s := Select("cluster", Type[scyllav1.ScyllaCluster]()).
		Select("pod", Type[v1.Pod]()).
		Filter("pod", func(_ v1.Pod) (bool, error) {
			return true, nil
		}).
		Relate("cluster", "pod", func(_ scyllav1.ScyllaCluster, _ v1.Pod) (bool, error) {
			return true, nil
		}).
		Collect(
			[]string{"cluster", "pod"},
			func(cluster scyllav1.ScyllaCluster, pod v1.Pod) {
				fmt.Printf("\"%s\" \"%s\"\n", cluster.Name, pod.Name)
			},
		)

	s(nil)

	// Output: "europe-central2" "scylla-operator-1"
	// "europe-central2" "scylla-operator-2"
	// "europe-central2" "scylla-operator-3"
	// "us-east1" "scylla-operator-1"
	// "us-east1" "scylla-operator-2"
	// "us-east1" "scylla-operator-3"
}
