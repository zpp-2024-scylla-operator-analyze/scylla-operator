package analyze

import (
	"context"
	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	"github.com/scylladb/scylla-operator/pkg/client/scylla/clientset/versioned/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"testing"
)

type dummyLister struct {
	indexer cache.Indexer
}

type dummyObject struct {
	metav1.ObjectMeta
	id int64
}

func (d dummyObject) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (d dummyObject) DeepCopyObject() runtime.Object {
	return d
}

func (d *dummyLister) List() ([]*dummyObject, error) {
	do := make([]*dummyObject, 0)
	for _, obj := range d.indexer.List() {
		do = append(do, obj.(*dummyObject))
	}
	return do, nil
}

type buildListerTest[T runtime.Object, U any] struct {
	ListerFactory func(i cache.Indexer) U
	Objects       []*T
}

func assertExpectedContainsPermutatedActual[T, U any](t *testing.T, expectedResources []T, actualResources []U) {
	if len(expectedResources) != len(actualResources) {
		t.Errorf("expected %d resources, got %d", len(expectedResources), len(actualResources))
		t.Logf("expected %v got %v", expectedResources, actualResources)
	}

	for _, expected := range expectedResources {
		found := false
		for _, actual := range actualResources {
			if reflect.DeepEqual(actual, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("object is missing: %v", expected)
		}
	}
}

func TestBuildLister(t *testing.T) {
	buildListerDummyTests := map[string]buildListerTest[dummyObject, dummyLister]{
		"empty dummy lister": {
			ListerFactory: func(i cache.Indexer) dummyLister {
				if i == nil {
					t.Fatal("expected lister.indexer to be non-nil")
				}
				return dummyLister{indexer: i}
			},
			Objects: []*dummyObject{},
		},
		"small dummy lister": {
			ListerFactory: func(i cache.Indexer) dummyLister {
				if i == nil {
					t.Fatal("expected lister.indexer to be non-nil")
				}
				return dummyLister{indexer: i}
			},
			Objects: []*dummyObject{
				{
					id: 3,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Obj3",
						Namespace: "NS3",
					},
				},
				{
					id: 1,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Obj1",
						Namespace: "NS1",
					},
				},
				{
					id: 2,
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Obj2",
						Namespace: "NS1",
					},
				},
			},
		},
	}

	for name, test := range buildListerDummyTests {
		t.Run(name, func(t *testing.T) {
			var objectsAsRawExtension []runtime.RawExtension
			for _, obj := range test.Objects {
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

			lister, err := BuildLister[dummyLister](ctx, test.ListerFactory, listFunc)

			if err != nil {
				t.Fatalf("BuildLister returned an error: %v", err)
			}
			if reflect.TypeOf(lister) != reflect.TypeFor[dummyLister]() {
				t.Fatalf("expected '%v' got: '%v'", reflect.TypeFor[dummyLister](), reflect.TypeOf(lister))
			}

			actualObjects, err := lister.List()
			if err != nil {
				t.Fatalf("dummyLister.List returned an error: %v", err)
			}

			assertExpectedContainsPermutatedActual(t, test.Objects, actualObjects)
		})
	}
}

func compareSingle[T any, U *T](
	t *testing.T,
	all []runtime.Object,
	listFunc func(selector labels.Selector) (ret []*T, err error),
) {
	actualResources, err := listFunc(labels.Everything())
	if err != nil {
		t.Fatalf("%s lister returned an error: %v", reflect.TypeFor[U]().Name(), err)
	}

	var expectedResources []U
	for _, res := range all {
		switch v := res.(type) {
		case U:
			expectedResources = append(expectedResources, v)
		}
	}

	assertExpectedContainsPermutatedActual(t, expectedResources, actualResources)
}

func TestNewDataSourceFromClients(t *testing.T) {
	newDataSourceTests := map[string]struct {
		kubernetesObjects []runtime.Object
		scyllaObjects     []runtime.Object
	}{
		"no objects": {},
		"some listers empty": {
			kubernetesObjects: []runtime.Object{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod1",
						Namespace: "NS1",
					},
				},
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod2",
						Namespace: "NS1",
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Svc1",
						Namespace: "NS1",
					},
				},
			},
			scyllaObjects: []runtime.Object{
				&scyllav1.ScyllaCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ScyllaCluster1",
						Namespace: "NS1",
					},
				},
			},
		},
		"all objects": {
			kubernetesObjects: []runtime.Object{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod1",
						Namespace: "NS1",
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Svc1",
						Namespace: "NS1",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Secret1",
						Namespace: "NS2",
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ConfigMap1",
						Namespace: "NS1",
					},
				},
			},
			scyllaObjects: []runtime.Object{
				&scyllav1.ScyllaCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ScyllaCluster1",
						Namespace: "NS1",
					},
				},
			},
		},
	}

	for name, test := range newDataSourceTests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			fakeClient := kubefake.NewSimpleClientset(test.kubernetesObjects...)
			scyllaFakeClient := fake.NewSimpleClientset(test.scyllaObjects...)
			defer cancel()

			dataSource, err := NewDataSourceFromClients(ctx, fakeClient, scyllaFakeClient)
			if err != nil {
				t.Fatalf("NewDataSourceFromClients returned an error: %v", err)
			}

			compareSingle(t, test.kubernetesObjects, dataSource.PodLister.List)
			compareSingle(t, test.kubernetesObjects, dataSource.ServiceLister.List)
			compareSingle(t, test.kubernetesObjects, dataSource.SecretLister.List)
			compareSingle(t, test.kubernetesObjects, dataSource.ConfigMapLister.List)
			compareSingle(t, test.kubernetesObjects, dataSource.ServiceAccountLister.List)

			compareSingle(t, test.scyllaObjects, dataSource.ScyllaClusterLister.List)
		})
	}
}
