package rules

import "github.com/scylladb/scylla-operator/pkg/analyze/symptoms"

var Symptoms = append([]symptoms.SymptomTreeNode{}, StorageSymptoms...)
