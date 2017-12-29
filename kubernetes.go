package lambda

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	api_app_v1 "k8s.io/api/apps/v1beta1"
	api_v1 "k8s.io/api/core/v1"
	api_ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	api_store_v1 "k8s.io/api/storage/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

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
			addFuncs = append(addFuncs, wf.function)
		case watch.Modified:
			updateFuncs = append(updateFuncs, wf.function)
		case watch.Deleted:
			deleteFuncs = append(deleteFuncs, wf.function)
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

	go func() {
		if len(entry.watchFunctions) > 0 {
			listwatch, err := getListWatch(op)
			_, controller := cache.NewInformer(
				listwatch,
				watchable.exec.Rs.GetObject(),
				time.Second*1,
				handlerFuncs,
			)
			if err != nil {
				panic(err)
			}
			controller.Run(entry.stopCh)
		}
	}()
	return nil
}

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

func (rs Resource) GeResourcetName() string {
	switch rs {
	// Resource not in any namespace
	case Namespace:
		return "namespaces"
	case Node:
		return "nodes"
	case StorageClass:
		return "storageclasses"
	// Resource in a namespace
	case Pod:
		return "pods"
	case ConfigMap:
		return "configmaps"
	case Service:
		return "services"
	case Endpoint:
		return "endpoints"
	case Secret:
		return "secrets"
	case Ingress:
		return "ingresses"
	case ReplicaSet:
		return "replicasets"
	case Deployment:
		return "deployments"
	case DaemonSet:
		return "daemonsets"
	case StatefulSet:
		return "statefulsets"
	case ReplicationController:
		return "replicationcontrollers"
	default:
		return ""
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

func (exec *kubernetesExecutable) WatchNamespace(namespace string) *kubernetesWatchable {
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
	case Secret:
		return clientset.CoreV1().Secrets(namespace), nil
	default:
		return nil, fmt.Errorf("unknown resource type %s", rs.String())
	}
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
	case Pod:
		return clientset.CoreV1(), nil
	case ConfigMap:
		return clientset.CoreV1(), nil
	case Service:
		return clientset.CoreV1(), nil
	case Endpoint:
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
