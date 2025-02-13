package selectors

import (
	"reflect"
	"slices"
)

type executor struct {
	resource    map[string]reflect.Type
	constraints map[string][]constraint
}

func newExecutor(
	resource map[string]reflect.Type,
	constraints map[string][]constraint,
) *executor {
	return &executor{resource: resource, constraints: constraints}
}

func (e *executor) execute(
	resources map[reflect.Type][]any,
	callback *function,
) {
	// TODO: Assert callback is ok

	ordered := make([]labeled[[]any], 0, len(e.resource))

	for label, resource := range e.resource {
		ordered = append(ordered, labeled[[]any]{
			Label: label,
			Value: resources[resource],
		})
	}

	// TODO: Filter

	slices.SortFunc(ordered, func(a, b labeled[[]any]) int {
		if len(a.Value) < len(b.Value) {
			return -1
		}

		if len(a.Value) > len(b.Value) {
			return 1
		}

		return 0
	})

	tuple := make([]labeled[any], 0, len(ordered))
	e.traverse(ordered, callback, tuple)
}

func (e *executor) traverse(
	resources []labeled[[]any],
	callback *function,
	tuple []labeled[any],
) {
	if len(tuple) >= cap(tuple) {
		e.process(callback, tuple)
		return
	}

	label := resources[len(tuple)].Label
	for _, resource := range resources[len(tuple)].Value {
		// TODO: Check relations

		e.traverse(resources, callback, append(tuple, labeled[any]{
			Label: label,
			Value: resource,
		}))
	}
}

func (e *executor) process(callback *function, tuple []labeled[any]) {
	labels := callback.Labels()
	args := make(map[string]any, len(labels))

	for _, resource := range tuple {
		if _, exists := labels[resource.Label]; !exists {
			continue
		}

		args[resource.Label] = resource.Value
	}

	callback.Call(args)

}
