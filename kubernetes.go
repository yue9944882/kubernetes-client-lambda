package lambda

import (
	"fmt"
	"reflect"

	"errors"

	api_app_v1 "k8s.io/api/apps/v1beta1"
	api_v1 "k8s.io/api/core/v1"
	api_ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	api_store_v1 "k8s.io/api/storage/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Resource string

type kubernetesResource interface{}
type kubernetesOpInterface interface{}

type kubernetesExecutable struct {
	clientset  kubernetes.Interface
	restconfig *rest.Config
	Namespace  string
	Rs         Resource
}

type KubernetesClient interface {
	InNamespace(namespace string) *Lambda
	All() *Lambda
}

const (
	Namespace             Resource = "Namespace"
	Node                  Resource = "Node"
	StorageClass          Resource = "StorageClass"
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

func getAllRuntimeObject() []runtime.Object {
	return []runtime.Object{
		Namespace.GetObject(),
		Node.GetObject(),
		StorageClass.GetObject(),
		Pod.GetObject(),
		ReplicaSet.GetObject(),
		ReplicationController.GetObject(),
		Deployment.GetObject(),
		ConfigMap.GetObject(),
		Ingress.GetObject(),
		Service.GetObject(),
		Endpoint.GetObject(),
		Secret.GetObject(),
		DaemonSet.GetObject(),
		StatefulSet.GetObject(),
	}
}

func (rs Resource) GetObject() runtime.Object {
	switch rs {
	// Resource not in any namespace
	case Namespace:
		return &api_v1.Namespace{}
	case Node:
		return &api_v1.Node{}
	case StorageClass:
		return &api_store_v1.StorageClass{}
	// Resource in a namespace
	case Pod:
		return &api_v1.Pod{}
	case ConfigMap:
		return &api_v1.ConfigMap{}
	case Service:
		return &api_v1.Service{}
	case Endpoint:
		return &api_v1.Endpoints{}
	case Secret:
		return &api_v1.Secret{}
	case Ingress:
		return &api_ext_v1beta1.Ingress{}
	case ReplicaSet:
		return &api_ext_v1beta1.ReplicaSet{}
	case Deployment:
		return &api_ext_v1beta1.Deployment{}
	case DaemonSet:
		return &api_ext_v1beta1.DaemonSet{}
	case StatefulSet:
		return &api_app_v1.StatefulSet{}
	case ReplicationController:
		return &api_v1.ReplicationController{}
	default:
		return nil
	}
}

func (rs Resource) InCluster() *kubernetesExecutable {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return rs.OutOfCluster(config)
}

func (rs Resource) OutOfCluster(config *rest.Config) *kubernetesExecutable {
	return &kubernetesExecutable{
		restconfig: config,
		Namespace:  meta_v1.NamespaceDefault,
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

func (exec *kubernetesExecutable) getClientset() kubernetes.Interface {
	if exec.clientset != nil {
		return exec.clientset
	}
	clientset, err := kubernetes.NewForConfig(exec.restconfig)
	if err != nil {
		return nil
	}
	return clientset
}

func opInterface(rs Resource, namespace string, clientset kubernetes.Interface) (kubernetesOpInterface, error) {
	if clientset == nil {
		return nil, errors.New("nil clientset proceed")
	}
	switch rs {

	// Resource not in any namespace
	case Namespace:
		return clientset.CoreV1().Namespaces(), nil
	case Node:
		return clientset.CoreV1().Nodes(), nil
	case StorageClass:
		return clientset.StorageV1().StorageClasses(), nil

	// Resource in a namespace
	case Pod:
		return clientset.CoreV1().Pods(namespace), nil
	case ConfigMap:
		return clientset.CoreV1().ConfigMaps(namespace), nil
	case Service:
		return clientset.CoreV1().Services(namespace), nil
	case Endpoint:
		return clientset.CoreV1().Endpoints(namespace), nil
	case Ingress:
		return clientset.ExtensionsV1beta1().Ingresses(namespace), nil
	case ReplicaSet:
		return clientset.ExtensionsV1beta1().ReplicaSets(namespace), nil
	case Deployment:
		return clientset.ExtensionsV1beta1().Deployments(namespace), nil
	case DaemonSet:
		return clientset.ExtensionsV1beta1().DaemonSets(namespace), nil
	case StatefulSet:
		return clientset.AppsV1beta1().StatefulSets(namespace), nil
	case ReplicationController:
		return clientset.CoreV1().ReplicationControllers(namespace), nil
	default:
		return nil, fmt.Errorf("unknown resource type %s", rs.String())
	}
}

func callListInterface(op kubernetesOpInterface) ([]kubernetesResource, error) {
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

func callGetInterface(op kubernetesOpInterface, name string) (kubernetesResource, error) {
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

func callCreateInterface(op kubernetesOpInterface, item kubernetesResource) (kubernetesResource, error) {
	method := reflect.ValueOf(op).MethodByName("Create")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(item),
	})
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}
	return ret[0].Interface(), nil
}

func callUpdateInterface(op kubernetesOpInterface, item kubernetesResource) (kubernetesResource, error) {
	method := reflect.ValueOf(op).MethodByName("Update")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(item),
	})
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}
	return ret[0].Interface(), nil
}

func callDeleteInterface(op kubernetesResource, name string) error {
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

func (exec *kubernetesExecutable) opListInterface() ([]kubernetesResource, error) {
	op, err := opInterface(exec.Rs, exec.Namespace, exec.getClientset())
	if err != nil {
		return nil, err
	}
	return callListInterface(op)
}

func (exec *kubernetesExecutable) opGetInterface(name string) (kubernetesResource, error) {
	op, err := opInterface(exec.Rs, exec.Namespace, exec.getClientset())
	if err != nil {
		return nil, err
	}
	return callGetInterface(op, name)
}

func (exec *kubernetesExecutable) opCreateInterface(item kubernetesResource) (kubernetesResource, error) {
	op, err := opInterface(exec.Rs, exec.Namespace, exec.getClientset())
	if err != nil {
		return nil, err
	}
	return callCreateInterface(op, item)
}

func (exec *kubernetesExecutable) opUpdateInterface(item kubernetesResource) (kubernetesResource, error) {
	op, err := opInterface(exec.Rs, exec.Namespace, exec.getClientset())
	if err != nil {
		return nil, err
	}
	return callUpdateInterface(op, item)
}

func (exec *kubernetesExecutable) opDeleteInterface(name string) error {
	op, err := opInterface(exec.Rs, exec.Namespace, exec.getClientset())
	if err != nil {
		return err
	}
	return callDeleteInterface(op, name)
}
