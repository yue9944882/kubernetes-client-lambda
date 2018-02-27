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
	indexer := &resourceIndexerImpl{
		store: make(map[Resource]metav1.APIResource),
	}
	grouplist, err := clientset.Discovery().ServerGroups()
	if err != nil {
		panic(err)
	}
	for _, g := range grouplist.Groups {
		for _, v := range g.Versions {
			rs, err := clientset.Discovery().ServerResourcesForGroupVersion(v.GroupVersion)
			if err != nil {
				panic(err)
			}
			for _, r := range rs.APIResources {
				for _, supportedResource := range GetResources() {
					supportedResource := supportedResource
					if supportedResource.GetCanonicalName() == r.Name {
						r.Group = g.Name
						r.Version = v.Version
						indexer.store[supportedResource] = r
					}
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
