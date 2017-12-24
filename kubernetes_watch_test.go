package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListerWatcher(t *testing.T) {
	clientset := fake.NewSimpleClientset(getAllRuntimeObject()...)
	var pod api_v1.Pod
	pod.Name = "foo"
	pod.Namespace = "default"
	op := clientset.CoreV1().Pods("default")
	resource, err := callCreateInterface(op, &pod)
	assert.NoError(t, err, "some error in create interface")
	assert.NotEmpty(t, resource, "resources create empty")
	op2, err := opInterface(Pod, "default", clientset)
	assert.NoError(t, err, "some error")
	lw, err := getListWatch(op2)
	assert.NoError(t, err, "some error")
	list, err := lw.List(meta_v1.ListOptions{})
	assert.NoError(t, err, "some error")
	podlist, ok := list.(*api_v1.PodList)
	assert.Equal(t, true, ok, "type assertion failure")
	assert.Equal(t, 1, len(podlist.Items), "creation failure")
}

func TestWatchManagerInstance(t *testing.T) {
	instance := getWatchManager()
	assert.NotNil(t, instance, "instance nil")
}

func TestWatchManagerFunc(t *testing.T) {
	f := func() {}
	entry := &watchEntry{
		stopCh:         nil,
		watchFunctions: []watchFunction{},
	}
	entry.AddFunc(watch.Added, f)
	assert.Equal(t, 1, len(entry.watchFunctions), "Failed to add function")
	entry.DelFunc(watch.Added, f)
	assert.Equal(t, 0, len(entry.watchFunctions), "Failed to delete function")
}

func TestWatchManagerRegister(t *testing.T) {
	f := func() {}
	e := getWatchManager().registerFunc(Pod, "default", watch.Added, f)
	assert.Equal(t, e, getWatchManager().getEntry(Pod, "default"), "Entry mismatch")
	assert.Equal(t, 1, len(e.watchFunctions), "Failed to register function")
	go func() {
		<-e.stopCh
	}()
	getWatchManager().unregisterFunc(Pod, "default", watch.Added, f)
	assert.Equal(t, 0, len(e.watchFunctions), "Failed to unregister function")
}
