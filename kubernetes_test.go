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
	assert.NotNil(t, resources, "resources list empty")
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

// ByPass following test there's underlying BUGs in client-go
// See PR: #57504

/*
func TestWatchCall(t *testing.T) {
	count := 0
	Pod.Mock().WatchNamespace("default").Register(watch.Added, func(pod *api_v1.Pod) {
		count++
	})
	ok, err := Pod.Mock().InNamespace("default").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "test"
		pod.Namespace = "default"
		return &pod
	}).CreateIfNotExist()
	assert.Equal(t, 1, count, "watch call missed")
	assert.NoError(t, err, "some error")
	assert.Equal(t, true, ok, "create failed")
}

func TestFakeWatchCall(t *testing.T) {
	clientset := fake.NewSimpleClientset(getAllRuntimeObject()...)
	watcher, err := clientset.CoreV1().Pods("default").Watch(meta_v1.ListOptions{})
	var pod api_v1.Pod
	count := 0
	go func() {
		evCh := watcher.ResultChan()
		for {
			select {
			case <-evCh:
				count++
			}
		}
	}()
	time.Sleep(time.Second * 1)
	_, err = clientset.CoreV1().Pods("default").Create(&pod)
	assert.NoError(t, err, "some error")
	pod.Name = "xxx"
	_, err = clientset.CoreV1().Pods("default").Create(&pod)
	assert.NoError(t, err, "some error")
	pod.Name = "xxx1"
	_, err = clientset.CoreV1().Pods("default").Create(&pod)
	assert.NoError(t, err, "some error")
	assert.Equal(t, err, 3, "watch event missed")
}
*/

func TestKubernetesInterface(t *testing.T) {
	clientset := fake.NewSimpleClientset(getAllRuntimeObject()...)
	resources := []Resource{
		Namespace,
		Node,
		StorageClass,
		Pod,
		ReplicaSet,
		ReplicationController,
		Deployment,
		ConfigMap,
		Ingress,
		Service,
		Endpoint,
		Secret,
		DaemonSet,
		StatefulSet,
	}

	for _, resource := range resources {
		obj := resource.GetObject()
		assert.NotNil(t, obj, "Failed to get object from resource")
		rsname := resource.GeResourcetName()
		assert.NotEqual(t, "", rsname, "Failed to get resource name")
		op, err := opInterface(resource, "default", clientset)
		assert.NotNilf(t, op, "Failed to get op interface from %s", resource.String())
		assert.NoError(t, err, "Some error in op interface")
		api, err := apiInterface(resource, clientset)
		assert.NotNil(t, api, "Failed to get api interface")
		assert.NoError(t, err, "Some error in api interface")
	}
}
