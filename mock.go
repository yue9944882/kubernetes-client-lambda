package lambda

import (
	"sync"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	mockclient = fake.NewSimpleClientset(getAllRuntimeObject()...)
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
	mockclient = fake.NewSimpleClientset(getAllRuntimeObject()...)
}
