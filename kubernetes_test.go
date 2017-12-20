package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	api_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func init() {
}

func TestReflectCall(t *testing.T) {
	clientset := fake.NewSimpleClientset(getAllRuntimeObject()...)
	var pod api_v1.Pod
	pod.Name = "foo"
	pod.Namespace = "default"
	op := clientset.CoreV1().Pods("default")
	resource, err := callCreateInterface(op, &pod)
	assert.NoError(t, err, "some error in create interface")
	assert.NotEmpty(t, resource, "resources create empty")
	resources, err := callListInterface(op)
	assert.NoError(t, err, "some error in list interface")
	assert.NotEmpty(t, resources, "resources list empty")
	pod.Labels = make(map[string]string)
	pod.Labels["test1"] = "v1"
	resource, err = callUpdateInterface(op, &pod)
	assert.NoError(t, err, "some error in update interface")
	assert.Equal(t, "foo", resource.(*api_v1.Pod).Name, "resource not updated")
	assert.Equal(t, "v1", resource.(*api_v1.Pod).Labels["test1"], "labels not updated")
	resource1, err := callGetInterface(op, "foo")
	assert.NoError(t, err, "some error in update interface")
	assert.NotNil(t, resource1, "getting a nil value")
	err = callDeleteInterface(op, "foo")
	assert.NoError(t, err, "some error happened")
}

func TestWatchCall(t *testing.T) {
	Pod.Mock().WatchNamespace("test")
}
