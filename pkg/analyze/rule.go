package analyze

import (
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

func getFieldValue(path string, obj interface{}) reflect.Value {
	fieldNames := strings.Split(path, ".")
	if fieldNames[0] == "Metadata" {
		fieldNames = fieldNames[1:]
	}
	checkpoints := new([]reflect.Value)
	
	value := reflect.ValueOf(obj).Elem()
	for _, f := range fieldNames {
		switch value.Kind() {
		case reflect.Map:
			value = value.MapIndex(reflect.ValueOf(f))
		case reflect.Struct:
			value = value.FieldByName(f)
		case reflect.Slice:

		default:
			panic(fmt.Sprintf("unknown field type %s", value.Type().String()))
		}
	}
	return value
}

type FieldRelation struct {
	LhsPath string
	RhsPath string
}

func (f *FieldRelation) EvaluateOn(a interface{}, b interface{}) (bool, error) {
	lhs := getFieldValue(f.LhsPath, a)
	rhs := getFieldValue(f.RhsPath, b)
	return reflect.DeepEqual(lhs, rhs), nil
}

type ObjectFields struct {
	obj interface{}
}

func (o *ObjectFields) Has(field string) (exists bool) {
	return getFieldValue(field, o.obj).IsValid()
}

func (o *ObjectFields) Get(field string) (value string) {
	return getFieldValue(field, o.obj).String()
}

var CsiDriverMissing = Rule{
	Diagnosis:   "local-csi-driver CSIDriver, referenced by scylladb-local-xfs StorageClass, is missing",
	Suggestions: "deploy local-csi-driver provisioner",
	Resources: []ResourceCondition{
		{
			Kind: "ScyllaCluster",
			Condition: fields.AndSelectors(
				fields.ParseSelectorOrDie("Status.Conditions.Type = 'StatefulSetControllerProgressing'"),
				fields.ParseSelectorOrDie("Status.Conditions.Type = 'Progressing'")),
		},
		{
			Kind:      "Pod",
			Condition: fields.ParseSelectorOrDie("status.conditions.type==\"PodScheduled\""),
		},
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
	},
}
