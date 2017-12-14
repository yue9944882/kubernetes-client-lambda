package example

import (
	"fmt"

	kubernetes "github.com/yue9944882/kubernetes-client-lambda"
	api_v1 "k8s.io/api/core/v1"
)

func in_cluster() {
	count := 0
	kubernetes.Pod.InCluster().InNamespace("devops").Each(func(pod *api_v1.Pod) {
		count++
		fmt.Println(pod.Name)
	})
	fmt.Println(count)
}
