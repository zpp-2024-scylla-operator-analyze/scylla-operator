package rules

import (
	"slices"
)

var Symptoms = slices.Concat(StorageSymptoms, GkeSymptoms)
