package lambda

import (
	"bytes"
	"reflect"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
	"k8s.io/client-go/util/flowcontrol"
)

// the mock KubernetesClient is statusful and if you want to reset its status then use MockReset
func Mock(objects ...runtime.Object) KubernetesClientLambda {
	fakePool, _ := NewFakes(objects...)
	return &kubernetesClientLambdaImpl{
		clientPool: fakePool,
	}
}

func NewFakes(objects ...runtime.Object) (dynamic.ClientPool, kubernetes.Interface) {
	fakeClientset := fake.NewSimpleClientset(objects...)
	fakeClientset.Fake.ReactionChain = []testing.Reactor{
		kclReactorWrapper(fakeClientset.ReactionChain[0]),
	}
	return &FakeClientPool{&(fakeClientset.Fake)}, fakeClientset
}

func kclReactorWrapper(reactor testing.Reactor) testing.Reactor {
	return &testing.SimpleReactor{
		Verb:     "*",
		Resource: "*",
		Reaction: func(action testing.Action) (handled bool, ret runtime.Object, err error) {
			isUnstructured := false
			switch typedAction := action.(type) {
			case testing.CreateActionImpl:
				_, ok := typedAction.GetObject().(*unstructured.Unstructured)
				if ok {
					isUnstructured = true
					obj, err := castUnstructuredToObject(
						typedAction.Object.GetObjectKind().GroupVersionKind(),
						typedAction.GetObject().(*unstructured.Unstructured),
					)
					typedAction.Object = obj
					if err != nil {
						return true, nil, err
					}
				}
				action = typedAction
			case testing.UpdateActionImpl:
				_, ok := typedAction.GetObject().(*unstructured.Unstructured)
				if ok {
					isUnstructured = true
					obj, err := castUnstructuredToObject(
						typedAction.Object.GetObjectKind().GroupVersionKind(),
						typedAction.GetObject().(*unstructured.Unstructured),
					)
					typedAction.Object = obj
					if err != nil {
						return true, nil, err
					}
				}
				action = typedAction
			}
			handled, ret, err = reactor.React(action)
			if err != nil {
				return
			}

			switch action := action.(type) {
			case testing.ListActionImpl:
				kind := action.GetKind().Kind + "List"
				reflect.ValueOf(ret).Elem().FieldByName("Kind").Set(reflect.ValueOf(kind))
				reflect.ValueOf(ret).Elem().FieldByName("APIVersion").Set(reflect.ValueOf(action.GetKind().Version))
				return
			case testing.DeleteActionImpl:
				return
			case testing.CreateActionImpl:
				kind := action.Object.GetObjectKind().GroupVersionKind().Kind
				apiVersion := action.Object.GetObjectKind().GroupVersionKind().Version
				reflect.ValueOf(ret).Elem().FieldByName("Kind").Set(reflect.ValueOf(kind))
				reflect.ValueOf(ret).Elem().FieldByName("APIVersion").Set(reflect.ValueOf(apiVersion))
			case testing.UpdateActionImpl:
				kind := action.Object.GetObjectKind().GroupVersionKind().Kind
				apiVersion := action.Object.GetObjectKind().GroupVersionKind().Version
				reflect.ValueOf(ret).Elem().FieldByName("Kind").Set(reflect.ValueOf(kind))
				reflect.ValueOf(ret).Elem().FieldByName("APIVersion").Set(reflect.ValueOf(apiVersion))
			}

			if isUnstructured {
				if reflect.DeepEqual(ret, &unstructured.Unstructured{}) {
					return
				}
				buffer := new(bytes.Buffer)
				err = unstructured.UnstructuredJSONScheme.Encode(ret, buffer)
				if err != nil {
					return
				}
				ret, _, err = unstructured.UnstructuredJSONScheme.Decode(buffer.Bytes(), nil, nil)
			}
			return
		},
	}
}

// FakeClientPool provides a fake implementation of dynamic.ClientPool.
// It assumes resource GroupVersions are the same as their corresponding kind GroupVersions.
type FakeClientPool struct {
	*testing.Fake
}

// ClientForGroupVersionKind returns a client configured for the specified groupVersionResource.
// Resource may be empty.
func (p *FakeClientPool) ClientForGroupVersionResource(resource schema.GroupVersionResource) (dynamic.Interface, error) {
	return p.ClientForGroupVersionKind(resource.GroupVersion().WithKind(""))
}

// ClientForGroupVersionKind returns a client configured for the specified groupVersionKind.
// Kind may be empty.
func (p *FakeClientPool) ClientForGroupVersionKind(kind schema.GroupVersionKind) (dynamic.Interface, error) {
	// we can just create a new client every time for testing purposes
	return &FakeClient{
		FakeClient: &dynamic_fake.FakeClient{
			GroupVersion: kind.GroupVersion(),
			Fake:         p.Fake,
		},
	}, nil
}

// FakeClient is a fake implementation of dynamic.Interface.
type FakeClient struct {
	*dynamic_fake.FakeClient
}

// GetRateLimiter returns the rate limiter for this client.
func (c *FakeClient) GetRateLimiter() flowcontrol.RateLimiter {
	return nil
}

// Resource returns an API interface to the specified resource for this client's
// group and version.  If resource is not a namespaced resource, then namespace
// is ignored.  The ResourceClient inherits the parameter codec of this client
func (c *FakeClient) Resource(resource *metav1.APIResource, namespace string) dynamic.ResourceInterface {
	return &FakeResourceClient{
		FakeResourceClient: &dynamic_fake.FakeResourceClient{
			Resource:  c.GroupVersion.WithResource(resource.Name),
			Kind:      c.GroupVersion.WithKind(resource.Kind),
			Namespace: namespace,

			Fake: c.Fake,
		},
	}
}

// ParameterCodec returns a client with the provided parameter codec.
func (c *FakeClient) ParameterCodec(parameterCodec runtime.ParameterCodec) dynamic.Interface {
	return &FakeClient{
		FakeClient: &dynamic_fake.FakeClient{
			Fake: c.Fake,
		},
	}
}

// FakeResourceClient is a fake implementation of dynamic.ResourceInterface
type FakeResourceClient struct {
	*dynamic_fake.FakeResourceClient
}

func (c *FakeResourceClient) List(opts metav1.ListOptions) (runtime.Object, error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(c.Resource, c.Kind, c.Namespace, opts), &unstructured.UnstructuredList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &unstructured.UnstructuredList{}
	items, err := meta.ExtractList(obj)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		accessor, _ := meta.Accessor(item)
		if label.Matches(labels.Set(accessor.GetLabels())) {
			unstructuredObj, err := castObjectToUnstructured(item)
			if err != nil {
				return nil, err
			}
			list.Items = append(list.Items, *unstructuredObj)
		}
	}
	return list, err
}
