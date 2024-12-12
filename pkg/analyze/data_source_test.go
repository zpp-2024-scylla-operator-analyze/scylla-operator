package analyze

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/client/scylla/clientset/versioned/fake"
	"github.com/scylladb/scylla-operator/pkg/helpers/slices"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"sort"
	"testing"
)

func TestBuildLister(t *testing.T) {
	t.Parallel()
	buildListerDummyTests := []struct {
		name               string
		listerFactory      func(i cache.Indexer) interface{}
		listerFunc         func(lister interface{}) ([]runtime.Object, error)
		expectedListerType reflect.Type
		expectedObjects    []runtime.Object
	}{
		{
			name: "empty pod lister",
			listerFactory: func(i cache.Indexer) interface{} {
				return v1.NewPodLister(i)
			},
			listerFunc: func(lister interface{}) ([]runtime.Object, error) {
				objects, err := lister.(v1.PodLister).List(labels.Everything())
				return slices.ConvertSlice(objects, convertToRuntimeObject), err
			},
			expectedListerType: reflect.TypeFor[*v1.PodLister](),
		},
		{
			name: "nonempty pod lister",
			listerFactory: func(i cache.Indexer) interface{} {
				return v1.NewPodLister(i)
			},
			listerFunc: func(lister interface{}) ([]runtime.Object, error) {
				objects, err := lister.(v1.PodLister).List(labels.Everything())
				return slices.ConvertSlice(objects, convertToRuntimeObject), err
			},
			expectedListerType: reflect.TypeFor[*v1.PodLister](),
			expectedObjects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "test",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "test",
					},
				},
			},
		},
	}

	for _, tc := range buildListerDummyTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var objectsAsRawExtension []runtime.RawExtension
			for _, obj := range tc.expectedObjects {
				objectsAsRawExtension = append(objectsAsRawExtension, runtime.RawExtension{
					Object: obj,
				})
			}

			ctx := context.Background()
			listFunc := func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
				return &metav1.List{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items:    objectsAsRawExtension,
				}, nil
			}

			lister, err := BuildLister(ctx, tc.listerFactory, listFunc)
			if err != nil {
				t.Fatalf("BuildLister returned an error: %v", err)
			}
			if equality.Semantic.DeepEqual(lister, tc.expectedListerType) {
				t.Fatalf("expected '%v' got: '%v'", tc.expectedListerType, reflect.TypeOf(lister))
			}

			objects, err := tc.listerFunc(lister)
			if err != nil {
				t.Fatalf("dummyLister.List returned an error: %v", err)
			}

			sort.Slice(objects, func(i, j int) bool {
				return compareRuntimeObjects(objects[i], objects[j])
			})

			if !equality.Semantic.DeepEqual(objects, tc.expectedObjects) {
				t.Errorf("expected and actual objects differ: %s", cmp.Diff(tc.expectedObjects, objects))
			}
		})
	}
}

func TestNewDataSourceFromClients_SingleLister(t *testing.T) {
	t.Parallel()
	newDataSourceTests := []struct {
		name              string
		kubernetesObjects []runtime.Object
		scyllaObjects     []runtime.Object
		listerFunc        func(*DataSource) ([]runtime.Object, error)
		expectedObjects   []runtime.Object
		expectedErr       error
	}{
		{
			name: "empty pod list",
			kubernetesObjects: func() []runtime.Object {
				return slices.FilterOut(newBasicKubernetesObjects(), func(obj runtime.Object) bool {
					_, ok := obj.(*corev1.Pod)
					return ok
				})
			}(),
			scyllaObjects: newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.PodLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list pods: %w", err)
				}
				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty pod list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.PodLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list pods: %w", err)
				}

				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "test",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "test",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "empty service list",
			kubernetesObjects: func() []runtime.Object {
				return slices.FilterOut(newBasicKubernetesObjects(), func(obj runtime.Object) bool {
					_, ok := obj.(*corev1.Service)
					return ok
				})
			}(),
			scyllaObjects: newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				pods, err := ds.ServiceLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list services: %w", err)
				}
				return slices.ConvertSlice(pods, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty service list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.ServiceLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list services: %w", err)
				}

				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service1",
						Namespace: "test",
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "service2",
						Namespace: "test",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "empty secret list",
			kubernetesObjects: func() []runtime.Object {
				return slices.FilterOut(newBasicKubernetesObjects(), func(obj runtime.Object) bool {
					_, ok := obj.(*corev1.Secret)
					return ok
				})
			}(),
			scyllaObjects: newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.SecretLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list secrets: %w", err)
				}
				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty secret list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.SecretLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list secrets: %w", err)
				}

				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret1",
						Namespace: "test",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret2",
						Namespace: "test",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "empty config map list",
			kubernetesObjects: func() []runtime.Object {
				return slices.FilterOut(newBasicKubernetesObjects(), func(obj runtime.Object) bool {
					_, ok := obj.(*corev1.ConfigMap)
					return ok
				})
			}(),
			scyllaObjects: newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.ConfigMapLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list config maps: %w", err)
				}
				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty config map list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.ConfigMapLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list config maps: %w", err)
				}

				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap1",
						Namespace: "test",
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap2",
						Namespace: "test",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "empty service account list",
			kubernetesObjects: func() []runtime.Object {
				return slices.FilterOut(newBasicKubernetesObjects(), func(obj runtime.Object) bool {
					_, ok := obj.(*corev1.ServiceAccount)
					return ok
				})
			}(),
			scyllaObjects: newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.ServiceAccountLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list service accounts: %w", err)
				}
				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty service account list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.ServiceAccountLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list service accounts: %w", err)
				}

				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "serviceaccount1",
						Namespace: "test",
					},
				},
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "serviceaccount2",
						Namespace: "test",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:              "empty scylla cluster list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects: func() []runtime.Object {
				return slices.FilterOut(newBasicScyllaObjects(), func(obj runtime.Object) bool {
					_, ok := obj.(*scyllav1.ScyllaCluster)
					return ok
				})
			}(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.ScyllaClusterLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list scylla clusters: %w", err)
				}
				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty scylla cluster list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
			listerFunc: func(ds *DataSource) ([]runtime.Object, error) {
				objects, err := ds.ScyllaClusterLister.List(labels.Everything())
				if err != nil {
					return nil, fmt.Errorf("can't list scylla clusters: %w", err)
				}

				return slices.ConvertSlice(objects, convertToRuntimeObject), nil
			},
			expectedObjects: []runtime.Object{
				&scyllav1.ScyllaCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "scyllacluster1",
						Namespace: "test",
					},
				},
				&scyllav1.ScyllaCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "scyllacluster2",
						Namespace: "test",
					},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range newDataSourceTests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			fakeClient := kubefake.NewSimpleClientset(tc.kubernetesObjects...)
			scyllaFakeClient := fake.NewSimpleClientset(tc.scyllaObjects...)

			dataSource, err := NewDataSourceFromClients(ctx, fakeClient, scyllaFakeClient)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected error: %v, got: %v", tc.expectedErr, err)
			}

			objects, err := tc.listerFunc(dataSource)
			if err != nil {
				t.Fatalf("unexpected error from lister func: %v", err)
			}

			sort.Slice(objects, func(i, j int) bool {
				return compareRuntimeObjects(objects[i], objects[j])
			})

			if !equality.Semantic.DeepEqual(objects, tc.expectedObjects) {
				t.Errorf("expected and actual objects differ: %s", cmp.Diff(tc.expectedObjects, objects))
			}
		})
	}
}

func compareRuntimeObjects(a runtime.Object, b runtime.Object) bool {
	valueA := reflect.ValueOf(a).Elem().FieldByName("ObjectMeta")
	valueB := reflect.ValueOf(b).Elem().FieldByName("ObjectMeta")
	metaA := valueA.Interface().(metav1.ObjectMeta)
	metaB := valueB.Interface().(metav1.ObjectMeta)
	return metaA.Namespace+metaA.Name < metaB.Namespace+metaB.Name
}

func convertToRuntimeObject[T runtime.Object](v T) runtime.Object {
	return runtime.Object(v)
}

func newBasicKubernetesObjects() []runtime.Object {
	return []runtime.Object{
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "test",
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "test",
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service1",
				Namespace: "test",
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service2",
				Namespace: "test",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret1",
				Namespace: "test",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret2",
				Namespace: "test",
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "configmap1",
				Namespace: "test",
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "configmap2",
				Namespace: "test",
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serviceaccount1",
				Namespace: "test",
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serviceaccount2",
				Namespace: "test",
			},
		},
	}
}

func newBasicScyllaObjects() []runtime.Object {
	return []runtime.Object{
		&scyllav1.ScyllaCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "scyllacluster1",
				Namespace: "test",
			},
		},
		&scyllav1.ScyllaCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "scyllacluster2",
				Namespace: "test",
			},
		},
	}
}
