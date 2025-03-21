package analyze

import (
	"context"
	"fmt"
	"github.com/scylladb/scylla-operator/pkg/analyze/front"
	"github.com/scylladb/scylla-operator/pkg/analyze/snapshot"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms/rules"
	"k8s.io/klog/v2"
	"os"
)

func Analyze(ctx context.Context, ds snapshot.Snapshot) error {
	klog.Infof("Analyzing the cluster for %d available symptom trees...", len(rules.Symptoms))

	foundIssues := 0
	for _, tree := range rules.Symptoms {
		diag, _, err := symptoms.MatchTree(tree, ds)
		if err != nil {
			return fmt.Errorf("Error when matching symptom %s: %v", tree.Symptom().Name(), err)
		}
		if diag != nil {
			foundIssues++
			for _, d := range diag {
				err = front.Print(os.Stdout, d)
				if err != nil {
					return err
				}
			}
		}
	}

	klog.Infof("Scanned the cluster for %d symptom trees, %d issue%s found", len(rules.Symptoms), foundIssues, map[bool]string{true: "s", false: ""}[foundIssues != 1])
	return nil
}
