package lambda

import (
	"sync"

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
		Rs: rs,
	}
}

// the mock KubernetesClient is statusful and if you want to reset its status then use MockReset
func Mock() KubernetesClientLambda {
	mutex.Lock()
	defer mutex.Unlock()
	initIndexer(mockclient)
	return &KubernetesClientLambdaMock{
		clientset: mockclient,
	}
}

func (rs Resource) ResetMock() {
	mutex.Lock()
	defer mutex.Unlock()
	mockclient = fake.NewSimpleClientset(getAllRuntimeObject()...)
}
