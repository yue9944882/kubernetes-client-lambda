package lambda

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/informers"
	"testing"
)

func TestFakeClientPool(t *testing.T) {

	cmIn1 := &corev1.ConfigMap{}
	cmIn1.Name = "foo1"
	cmIn1.Namespace = "foo1"
	cmIn1.Kind = ConfigMap.GetKind()
	cmIn1.APIVersion = ConfigMap.GetAPIVersion()

	fakePool, _ := NewFakes(cmIn1)

	fakeDynamicClient, err := fakePool.ClientForGroupVersionKind(
		GetResouceIndexerInstance().GetGroupVersionKind(ConfigMap),
	)
	assert.NoError(t, err, "some error")
	// Test Get Call
	fakeClient := fakeDynamicClient.Resource(indexerInstance.GetAPIResource(ConfigMap), "foo1")
	cmOut1, err := fakeClient.Get("foo1", metav1.GetOptions{})
	assert.NoError(t, err, "some error")
	usCm1, err := castObjectToUnstructured(cmIn1)
	assert.NoError(t, err, "some error")
	assert.Equal(t, usCm1, cmOut1, "object not equal")

	// Test List Call
	cms, err := fakeClient.List(metav1.ListOptions{})
	assert.NoError(t, err, "some error")
	assert.Equal(t, *usCm1, cms.(*unstructured.UnstructuredList).Items[0], "object not equal")

	cmIn2 := &corev1.ConfigMap{}
	cmIn2.Name = "foo2"
	cmIn2.Namespace = "foo1"
	cmIn2.Kind = ConfigMap.GetKind()
	cmIn2.APIVersion = ConfigMap.GetAPIVersion()

	cm1Watch, err := fakeClient.Watch(metav1.ListOptions{})
	wch1 := cm1Watch.ResultChan()

	// Test Add Call
	usCm2, err := castObjectToUnstructured(cmIn2)
	assert.NoError(t, err, "some error")
	_, err = fakeClient.Create(usCm2)
	assert.NoError(t, err, "some error")

	// Test Get Call
	cmOut2, err := fakeClient.Get("foo2", metav1.GetOptions{})
	assert.NoError(t, err, "some error")
	assert.NoError(t, err, "some error")
	assert.Equal(t, usCm2, cmOut2, "object not equal")

	// Test Watch Call
	event := <-wch1
	assert.Equal(t, event.Object, cmIn2, "watched object not equal")
	assert.NoError(t, err, "some error")
}

func TestFakeClientInformer(t *testing.T) {
	cmIn1 := &corev1.ConfigMap{}
	cmIn1.Name = "foo1"
	cmIn1.Namespace = "foo1"
	cmIn1.Kind = ConfigMap.GetKind()
	cmIn1.APIVersion = ConfigMap.GetAPIVersion()
	_, fakeClient := NewFakes(cmIn1)
	fakeFactory := informers.NewSharedInformerFactory(fakeClient, 0)
	fakeFactory.Core().V1().ConfigMaps().Informer()
	fakeFactory.Start(make(chan struct{}))
	fakeFactory.WaitForCacheSync(make(chan struct{}))
	cmOut1, err := fakeFactory.Core().V1().ConfigMaps().Lister().ConfigMaps("foo1").Get("foo1")
	assert.NoError(t, err, "some error")
	assert.Equal(t, cmOut1, cmIn1, "configmap not equal")
	cmIn2 := &corev1.ConfigMap{}
	cmIn2.Name = "foo2"
	cmIn2.Namespace = "foo1"
	cmIn2.Kind = ConfigMap.GetKind()
	cmIn2.APIVersion = ConfigMap.GetAPIVersion()
	_, err = fakeClient.Core().ConfigMaps("foo1").Create(cmIn2)
	fakeFactory.WaitForCacheSync(make(chan struct{}))
	assert.NoError(t, err, "some error")
	cmOut2, err := fakeFactory.Core().V1().ConfigMaps().Lister().ConfigMaps("foo1").Get("foo2")
	assert.Equal(t, cmIn2, cmOut2, "configmap not equal")
	assert.NoError(t, err, "some error")
}
