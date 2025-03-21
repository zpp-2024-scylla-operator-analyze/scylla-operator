package rules

import (
	"regexp"
	"strings"

	"github.com/scylladb/scylla-operator/pkg/analyze/selector"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	scyllav1alpha1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1alpha1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

const (
	csiDriverContainerName                     = "local-csi-driver"
	nodeConfigConditionPattern                 = "NodeSetup(?P<node>.*)Degraded"
	nodeConfigNonexistentVolumeMessageFragment = "resolve RAID device"
)

var (
	nodeConfigConditionRegexp = regexp.MustCompile(nodeConfigConditionPattern)
)

var StorageSymptoms = []symptoms.SymptomTreeNode{
	buildLocalCsiDriverMissingSymptoms(),
	buildStorageClassMissingSymptoms(),
	buildNodeConfigSymptoms(),
}

func buildLocalCsiDriverMissingSymptoms() symptoms.SymptomTreeNode {
	// local-csi-driver CSIDriver, referenced by scylladb-local-xfs StorageClass, is missing
	csiDriverMissing := symptoms.NewSymptom("CSIDriver is missing",
		[]string{"local-csi-driver CSIDriver, referenced by scylladb-local-xfs StorageClass, is missing"},
		[]string{"deploy local-csi-driver provisioner (or change StorageClass)"},
		selector.
			New().
			Select("scylla-cluster", selector.Type[*scyllav1.ScyllaCluster](), func(c *scyllav1.ScyllaCluster) (bool, error) {
				statefulSetControllerProgressing := false
				progressing := false
				storageClassXfs := false
				for _, rack := range c.Spec.Datacenter.Racks {
					if *rack.Storage.StorageClassName == "scylladb-local-xfs" {
						storageClassXfs = true
					}
				}
				for _, cond := range c.Status.Conditions {
					if cond.Type == "StatefulSetControllerProgressing" && cond.Status == "True" {
						statefulSetControllerProgressing = true
					} else if cond.Type == "Progressing" && cond.Status == "True" {
						progressing = true
					}
				}
				return storageClassXfs && (statefulSetControllerProgressing || progressing), nil
			}).
			Select("scylla-pod", selector.Type[*v1.Pod](), func(p *v1.Pod) (bool, error) {
				return p.Status.Phase == v1.PodPending, nil
			}).
			Select("storage-class", selector.Type[*storagev1.StorageClass](), nil).
			SelectWithNil("csi-driver", selector.Type[*storagev1.CSIDriver](), nil).
			Relate("scylla-cluster", "storage-class", func(c *scyllav1.ScyllaCluster, sc *storagev1.StorageClass) (bool, error) {
				for _, rack := range c.Spec.Datacenter.Racks {
					if *rack.Storage.StorageClassName == sc.Name {
						return true, nil
					}
				}
				return false, nil
			}).
			Relate("scylla-cluster", "scylla-pod", func(c *scyllav1.ScyllaCluster, p *v1.Pod) (bool, error) {
				return c.Name == p.Labels["scylla/cluster"], nil
			}).
			Relate("storage-class", "csi-driver", func(sc *storagev1.StorageClass, d *storagev1.CSIDriver) (bool, error) {
				return sc.Provisioner == d.Name, nil
			}))

	csiDriverMissingSymptoms := symptoms.NewSymptomTreeLeaf("csi-driver-missing", csiDriverMissing)
	return csiDriverMissingSymptoms
}

func buildStorageClassMissingSymptoms() symptoms.SymptomTreeNode {
	// scylladb-local-xfs StorageClass used by a ScyllaCluster is missing
	notDeployedStorageClass := symptoms.NewSymptom("StorageClass is missing",
		[]string{"scylladb-local-xfs StorageClass used by a ScyllaCluster is missing"},
		[]string{"deploy scylladb-local-xfs StorageClass (or change StorageClass)"},
		selector.
			New().
			Select("scylla-cluster", selector.Type[*scyllav1.ScyllaCluster](), func(c *scyllav1.ScyllaCluster) (bool, error) {
				statefulSetControllerProgressing := false
				progressing := false
				storageClassXfs := false
				for _, rack := range c.Spec.Datacenter.Racks {
					if *rack.Storage.StorageClassName == "scylladb-local-xfs" {
						storageClassXfs = true
					}
				}
				for _, cond := range c.Status.Conditions {
					if cond.Type == "StatefulSetControllerProgressing" && cond.Status == "True" {
						statefulSetControllerProgressing = true
					} else if cond.Type == "Progressing" && cond.Status == "True" {
						progressing = true
					}
				}
				return storageClassXfs && (statefulSetControllerProgressing || progressing), nil
			}).
			SelectWithNil("storage-class", selector.Type[*storagev1.StorageClass](), nil).
			Select("scylla-pod", selector.Type[*v1.Pod](), func(p *v1.Pod) (bool, error) {
				podScheduled := true
				unboundPVC := false
				for _, cond := range p.Status.Conditions {
					if cond.Type == "PodScheduled" {
						if strings.Contains(cond.Message, "pod has unbound immediate PersistentVolumeClaims") {
							unboundPVC = true
						}
						if cond.Status == "False" {
							podScheduled = false
						}
					}
				}
				return !podScheduled && unboundPVC, nil
			}).
			Select("csi-driver", selector.Type[*storagev1.CSIDriver](), nil).
			Select("pod-pvc", selector.Type[*v1.PersistentVolumeClaim](), func(pvc *v1.PersistentVolumeClaim) (bool, error) {
				return pvc.Status.Phase == v1.ClaimPending, nil
			}).
			Relate("scylla-cluster", "scylla-pod", func(c *scyllav1.ScyllaCluster, p *v1.Pod) (bool, error) {
				return c.Name == p.Labels["scylla/cluster"], nil
			}).
			Relate("scylla-cluster", "storage-class", func(c *scyllav1.ScyllaCluster, sc *storagev1.StorageClass) (bool, error) {
				for _, rack := range c.Spec.Datacenter.Racks {
					if *rack.Storage.StorageClassName == sc.Name {
						return true, nil
					}
				}
				return false, nil
			}).
			Relate("storage-class", "csi-driver", func(sc *storagev1.StorageClass, d *storagev1.CSIDriver) (bool, error) {
				return sc.Provisioner == d.Name, nil
			}).
			Relate("pod-pvc", "scylla-pod", func(pvc *v1.PersistentVolumeClaim, p *v1.Pod) (bool, error) {
				for _, volume := range p.Spec.Volumes {
					vPvc := volume.PersistentVolumeClaim
					if vPvc != nil && (*vPvc).ClaimName == pvc.Name {
						return true, nil
					}
				}
				return false, nil
			}).
			Relate("pod-pvc", "storage-class", func(pvc *v1.PersistentVolumeClaim, sc *storagev1.StorageClass) (bool, error) {
				return pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == sc.Name, nil
			}))

	return symptoms.NewSymptomTreeLeaf("StorageClass missing", notDeployedStorageClass)
}

func buildNodeConfigSymptoms() symptoms.SymptomTreeNode {
	// E2Es for misconfigured NodeConfigs
	nodeConfigClusterWideNonexistentVolume := symptoms.NewSymptom(
		"NodeConfig doesn't provision storage for local-csi-driver's volumes-dir",
		[]string{
			"driver is unable to provide storage due to NodeConfig being misconfigured",
			"this may be a false-positive in clusters which don't use NodeConfig to mount `volumes-dir` of local-csi-drivers' Pods",
		},
		[]string{
			"fix the NodeConfig (see node-config's error conditions), then rolling restart local-csi-driver",
		},
		selector.
			New().
			Select("scylla-cluster", selector.Type[*scyllav1.ScyllaCluster](), func(c *scyllav1.ScyllaCluster) (bool, error) {
				return !symptoms.MeetsCondition(c, "Available", "True"), nil
			}).
			Select("scylla-pod", selector.Type[*v1.Pod](), func(p *v1.Pod) (bool, error) {
				return p.Status.Phase != v1.PodSucceeded, nil
			}).
			Select("pod-pvc", selector.Type[*v1.PersistentVolumeClaim](), func(pvc *v1.PersistentVolumeClaim) (bool, error) {
				return pvc.Status.Phase == v1.ClaimPending && pvc.Spec.StorageClassName != nil, nil
			}).
			// TODO: check every storage class
			Select("storage-class", selector.Type[*storagev1.StorageClass](), nil).
			Select("csi-driver", selector.Type[*storagev1.CSIDriver](), nil).
			Select("scylla-node", selector.Type[*v1.Node](), nil).
			Select("node-config", selector.Type[*scyllav1alpha1.NodeConfig](), nil).
			Select("csi-pod", selector.Type[*v1.Pod](), func(d *v1.Pod) (bool, error) {
				for _, cont := range d.Spec.Containers {
					if cont.Name == csiDriverContainerName {
						return true, nil
					}
				}
				return false, nil
			}).
			Relate("scylla-cluster", "scylla-pod", func(c *scyllav1.ScyllaCluster, p *v1.Pod) (bool, error) {
				return c.Name == p.Labels["scylla/cluster"], nil
			}).
			Relate("scylla-pod", "pod-pvc", func(p *v1.Pod, pvc *v1.PersistentVolumeClaim) (bool, error) {
				for _, volume := range p.Spec.Volumes {
					if volume.PersistentVolumeClaim != nil {
						volumePvc := *volume.PersistentVolumeClaim
						if volumePvc.ClaimName == pvc.Name {
							return true, nil
						}
					}
				}
				return false, nil
			}).
			Relate("pod-pvc", "storage-class", func(pvc *v1.PersistentVolumeClaim, s *storagev1.StorageClass) (bool, error) {
				return *pvc.Spec.StorageClassName == s.Name, nil
			}).
			Relate("storage-class", "csi-driver", func(s *storagev1.StorageClass, csi *storagev1.CSIDriver) (bool, error) {
				return s.Provisioner == csi.Name, nil
			}).
			Relate("scylla-node", "scylla-cluster", func(n *v1.Node, c *scyllav1.ScyllaCluster) (bool, error) {
				// TODO: relate the node to the cluster (with node placement rules?)
				return true, nil
			}).
			Relate("node-config", "scylla-node", func(nc *scyllav1alpha1.NodeConfig, n *v1.Node) (bool, error) {
				return symptoms.MeetsNodeSelectorPlacementRules(n, nc.Spec.Placement.NodeSelector), nil
			}).
			// node-config tunes the node on which csi-pod is placed
			Relate("csi-pod", "node-config", func(csiPod *v1.Pod, nodeConfig *scyllav1alpha1.NodeConfig) (bool, error) {
				matchesPod := false
				dirsMounted := false
				for _, status := range nodeConfig.Status.NodeStatuses {
					if status.Name == csiPod.Spec.NodeName {
						matchesPod = true
					}
				}
				if !matchesPod {
					return false, nil
				}
				volumesDir := ""
				volumesMountPoint := ""
				for _, container := range csiPod.Spec.Containers {
					csiContainer := false
					for _, arg := range container.Args {
						if strings.Contains(arg, "--volumes-dir=") {
							volumesDir = strings.Split(arg, "--volumes-dir=")[1]
							volumesDir = strings.TrimSpace(volumesDir)
							csiContainer = true
						}
					}
					if csiContainer {
						for _, mnt := range container.VolumeMounts {
							if mnt.MountPath == volumesDir {
								volumesMountPoint = mnt.Name
							}
						}
					}
				}

				if volumesMountPoint != "" {
					for _, vol := range csiPod.Spec.Volumes {
						if vol.Name == volumesMountPoint {
							for _, mp := range nodeConfig.Spec.LocalDiskSetup.Mounts {
								if (*vol.HostPath).Path == mp.MountPoint {
									dirsMounted = true
								}
							}
						}
					}
				}

				return !dirsMounted, nil
			}))

	// Detects the non-existent device just by conditions
	nodeConfigNonexistentDevice := symptoms.NewSymptom(
		"NodeConfig non-existent device condition",
		[]string{"NodeConfig configured with a non-existent device"},
		[]string{"fix NodeConfig"},
		selector.
			New().
			Select("node-config", selector.Type[*scyllav1alpha1.NodeConfig](), func(nc *scyllav1alpha1.NodeConfig) (bool, error) {
				for _, cond := range nc.Status.Conditions {
					match := nodeConfigConditionRegexp.FindStringSubmatch(cond.Message)
					if len(match) == 0 {
						continue
					}
					if cond.Status != v1.ConditionTrue {
						continue
					}
					if strings.Contains(cond.Message, nodeConfigNonexistentVolumeMessageFragment) {
						return true, nil
					}
				}
				return false, nil
			}))

	return symptoms.NewSymptomTreeNodeGroup(
		"NodeConfig nonexistent device",
		symptoms.OrCondition,
		symptoms.NewSymptomTreeLeaf(nodeConfigClusterWideNonexistentVolume.Name(), nodeConfigClusterWideNonexistentVolume),
		symptoms.NewSymptomTreeLeaf(nodeConfigNonexistentDevice.Name(), nodeConfigNonexistentDevice))
}
