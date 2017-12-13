package lambda

import (
	"reflect"

	"errors"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type kubernetesResource interface{}
type kubernetesOpInterface interface{}

type kubernetesExecutable struct {
	Clientset *kubernetes.Clientset
	Namespace string
	Rs        Resource
}

type KubernetesClient interface {
	InNamespace(namespace string) *Lambda
	All() *Lambda
}

func (exec *kubernetesExecutable) InNamespace(namespace string) (l *Lambda) {
	exec.Namespace = namespace
	ch := make(chan interface{})
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
	ch := make(chan interface{})
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

func (exec *kubernetesExecutable) opInterface() kubernetesOpInterface {
	switch exec.Rs {
	case Namespace:
		return exec.Clientset.CoreV1().Namespaces()
	case Pod:
		return exec.Clientset.CoreV1().Pods(exec.Namespace)
	case ConfigMap:
		return exec.Clientset.CoreV1().ConfigMaps(exec.Namespace)
	case Service:
		return exec.Clientset.CoreV1().Services(exec.Namespace)
	case Endpoint:
		return exec.Clientset.CoreV1().Endpoints(exec.Namespace)
	case Ingress:
		return exec.Clientset.ExtensionsV1beta1().Ingresses(exec.Namespace)
	case ReplicaSet:
		return exec.Clientset.ExtensionsV1beta1().ReplicaSets(exec.Namespace)
	case Deployment:
		return exec.Clientset.ExtensionsV1beta1().Deployments(exec.Namespace)
	case DaemonSet:
		return exec.Clientset.ExtensionsV1beta1().DaemonSets(exec.Namespace)
	case StatefulSet:
		return exec.Clientset.AppsV1beta1().StatefulSets(exec.Namespace)
	default:
		return nil
	}
}

func (exec *kubernetesExecutable) opListInterface() ([]kubernetesResource, error) {
	var resources []kubernetesResource
	method := reflect.ValueOf(exec.opInterface()).MethodByName("List")
	var ret []reflect.Value
	if method.Type().NumIn() == 0 {
		ret = method.Call([]reflect.Value{})
	} else if method.Type().NumIn() == 1 {
		ret = method.Call([]reflect.Value{
			reflect.ValueOf(meta_v1.ListOptions{}),
		})
	}
	if err := ret[1].Interface().(error); err != nil {
		return nil, err
	}
	items := ret[0].FieldByName("Items")
	if reflect.TypeOf(items).Kind() != reflect.Slice {
		return nil, errors.New("tainted results from list method")
	}
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Interface()
		resources = append(resources, item)
	}
	return resources, nil
}

func (exec *kubernetesExecutable) opGetInterface(name string) (kubernetesResource, error) {
	method := reflect.ValueOf(exec.opInterface()).MethodByName("Get")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(name),
		reflect.ValueOf(meta_v1.GetOptions{}),
	})
	if err := ret[1].Interface().(error); err != nil {
		return nil, err
	}
	return ret[0].Interface(), nil
}

func (exec *kubernetesExecutable) opCreateInterface(item kubernetesResource) (kubernetesResource, error) {
	method := reflect.ValueOf(exec.opInterface()).MethodByName("Create")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(item),
	})
	if err := ret[1].Interface().(error); err != nil {
		return nil, err
	}
	return ret[0].Interface(), nil
}

func (exec *kubernetesExecutable) opUpdateInterface(item kubernetesResource) (kubernetesResource, error) {
	method := reflect.ValueOf(exec.opInterface()).MethodByName("Update")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(item),
	})
	if err := ret[1].Interface().(error); err != nil {
		return nil, err
	}
	return ret[0].Interface(), nil
}

func (exec *kubernetesExecutable) opDeleteInterface(name string) error {
	method := reflect.ValueOf(exec.opInterface()).MethodByName("Delete")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(name),
		reflect.ValueOf(&meta_v1.DeleteOptions{}),
	})
	return ret[0].Interface().(error)
}
