package lambda

import (
	"bytes"
	"os"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// Resource is kubernetes resource enumeration hiding api version
type Resource string

type kubernetesExecutable struct {
	getFunc         func(namespace, name string) (runtime.Object, error)
	listFunc        func(namespace string) ([]runtime.Object, error)
	Rs              Resource
	clientInterface dynamic.Interface
}

// KubernetesClientLambda provides manipulation interface for resources
type KubernetesClientLambda interface {
	Type(Resource) KubernetesLambda
}

type kubernetesClientLambdaImpl struct {
	informerFactory informers.SharedInformerFactory
	clientPool      dynamic.ClientPool
}

func (kcl *kubernetesClientLambdaImpl) Type(rs Resource) KubernetesLambda {
	gvr := GetResouceIndexerInstance().GetGroupVersionResource(rs)
	gvk := GetResouceIndexerInstance().GetGroupVersionKind(rs)
	api := GetResouceIndexerInstance().GetAPIResource(rs)
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
			tmpObj, err := i.Resource(api, namespace).Get(name, metav1.GetOptions{})
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
				if err == nil {
					if informer.Informer().LastSyncResourceVersion() == "" {
						kcl.informerFactory.Start(make(chan struct{}))
						// TODO: set timeout for waiting cache sync
						cache.WaitForCacheSync(make(chan struct{}), informer.Informer().HasSynced)
					}
					return informer.Lister().ByNamespace(namespace).List(labels.Everything())
				}
			}
			tmpObjList, err := i.Resource(api, namespace).List(metav1.ListOptions{})
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
					return nil, err
				}
				buffer := new(bytes.Buffer)
				if err := unstructured.UnstructuredJSONScheme.Encode(tmpObj, buffer); err != nil {
					return nil, err
				}
				if _, _, err := unstructured.UnstructuredJSONScheme.Decode(buffer.Bytes(), nil, obj); err != nil {
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

	// InNamespace list the resources in the namespace with a default pager
	// and put them into lambda pipeline.
	InNamespace(namespaces ...string) *Lambda
}

func getKCLFromConfig(config *rest.Config) *kubernetesClientLambdaImpl {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Resource discovery
	// Must succeed: panics on failure
	func() {
		initIndexer(clientset)
	}()

	factory := informers.NewSharedInformerFactory(clientset, time.Minute)
	return &kubernetesClientLambdaImpl{
		informerFactory: factory,
		clientPool:      dynamic.NewDynamicClientPool(config),
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
		getFunc:         exec.getFunc,
		rs:              exec.Rs,
		val:             ch,
		clientInterface: exec.clientInterface,
	}

	var wg sync.WaitGroup
	go func() {
		for _, namespace := range namespaces {
			wg.Add(1)
			go func() {
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
	}()
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
	gvr := GetResouceIndexerInstance().GetGroupVersionResource(rs)
	informer, err := informerFactory.ForResource(gvr)
	isNamespaced := GetResouceIndexerInstance().IsNamespaced(rs)
	if isNamespaced {
		objs, err = informer.Lister().ByNamespace(namespace).List(labels.Everything())
	} else {
		objs, err = informer.Lister().List(labels.Everything())
	}
	return
}
