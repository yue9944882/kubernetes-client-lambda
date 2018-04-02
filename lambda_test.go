package lambda

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLatestObject(t *testing.T) {
	mock := Mock()
	_, err := mock.Type(ConfigMap).
		InNamespace("foons").
		Add(func() *corev1.ConfigMap {
			cm := &corev1.ConfigMap{}
			cm.Name = "testcm1"
			cm.Namespace = "foons"
			cm.CreationTimestamp = metav1.Time{time.Now()}
			return cm
		}).
		Add(func() *corev1.ConfigMap {
			cm := &corev1.ConfigMap{}
			cm.Name = "testcm2"
			cm.Namespace = "foons"
			time.Sleep(time.Second)
			cm.CreationTimestamp = metav1.Time{time.Now()}
			return cm
		}).Create()
	assert.NoError(t, err, "creation failed")
	mock.Type(ConfigMap).
		InNamespace("foons").
		List().
		LatestCreated().
		Each(func(cm *corev1.ConfigMap) {
			assert.Equal(t, "testcm2", cm.Name, "name wrong")
		})
}
