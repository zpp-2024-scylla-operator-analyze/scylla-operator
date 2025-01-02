package analyze

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/fields"
	"reflect"
	"strings"
)

type ResourceCondition struct {
	Kind      string
	Condition fields.Selector
}

type Relation interface {
	EvaluateOn(a interface{}, b interface{}) (bool, error)
}

type resourceConnection struct {
	Lhs int
	Rhs int
	Rel Relation
}

type Rule struct {
	Diagnosis   string
	Suggestions string
	Resources   []ResourceCondition
	Relations   []resourceConnection
}

func getFieldValueIterator(node reflect.Value, fields []string) func() *reflect.Value {
	switch node.Kind() {
	case reflect.Ptr:
		return getFieldValueIterator(node.Elem(), fields)
	}

	if len(fields) == 0 {
		called := false
		return func() *reflect.Value {
			if !called {
				called = true
				return &node
			}
			return nil
		}
	}

	switch node.Kind() {
	case reflect.Map:
		return getFieldValueIterator(node.MapIndex(reflect.ValueOf(fields[0])), fields[1:])
	case reflect.Struct:
		return getFieldValueIterator(node.FieldByName(fields[0]), fields[1:])
	case reflect.Slice | reflect.Array:
		i := -1
		iter := func() *reflect.Value { return nil }
		return func() *reflect.Value {
			val := iter()
			for val == nil && i+1 < node.Len() {
				i++
				iter = getFieldValueIterator(node.Index(i), fields)
				val = iter()
			}
			return val
		}
	default:
		panic(errors.New(fmt.Sprintf("unknown field type %s for %v", node.Type().String(), node)))
	}
	return nil
}

func GetFieldValueIterator(path string, obj interface{}) func() *reflect.Value {
	path = strings.Map(func(ch rune) rune {
		if ch == ' ' {
			return -1
		}
		return ch
	}, path)
	fieldNames := strings.Split(path, ".")
	if fieldNames[0] == "Metadata" {
		fieldNames = fieldNames[1:]
	}
	return getFieldValueIterator(reflect.ValueOf(obj), fieldNames)
}

type FieldRelation struct {
	LhsPath string
	RhsPath string
}

func (r *FieldRelation) EvaluateOn(a interface{}, b interface{}) (bool, error) {
	lhsIter := GetFieldValueIterator(r.LhsPath, a)
	rhsIter := GetFieldValueIterator(r.RhsPath, b)
	lhsAll := make([]*reflect.Value, 0)
	for lhs := lhsIter(); lhs != nil; lhs = lhsIter() {
		lhsAll = append(lhsAll, lhs)
	}
	for rhs := rhsIter(); rhs != nil; rhs = rhsIter() {
		for _, lhs := range lhsAll {
			if reflect.DeepEqual(lhs, rhs) {
				return true, nil
			}
		}
	}
	return false, nil
}

type ExistentialRelation struct {
	Exists bool
}

func (r *ExistentialRelation) EvaluateOn(a interface{}, b interface{}) (bool, error) {
	if a != b {
		panic("a and b have to be equal")
	}
	return r.Exists && a != nil || !r.Exists && a == nil, nil
}

var CsiDriverMissing = Rule{
	Diagnosis:   "local-csi-driver CSIDriver, referenced by scylladb-local-xfs StorageClass, is missing",
	Suggestions: "deploy local-csi-driver provisioner",
	Resources: []ResourceCondition{
		{
			Kind: "ScyllaCluster",
			Condition: fields.AndSelectors(
				fields.ParseSelectorOrDie("Status.Conditions.Type=StatefulSetControllerProgressing"),
				fields.ParseSelectorOrDie("Status.Conditions.Type=Progressing"),
				fields.ParseSelectorOrDie("Spec.Datacenter.Racks.Storage.StorageClassName=scylladb-local-xfs")),
		},
		{
			Kind:      "Pod",
			Condition: fields.ParseSelectorOrDie("Status.Conditions.Type=PodScheduled"),
		},
		//{
		//	Kind:      "LocalCsiDriver",
		//	Condition: fields.Everything(),
		//},
	},
	Relations: []resourceConnection{
		{
			Lhs: 0,
			Rhs: 1,
			Rel: &FieldRelation{
				LhsPath: "Metadata.Name",
				RhsPath: "Metadata.Labels.scylla/cluster",
			},
		},
		//{
		//	Lhs: 2,
		//	Rhs: 2,
		//	Rel: &ExistentialRelation{},
		//},
	},
}
