package snapshot

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
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

func NewSnapshotFromFS(fsys fs.FS, decoder runtime.Decoder) (Snapshot, error) {
	ds := NewDefaultSnapshot()

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
		return nil, fmt.Errorf("can't walk the file tree: %w", err)
	}
	return &ds, nil
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
