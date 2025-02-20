package sources

import (
	"github.com/google/go-cmp/cmp"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/tools/cache"
)

func sortByName(objs []interface{}) {
	sort.Slice(objs, func(i, j int) bool {
		iObj, iOK := objs[i].(metav1.Object)
		jObj, jOK := objs[j].(metav1.Object)
		if !iOK || !jOK {
			panic("can't cast indexer object to metav1.Object")
		}
		return iObj.GetName() < jObj.GetName()
	})
}

// Checks if all object listed by .List() can be also retrieved using .GetByKey()
func checkIndexer(t *testing.T, indexer cache.Indexer, keyFunc func(obj interface{}) (string, error)) bool {
	for _, obj := range indexer.List() {
		key, err := keyFunc(obj)
		if err != nil {
			t.Errorf("can't apply keyFunc to object in list: %v", err)
			return false
		}
		gbkObj, exists, err := indexer.GetByKey(key)
		if err != nil {
			t.Errorf("can't retrieve item using GetByKey: %v", err)
			return false
		}
		if !exists {
			t.Errorf("no object with key exists in indexer %s", key)
			return false
		}
		if !reflect.DeepEqual(obj, gbkObj) {
			t.Errorf("expected and got indexer.GetByKey objects differ %s", cmp.Diff(obj, gbkObj))
			return false
		}
	}
	return true
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
		name                   string
		archive                fstest.MapFS
		expectedIndexerObjects map[reflect.Type][]interface{}
		expectedError          error
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
			expectedIndexerObjects: map[reflect.Type][]interface{}{
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
			expectedIndexerObjects: map[reflect.Type][]interface{}{},
			expectedError:          nil,
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
			expectedIndexerObjects: map[reflect.Type][]interface{}{
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
			expectedIndexerObjects: map[reflect.Type][]interface{}{},
			expectedError:          nil,
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
			expectedIndexerObjects: map[reflect.Type][]interface{}{},
			expectedError:          nil,
		},
		{
			name:                   "empty fs is valid",
			archive:                fstest.MapFS{},
			expectedIndexerObjects: map[reflect.Type][]interface{}{},
			expectedError:          nil,
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
			expectedIndexerObjects: map[reflect.Type][]interface{}{
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
			indexers, err := IndexersFromFS(tc.archive, testDecoder)

			if !reflect.DeepEqual(err, tc.expectedError) {
				t.Fatalf("got error %v expected error %v", err, tc.expectedError)
			}

			if err == nil {
				gotIndexerObjects := map[reflect.Type][]interface{}{}

				for indexerType, indexer := range indexers {
					checkIndexer(t, indexer, cache.MetaNamespaceKeyFunc)

					gotIndexerObjects[indexerType] = indexer.List()
					sortByName(gotIndexerObjects[indexerType])
					sortByName(tc.expectedIndexerObjects[indexerType])
				}

				if !reflect.DeepEqual(gotIndexerObjects, tc.expectedIndexerObjects) {
					t.Fatalf("expected and got indexer objects differ: %s", cmp.Diff(tc.expectedIndexerObjects, gotIndexerObjects))
				}
			}
		})
	}
}
