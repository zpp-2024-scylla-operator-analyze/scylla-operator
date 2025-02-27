package selectors

import (
	"fmt"
	"github.com/scylladb/scylla-operator/pkg/analyze/sources"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	// "strings"
)

func nameScyllaCluster(cluster any) string {
	if cluster == nil {
		return "nil"
	}

	return cluster.(*scyllav1.ScyllaCluster).Name
}

func nameStorageClass(class any) string {
	if class == nil {
		return "nil"
	}

	return class.(*storagev1.StorageClass).Name
}

func nameCSIDriver(driver any) string {
	if driver == nil {
		return "nil"
	}

	return driver.(*storagev1.CSIDriver).Name
}

func strptr(s string) *string {
	return &s
}

func ExampleMissingCSIDriver() {
	resources := map[reflect.Type][]any{
		Type[*scyllav1.ScyllaCluster](): []any{
			&scyllav1.ScyllaCluster{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "europe-central2",
				},
				Spec: scyllav1.ScyllaClusterSpec{
					Datacenter: scyllav1.DatacenterSpec{
						Racks: []scyllav1.RackSpec{
							scyllav1.RackSpec{
								Storage: scyllav1.Storage{
									StorageClassName: strptr("scylladb-local-xfs"),
								},
							},
						},
					},
				},
				Status: scyllav1.ScyllaClusterStatus{},
			},
			&scyllav1.ScyllaCluster{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "us-east1",
				},
				Spec: scyllav1.ScyllaClusterSpec{
					Datacenter: scyllav1.DatacenterSpec{
						Racks: []scyllav1.RackSpec{
							scyllav1.RackSpec{
								Storage: scyllav1.Storage{
									StorageClassName: strptr("scylladb-dummy"),
								},
							},
						},
					},
				},
				Status: scyllav1.ScyllaClusterStatus{},
			},
		},
		Type[*v1.Pod](): []any{},
		Type[*storagev1.StorageClass](): []any{
			&storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "scylladb-local-xfs",
				},
				Provisioner: "local.csi.scylladb.com",
			},
			&storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "scylladb-dummy",
				},
				Provisioner: "dummy.csi.scylladb.com",
			},
		},
		Type[*storagev1.CSIDriver](): []any{
			&storagev1.CSIDriver{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummy.csi.scylladb.com",
				},
			},
		},
	}

	snapshot := &sources.DataSource2{
		Objects: resources,
	}

	builder := Select("scylla-cluster", Type[*scyllav1.ScyllaCluster]()).
		Select("storage-class", Type[*storagev1.StorageClass]()).
		Select("csi-driver", Type[*storagev1.CSIDriver]()).
		Filter("scylla-cluster", func(c *scyllav1.ScyllaCluster) bool {
			return c != nil
		}).
		Filter("storage-class", func(s *storagev1.StorageClass) bool {
			return s != nil
		}).
		Assert("csi-driver", func(d *storagev1.CSIDriver) bool {
			return d == nil
		}).
		Relate("scylla-cluster", "storage-class", func(c *scyllav1.ScyllaCluster, s *storagev1.StorageClass) bool {
			for _, rack := range c.Spec.Datacenter.Racks {
				if *rack.Storage.StorageClassName == s.Name {
					return true
				}
			}
			return false
		}).
		Relate("storage-class", "csi-driver", func(s *storagev1.StorageClass, d *storagev1.CSIDriver) bool {
			return s.Provisioner == d.Name
		})

	selectorAny := builder.Any()
	selectorCollect := builder.CollectAll()

	fmt.Printf("%t\n", selectorAny(snapshot))

	for _, tuple := range selectorCollect(snapshot) {
		fmt.Printf("%s %s %s\n",
			nameScyllaCluster(tuple["scylla-cluster"]),
			nameStorageClass(tuple["storage-class"]),
			nameCSIDriver(tuple["csi-driver"]),
		)
	}

	// Output: true
	// europe-central2 scylladb-local-xfs nil
}
