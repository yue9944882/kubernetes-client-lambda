package lambda

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	api_v1 "k8s.io/api/core/v1"
)

func TestSimpleMock(t *testing.T) {
	ns := "test"
	mockPod := Pod.Mock(true)
	var pod api_v1.Pod
	name := "pod"
	pod.Namespace = ns
	// Add 10 pods
	for i := 0; i < 10; i++ {
		mockPod.InNamespace(ns).Add(func() *api_v1.Pod {
			pod.Name = name + strconv.Itoa(i)
			return &pod
		}).Create()
	}
	cnt := 0
	mockPod.InNamespace(ns).Each(func(pod *api_v1.Pod) {
		if pod != nil {
			cnt++
		}
	})
	assert.Equal(t, 10, cnt, "pod add failed")
}

func TestNamespaceMock(t *testing.T) {
	success, err := Namespace.Mock(false).All().Add(func() *api_v1.Namespace {
		var ns api_v1.Namespace
		ns.Name = "testns"
		return &ns
	}).CreateIfNotExist()

	assert.Equal(t, true, success, "create success")
	assert.NoError(t, err, "creation failure")

	count := 0
	Namespace.Mock(false).All().Each(func(ns *api_v1.Namespace) {
		if ns != nil {
			count++
		}
	})
	assert.Equal(t, 1, count, "namespace not created")
}
