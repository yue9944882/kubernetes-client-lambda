package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	OutOfClusterDefault()
}

func TestResourceUpdate(t *testing.T) {
	fakePool := newFakeClientPool()
	i, err := fakePool.ClientForGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    string(ConfigMap),
	})
	cm := &v1.ConfigMap{}
	cm.Kind = string(ConfigMap)
	cm.Name = "test"
	cm.Namespace = "test"
	cm.Data = map[string]string{"a": "b"}
	assert.NoError(t, err, "some error")
	err = create(i, ConfigMap, cm)
	assert.NoError(t, err, "some error")
	err = update(i, ConfigMap, cm)
	assert.NoError(t, err, "some error")
	err = delete(i, ConfigMap, cm)
	assert.NoError(t, err, "some error")
}
