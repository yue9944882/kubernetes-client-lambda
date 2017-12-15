package lambda

import (
	"fmt"
)

type MockResource interface{}

type NamedMockResource map[string]MockResource

type NamespacedMockResource map[string]NamedMockResource

type MockKubernetes struct {
	namespaceAutoCreate bool
	rs                  Resource
	namespace           string
}

var (
	mock = make(map[string]NamespacedMockResource)
)

func (mk *MockKubernetes) InNamespace(namespace string) (l *Lambda) {
	mk.namespace = namespace
	resources := mk.fetch()
	ch := make(chan kubernetesResource)
	l = &Lambda{
		op:  mk,
		val: ch,
	}
	go func() {
		for _, resource := range resources {
			ch <- resource
		}
		close(ch)
	}()
	return
}

func (mk *MockKubernetes) All() (l *Lambda) {
	resources := mk.fetch()
	ch := make(chan kubernetesResource)
	l = &Lambda{
		op:  mk,
		val: ch,
	}
	go func() {
		for _, resource := range resources {
			ch <- resource
		}
		close(ch)
	}()
	return
}

// Putting all resource which doesn't belong to any namespace under "" key
func (mk *MockKubernetes) fetch() NamedMockResource {
	if _, exists := mock[mk.rs.String()]; !exists {
		mock[mk.rs.String()] = make(NamespacedMockResource)
	}
	resourceMock := mock[mk.rs.String()]
	// HACK: "" namespace is for namespace-not-awared resources
	if _, exists := resourceMock[""]; !exists {
		resourceMock[""] = make(NamedMockResource)
	}
	if mk.namespaceAutoCreate {
		if _, exists := resourceMock[mk.namespace]; !exists {
			resourceMock[mk.namespace] = make(NamedMockResource)
		}
	}
	ns := resourceMock[mk.namespace]
	return ns
}

func (mk *MockKubernetes) opCreateInterface(item kubernetesResource) (kubernetesResource, error) {
	if _, exists := mk.fetch()[getNameOfResource(item)]; exists {
		return nil, fmt.Errorf("create failed: resource %s already exists", getNameOfResource(item))
	}
	mk.fetch()[getNameOfResource(item)] = item
	return item, nil
}

func (mk *MockKubernetes) opDeleteInterface(name string) error {
	if _, exists := mk.fetch()[name]; !exists {
		return fmt.Errorf("delete failed: resource %s doesn't exists", name)
	}
	delete(mk.fetch(), name)
	return nil
}

func (mk *MockKubernetes) opUpdateInterface(item kubernetesResource) (kubernetesResource, error) {
	if _, exists := mk.fetch()[getNameOfResource(item)]; !exists {
		return nil, fmt.Errorf("update failed: resource %s doesn't exists", getNameOfResource(item))
	}
	mk.fetch()[getNameOfResource(item)] = item
	return item, nil
}

func (mk *MockKubernetes) opGetInterface(name string) (kubernetesResource, error) {
	if rs, exists := mk.fetch()[name]; !exists {
		return nil, fmt.Errorf("get failed: resource %s doesn't exists", name)
	} else {
		return rs, nil
	}
}

func (mk *MockKubernetes) opListInterface() ([]kubernetesResource, error) {
	var items []kubernetesResource
	for _, v := range mk.fetch() {
		items = append(items, v)
	}
	return items, nil
}
