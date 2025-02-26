package sources

import (
	"context"
	"fmt"
	"io/fs"
	storagev1 "k8s.io/api/storage/v1"
	storagev1listers "k8s.io/client-go/listers/storage/v1"
	"os"
	"path/filepath"
	"reflect"

	scyllav1 "github.com/scylladb/scylla-operator/pkg/api/scylla/v1"
	scyllav1listers "github.com/scylladb/scylla-operator/pkg/client/scylla/listers/scylla/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func getIndexerForType(indexers map[reflect.Type]cache.Indexer, objType reflect.Type) cache.Indexer {
	if indexer, exists := indexers[objType]; exists {
		return indexer
	}
	// Create new indexer for this type if it doesn't exist
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	indexers[objType] = indexer
	return indexer
}


func NewDataSource2FromFS(fsys fs.FS, decoder runtime.Decoder) (DataSource2, error) {
	ds := NewDataSource2()

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("can't read file %q: %w", path, err)
		}
		obj, _, err := decoder.Decode(content, nil, nil)

		if err != nil {
			if !runtime.IsNotRegisteredError(err) {
				return fmt.Errorf("can't deserialize file %q: %w", path, err)
			}
			return nil
		}
		ds.Add(obj)
		return nil
	})

	if err != nil {
		return ds, fmt.Errorf("can't walk the file tree: %w", err)
	}
	return ds, nil
}

func IndexersFromFS(fsys fs.FS, decoder runtime.Decoder) (map[reflect.Type]cache.Indexer, error) {
	indexers := make(map[reflect.Type]cache.Indexer)

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("can't read file %q: %w", path, err)
		}
		obj, _, err := decoder.Decode(content, nil, nil)

		if err != nil {
			if !runtime.IsNotRegisteredError(err) {
				return fmt.Errorf("can't deserialize file %q: %w", path, err)
			}
			return nil
		}
		objType := reflect.TypeOf(obj)
		indexer := getIndexerForType(indexers, objType)
		err = indexer.Add(obj)
		if err != nil {
			return fmt.Errorf("can't add object to indexer: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("can't walk the file tree: %w", err)
	}
	return indexers, nil
}

func NewDataSourceFromFS(ctx context.Context, archivePath string, decoder runtime.Decoder) (*DataSource, error) {
	if decoder == nil {
		return nil, fmt.Errorf("decoder must be specified")
	}

	indexers, err := IndexersFromFS(os.DirFS(archivePath), decoder)
	if err != nil {
		return nil, fmt.Errorf("can't build indexers from fs: %w", err)
	}

	types := []interface{}{
		&corev1.Pod{},
		&corev1.Service{},
		&corev1.Secret{},
		&corev1.ConfigMap{},
		&corev1.ServiceAccount{},
		&scyllav1.ScyllaCluster{},
	}

	// Create an indexer for each type if it wasn't created already
	for _, obj := range types {
		t := reflect.TypeOf(obj)
		getIndexerForType(indexers, t)
	}

	return &DataSource{
		PodLister:            corev1listers.NewPodLister(indexers[reflect.TypeOf(&corev1.Pod{})]),
		ServiceLister:        corev1listers.NewServiceLister(indexers[reflect.TypeOf(&corev1.Service{})]),
		SecretLister:         corev1listers.NewSecretLister(indexers[reflect.TypeOf(&corev1.Secret{})]),
		ConfigMapLister:      corev1listers.NewConfigMapLister(indexers[reflect.TypeOf(&corev1.ConfigMap{})]),
		ServiceAccountLister: corev1listers.NewServiceAccountLister(indexers[reflect.TypeOf(&corev1.ServiceAccount{})]),
		ScyllaClusterLister:  scyllav1listers.NewScyllaClusterLister(indexers[reflect.TypeOf(&scyllav1.ScyllaCluster{})]),
		StorageClassLister:   storagev1listers.NewStorageClassLister(getIndexerForType(indexers, reflect.TypeOf(&storagev1.StorageClass{}))),
		CSIDriverLister:      storagev1listers.NewCSIDriverLister(getIndexerForType(indexers, reflect.TypeOf(&storagev1.CSIDriver{}))),
	}, nil
}
