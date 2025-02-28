package snapshot

import (
	"context"
	"github.com/google/go-cmp/cmp"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/client/scylla/clientset/versioned/fake"
	"github.com/scylladb/scylla-operator/pkg/helpers/slices"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

func compareRuntimeObjects(a interface{}, b interface{}) bool {
	aObj, aOK := a.(metav1.Object)
	bObj, bOK := b.(metav1.Object)
	if !aOK || !bOK {
		panic("can't cast indexer object to metav1.Object")
	}
	valueA := reflect.ValueOf(aObj).Elem().FieldByName("ObjectMeta")
	valueB := reflect.ValueOf(bObj).Elem().FieldByName("ObjectMeta")
	metaA := valueA.Interface().(metav1.ObjectMeta)
	metaB := valueB.Interface().(metav1.ObjectMeta)
	return metaA.Namespace+metaA.Name < metaB.Namespace+metaB.Name
}

func compareSnapshotObjects(got map[reflect.Type][]interface{}, expected map[reflect.Type][]interface{}) bool {
	for t, gotObjectsList := range got {
		expectedObjectsList := expected[t]
		sort.Slice(gotObjectsList, func(i, j int) bool {
			return compareRuntimeObjects(gotObjectsList[i], gotObjectsList[j])
		})
		sort.Slice(expectedObjectsList, func(i, j int) bool {
			return compareRuntimeObjects(expectedObjectsList[i], expectedObjectsList[j])
		})
	}

	return equality.Semantic.DeepEqual(got, expected)
}

func TestNewDataSourceFromClients_SingleLister(t *testing.T) {
	t.Parallel()
	newDataSourceTests := []struct {
		name              string
		kubernetesObjects []runtime.Object
		scyllaObjects     []runtime.Object
		checkedType       reflect.Type
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
			scyllaObjects:   newBasicScyllaObjects(),
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty pod list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
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
			scyllaObjects:   newBasicScyllaObjects(),
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty service list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
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
			scyllaObjects:   newBasicScyllaObjects(),
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty secret list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
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
			scyllaObjects:   newBasicScyllaObjects(),
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty config map list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
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
			scyllaObjects:   newBasicScyllaObjects(),
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty service account list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
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
			kubernetesObjects: []runtime.Object{},
			scyllaObjects: func() []runtime.Object {
				return slices.FilterOut(newBasicScyllaObjects(), func(obj runtime.Object) bool {
					_, ok := obj.(*scyllav1.ScyllaCluster)
					return ok
				})
			}(),
			expectedObjects: []runtime.Object{},
			expectedErr:     nil,
		},
		{
			name:              "nonempty scylla cluster list",
			kubernetesObjects: newBasicKubernetesObjects(),
			scyllaObjects:     newBasicScyllaObjects(),
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

			snapshot, err := NewSnapshotFromListers(ctx, fakeClient, scyllaFakeClient)

			gotObjects := snapshot.Objects[tc.checkedType]

			sort.Slice(gotObjects, func(i, j int) bool {
				return compareRuntimeObjects(gotObjects[i], gotObjects[j])
			})
			sort.Slice(tc.expectedObjects, func(i, j int) bool {
				return compareRuntimeObjects(tc.expectedObjects[i], tc.expectedObjects[j])
			})
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("expected error: %v, got: %v", tc.expectedErr, err)
			}

			if !equality.Semantic.DeepEqual(snapshot.Objects, tc.expectedObjects) {
				t.Errorf("expected and actual objects differ: %s", cmp.Diff(tc.expectedObjects, snapshot.Objects))
			}
		})
	}
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

func TestArchiveReader(t *testing.T) {
	t.Parallel()

	testScheme := runtime.NewScheme()

	testSchemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		scyllav1.Install,
	}

	err := testSchemeBuilder.AddToScheme(testScheme)
	if err != nil {
		t.Fatal(err)
	}

	testDecoder := serializer.NewCodecFactory(testScheme).UniversalDeserializer()

	tt := []struct {
		name            string
		archive         fstest.MapFS
		expectedObjects map[reflect.Type][]interface{}
		expectedError   error
	}{
		{
			name: "deserializes .yaml files",
			archive: fstest.MapFS{
				"file1.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod1
`))},
				"file2.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod2
`))},
			},
			expectedObjects: map[reflect.Type][]interface{}{
				reflect.TypeOf(&corev1.Pod{}): {
					&corev1.Pod{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Pod",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod1",
							Namespace: "scylla",
						},
					},
					&corev1.Pod{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Pod",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod2",
							Namespace: "scylla",
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "ignores extensions other than .yaml and files with no extension",
			archive: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod1
				`))},
				"file.json": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod2
`))},
				"fileyyaml.notyaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod3
`))},
				"file": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod4
`))},
				"file.yamll": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod5
`))},
			},
			expectedObjects: map[reflect.Type][]interface{}{},
			expectedError:   nil,
		},
		{
			name: "deserializes nested files",
			archive: fstest.MapFS{
				"dir/file.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod1
`))},
				"dir/dir2/file.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  namespace: scylla
  name: pod2
				`))},
			},
			expectedObjects: map[reflect.Type][]interface{}{
				reflect.TypeOf(&corev1.Pod{}): {
					&corev1.Pod{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Pod",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod1",
							Namespace: "scylla",
						},
					},
					&corev1.Pod{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Pod",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pod2",
							Namespace: "scylla",
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "doesn't try to read files which are directories with names ending with .yaml",
			archive: fstest.MapFS{
				"dir1.yaml/dir2.yaml/file.txt": &fstest.MapFile{Data: []byte("test")},
			},
			expectedObjects: map[reflect.Type][]interface{}{},
			expectedError:   nil,
		},
		{
			name: "ignores unknown resource kinds",
			archive: fstest.MapFS{
				"ciliumFile.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: cilium.io/v2
kind: CiliumEndpoint
metadata:
  name: cilium
`))},
				"otherFile.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: testApi
kind: testEndpoint
metadata:
  name: test
`))},
			},
			expectedObjects: map[reflect.Type][]interface{}{},
			expectedError:   nil,
		},
		{
			name:            "empty fs is valid",
			archive:         fstest.MapFS{},
			expectedObjects: map[reflect.Type][]interface{}{},
			expectedError:   nil,
		},
		{
			name: "multiple objects type",
			archive: fstest.MapFS{
				"pod.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Pod
metadata:
  name: testPod
  namespace: test
`))},
				"service.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Service
metadata:
  name: testService
  namespace: test
`))},
				"secret.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: Secret
metadata:
  name: testSecret
  namespace: test
`))},
				"configMap.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: testConfigMap
  namespace: test
`))},
				"serviceAccount.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: testServiceAccount
  namespace: test
`))},
				"scyllaCluster1.yaml": &fstest.MapFile{Data: []byte(strings.TrimSpace(`
apiVersion: scylla.scylladb.com/v1
kind: ScyllaCluster
metadata:
  name: testScyllaCluster
  namespace: test
`))},
			},
			expectedObjects: map[reflect.Type][]interface{}{
				reflect.TypeOf(&corev1.Pod{}): {
					&corev1.Pod{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Pod",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testPod",
							Namespace: "test",
						},
					},
				},
				reflect.TypeOf(&corev1.Service{}): {
					&corev1.Service{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Service",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testService",
							Namespace: "test",
						},
					},
				},
				reflect.TypeOf(&corev1.Secret{}): {
					&corev1.Secret{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testSecret",
							Namespace: "test",
						},
					},
				},
				reflect.TypeOf(&corev1.ConfigMap{}): {
					&corev1.ConfigMap{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ConfigMap",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testConfigMap",
							Namespace: "test",
						},
					},
				},
				reflect.TypeOf(&corev1.ServiceAccount{}): {
					&corev1.ServiceAccount{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ServiceAccount",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testServiceAccount",
							Namespace: "test",
						},
					},
				},
				reflect.TypeOf(&scyllav1.ScyllaCluster{}): {
					&scyllav1.ScyllaCluster{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ScyllaCluster",
							APIVersion: "scylla.scylladb.com/v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testScyllaCluster",
							Namespace: "test",
						},
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			snapshot, err := NewSnapshotFromFS(tc.archive, testDecoder)

			if !reflect.DeepEqual(err, tc.expectedError) {
				t.Fatalf("got error %v expected error %v", err, tc.expectedError)
			}

			if err == nil {

				if !compareSnapshotObjects(snapshot.Objects, tc.expectedObjects) {
					t.Errorf("expected and actual objects differ: %s", cmp.Diff(tc.expectedObjects, snapshot.Objects))
				}
			}
		})
	}
}
