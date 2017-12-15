package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	kubernetes "github.com/yue9944882/kubernetes-client-lambda"
	api_v1 "k8s.io/api/core/v1"
)

func TestPodOperation(t *testing.T) {
	ns := "default"

	success, err := kubernetes.Pod.OutOfCluster(restconfig).InNamespace(ns).Add(
		func() *api_v1.Pod {
			var pod api_v1.Pod
			pod.APIVersion = "v1"
			pod.Kind = "Pod"
			pod.Name = "test1"
			pod.Namespace = ns
			pod.Spec.Containers = []api_v1.Container{
				api_v1.Container{
					Name:  "test-container",
					Image: "alpine:3.6",
				},
			}
			return &pod
		},
	).CreateIfNotExist()
	assert.Equal(t, true, success, "failed to create pod")
	assert.NoError(t, err, "some error happened")
	success, err = kubernetes.Pod.OutOfCluster(restconfig).InNamespace(ns).Grep(func(pod *api_v1.Pod) bool {
		return pod.Name == "test1"
	}).Map(func(pod *api_v1.Pod) *api_v1.Pod {
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		pod.Labels["test"] = "yes"
		return pod
	}).Update()
	assert.Equal(t, true, success, "failed to update pod")
	assert.NoError(t, err, "some error happened")
	success, err = kubernetes.Pod.OutOfCluster(restconfig).InNamespace(ns).Grep(func(pod *api_v1.Pod) bool {
		return pod.Name == "test1"
	}).Delete()
	assert.Equal(t, true, success, "failed to delete pod")
	assert.NoError(t, err, "some error happened")
}
