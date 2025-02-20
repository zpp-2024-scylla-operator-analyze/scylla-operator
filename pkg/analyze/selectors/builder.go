package selectors

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"reflect"
)

// Mock's dependencies
import (
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mockDataSourceProcessing(_ *sources.DataSource) map[reflect.Type][]any {
	return map[reflect.Type][]any{
		reflect.TypeFor[scyllav1.ScyllaCluster](): []any{
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
		reflect.TypeFor[v1.Pod](): []any{
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
					Name: "dns",
				},
				Spec: v1.PodSpec{
					NodeName: "us-east1",
				},
				Status: v1.PodStatus{},
			},
		},
	}
}

type builder struct {
	resources   map[string]reflect.Type
	constraints map[string][]*constraint
	relations   []*relation
}

func Type[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}

func Select(label string, typ reflect.Type) *builder {
	return (&builder{
		resources:   make(map[string]reflect.Type),
		constraints: make(map[string][]*constraint),
		relations:   make([]*relation, 0),
	}).Select(label, typ)
}

func (b *builder) Select(label string, typ reflect.Type) *builder {
	if _, exists := b.resources[label]; exists {
		panic("TODO: Handle duplicate labels")
	}

	b.resources[label] = typ

	return b
}

func (b *builder) Filter(label string, f any) *builder {
	typ, defined := b.resources[label]
	if !defined {
		panic("TODO: Handle undefined labels in Filter")
	}

	constraint := newConstraint(label, f)
	if constraint.Labels()[label] != reflect.PointerTo(typ) {
		panic("TODO: Handle mismatched type in Filter")
	}

	b.constraints[label] = append(b.constraints[label], constraint)

	return b
}

func (b *builder) Relate(lhs, rhs string, f any) *builder {
	// TODO: Check input

	relation := newRelation(lhs, rhs, f)

	b.relations = append(b.relations, relation)

	return b
}

func (b *builder) Collect(labels []string, function any) func(*sources.DataSource) {
	for _, label := range labels {
		if _, contains := b.resources[label]; !contains {
			panic("TODO: Handle undefined label")
		}
	}

	callback := newFunction[bool](labels, function)
	executor := newExecutor(b.resources, b.constraints, b.relations)

	return func(ds *sources.DataSource) {
		executor.execute(mockDataSourceProcessing(ds), callback)
	}
}
