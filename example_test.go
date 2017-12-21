package lambda_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	kubernetes "github.com/yue9944882/kubernetes-client-lambda"
	api_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func ExampleLambda_OutOfCluster() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	count := 0
	kubernetes.Pod.OutOfCluster(config).InNamespace("devops").NameEqual("test-pod").Each(
		func(pod *api_v1.Pod) {
			count++
		},
	)
	fmt.Println(count)
}

func ExampleLambda_InCluster() {
	count := 0
	kubernetes.Pod.InCluster().InNamespace("devops").Each(func(pod *api_v1.Pod) {
		count++
		fmt.Println(pod.Name)
	})
	fmt.Println(count)
}

func ExampleLambda_Watch() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	count := 0
	kubernetes.Pod.OutOfCluster(config).InNamespace("devops").NameEqual("test-pod").Each(
		func(pod *api_v1.Pod) {
			count++
		},
	)
	fmt.Println(count)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
