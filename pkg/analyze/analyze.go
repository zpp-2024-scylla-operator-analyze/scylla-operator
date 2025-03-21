package analyze

import (
	"context"
	"fmt"
	"github.com/scylladb/scylla-operator/pkg/analyze/snapshot"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms/rules"
	"k8s.io/klog/v2"
)

func Analyze(ctx context.Context, ds snapshot.Snapshot) error {
	klog.Infof("Analyzing the cluster for %d available symptom trees...", len(rules.Symptoms))

	for _, tree := range rules.Symptoms {
		diag, _, err := symptoms.MatchTree(tree, ds)
		if err != nil {
			return fmt.Errorf("Error when matching symptom %s: %v", tree.Symptom().Name(), err)
		}
		if diag != nil {
			for _, d := range diag {
				fmt.Println(d)
				if err != nil {
					return err
				}
			}
		}
	}

	klog.Infof("Scanned the cluster for %d symptom trees", len(rules.Symptoms))
	return nil
}
