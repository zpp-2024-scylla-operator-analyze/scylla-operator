package selectors

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/snapshot"
	"reflect"
	"slices"
)

type executor struct {
	resource    map[string]reflect.Type
	constraints map[string][]*constraint
	assertion   map[string]*predicate
	relations   []*relation
}

func newExecutor(
	resource map[string]reflect.Type,
	constraints map[string][]*constraint,
	assertion map[string]*predicate,
	relations []*relation,
) *executor {
	return &executor{
		resource:    resource,
		constraints: constraints,
		assertion:   assertion,
		relations:   relations,
	}
}

func filter(resources []any, label string, constraints []*constraint) []any {
	result := make([]any, 0, len(resources)+1)

resourceLoop:
	for _, resource := range append(resources, nil) {
		for _, constraint := range constraints {
			if !constraint.Check(label, resource) {
				continue resourceLoop
			}
		}

		result = append(result, resource)
	}

	return result
}

func (e *executor) execute(
	ds snapshot.Snapshot,
	callback func(map[string]any) bool,
) {
	// TODO: Assert callback is ok

	ordered := make([]labeled[[]any], 0, len(e.resource))

	for label, resource := range e.resource {
		ordered = append(ordered, labeled[[]any]{
			Label: label,
			Value: filter(ds.List(resource), label, e.constraints[label]),
		})
	}

	slices.SortFunc(ordered, func(a, b labeled[[]any]) int {
		if len(a.Value) < len(b.Value) {
			return -1
		}

		if len(a.Value) > len(b.Value) {
			return 1
		}

		return 0
	})

	position := make(map[string]int)
	for idx, resources := range ordered {
		position[resources.Label] = idx
	}

	relations := make([]map[int]*relation, len(ordered))
	for i := range relations {
		relations[i] = make(map[int]*relation)
	}

	for _, relation := range e.relations {
		firstLabel, secondLabel := relation.Labels()

		first := position[firstLabel]
		second := position[secondLabel]

		if first > second {
			first, second = second, first
		}

		relations[second][first] = relation
	}

	tuple := make([]labeled[any], 0, len(ordered))
	e.traverse(ordered, relations, callback, tuple)
}

func (e *executor) check(
	lhs, rhs labeled[any],
	lhsResources, rhsResources []any,
	relation *relation,
) bool {
	if lhs.Value == nil && rhs.Value == nil {
		return false
	}

	if lhs.Value == nil && rhs.Value != nil {
		for _, resource := range lhsResources {
			if resource == nil {
				continue
			}

			labeled := labeled[any]{
				Label: lhs.Label,
				Value: resource,
			}

			if relation.Check(labeled, rhs) {
				return false
			}
		}

		return true
	}

	if lhs.Value != nil && rhs.Value == nil {
		for _, resource := range rhsResources {
			if resource == nil {
				continue
			}

			labeled := labeled[any]{
				Label: rhs.Label,
				Value: resource,
			}

			if relation.Check(lhs, labeled) {
				return false
			}
		}

		return true
	}

	if lhs.Value != nil && rhs.Value != nil {
		return relation.Check(lhs, rhs)
	}

	panic("Unreachable")
}

func (e *executor) traverse(
	resources []labeled[[]any],
	relations []map[int]*relation,
	callback func(map[string]any) bool,
	tuple []labeled[any],
) bool {
	if len(tuple) >= cap(tuple) {
		return e.process(callback, tuple)
	}

rhs:
	for _, selected := range resources[len(tuple)].Value {
		rhs := labeled[any]{
			Label: resources[len(tuple)].Label,
			Value: selected,
		}

		assertion, exists := e.assertion[rhs.Label]
		if exists && !assertion.Check(rhs.Label, rhs.Value) {
			continue rhs
		}

	lhs:
		for i, lhs := range tuple {
			relation, exists := relations[len(tuple)][i]
			if !exists {
				continue lhs
			}

			if !e.check(
				lhs, rhs,
				resources[i].Value, resources[len(tuple)].Value,
				relation,
			) {
				continue rhs
			}
		}

		if !e.traverse(
			resources, relations, callback, append(tuple, rhs),
		) {
			return false
		}
	}

	return true
}

func (e *executor) process(
	callback func(map[string]any) bool,
	tuple []labeled[any],
) bool {
	args := make(map[string]any, len(tuple))

	for _, resource := range tuple {
		args[resource.Label] = resource.Value
	}

	return callback(args)
}
