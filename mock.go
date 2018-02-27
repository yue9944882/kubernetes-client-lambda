package lambda

import (
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"
)

// the mock KubernetesClient is statusful and if you want to reset its status then use MockReset
func Mock() KubernetesClientLambda {
	o := testing.NewObjectTracker(scheme.Scheme, scheme.Codecs.UniversalDecoder())

	fakePtr := testing.Fake{}
	fakePtr.AddReactor("*", "*", testing.ObjectReaction(o))
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
	return &KubernetesClientLambdaImpl{
		clientPool: &fake.FakeClientPool{fakePtr},
	}
}
