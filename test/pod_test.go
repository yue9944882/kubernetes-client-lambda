package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	kubernetes "github.com/yue9944882/kubernetes-client-lambda"
	api_v1 "k8s.io/api/core/v1"
)

func TestPodCreate(t *testing.T) {
	ns := "default"
	success, err := kubernetes.Pod.InCluster().InNamespace(ns).Add(
		func() *api_v1.Pod {
			var pod api_v1.Pod
			pod.APIVersion = "v1"
			pod.Kind = "Pod"
			pod.Name = "test1"
			pod.Namespace = ns
			pod.Spec.Containers = []api_v1.Container{
				api_v1.Container{
					Name:  "test_container",
					Image: "alpine:3.6",
				},
			}
			return &pod
		},
	).Create()
	assert.Equal(t, true, success, "failed to create pod")
	assert.NoError(t, err, "some error happened")
}
