package analyze

import (
	"reflect"

	scylla_scheme "github.com/scylladb/scylla-operator/pkg/scheme"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"
)

var (
	scheme       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(scheme)
	deserializer runtime.Decoder
	objects      []runtime.Object

	indexers = make(map[reflect.Type]cache.Indexer)
)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(scylla_scheme.AddToScheme(scheme))
}

func getIndexerForType(objType reflect.Type) cache.Indexer {
	if indexer, exists := indexers[objType]; exists {
		return indexer
	}
	// Create a new indexer for this type if it doesn't exist
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	indexers[objType] = indexer
	return indexer
}

func deserialize_file(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".yaml") {
		content, err := os.ReadFile(path)
		if err != nil {
			klog.Errorf("ReadFile error: %v", err)
			return nil
		}
		obj, _, err := deserializer.Decode(content, nil, nil)
		if err != nil {
			return nil
		}

		objType := reflect.TypeOf(obj)
		indexer := getIndexerForType(objType)
		indexer.Add(obj)

	}
	return nil
}

func IndexersFromArchive(archivePath string) map[reflect.Type]cache.Indexer {

	klog.Infof("Archive reader %s", archivePath)

	deserializer = codecs.UniversalDeserializer()

	err := filepath.Walk(archivePath, deserialize_file)
	if err != nil {
		klog.Errorf("Walk error: %v", err)
	}

	return indexers
}
