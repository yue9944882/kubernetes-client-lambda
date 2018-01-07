package lambda

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// Resource is kubernetes resource enumeration hiding api version
type Resource string

type kubernetesResource interface{}
type kubernetesOpInterface interface{}
type kubernetesVersionInterface interface{}

type kubernetesExecutable struct {
	clientset  kubernetes.Interface
	restconfig *rest.Config
	Namespace  string
	Rs         Resource
}

type kubernetesWatchable struct {
	exec *kubernetesExecutable
}

// KubernetesLambda provides access entry function for kubernetes
type KubernetesLambda interface {
	// WatchNamespace watches a namespace.
	// register or unregister "function"-typed lambda.
	WatchNamespace(namespace string) KubernetesWatch

	// InNamespace list the resources in the namespace with a default pager
	// and put them into lambda pipeline.
	InNamespace(namespace string) *Lambda

	// All list all the resources.
	// ** Note that this method should only be used for resource which doesn't
	// ** belong to any namespace. e.g. Node, Namespace..
	All() *Lambda

	// WatchAll watches the change from resources.
	// ** Note that this method should only be used for resource which doesn't
	// ** belong to any namespace. e.g. Node, Namespace..
	WatchAll() KubernetesWatch
}

// KubernetesWatch provides watch registry for kubernetes
type KubernetesWatch interface {
	Register(t watch.EventType, function Function) error
	Unregister(t watch.EventType, function Function) error
}

// Register appends the function and it will be invoked as long as any event matches
// the event type arrives.
func (watchable *kubernetesWatchable) Register(t watch.EventType, function Function) error {
	entry := getWatchManager().registerFunc(watchable.exec.Rs, watchable.exec.Namespace, t, function)
	op, err := opInterface(watchable.exec.Rs, watchable.exec.Namespace, watchable.exec.getClientset())
	if err != nil {
		return err
	}

	addFuncs := []Function{}
	updateFuncs := []Function{}
	deleteFuncs := []Function{}

	for _, wf := range entry.watchFunctions {
		switch wf.t {
		case watch.Added:
			addFuncs = append(addFuncs, wf.function.(Function))
		case watch.Modified:
			updateFuncs = append(updateFuncs, wf.function.(Function))
		case watch.Deleted:
			deleteFuncs = append(deleteFuncs, wf.function.(Function))
		}
	}

	handlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			for _, addf := range addFuncs {
				callFunction(addf, obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			for _, updatef := range updateFuncs {
				callFunction(updatef, newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			for _, deletef := range deleteFuncs {
				callFunction(deletef, obj)
			}
		},
	}

	if len(entry.watchFunctions) > 0 {
		go func() {
			listwatch, err := getListWatch(op)
			_, controller := cache.NewInformer(
				listwatch,
				watchable.exec.Rs.GetObject(),
				time.Second*0,
				handlerFuncs,
			)
			if err != nil {
				panic(err)
			}
			controller.Run(entry.stopCh)
		}()
	}
	return nil
}

// Unregister make sure the function won't be invoked again even if any event matching the event
// type arrives.
func (watchable *kubernetesWatchable) Unregister(t watch.EventType, function Function) error {
	entry := getWatchManager().unregisterFunc(watchable.exec.Rs, watchable.exec.Namespace, t, function)
	op, err := opInterface(watchable.exec.Rs, watchable.exec.Namespace, watchable.exec.getClientset())
	if err != nil {
		return err
	}
	go func() {
		if len(entry.watchFunctions) > 0 {
			listwatch, err := getListWatch(op)
			_, controller := cache.NewInformer(
				listwatch,
				watchable.exec.Rs.GetObject(),
				time.Second*0,
				cache.ResourceEventHandlerFuncs{},
			)
			if err != nil {
				panic(err)
			}
			controller.Run(entry.stopCh)
		}
	}()
	return nil
}

// InCluster establishes connection with kube-apiserver if the program is
// running in a kubernetes cluster.
func (rs Resource) InCluster() KubernetesLambda {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return rs.OutOfCluster(config)
}

// OutOfCluster establishe connection witha kube-apiserver by loading specific
// kube-config.
func (rs Resource) OutOfCluster(config *rest.Config) KubernetesLambda {
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
			l.addError(err)
			return
		}
		for resource := range resources {
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
			l.addError(err)
			return
		}
		for resource := range resources {
			ch <- resource
		}
	}()
	return l
}

func (exec *kubernetesExecutable) WatchNamespace(namespace string) KubernetesWatch {
	exec.Namespace = namespace
	return &kubernetesWatchable{
		exec: exec,
	}
}

func (exec *kubernetesExecutable) WatchAll() KubernetesWatch {
	return &kubernetesWatchable{
		exec: exec,
	}
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

func getListWatch(op kubernetesOpInterface) (*cache.ListWatch, error) {
	listFunc := func(options meta_v1.ListOptions) (runtime.Object, error) {
		method := reflect.ValueOf(op).MethodByName("List")
		ret := method.Call([]reflect.Value{
			reflect.ValueOf(meta_v1.ListOptions{}),
		})
		if err := ret[1].Interface(); err != nil {
			return nil, err.(error)
		}
		return ret[0].Interface().(runtime.Object), nil
	}
	watchFunc := func(options meta_v1.ListOptions) (watch.Interface, error) {
		method := reflect.ValueOf(op).MethodByName("Watch")
		ret := method.Call([]reflect.Value{
			reflect.ValueOf(meta_v1.ListOptions{}),
		})
		if err := ret[1].Interface(); err != nil {
			return nil, err.(error)
		}
		return ret[0].Interface().(watch.Interface), nil
	}
	return &cache.ListWatch{
		ListFunc:  listFunc,
		WatchFunc: watchFunc,
	}, nil
}

func apiInterface(rs Resource, clientset kubernetes.Interface) (kubernetesVersionInterface, error) {
	if clientset == nil {
		return nil, errors.New("nil clientset proceed")
	}
	switch rs {
	case Namespace:
		return clientset.CoreV1(), nil
	case Node:
		return clientset.CoreV1(), nil
	case StorageClass:
		return clientset.StorageV1(), nil
	// Resource in a namespace

	// core/v1
	case Pod:
		return clientset.CoreV1(), nil
	case ConfigMap:
		return clientset.CoreV1(), nil
	case Service:
		return clientset.CoreV1(), nil
	case Endpoints:
		return clientset.CoreV1(), nil
	case LimitRange:
		return clientset.CoreV1(), nil
	case Ingress:
		return clientset.ExtensionsV1beta1(), nil
	case ReplicaSet:
		return clientset.ExtensionsV1beta1(), nil
	case Deployment:
		return clientset.ExtensionsV1beta1(), nil
	case DaemonSet:
		return clientset.ExtensionsV1beta1(), nil
	case StatefulSet:
		return clientset.AppsV1beta1(), nil
	case ReplicationController:
		return clientset.CoreV1(), nil
	case Secret:
		return clientset.CoreV1(), nil
	case Job:
		return clientset.BatchV1(), nil
	case CronJob:
		return clientset.BatchV2alpha1(), nil
	default:
		return nil, fmt.Errorf("unknown resource type %s", rs.String())
	}
}

func callListInterface(op kubernetesOpInterface) (<-chan kubernetesResource, error) {
	method := reflect.ValueOf(op).MethodByName("List")
	pgr := &ListPager{
		PageSize: 128,
		PageFn:   method.Interface(),
	}
	return pgr.List(meta_v1.ListOptions{})
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

func callDeleteInterface(op kubernetesOpInterface, name string) error {
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

func callWatchInterface(op kubernetesOpInterface) (<-chan watch.Event, error) {
	method := reflect.ValueOf(op).MethodByName("Watch")
	ret := method.Call([]reflect.Value{
		reflect.ValueOf(meta_v1.ListOptions{}),
	})
	if err := ret[1].Interface(); err != nil {
		return nil, err.(error)
	}
	watcher := ret[0].Interface().(watch.Interface)
	return watcher.ResultChan(), nil
}

func callRESTClientInterface(api kubernetesVersionInterface) rest.Interface {
	method := reflect.ValueOf(api).MethodByName("RESTClient")
	ret := method.Call([]reflect.Value{})
	client := ret[0].Interface().(rest.Interface)
	return client
}

func (exec *kubernetesExecutable) opListInterface() (<-chan kubernetesResource, error) {
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

func (exec *kubernetesExecutable) opWatchInterface(t watch.EventType) (<-chan kubernetesResource, error) {
	op, err := opInterface(exec.Rs, exec.Namespace, exec.getClientset())
	if err != nil {
		return nil, err
	}
	eventCh, err := callWatchInterface(op)
	if err != nil {
		return nil, err
	}
	rsCh := make(chan kubernetesResource)
	go func() {
		defer close(rsCh)
		for event := range eventCh {
			if event.Type == t {
				rsCh <- event.Object
			}
		}
	}()
	return rsCh, nil
}

func (exec *kubernetesExecutable) opGetRESTClient() (rest.Interface, error) {
	api, err := apiInterface(exec.Rs, exec.getClientset())
	if err != nil {
		return nil, err
	}
	return callRESTClientInterface(api), nil
}
