package selectors

import (
	"fmt"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	scyllav1listers "github.com/scylladb/scylla-operator/pkg/client/scylla/listers/scylla/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"strings"
)

type mockScyllaClusterLister struct {
	clusters []*scyllav1.ScyllaCluster
}

func (l mockScyllaClusterLister) List(_ labels.Selector) ([]*scyllav1.ScyllaCluster, error) {
	return l.clusters, nil
}

func (l mockScyllaClusterLister) ScyllaClusters(_ string) scyllav1listers.ScyllaClusterNamespaceLister {
	panic("")
}

type mockPodLister struct {
	pods []*v1.Pod
}

func (l mockPodLister) List(_ labels.Selector) ([]*v1.Pod, error) {
	return l.pods, nil
}

func (l mockPodLister) Pods(_ string) corev1listers.PodNamespaceLister {
	panic("")
}

func mockDataSource() *sources.DataSource {
	return &sources.DataSource{
		PodLister: mockPodLister{
			pods: []*v1.Pod{
				&v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "scylla-operator-1",
					},
					Spec: v1.PodSpec{
						NodeName: "europe-central2",
					},
					Status: v1.PodStatus{},
				},
				&v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "scylla-operator-2",
					},
					Spec: v1.PodSpec{
						NodeName: "europe-central2",
					},
					Status: v1.PodStatus{},
				},
				&v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "scylla-operator-3",
					},
					Spec: v1.PodSpec{
						NodeName: "us-east1",
					},
					Status: v1.PodStatus{},
				},
				&v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "dns-1",
					},
					Spec: v1.PodSpec{
						NodeName: "us-east1",
					},
					Status: v1.PodStatus{},
				},
			},
		},
		ServiceLister:        nil,
		SecretLister:         nil,
		ConfigMapLister:      nil,
		ServiceAccountLister: nil,
		ScyllaClusterLister: mockScyllaClusterLister{
			clusters: []*scyllav1.ScyllaCluster{
				&scyllav1.ScyllaCluster{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "europe-central2",
					},
					Spec:   scyllav1.ScyllaClusterSpec{},
					Status: scyllav1.ScyllaClusterStatus{},
				},
				&scyllav1.ScyllaCluster{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "us-east1",
					},
					Spec:   scyllav1.ScyllaClusterSpec{},
					Status: scyllav1.ScyllaClusterStatus{},
				},
			},
		},
	}

}

func ExampleCollect() {
	ds := mockDataSource()

	s := Select("cluster", Type[scyllav1.ScyllaCluster]()).
		Select("pod", Type[v1.Pod]()).
		Filter("pod", func(pod *v1.Pod) bool {
			return strings.HasPrefix(pod.Name, "scylla-operator-")
		}).
		Relate("cluster", "pod", func(cluster *scyllav1.ScyllaCluster, pod *v1.Pod) bool {
			return cluster.Name == pod.Spec.NodeName
		}).
		Collect()

	result := s(ds)
	for _, match := range result {
		cluster := match["cluster"].(*scyllav1.ScyllaCluster)
		pod := match["pod"].(*v1.Pod)
		fmt.Printf("%s %s\n", cluster.Name, pod.Name)
	}

	// Output: europe-central2 scylla-operator-1
	// europe-central2 scylla-operator-2
	// us-east1 scylla-operator-3
}

func ExampleForEach() {
	ds := mockDataSource()

	s := Select("cluster", Type[scyllav1.ScyllaCluster]()).
		Select("pod", Type[v1.Pod]()).
		Filter("pod", func(pod *v1.Pod) bool {
			return strings.HasPrefix(pod.Name, "scylla-operator-")
		}).
		Relate("cluster", "pod", func(cluster *scyllav1.ScyllaCluster, pod *v1.Pod) bool {
			return cluster.Name == pod.Spec.NodeName
		}).
		ForEach(
			[]string{"cluster", "pod"},
			func(cluster *scyllav1.ScyllaCluster, pod *v1.Pod) bool {
				fmt.Printf("%s %s\n", cluster.Name, pod.Name)
				return true
			},
		)

	s(ds)

	// Output: europe-central2 scylla-operator-1
	// europe-central2 scylla-operator-2
	// us-east1 scylla-operator-3
}

func ExampleAny() {
	ds := mockDataSource()

	s1 := Select("cluster", Type[scyllav1.ScyllaCluster]()).
		Filter("cluster", func(cluster *scyllav1.ScyllaCluster) bool {
			return cluster.Name == "europe-central2"
		}).
		Select("pod", Type[v1.Pod]()).
		Filter("pod", func(pod *v1.Pod) bool {
			return strings.HasPrefix(pod.Name, "dns-")
		}).
		Relate("cluster", "pod", func(cluster *scyllav1.ScyllaCluster, pod *v1.Pod) bool {
			return cluster.Name == pod.Spec.NodeName
		}).
		Any()

	s2 := Select("cluster", Type[scyllav1.ScyllaCluster]()).
		Filter("cluster", func(cluster *scyllav1.ScyllaCluster) bool {
			return cluster.Name == "us-east1"
		}).
		Select("pod", Type[v1.Pod]()).
		Filter("pod", func(pod *v1.Pod) bool {
			return strings.HasPrefix(pod.Name, "dns-")
		}).
		Relate("cluster", "pod", func(cluster *scyllav1.ScyllaCluster, pod *v1.Pod) bool {
			return cluster.Name == pod.Spec.NodeName
		}).
		Any()

	fmt.Printf("%t\n", s1(ds))
	fmt.Printf("%t\n", s2(ds))

	// Output: false
	// true
}
