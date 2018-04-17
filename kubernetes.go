package lambda

import (
	"os"
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	versionParseRegexp = regexp.MustCompile(`v(\d+)((alpha|beta)(\d+))?`)
)

type Version string

func (v Version) GetNumericVersion() string {
	matches := versionParseRegexp.FindStringSubmatch(string(v))
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func (v Version) GetSuffix() string {
	// e.g. v1beta1,
	matches := versionParseRegexp.FindStringSubmatch(string(v))
	if len(matches) < 5 {
		return ""
	}
	return matches[3] + matches[4]
}

// Resource is kubernetes resource enumeration hiding api version
type Resource struct {
	Name    string
	Version string
}

type kubernetesExecutable struct {
	Rs              Resource
	namespaces      []string
	clientInterface dynamic.Interface
	informer        informers.GenericInformer
}

// KubernetesClientLambda provides manipulation interface for resources
type KubernetesClientLambda interface {
	Type(Resource) *kubernetesExecutable
	GetRestConfig() *rest.Config
}

type kubernetesClientLambdaImpl struct {
	informerFactory informers.SharedInformerFactory
	clientPool      dynamic.ClientPool
	restConfig      *rest.Config
}

func (kcl *kubernetesClientLambdaImpl) GetRestConfig() *rest.Config {
	return kcl.restConfig
}

func (kcl *kubernetesClientLambdaImpl) Type(rs Resource) *kubernetesExecutable {
	gvr := GetResouceIndexerInstance().GetGroupVersionResource(rs)
	i, err := kcl.clientPool.ClientForGroupVersionResource(gvr)
	if err != nil {
		panic(err)
	}

	exec := &kubernetesExecutable{
		Rs:              rs,
		clientInterface: i,
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
		exec.informer = informer
	}
	return exec
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
		restConfig:      config,
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
			return exec.informer.Lister().ByNamespace(namespace).Get(name)
		},
		listFunc: func(namespace string, selector labels.Selector) ([]runtime.Object, error) {
			return exec.informer.Lister().ByNamespace(namespace).List(selector)
		},
		createFunc: func(object runtime.Object) error {
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
			if _, err := exec.clientInterface.Resource(api, accessor.GetNamespace()).Create(tmpObj); err != nil {
				return err
			}
			cache.WaitForCacheSync(make(chan struct{}), exec.informer.Informer().HasSynced)
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
			cache.WaitForCacheSync(make(chan struct{}), exec.informer.Informer().HasSynced)
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
			cache.WaitForCacheSync(make(chan struct{}), exec.informer.Informer().HasSynced)
			return nil
		},
	}
	close(ch)

	return l
}

func (exec *kubernetesExecutable) OnAdd(f func(interface{})) {
	exec.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: f,
	})
}

func (exec *kubernetesExecutable) OnUpdate(f func(interface{}, interface{})) {
	exec.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: f,
	})
}
func (exec *kubernetesExecutable) OnDelete(f func(interface{})) {
	exec.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: f,
	})
}
