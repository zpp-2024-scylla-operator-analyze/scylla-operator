package analyze

import (
	"reflect"

	"github.com/scylladb/scylla-operator/pkg/scheme"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	apierrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"path/filepath"
)

func init() {
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme.Scheme))
}

func getIndexerForType(indexers map[reflect.Type]cache.Indexer, objType reflect.Type) cache.Indexer {
	if indexer, exists := indexers[objType]; exists {
		return indexer
	}
	// Create a new indexer for this type if it doesn't exist
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

func IndexersFromArchive(fsys fs.FS) (map[reflect.Type]cache.Indexer, error) {
	var errs []error

	deserializer := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()

	indexers := make(map[reflect.Type]cache.Indexer)

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if isYamlFile(fsys, path, d) {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				klog.Errorf("ReadFile error: %v", err)
				return fs.SkipAll
			}
			obj, _, err := deserializer.Decode(content, nil, nil)
			if (err != nil) && (!runtime.IsNotRegisteredError(err)) {
				klog.Errorf("Deserialize error: %v", err)
				klog.Errorf("Content: %s", content)
				return fs.SkipAll
			}
			objType := reflect.TypeOf(obj)
			indexer := getIndexerForType(indexers, objType)
			indexer.Add(obj)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Walk error:", err)
		errs = append(errs, fmt.Errorf("Cant read/deserialize all files in directory"))
	}
	if len(indexers) == 0 {
		errs = append(errs, fmt.Errorf("No objects to deserialize"))
	}
	return indexers, apierrors.NewAggregate(errs)
}
