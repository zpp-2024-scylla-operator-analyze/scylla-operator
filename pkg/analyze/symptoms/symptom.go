package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
)

type Symptom interface {
	Check(source *sources.DataSource) (bool, error)
}
