package symptoms

import (
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	"k8s.io/apimachinery/pkg/labels"
)

func NoOperator(ds *sources.DataSource) (bool, error) {
	sel, err := labels.Parse("app.kubernetes.io/name == scylla-operator")
	if err != nil {
		return false, err
	}

	list, err := ds.PodLister.List(sel)
	if err != nil {
		return false, err
	}

	return len(list) == 0, nil
}
