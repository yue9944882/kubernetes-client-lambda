package lambda

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

var indexerInitialized bool
var indexerInstance ResourceIndexer

var _ ResourceIndexer = &resourceIndexerImpl{}

type ResourceIndexer interface {
	IsNamespaced(resource Resource) bool
	GetAPIResource(resource Resource) *metav1.APIResource
	GetGroupVersionKind(resource Resource) schema.GroupVersionKind
	GetGroupVersionResource(resource Resource) schema.GroupVersionResource
}

type resourceIndexerImpl struct {
	store map[Resource]*metav1.APIResource
}

func init() {
	initIndexer()
}

type SortedGVKs []schema.GroupVersionKind

func (s SortedGVKs) Len() int { return len(s) }

func (s SortedGVKs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s SortedGVKs) Less(i, j int) bool {
	vi := s[i].Version
	vj := s[j].Version

	iLen := len(vi)
	jLen := len(vj)

	minLen := iLen

	if minLen > jLen {
		minLen = jLen
	}

	for idx := 0; idx < minLen; idx++ {
		if vi[idx] < vj[idx] {
			return true
		} else if vi[idx] > vj[idx] {
			return false
		}
	}

	return iLen < jLen
}

func initIndexer() {
	i := fake.NewSimpleClientset()
	indexer := &resourceIndexerImpl{
		store: make(map[Resource]*metav1.APIResource),
	}
	knownGvks := []schema.GroupVersionKind{}
	for gvk := range scheme.Scheme.AllKnownTypes() {
		knownGvks = append(knownGvks, gvk)
	}
	gvks := SortedGVKs(knownGvks)
	sort.Sort(gvks)
	indexedMap := map[Resource]bool{}
	for _, gvk := range knownGvks {
		gvk := gvk
		for _, supportedResource := range GetResources() {
			pluralGvr, singularGvr := meta.UnsafeGuessKindToResource(gvk)
			if pluralGvr.Resource == strings.ToLower(string(supportedResource)) {
				if !indexedMap[supportedResource] {
					gvkGroup := strings.SplitN(gvk.Group, ".", 2)[0]
					if gvkGroup == "" {
						gvkGroup = "core"
					}
					gvMethod := reflect.ValueOf(i).MethodByName(
						capitalizeFirstLetter(gvkGroup) + capitalizeFirstLetter(gvk.Version),
					).Call([]reflect.Value{})[0]
					namespaced := gvMethod.MethodByName(string(supportedResource)).Type().NumIn() != 0
					apiRs := metav1.APIResource{
						Name:         pluralGvr.Resource,
						SingularName: singularGvr.Resource,
						Namespaced:   namespaced,
						Group:        gvk.Group,
						Version:      gvk.Version,
						Kind:         gvk.Kind,
					}
					indexer.store[supportedResource] = &apiRs
					indexedMap[supportedResource] = true
				}
			}
		}
	}
	indexerInstance = indexer
	indexerInitialized = true
}

func capitalizeFirstLetter(s string) string {
	if len(s) > 1 {
		return strings.ToUpper(string(s[0])) + string(s[1:])
	}
	return strings.ToUpper(s)
}

// getResouceIndexerInstance blocks until initIndexer is invoked
func GetResouceIndexerInstance() (indexer ResourceIndexer) {
	// Singleton
	return indexerInstance
}

func (indexer *resourceIndexerImpl) IsNamespaced(resource Resource) bool {
	return indexer.store[resource].Namespaced
}

func (indexer *resourceIndexerImpl) GetGroupVersionResource(resource Resource) schema.GroupVersionResource {
	apiResource := indexer.store[resource]
	if apiResource == nil {
		panic(fmt.Sprintf("unindexed resource %s", resource))
	}
	return schema.GroupVersionResource{
		Group:    apiResource.Group,
		Version:  apiResource.Version,
		Resource: apiResource.Name,
	}
}

func (indexer *resourceIndexerImpl) GetAPIResource(resource Resource) *metav1.APIResource {
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
