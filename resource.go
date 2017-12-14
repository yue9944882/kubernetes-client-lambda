package lambda

import (
	"k8s.io/client-go/kubernetes"
)

type Resource string

const (
	Namespace   Resource = " Namespace"
	Pod         Resource = "Pod"
	ReplicaSet  Resource = "ReplicaSet"
	Deployment  Resource = "Deployment"
	ConfigMap   Resource = "ConfigMap"
	Ingress     Resource = "Ingress"
	Service     Resource = "Service"
	Endpoint    Resource = "Endpoint"
	Secret      Resource = "Secret"
	DaemonSet   Resource = "DaemonSet"
	StatefulSet Resource = "StatefulSet"
)

func (rs Resource) Bind(clientset *kubernetes.Clientset) KubernetesClient {
	return &kubernetesExecutable{
		Clientset: clientset,
		Namespace: "default",
	}
}

func (rs Resource) Mock(namespaceAutoCreate bool) KubernetesClient {
	return &MockKubernetes{
		rs:                  rs,
		namespace:           "",
		namespaceAutoCreate: namespaceAutoCreate,
	}
}

func (rs Resource) String() string {
	return string(rs)
}
