package selectors

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"reflect"
)

type Selector struct {
	resources   map[string]reflect.Type
	constraints map[string][]*constraint
	assertion   map[string]*predicate
	relations   []*relation
}

func Type[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}

func Select(label string, typ reflect.Type) *Selector {
	return (&Selector{
		resources:   make(map[string]reflect.Type),
		constraints: make(map[string][]*constraint),
		assertion:   make(map[string]*predicate),
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

func (b *Selector) Assert(label string, f any) *Selector {
	typ, defined := b.resources[label]
	if !defined {
		panic("TODO: Handle undefined labels in Filter")
	}

	assertion := newPredicate(label, f)
	if assertion.Labels()[label] != reflect.PointerTo(typ) {
		panic("TODO: Handle mismatched type in Filter")
	}

	b.assertion[label] = assertion

	return b
}

func (b *Selector) Relate(lhs, rhs string, f any) *Selector {
	// TODO: Check input

	relation := newRelation(lhs, rhs, f)

	b.relations = append(b.relations, relation)

	return b
}

func (b *Selector) Collect() func(*sources.DataSource2) []map[string]any {
	executor := newExecutor(
		b.resources,
		b.constraints,
		b.assertion,
		b.relations,
	)

	return func(ds *sources.DataSource2) []map[string]any {
		result := make([]map[string]any, 0)

		executor.execute(ds, func(resources map[string]any) bool {
			result = append(result, resources)
			return true
		})

		return result
	}
}

func (b *Selector) ForEach(labels []string, function any) func(*sources.DataSource2) {
	for _, label := range labels {
		if _, contains := b.resources[label]; !contains {
			panic("TODO: Handle undefined label")
		}
	}

	callback := newFunction[bool](labels, function)
	executor := newExecutor(
		b.resources,
		b.constraints,
		b.assertion,
		b.relations,
	)

	return func(ds *sources.DataSource2) {
		executor.execute(ds, func(resources map[string]any) bool {
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

func (b *Selector) Any() func(*sources.DataSource2) bool {
	executor := newExecutor(
		b.resources,
		b.constraints,
		b.assertion,
		b.relations,
	)

	return func(ds *sources.DataSource2) bool {
		result := false

		executor.execute(ds, func(_ map[string]any) bool {
			result = true
			return false
		})

		return result
	}
}
