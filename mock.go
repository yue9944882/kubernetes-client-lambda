package lambda

import (
	"bytes"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	discovery_fake "k8s.io/client-go/discovery/fake"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"
)

// the mock KubernetesClient is statusful and if you want to reset its status then use MockReset
func Mock() KubernetesClientLambda {
	return &kubernetesClientLambdaImpl{
		clientPool: NewFakeClientPool(),
	}
}

func NewFakeClientPool() *dynamic_fake.FakeClientPool {
	/*
		fakePtr.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
			gvr := action.GetResource()
			ns := action.GetNamespace()
			watch, err := o.Watch(gvr, ns)
			if err != nil {
				return false, nil, err
			}
			return true, watch, nil
		})
	*/
	return &dynamic_fake.FakeClientPool{
		Fake: newFakeTracker(),
	}
}

func newFakeTracker() testing.Fake {
	return testing.Fake{
		ReactionChain: []testing.Reactor{newUnstructuredReactor()},
	}
}

func newUnstructuredReactor() *testing.SimpleReactor {
	o := testing.NewObjectTracker(scheme.Scheme, scheme.Codecs.UniversalDecoder())
	return &testing.SimpleReactor{
		Verb:     "*",
		Resource: "*",
		Reaction: func(action testing.Action) (handled bool, ret runtime.Object, err error) {
			switch typedAction := action.(type) {
			case testing.CreateActionImpl:
				gvk := typedAction.GetObject().(*unstructured.Unstructured).GroupVersionKind()
				obj, err := castUnstructuredToObject(gvk, typedAction.GetObject().(*unstructured.Unstructured))
				if err != nil {
					return true, nil, err
				}
				typedAction.Object = obj
				action = typedAction
			case testing.UpdateActionImpl:
				gvk := typedAction.GetObject().(*unstructured.Unstructured).GroupVersionKind()
				obj, err := castUnstructuredToObject(gvk, typedAction.GetObject().(*unstructured.Unstructured))
				if err != nil {
					return true, nil, err
				}
				typedAction.Object = obj
				action = typedAction
			}
			originalReactFunc := testing.ObjectReaction(o)
			handled, ret, err = originalReactFunc(action)
			if err != nil {
				return
			}

			switch action := action.(type) {
			case testing.ListActionImpl:
				kind := action.GetKind().Kind + "List"
				reflect.ValueOf(ret).Elem().FieldByName("Kind").Set(reflect.ValueOf(kind))
				reflect.ValueOf(ret).Elem().FieldByName("APIVersion").Set(reflect.ValueOf(action.GetKind().Version))
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

			if reflect.DeepEqual(ret, &unstructured.Unstructured{}) {
				return
			}

			buffer := new(bytes.Buffer)
			err = unstructured.UnstructuredJSONScheme.Encode(ret, buffer)
			if err != nil {
				return
			}
			ret, _, err = unstructured.UnstructuredJSONScheme.Decode(buffer.Bytes(), nil, nil)
			return
		},
	}
}

func newFakeDiscovery() *discovery_fake.FakeDiscovery {
	fakePtr := newFakeTracker()
	o := testing.NewObjectTracker(scheme.Scheme, scheme.Codecs.UniversalDecoder())
	fakePtr.AddReactor("*", "*", testing.ObjectReaction(o))
	return &discovery_fake.FakeDiscovery{
		Fake: &fakePtr,
	}
}
