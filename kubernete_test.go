package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
)

func TestKubernetesSimple(t *testing.T) {
	kcl := OutOfClusterDefault()
	err := kcl.Type(Pod).InNamespace("default").Each(func(pod *corev1.Pod) {})
	assert.NoError(t, err, "some error")
}
