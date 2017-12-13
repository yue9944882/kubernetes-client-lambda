package lambda

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	api_v1 "k8s.io/api/core/v1"
)

func TestSimpleMock(t *testing.T) {
	ns := "test"
	mockPod := Pod.Mock()
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
