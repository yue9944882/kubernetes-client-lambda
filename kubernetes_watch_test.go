package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestWatchManager(t *testing.T) {
}

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
