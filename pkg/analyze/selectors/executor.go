package selectors

import (
	"reflect"
	"slices"
)

type executor struct {
	resource    map[string]reflect.Type
	constraints map[string][]*constraint
	relations   []*relation
}

func newExecutor(
	resource map[string]reflect.Type,
	constraints map[string][]*constraint,
	relations []*relation,
) *executor {
	return &executor{
		resource:    resource,
		constraints: constraints,
		relations:   relations,
	}
}

func filter(resources []any, label string, constraints []*constraint) []any {
	result := make([]any, 0, len(resources))

resourceLoop:
	for _, resource := range resources {
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
	resources map[reflect.Type][]any,
	callback *function[bool],
) {
	// TODO: Assert callback is ok

	ordered := make([]labeled[[]any], 0, len(e.resource))

	for label, resource := range e.resource {
		ordered = append(ordered, labeled[[]any]{
			Label: label,
			Value: filter(resources[resource], label, e.constraints[label]),
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

	relations := make([]map[int][]*relation, len(ordered))
	for i, _ := range relations {
		relations[i] = make(map[int][]*relation)
	}

	for _, relation := range e.relations {
		firstLabel, secondLabel := relation.Labels()

		first := position[firstLabel]
		second := position[secondLabel]

		if first > second {
			first, second = second, first
		}

		relations[second][first] = append(relations[second][first], relation)
	}

	tuple := make([]labeled[any], 0, len(ordered))
	e.traverse(ordered, relations, callback, tuple)
}

func (e *executor) traverse(
	resources []labeled[[]any],
	relations []map[int][]*relation,
	callback *function[bool],
	tuple []labeled[any],
) bool {

	if len(tuple) >= cap(tuple) {
		return e.process(callback, tuple)
	}

	label := resources[len(tuple)].Label
outer:
	for _, resource := range resources[len(tuple)].Value {
		labeled := labeled[any]{Label: label, Value: resource}

		for other, relations := range relations[len(tuple)] {
			for _, relation := range relations {
				if !relation.Check(tuple[other], labeled) {
					continue outer
				}
				//fmt.Println(relation.Check(tuple[other], labeled))
			}
		}

		if !e.traverse(resources, relations, callback, append(tuple, labeled)) {
			return false
		}
	}

	return true
}

func (e *executor) process(callback *function[bool], tuple []labeled[any]) bool {
	labels := callback.Labels()
	args := make(map[string]any, len(labels))

	for _, resource := range tuple {
		if _, exists := labels[resource.Label]; !exists {
			continue
		}

		args[resource.Label] = resource.Value
	}

	return callback.CallAsOne(args)
}
