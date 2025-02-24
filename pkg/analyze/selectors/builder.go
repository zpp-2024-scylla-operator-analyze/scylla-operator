package selectors

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"
)

type Selector struct {
	resources   map[string]reflect.Type
	constraints map[string][]*constraint
	relations   []*relation
}

func Type[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}

func Select(label string, typ reflect.Type) *Selector {
	return (&Selector{
		resources:   make(map[string]reflect.Type),
		constraints: make(map[string][]*constraint),
		relations:   make([]*relation, 0),
	}).Select(label, typ)
}

func (b *Selector) Select(label string, typ reflect.Type) *Selector {
	if _, exists := b.resources[label]; exists {
		panic("TODO: Handle duplicate labels")
	}

	b.resources[label] = typ

	return b
}

// SelectPhantom Defines a resource that should not exist in the cluster.
func (b *Selector) SelectPhantom(label string, typ reflect.Type) *Selector {
	panic("not implemented")
}

func (b *Selector) Filter(label string, f any) *Selector {
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

// Relate Relates two resources labeled with lhs, rhs. f need not be commutative.
// For phantom resources, the following rules apply:
//   - phantom - phantom: this relation has no effect,
//   - phantom - non-phantom: guarantees that the non-phantom resource is not related to a non-phantom resource
//     that satisfies all requirements imposed on the phantom resource.
//   - non-phantom - non-phantom: guarantees that the lhs resource is related to the rhs resource which satisfies
//     all requirements imposed on the rhs resource.
func (b *Selector) Relate(lhs, rhs string, f any) *Selector {
	// TODO: Check input

	relation := newRelation(lhs, rhs, f)

	b.relations = append(b.relations, relation)

	return b
}

func eraseSliceType[T any](slice []T, _ error) []any {
	result := make([]any, len(slice))

	for i, _ := range slice {
		result[i] = slice[i]
	}

	return result
}

func fromDataSource(ds *sources.DataSource) map[reflect.Type][]any {
	result := make(map[reflect.Type][]any)

	if ds.PodLister != nil {
		result[reflect.TypeFor[v1.Pod]()] =
			eraseSliceType(ds.PodLister.List(labels.Everything()))
	}

	if ds.ServiceLister != nil {
		result[reflect.TypeFor[v1.Service]()] =
			eraseSliceType(ds.ServiceLister.List(labels.Everything()))
	}

	if ds.SecretLister != nil {
		result[reflect.TypeFor[v1.Secret]()] =
			eraseSliceType(ds.SecretLister.List(labels.Everything()))
	}

	if ds.ConfigMapLister != nil {
		result[reflect.TypeFor[v1.ConfigMap]()] =
			eraseSliceType(ds.ConfigMapLister.List(labels.Everything()))
	}

	if ds.ServiceAccountLister != nil {
		result[reflect.TypeFor[v1.ServiceAccount]()] =
			eraseSliceType(ds.ServiceAccountLister.List(labels.Everything()))
	}

	if ds.ScyllaClusterLister != nil {
		result[reflect.TypeFor[scyllav1.ScyllaCluster]()] =
			eraseSliceType(ds.ScyllaClusterLister.List(labels.Everything()))
	}

	if ds.StorageClassLister != nil {
		result[reflect.TypeFor[storagev1.StorageClass]()] =
			eraseSliceType(ds.StorageClassLister.List(labels.Everything()))
	}

	if ds.CSIDriverLister != nil {
		result[reflect.TypeFor[storagev1.CSIDriver]()] =
			eraseSliceType(ds.CSIDriverLister.List(labels.Everything()))
	}

	return result
}

func (b *Selector) Collect(labels []string, function any) func(*sources.DataSource) {
	for _, label := range labels {
		if _, contains := b.resources[label]; !contains {
			panic("TODO: Handle undefined label")
		}
	}

	callback := newFunction[bool](labels, function)
	executor := newExecutor(b.resources, b.constraints, b.relations)
	proxy := newFunction[bool]([]string{"allArgs"}, func(allArgs map[string]any) {
		callback.Call(allArgs)
	})

	return func(ds *sources.DataSource) {
		executor.execute(fromDataSource(ds), proxy)
	}
}

func (b *Selector) CollectAll(function any) func(*sources.DataSource) {
	callback := newFunction[bool]([]string{"allArgs"}, function)
	executor := newExecutor(b.resources, b.constraints, b.relations)

	return func(ds *sources.DataSource) {
		executor.execute(fromDataSource(ds), callback)
	}
}
