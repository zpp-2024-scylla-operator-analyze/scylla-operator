package selectors

import (
	"reflect"
)

type function struct {
	labels []string
	value  reflect.Value
}

func newFunction(labels []string, lambda any) *function {
	typ := reflect.TypeOf(lambda)

	if typ.Kind() != reflect.Func {
		return nil
	}

	if typ.NumIn() != len(labels) {
		return nil
	}

	// TODO: Assert that function is ok

	result := &function{
		labels: labels,
		value:  reflect.ValueOf(lambda),
	}

	// Check if all labels are unique
	if len(labels) != len(result.Labels()) {
		return nil
	}

	return result
}

func (f *function) Labels() map[string]reflect.Type {
	result := make(map[string]reflect.Type, len(f.labels))

	typ := f.value.Type()
	for i := range typ.NumIn() {
		result[f.labels[i]] = typ.In(i)
	}

	return result
}

func (f *function) Call(args map[string]any) {
	in := make([]reflect.Value, 0, len(f.labels))

	for _, label := range f.labels {
		arg, exists := args[label]
		if !exists {
			panic("TODO: Missing argument")
		}

		in = append(in, reflect.ValueOf(arg))
	}

	f.value.Call(in)
}
