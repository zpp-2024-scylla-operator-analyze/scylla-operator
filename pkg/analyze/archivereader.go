package analyze

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"

	"github.com/scylladb/scylla-operator/pkg/scheme"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
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

func isYamlFile(fsys fs.FS, path string, d fs.DirEntry) bool {
	if d.IsDir() {
		return false
	}
	extension := filepath.Ext(path)
	if (extension == ".yaml") || (extension == ".yml") {
		return true // Assume it is a valid yaml file
	}
	if extension == "" {
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return false
		}
		var unmarshalledContent interface{}
		err = yaml.Unmarshal(content, &unmarshalledContent)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func IndexersFromFS(fsys fs.FS) (map[reflect.Type]cache.Indexer, error) {

	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()

	indexers := make(map[reflect.Type]cache.Indexer)

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if isYamlFile(fsys, path, d) {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return fmt.Errorf("ReadFile error: %v", err)
			}
			obj, _, err := deserializer.Decode(content, nil, nil)
			if (err != nil) && (!runtime.IsNotRegisteredError(err)) {
				return fmt.Errorf("deserialize error: %v", err)
			}
			if err == nil {
				objType := reflect.TypeOf(obj)
				indexer := getIndexerForType(indexers, objType)
				err = indexer.Add(obj)
				if err != nil {
					return fmt.Errorf("indexer add error: %v", err)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("can't walk the file tree: %w", err)
	}
	return indexers, nil
}
