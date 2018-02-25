package lambda

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

var indexerInitialized bool
var indexerInstance ResourceIndexer

var _ ResourceIndexer = &resourceIndexerImpl{}

type ResourceIndexer interface {
	IsNamespaced(resource Resource) bool
	GetAPIResource(resource Resource) metav1.APIResource
	GetGroupVersionKind(resource Resource) schema.GroupVersionKind
	GetGroupVersionResource(resource Resource) schema.GroupVersionResource
}

type resourceIndexerImpl struct {
	store map[Resource]metav1.APIResource
}

func initIndexer(clientset kubernetes.Interface) {
	apiResources, err := clientset.Discovery().ServerResources()
	if err != nil {
		panic(err)
	}
	indexer := &resourceIndexerImpl{
		store: make(map[Resource]metav1.APIResource),
	}
	for _, apiResource := range apiResources {
		for _, resource := range apiResource.APIResources {
			for _, supportedResource := range GetResources() {
				if supportedResource.GetCanonicalName() == resource.Name {
					indexer.store[supportedResource] = resource
				}
			}
		}
	}
	indexerInstance = indexer
	indexerInitialized = true
}

// getResouceIndexerInstance blocks until initIndexer is invoked
func getResouceIndexerInstance() (indexer ResourceIndexer) {
	// Singleton
	for !indexerInitialized {
	}
	indexer = indexerInstance
	return
}

func (indexer *resourceIndexerImpl) IsNamespaced(resource Resource) bool {
	return indexer.store[resource].Namespaced
}

func (indexer *resourceIndexerImpl) GetGroupVersionResource(resource Resource) schema.GroupVersionResource {
	apiResource := indexer.store[resource]
	return schema.GroupVersionResource{
		Group:    apiResource.Group,
		Version:  apiResource.Version,
		Resource: apiResource.Name,
	}
}

func (indexer *resourceIndexerImpl) GetAPIResource(resource Resource) metav1.APIResource {
	return indexer.store[resource]
}

func (indexer *resourceIndexerImpl) GetGroupVersionKind(resource Resource) schema.GroupVersionKind {
	apiResource := indexer.store[resource]
	return schema.GroupVersionKind{
		Group:   apiResource.Group,
		Version: apiResource.Version,
		Kind:    apiResource.Kind,
	}
}
