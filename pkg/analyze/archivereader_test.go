package analyze

import (
	"fmt"
	"io/fs"
	"k8s.io/klog/v2"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"encoding/json"
)

func brokenOpenFunc(name string) (fs.File, error) {
	return nil, fmt.Errorf("Broken file")
}

type wrapperFS struct {
	base     fstest.MapFS
	openFunc func(name string) (fs.File, error)
}

func (w wrapperFS) Glob(pattern string) ([]string, error) {
	return w.base.Glob(pattern)
}

func (w wrapperFS) Open(name string) (fs.File, error) {
	if w.openFunc != nil {
		return w.openFunc(name)
	}
	return w.base.Open(name)
}

func (w wrapperFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return w.base.ReadDir(name)
}

func (w wrapperFS) ReadFile(name string) ([]byte, error) {
	return w.base.ReadFile(name)
}

func (w wrapperFS) Stat(name string) (fs.FileInfo, error) {
	return w.base.Stat(name)
}

func (w wrapperFS) Sub(name string) (fs.FS, error) {
	return w.base.Sub(name)
}

func defaultPodYaml(name string) []byte {
	return []byte(strings.TrimPrefix(`
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: scylla
  namespace: scylla
  name: `, "\n") + name + "\n")
}

func errorsEqual(err1 error, err2 error) bool {
	if (err1 == nil) && (err2 == nil) {
		return true
	}
	if (err1 == nil) || (err2 == nil) {
		return false
	}
	return (err1.Error() == err2.Error())
}

func TestArchiveReaderErrors(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name                      string
		testType                  string
		archive                   wrapperFS
		expectedError             error
		ExpectedIndexerObjectsNum int
		expectedIndexerObj        cache.Indexer
	}{
		{
			name:     " deserializes .yaml and .yml files",
			testType: "error check",
			archive: wrapperFS{
				base: fstest.MapFS{
					"file1.yaml": &fstest.MapFile{Data: defaultPodYaml("pod1")},
					"file2.yml":  &fstest.MapFile{Data: defaultPodYaml("pod2")},
				},
			},
			expectedError:             nil,
			ExpectedIndexerObjectsNum: 2,
		}, {
			testType: "error check",

			name: "deserializes yaml files with no extension",
			archive: wrapperFS{
				base: fstest.MapFS{
					"file": &fstest.MapFile{Data: defaultPodYaml("pod1")},
				},
			},
			expectedError:             nil,
			ExpectedIndexerObjectsNum: 1,
		}, {
			testType: "error check",

			name: "ignores extensions other than .yaml and .yml",
			archive: wrapperFS{
				base: fstest.MapFS{
					"file.txt":           &fstest.MapFile{Data: defaultPodYaml("pod1")},
					"file.json":          &fstest.MapFile{Data: defaultPodYaml("pod2")},
					"surelyyaml.notyaml": &fstest.MapFile{Data: defaultPodYaml("pod3")},
					"file.yaml":          &fstest.MapFile{Data: defaultPodYaml("pod4")},
					"file.yamll":         &fstest.MapFile{Data: defaultPodYaml("pod5")},
				},
			},
			expectedError:             nil,
			ExpectedIndexerObjectsNum: 1,
		}, {
			testType: "error check",

			name: "ignores non-yaml files with no extension",
			archive: wrapperFS{
				base: fstest.MapFS{
					"file": &fstest.MapFile{Data: []byte(strings.TrimPrefix(`
{
	"apiVersion": "v1",
	"kind": "ServiceAccount",
	"metadata": {
		"creationTimestamp": "2024-11-02T18:24:35Z",
		"name": "default",
		"namespace": "scylla",
		"resourceVersion": "10469",
		"uid": "6b5c377e-0fa0-42f4-afa7-54a68039ea38"
	}
}
	`, "\n"),
					)},
					"file.yaml": &fstest.MapFile{Data: defaultPodYaml("pod1")},
				},
			},
			expectedError:             nil,
			ExpectedIndexerObjectsNum: 1,
		}, {
			testType: "error check",

			name: " deserializes nested files",
			archive: wrapperFS{
				base: fstest.MapFS{
					"dir/file.yaml":            &fstest.MapFile{Data: defaultPodYaml("pod1")},
					"dir/dir2/file.yaml":       &fstest.MapFile{Data: defaultPodYaml("pod2")},
					"otherdir/a/b/c/file.yaml": &fstest.MapFile{Data: defaultPodYaml("pod3")},
				},
			},
			expectedError:             nil,
			ExpectedIndexerObjectsNum: 3,
		}, {
			testType: "error check",

			name: " doesnt try to read files which are directories with names ending with .yaml and .yml",
			archive: wrapperFS{
				base: fstest.MapFS{
					"dir1.yaml/dir2.yml/file.yaml": &fstest.MapFile{Data: defaultPodYaml("pod1")},
				},
			},
			expectedError:             nil,
			ExpectedIndexerObjectsNum: 1,
		}, {
			testType: "error check",
			name:     " returns error when some error",
			archive: wrapperFS{
				base:     nil,
				openFunc: brokenOpenFunc,
			},

			expectedError:             fmt.Errorf("No objects to deserialize"),
			ExpectedIndexerObjectsNum: 0,
		}, {
			testType: "error check",

			name: " returns error when path is empty",
			archive: wrapperFS{
				base: fstest.MapFS{},
			},
			expectedError: fmt.Errorf("No objects to deserialize"),
		}, {
			testType: "error check",

			name: " ignores unknown resource kinds",
			archive: wrapperFS{
				base: fstest.MapFS{
					"ciliumFile.yaml": &fstest.MapFile{Data: []byte(strings.TrimPrefix(`
apiVersion: cilium.io/v2
kind: CiliumEndpoint
metadata:
  name: cilium
`, "\n"))},
					"otherFile.yml": &fstest.MapFile{Data: []byte(strings.TrimPrefix(`
apiVersion: testApi
kind: testEndpoint
metadata:
  name: test
`, "\n"))},
					"knownFile.yaml":  &fstest.MapFile{Data: defaultPodYaml("pod1")},
					"knownFile2.yaml": &fstest.MapFile{Data: defaultPodYaml("pod2")},
				},
			},
			expectedError:             nil,
			ExpectedIndexerObjectsNum: 2,
		}, {
			testType: "error check",

			name: " returns error when directory is empty/ no yaml files/ yaml files are empty",
			archive: wrapperFS{
				base: fstest.MapFS{
					"file.txt": &fstest.MapFile{Data: []byte("test")},
				},
			},
			expectedError: fmt.Errorf("No objects to deserialize"),
		}, {
			name:     "webhook server",
			testType: "data check",
			archive: wrapperFS{
				base: fstest.MapFS{
					"webhook-server.yaml": &fstest.MapFile{Data: []byte(strings.TrimPrefix(`
apiVersion: v1
kind: Pod
metadata:
  name: webhook-server-7cbd674df9-782zm
  namespace: scylla-operator
  labels:
    app.kubernetes.io/name: webhook-server
spec:
  containers:
  - name: webhook-server
    image: docker.io/scylladb/scylla-operator:latest
    ports:
    - containerPort: 5000
      protocol: TCP
  volumes:
  - name: cert
    secret:
      secretName: scylla-operator-serving-cert
status:
  conditions:
  - type: Ready
    status: "True"
  containerStatuses:
  - name: webhook-server
    ready: true
    state:
      running: {}
`, "\n"))},
				},
			},
			expectedError: nil,
			expectedIndexerObj: func() cache.Indexer {
				indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})

				pod := &corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "webhook-server-7cbd674df9-782zm",
						Namespace: "scylla-operator",
						Labels: map[string]string{
							"app.kubernetes.io/name": "webhook-server",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "webhook-server",
								Image: "docker.io/scylladb/scylla-operator:latest",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 5000,
										Protocol:      corev1.ProtocolTCP,
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "cert",
								VolumeSource: corev1.VolumeSource{
									Secret: &corev1.SecretVolumeSource{
										SecretName: "scylla-operator-serving-cert",
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodConditionType("Ready"),
								Status: corev1.ConditionTrue,
							},
						},
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name:  "webhook-server",
								Ready: true,
								State: corev1.ContainerState{
									Running: &corev1.ContainerStateRunning{},
								},
							},
						},
					},
				}

				if err := indexer.Add(pod); err != nil {
					panic(err)
				}

				return indexer
			}(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			indexers, err := IndexersFromArchive(tc.archive)

			if tc.testType == "error check" {
				if !errorsEqual(err, tc.expectedError) {
					klog.Errorf("got error %v expected error %v", err, tc.expectedError)
					t.Fatal(err)
				}
				count := 0
				for _, indexer := range indexers {
					count += len(indexer.List())
				}
				if count != tc.ExpectedIndexerObjectsNum {
					t.Errorf("expected and got indexer items differ (got %d expected %d)", count, tc.ExpectedIndexerObjectsNum)
				}
			}
			if tc.testType == "data check" {
				if err != nil {
					t.Fatal(err)
				}

				//fmt.Println(reflect.TypeOf(indexers[reflect.TypeOf(&corev1.Pod{})]))
				actualJson, err := json.MarshalIndent(indexers[reflect.TypeOf(&corev1.Pod{})].List(), "", " ")
				if err != nil {
					t.Errorf("marshall1 %s", err)
					panic("marshall failed")
				}
				expectedJson, err := json.MarshalIndent(tc.expectedIndexerObj.List(), "", " ")
				if err != nil {
					t.Errorf("marshall2 %s", err)

					panic("marshall failed")
				}

				if string(actualJson) != string(expectedJson) {
					t.Errorf("indexer differs\n")
				}
			}
		})

	}

}
