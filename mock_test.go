package lambda

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	api_v1 "k8s.io/api/core/v1"
)

func TestSimpleMock(t *testing.T) {
	ns := "test-ns"
	mockPod := Pod.Mock()
	// Add 10 pods
	for i := 0; i < 10; i++ {
		mockPod.InNamespace(ns).Add(func() *api_v1.Pod {
			var pod api_v1.Pod
			name := "pod"
			pod.Namespace = ns
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
	success, err := Namespace.Mock().All().Add(func() *api_v1.Namespace {
		var ns api_v1.Namespace
		ns.Name = "testns"
		return &ns
	}).CreateIfNotExist()

	assert.Equal(t, true, success, "create success")
	assert.NoError(t, err, "creation failure")

	count := 0
	Namespace.Mock().All().Each(func(ns *api_v1.Namespace) {
		if ns != nil {
			count++
		}
	})
	assert.Equal(t, 2, count, "namespace not created")

	deleted, err := Namespace.Mock().All().Grep(func(ns *api_v1.Namespace) bool {
		return ns.Name == "testns"
	}).DeleteIfExist()
	assert.Equal(t, true, deleted, "create success")
	assert.NoError(t, err, "creation failure")
}

func TestMockWatch(t *testing.T) {

}
