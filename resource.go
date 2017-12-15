package lambda

import (
	"k8s.io/client-go/rest"
)

type Resource string

func init() {

}

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

func (rs Resource) InCluster() KubernetesClient {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return rs.OutOfCluster(config)
}

func (rs Resource) OutOfCluster(config *rest.Config) KubernetesClient {
	return &kubernetesExecutable{
		restconfig: config,
		Namespace:  "default",
		Rs:         rs,
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
