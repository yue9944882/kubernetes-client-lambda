package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	kcl "github.com/yue9944882/kubernetes-client-lambda"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSimpleConfigMapManipulation(t *testing.T) {
	testFunc := func(kclInterface kcl.KubernetesClientLambda) {
		testConfigMapName := "test-xyz"
		created, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NamePrefix("test-").
			Add(
				func() *corev1.ConfigMap {
					cm := &corev1.ConfigMap{}
					cm.Name = testConfigMapName
					cm.Kind = "ConfigMap"
					cm.Namespace = metav1.NamespaceDefault
					return cm
				},
			).Create()
		assert.Equal(t, true, created, "not created")
		assert.NoError(t, err, "some error")
		notempty, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NotEmpty()
		assert.Equal(t, true, notempty, "shouldn't be empty")
		assert.NoError(t, err, "some error")
		deleted, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NameEqual(testConfigMapName).
			Delete()
		assert.Equal(t, true, deleted, "deletion failed")
		assert.NoError(t, err, "some error")
	}
	testFunc(kcl.OutOfClusterDefault())
	testFunc(kcl.Mock())
}

func TestCreateIfNotExist(t *testing.T) {
	testFunc := func(kclInterface kcl.KubernetesClientLambda) {
		testConfigMapName := "test-abc"
		created, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NamePrefix("test-").
			Add(
				func() *corev1.ConfigMap {
					cm := &corev1.ConfigMap{}
					cm.Name = testConfigMapName
					cm.Kind = "ConfigMap"
					cm.Namespace = metav1.NamespaceDefault
					return cm
				},
			).Create()
		assert.Equal(t, true, created, "not created")
		assert.NoError(t, err, "some error")
		time.Sleep(time.Second)

		created, existed, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NameEqual(testConfigMapName).
			Add(
				func() *corev1.ConfigMap {
					cm := &corev1.ConfigMap{}
					cm.Name = testConfigMapName
					cm.Kind = "ConfigMap"
					cm.Namespace = metav1.NamespaceDefault
					return cm
				},
			).CreateIfNotExist()
		assert.Equal(t, true, created, "not created")
		assert.Equal(t, true, existed, "not created")
		assert.NoError(t, err, "some error")
		deleted, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NameEqual(testConfigMapName).
			Delete()
		assert.Equal(t, true, deleted, "deletion failed")
		assert.NoError(t, err, "some error")
	}
	testFunc(kcl.OutOfClusterDefault())
	testFunc(kcl.Mock())
}

func TestUpdateIfExist(t *testing.T) {
	testFunc := func(kclInterface kcl.KubernetesClientLambda) {
		testConfigMapName := "test-abc"
		created, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NameEqual("test-abc").
			Add(
				func() *corev1.ConfigMap {
					cm := &corev1.ConfigMap{}
					cm.Name = testConfigMapName
					cm.Kind = "ConfigMap"
					cm.Namespace = metav1.NamespaceDefault
					return cm
				},
			).Create()
		assert.Equal(t, true, created, "not created")
		assert.NoError(t, err, "some error")
		time.Sleep(time.Second)

		updated, existed, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NameEqual(testConfigMapName).
			Iter(
				func(cm *corev1.ConfigMap) {
					fmt.Println("=====")
					cm.Data = map[string]string{
						"key": "value",
					}
				},
			).UpdateIfExist()
		assert.Equal(t, true, updated, "not updated")
		assert.Equal(t, true, existed, "not existed")
		assert.NoError(t, err, "some error")

		deleted, err := kclInterface.Type(kcl.ConfigMap).
			InNamespace(metav1.NamespaceDefault).
			NameEqual(testConfigMapName).
			Delete()
		assert.Equal(t, true, deleted, "deletion failed")
		assert.NoError(t, err, "some error")
	}
	testFunc(kcl.OutOfClusterDefault())
	// testFunc(kcl.Mock())
}
