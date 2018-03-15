package lambda

import (
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Rs              Resource
	namespaces      []string
	clientInterface dynamic.Interface
	informer        informers.GenericInformer
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
	i, err := kcl.clientPool.ClientForGroupVersionResource(gvr)
	if err != nil {
		panic(err)
	}
	if kcl.informerFactory != nil {
		informer, err := kcl.informerFactory.ForResource(gvr)
		if err != nil {
			panic(err)
		}
		if informer.Informer().LastSyncResourceVersion() == "" {
			kcl.informerFactory.Start(make(chan struct{}))
			// TODO: set timeout for waiting cache sync
			cache.WaitForCacheSync(make(chan struct{}), informer.Informer().HasSynced)
		}
	}

	return &kubernetesExecutable{
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
	rs := exec.Rs
	gvk := GetResouceIndexerInstance().GetGroupVersionKind(rs)
	api := GetResouceIndexerInstance().GetAPIResource(rs)

	exec.namespaces = namespaces

	ch := make(chan runtime.Object)

	if len(namespaces) == 0 {
		exec.namespaces = []string{metav1.NamespaceAll}
	}

	l := &Lambda{
		rs:         exec.Rs,
		namespaces: exec.namespaces,
		val:        ch,
		getFunc: func(namespace, name string) (runtime.Object, error) {
			if exec.informer != nil {
				return exec.informer.Lister().ByNamespace(namespace).Get(name)
			}
			tmpObj, err := exec.clientInterface.Resource(api, namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			obj, err := scheme.Scheme.New(gvk)
			if err != nil {
				panic(err)
			}
			tmpObj.GetObjectKind().SetGroupVersionKind(gvk)
			if err := scheme.Scheme.Convert(tmpObj, obj, nil); err != nil {
				return nil, err
			}
			return obj, nil
		},
		listFunc: func(namespace string) ([]runtime.Object, error) {
			if exec.informer != nil {
				return exec.informer.Lister().ByNamespace(namespace).List(labels.Everything())
			}
			tmpObjList, err := exec.clientInterface.Resource(api, namespace).List(metav1.ListOptions{})
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
				tmpObj.GetObjectKind().SetGroupVersionKind(gvk)
				if err := scheme.Scheme.Convert(tmpObj, obj, nil); err != nil {
					return nil, err
				}
				retObjs = append(retObjs, obj)
			}
			return retObjs, nil
		},
		createFunc: func(object runtime.Object) error {
			api := GetResouceIndexerInstance().GetAPIResource(rs)
			accessor, err := meta.Accessor(object)
			if err != nil {
				return err
			}
			tmpObj, err := castObjectToUnstructured(object)
			if err != nil {
				return err
			}
			if _, err := exec.clientInterface.Resource(api, accessor.GetNamespace()).Create(tmpObj); err != nil {
				return err
			}
			return nil
		},
		updateFunc: func(object runtime.Object) error {
			api := GetResouceIndexerInstance().GetAPIResource(rs)
			accessor, err := meta.Accessor(object)
			if err != nil {
				return err
			}
			object.GetObjectKind().SetGroupVersionKind(gvk)
			tmpObj, err := castObjectToUnstructured(object)
			if err != nil {
				return err
			}
			if _, err := exec.clientInterface.Resource(api, accessor.GetNamespace()).Update(tmpObj); err != nil {
				return err
			}
			return nil
		},
		deleteFunc: func(object runtime.Object) error {
			api := GetResouceIndexerInstance().GetAPIResource(rs)
			accessor, err := meta.Accessor(object)
			if err != nil {
				return err
			}
			if err := exec.clientInterface.Resource(api, accessor.GetNamespace()).Delete(accessor.GetName(), &metav1.DeleteOptions{}); err != nil {
				return err
			}
			return nil
		},
	}
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
