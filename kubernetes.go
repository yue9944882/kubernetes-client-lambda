package lambda

import (
	"os"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Resource is kubernetes resource enumeration hiding api version
type Resource string

type kubernetesOpInterface interface{}
type kubernetesApiGroupInterface interface{}
type kubernetesVersionInterface interface{}

type kubernetesExecutable struct {
	getFunc         func(namespace, name string) (runtime.Object, error)
	listFunc        func(namespace string) ([]runtime.Object, error)
	Rs              Resource
	clientInterface dynamic.Interface
}

type kubernetesWatchable struct {
	exec *kubernetesExecutable
}

// KubernetesClientLambda provides manipulation interface for resources
type KubernetesClientLambda interface {
	Type(Resource) KubernetesLambda
}

type KubernetesClientLambdaImpl struct {
	config          *rest.Config
	informerFactory informers.SharedInformerFactory
	clientPool      dynamic.ClientPool
}

func (kcl *KubernetesClientLambdaImpl) Type(rs Resource) KubernetesLambda {
	gvr := getResouceIndexerInstance().GetGroupVersionResource(rs)
	gvk := getResouceIndexerInstance().GetGroupVersionKind(rs)
	api := getResouceIndexerInstance().GetAPIResource(rs)
	i, err := kcl.clientPool.ClientForGroupVersionResource(gvr)
	if err != nil {
		panic(err)
	}
	return &kubernetesExecutable{
		getFunc: func(namespace, name string) (runtime.Object, error) {
			if kcl.informerFactory != nil {
				informer, err := kcl.informerFactory.ForResource(gvr)
				if err != nil {
					return nil, err
				}
				return informer.Lister().ByNamespace(namespace).Get(name)
			}
			tmpObj, err := i.Resource(&api, namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			obj, err := scheme.Scheme.New(gvk)
			if err != nil {
				panic(err)
			}
			if err := scheme.Scheme.Convert(tmpObj, obj, nil); err != nil {
				return nil, err
			}
			return obj, nil
		},
		listFunc: func(namespace string) ([]runtime.Object, error) {
			if kcl.informerFactory != nil {
				informer, err := kcl.informerFactory.ForResource(gvr)
				if err != nil {
					return nil, err
				}
				return informer.Lister().ByNamespace(namespace).List(labels.Everything())
			}
			tmpObjList, err := i.Resource(&api, namespace).List(metav1.ListOptions{})
			if err != nil {
				return nil, err
			}
			tmpObjs, err := meta.ExtractList(tmpObjList)
			if err != nil {
				return nil, err
			}
			retObjs := []runtime.Object{}
			for _, tmpObj := range tmpObjs {
				obj, err := scheme.Scheme.New(gvk)
				if err != nil {
					panic(err)
				}
				if err := scheme.Scheme.Convert(tmpObj, obj, nil); err != nil {
					return nil, err
				}
				retObjs = append(retObjs, obj)
			}
			return retObjs, nil
		},
		Rs:              rs,
		clientInterface: i,
	}
}

// KubernetesLambda provides access entry function for kubernetes
type KubernetesLambda interface {
	// WatchNamespace watches a namespace.
	// register or unregister "function"-typed lambda.
	// WatchNamespace(namespaces ...string) KubernetesWatch

	// InNamespace list the resources in the namespace with a default pager
	// and put them into lambda pipeline.
	InNamespace(namespaces ...string) *Lambda
}

// KubernetesWatch provides watch registry for kubernetes
type KubernetesWatch interface {
	Register(t watch.EventType, function Function) error
	Unregister(t watch.EventType, function Function) error
}

/*
// Register appends the function and it will be invoked as long as any event matches
// the event type arrives.
func (watchable *kubernetesWatchable) Register(t watch.EventType, function Function) error {
	entry := getWatchManager().registerFunc(watchable.exec.Rs, watchable.exec.Namespace, t, function)
	op, err := opInterface(watchable.exec.Rs, watchable.exec.Namespace, watchable.exec.clientset)
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
	op, err := opInterface(watchable.exec.Rs, watchable.exec.Namespace, watchable.exec.clientset)
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
func InCluster() KubernetesClientLambda {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return OutOfCluster(config)
}

// OutOfCluster establishe connection witha kube-apiserver by loading specific
// kube-config.
func OutOfCluster(config *rest.Config) *KubernetesClientLambdaImpl {
	return getKCLFromConfig(config)
}
*/

func getKCLFromConfig(config *rest.Config) *KubernetesClientLambdaImpl {
	// Resource discovery
	// Must succeed: panics on failure
	func() {
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}
		initIndexer(clientset)
	}()

	return &KubernetesClientLambdaImpl{
		config: config,
	}
}

// OutOfClusterDefault loads configuration from ~/.kube/config
func OutOfClusterDefault() KubernetesClientLambda {
	return OutOfClusterInContext("")
}

// OutOfClusterInContext is used to switch context of multi-cluster kubernetes
func OutOfClusterInContext(context string) KubernetesClientLambda {
	config, err := clientcmd.LoadFromFile(os.Getenv("HOME") + "/.kube/config")
	if err != nil {
		panic(err)
	}
	if config == nil || config.Contexts == nil {
		panic("something's wrong with kube config file")
	}
	if context != "" {
		config.CurrentContext = context
	}
	clientConfig, err := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		panic(err)
	}
	return getKCLFromConfig(clientConfig)
}

func (exec *kubernetesExecutable) InNamespace(namespaces ...string) *Lambda {

	if len(namespaces) == 0 {
		namespaces = []string{metav1.NamespaceAll}
	}

	ch := make(chan runtime.Object)

	l := &Lambda{
		rs:  exec.Rs,
		val: ch,
	}

	var wg sync.WaitGroup
	for _, namespace := range namespaces {
		go func() {
			wg.Add(1)
			objs, err := exec.listFunc(namespace)
			if err != nil {
				panic(err)
			}
			for _, obj := range objs {
				ch <- obj
			}
			wg.Done()
		}()
	}
	wg.Wait()
	close(ch)
	return l
}

/*
func (exec *kubernetesExecutable) WatchNamespace(namespace string) KubernetesWatch {
	exec.Namespace = namespace
	return &kubernetesWatchable{
		exec: exec,
	}
}
*/

func listResourceViaInformer(informerFactory informers.SharedInformerFactory, rs Resource, namespace string) (objs []runtime.Object, err error) {
	gvr := getResouceIndexerInstance().GetGroupVersionResource(rs)
	informer, err := informerFactory.ForResource(gvr)
	isNamespaced := getResouceIndexerInstance().IsNamespaced(rs)
	if isNamespaced {
		objs, err = informer.Lister().ByNamespace(namespace).List(labels.Everything())
	} else {
		objs, err = informer.Lister().List(labels.Everything())
	}
	return
}
