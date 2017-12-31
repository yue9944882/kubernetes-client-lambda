package lambda

import (
	"sync"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
)

var (
	mockclient = newFakeClientSet(getAllRuntimeObject()...)
	mutex      = sync.Mutex{}
)

// Mock return a mock interface of lambda KubernetesClient
// the mock KubernetesClient is statusful and if you want to reset its status then use MockReset
func (rs Resource) Mock() *kubernetesExecutable {
	mutex.Lock()
	defer mutex.Unlock()
	return &kubernetesExecutable{
		Rs:        rs,
		Namespace: meta_v1.NamespaceDefault,
		clientset: mockclient,
	}
}

func (rs Resource) ResetMock() {
	mutex.Lock()
	defer mutex.Unlock()
	mockclient = newFakeClientSet(getAllRuntimeObject()...)

}

func newFakeClientSet(objects ...runtime.Object) kubernetes.Interface {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)

	meta_v1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	fake.AddToScheme(scheme)
	o := &kubernetesTracker{
		scheme:   scheme,
		decoder:  codecs.UniversalDecoder(),
		objects:  make(map[schema.GroupVersionResource][]runtime.Object),
		watchers: make(map[schema.GroupVersionResource]map[string]*watch.FakeWatcher),
	}
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}
	fakePtr := testing.Fake{}
	fakePtr.AddReactor("*", "*", testing.ObjectReaction(o))
	fakePtr.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, watch, nil
	})
	return &fake.Clientset{Fake: fakePtr}
}
