/*
 * GKE-specific issues.
 */
package rules

import (
	"regexp"
	"strings"

	"github.com/scylladb/scylla-operator/pkg/analyze/selector"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms/helpers"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

var GkeSymptoms = []symptoms.SymptomTreeNode{
	buildGkeNodeSetupFsCreationFailureSymptoms(),
}

const (
	nodeSetupFsDegradedPattern = "FilesystemControllerNodeSetup(.*?)Degraded_Error"
)

func buildGkeNodeSetupFsCreationFailureSymptoms() symptoms.SymptomTreeNode {
	mkfsStderr := "mkfs: failed to execute mkfs.xfs: No such file or directory"
	affectedGkeVersions := []string{
		"1.32.2-gke.1182001",
		"v1.32.3-gke.1170000",
	}
	nodeSetupFsDegradedRegexp := regexp.MustCompile(nodeSetupFsDegradedPattern)

	return symptoms.NewSymptomTreeLeaf(
		"gke node setup failure on filesystem creation",
		symptoms.NewSymptomWithExternalIssues(
			"Node setup filesystem creation failure at specific GKE versions",
			[]string{
				"Node setup fails on filesystem creation in specific GKE versions.",
				"Affected versions: " + strings.Join(affectedGkeVersions, ",") + ".",
			},
			nil,
			[]symptoms.ExternalIssueReference{
				// #2579: Node setup fails on filesystem creation in v1.32.3-gke.1170000
				symptoms.NewScyllaOperatorIssueReference(2579),
			},
			selector.
				New().
				Select("node", selector.Type[*v1.Node](), func(node *v1.Node) (bool, error) {
					for _, gkeVersion := range affectedGkeVersions {
						if helpers.RunsGkeVersion(node, gkeVersion) {
							return true, nil
						}
					}
					return false, nil
				}).
				Select("node-config", selector.Type[*scyllav1alpha1.NodeConfig](), nil).
				Relate("node", "node-config", func(node *v1.Node, nodeConfig *scyllav1alpha1.NodeConfig) (bool, error) {
					if !helpers.NodeConfigTunesNode(nodeConfig, node) {
						return false, nil
					}

					for _, cond := range nodeConfig.Status.Conditions {
						if cond.Status == v1.ConditionFalse {
							continue
						}
						fsControllerDegradedReason := nodeSetupFsDegradedRegexp.FindStringSubmatch(cond.Reason)
						if len(fsControllerDegradedReason) == 0 {
							continue
						}
						degradedNode := fsControllerDegradedReason[1]
						if node.Name != degradedNode {
							continue
						}
						if strings.Contains(cond.Message, mkfsStderr) {
							return true, nil
						}
					}
					return false, nil
				}),
		))
}
