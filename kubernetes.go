package lambda

import (
	"fmt"
	"reflect"

	"errors"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Resource string

type kubernetesResource interface{}
type kubernetesOpInterface interface{}

type kubernetesExecutable struct {
	restconfig *rest.Config
	Namespace  string
	Rs         Resource
}

type KubernetesClient interface {
	InNamespace(namespace string) *Lambda
	All() *Lambda
}

func init() {

}

const (
	Namespace             Resource = " Namespace"
	StorageClass          Resource = "StorageClass"
	Node                  Resource = "Node"
	Pod                   Resource = "Pod"
	ReplicaSet            Resource = "ReplicaSet"
	ReplicationController Resource = "ReplicationController"
	Deployment            Resource = "Deployment"
	ConfigMap             Resource = "ConfigMap"
	Ingress               Resource = "Ingress"
	Service               Resource = "Service"
	Endpoint              Resource = "Endpoint"
	Secret                Resource = "Secret"
	DaemonSet             Resource = "DaemonSet"
	StatefulSet           Resource = "StatefulSet"
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

func (rs Resource) String() string {
	return string(rs)
}

func (exec *kubernetesExecutable) InNamespace(namespace string) (l *Lambda) {
	exec.Namespace = namespace
	ch := make(chan kubernetesResource)
	l = &Lambda{
		op:  exec,
		val: ch,
	}
	go func() {
		defer close(ch)
		resources, err := exec.opListInterface()
		if err != nil {
			l.Error = err
			return
		}
		for _, resource := range resources {
			ch <- resource
		}
	}()
	return l
}

func (exec *kubernetesExecutable) All() (l *Lambda) {
	ch := make(chan kubernetesResource)
	l = &Lambda{
		op:  exec,
		val: ch,
	}
	go func() {
		defer close(ch)
		resources, err := exec.opListInterface()
		if err != nil {
			l.Error = err
			return
		}
		for _, resource := range resources {
			ch <- resource
		}
	}()
	return l
}

func (exec *kubernetesExecutable) opInterface() (kubernetesOpInterface, error) {
	clientset, err := kubernetes.NewForConfig(exec.restconfig)
	if err != nil {
		return nil, err
	}
	switch exec.Rs {

	// Resource not in any namespace
	case Namespace:
		return clientset.CoreV1().Namespaces(), nil
	case Node:
		return clientset.CoreV1().Nodes(), nil
	case StorageClass:
		return clientset.StorageV1().StorageClasses(), nil

	// Resource in a namespace
	case Pod:
		return clientset.CoreV1().Pods(exec.Namespace), nil
	case ConfigMap:
		return clientset.CoreV1().ConfigMaps(exec.Namespace), nil
	case Service:
		return clientset.CoreV1().Services(exec.Namespace), nil
	case Endpoint:
		return clientset.CoreV1().Endpoints(exec.Namespace), nil
	case Ingress:
		return clientset.ExtensionsV1beta1().Ingresses(exec.Namespace), nil
	case ReplicaSet:
		return clientset.ExtensionsV1beta1().ReplicaSets(exec.Namespace), nil
	case Deployment:
		return clientset.ExtensionsV1beta1().Deployments(exec.Namespace), nil
	case DaemonSet:
		return clientset.ExtensionsV1beta1().DaemonSets(exec.Namespace), nil
	case StatefulSet:
		return clientset.AppsV1beta1().StatefulSets(exec.Namespace), nil
	case ReplicationController:
		return clientset.CoreV1().ReplicationControllers(exec.Namespace), nil
	default:
		return nil, fmt.Errorf("unknown resource type %s", exec.Rs.String())
	}
}

func (exec *kubernetesExecutable) opListInterface() ([]kubernetesResource, error) {
	op, err := exec.opInterface()
	if err != nil {
		return nil, err
	}
	var resources []kubernetesResource
	method := reflect.ValueOf(op).MethodByName("List")
	var ret []reflect.Value
	if method.Type().NumIn() == 0 {
		ret = method.Call([]reflect.Value{})
	} else if method.Type().NumIn() == 1 {
		ret = method.Call([]reflect.Value{
			reflect.ValueOf(meta_v1.ListOptions{}),
		})
	}
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}
	items := reflect.Indirect(ret[0]).FieldByName("Items")
	if items.Type().Kind() != reflect.Slice {
		return nil, errors.New("tainted results from list method")
	}
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Addr().Interface()
		resources = append(resources, item)
	}
	return resources, nil
}

func (exec *kubernetesExecutable) opGetInterface(name string) (kubernetesResource, error) {
	op, err := exec.opInterface()
	if err != nil {
		return nil, err
	}
	method := reflect.ValueOf(op).MethodByName("Get")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(name),
		reflect.ValueOf(meta_v1.GetOptions{}),
	})
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}
	return ret[0].Interface(), nil
}

func (exec *kubernetesExecutable) opCreateInterface(item kubernetesResource) (kubernetesResource, error) {
	op, err := exec.opInterface()
	if err != nil {
		return nil, err
	}
	method := reflect.ValueOf(op).MethodByName("Create")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(item),
	})
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}
	return ret[0].Interface(), nil
}

func (exec *kubernetesExecutable) opUpdateInterface(item kubernetesResource) (kubernetesResource, error) {
	op, err := exec.opInterface()
	if err != nil {
		return nil, err
	}
	method := reflect.ValueOf(op).MethodByName("Update")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(item),
	})
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}
	return ret[0].Interface(), nil
}

func (exec *kubernetesExecutable) opDeleteInterface(name string) error {
	op, err := exec.opInterface()
	if err != nil {
		return err
	}
	method := reflect.ValueOf(op).MethodByName("Delete")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(name),
		reflect.ValueOf(&meta_v1.DeleteOptions{}),
	})
	if err := ret[0].Interface(); err != nil {
		return err.(error)
	}
	return nil
}
