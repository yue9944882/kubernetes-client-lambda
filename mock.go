package lambda

import (
	"sync"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	mockclient = fake.NewSimpleClientset(getAllRuntimeObject()...)
	mutex      = sync.Mutex{}
)

type KubernetesClientLambdaMock struct {
	clientset kubernetes.Interface
}

func (kcl *KubernetesClientLambdaMock) Type(rs Resource) KubernetesLambda {
	return &kubernetesExecutable{
		clientset: kcl.clientset,
		Namespace: meta_v1.NamespaceDefault,
		Rs:        rs,
	}
}

// Mock return a mock interface of lambda KubernetesClient
// the mock KubernetesClient is statusful and if you want to reset its status then use MockReset
func Mock() KubernetesClientLambda {
	mutex.Lock()
	defer mutex.Unlock()
	return &KubernetesClientLambdaMock{
		clientset: mockclient,
	}
}

func (rs Resource) ResetMock() {
	mutex.Lock()
	defer mutex.Unlock()
	mockclient = fake.NewSimpleClientset(getAllRuntimeObject()...)
}
