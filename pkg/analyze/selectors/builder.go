package selectors

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"
)

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

func (b *builder) Collect() func(*sources.DataSource) []map[string]any {
	executor := newExecutor(b.resources, b.constraints, b.relations)

	return func(ds *sources.DataSource) []map[string]any {
		result := make([]map[string]any, 0)

		executor.execute(fromDataSource(ds), func(resources map[string]any) bool {
			result = append(result, resources)
			return true
		})

		return result
	}
}

func (b *builder) ForEach(labels []string, function any) func(*sources.DataSource) {
	for _, label := range labels {
		if _, contains := b.resources[label]; !contains {
			panic("TODO: Handle undefined label")
		}
	}

	callback := newFunction[bool](labels, function)
	executor := newExecutor(b.resources, b.constraints, b.relations)

	return func(ds *sources.DataSource) {
		executor.execute(fromDataSource(ds), func(resources map[string]any) bool {
			labels := callback.Labels()
			args := make(map[string]any, len(labels))

			for label, resource := range resources {
				if _, exists := labels[label]; !exists {
					continue
				}

				args[label] = resource
			}

			return callback.Call(args)
		})
	}
}

func (b *builder) Any() func(*sources.DataSource) bool {
	executor := newExecutor(b.resources, b.constraints, b.relations)

	return func(ds *sources.DataSource) bool {
		result := false

		executor.execute(fromDataSource(ds), func(_ map[string]any) bool {
			result = true
			return false
		})

		return result
	}

}
