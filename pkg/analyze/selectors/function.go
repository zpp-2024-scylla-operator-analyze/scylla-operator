package selectors

import (
	"fmt"
	"reflect"
)

type function[R any] struct {
	labels []string
	value  reflect.Value
}

func newFunction[R any](labels []string, lambda any) *function[R] {
	typ := reflect.TypeOf(lambda)

	if typ.Kind() != reflect.Func {
		return nil
	}

	if typ.NumIn() != len(labels) {
		return nil
	}

	if typ.NumOut() != 1 {
		return nil
	}

	// TODO: Assert that function is ok

	result := &function[R]{
		labels: labels,
		value:  reflect.ValueOf(lambda),
	}

	// Check if all labels are unique
	if len(labels) != len(result.Labels()) {
		return nil
	}

	return result
}

func (f *function[R]) Labels() map[string]reflect.Type {
	result := make(map[string]reflect.Type, len(f.labels))

	typ := f.value.Type()
	for i := range typ.NumIn() {
		result[f.labels[i]] = typ.In(i)
	}

	return result
}

func (f *function[R]) Call(args map[string]any) R {
	in := make([]reflect.Value, 0, len(f.labels))

	for i, label := range f.labels {
		arg, exists := args[label]
		if !exists {
			fmt.Printf("expected labels: %v got args: %v", f.labels, args)
			panic("TODO: Missing argument")
		}

		if arg == nil {
			in = append(in, reflect.Zero(f.value.Type().In(i)))
		} else {
			in = append(in, reflect.ValueOf(arg))
		}
	}

	result := f.value.Call(in)

	return result[0].Interface().(R)
}
