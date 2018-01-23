package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	api_v1 "k8s.io/api/core/v1"
)

func TestJudgeResourceFromLambda(t *testing.T) {
	producer := func() *api_v1.Pod { return nil }
	rs, err := getResourceTypeOfProducer(producer)
	assert.NoError(t, err, "some error")
	assert.Equal(t, Pod, rs, "wrong type")
	predicate := func(*api_v1.Pod) bool { return true }
	rs, err = getResourceTypeOfPredicate(predicate)
	assert.NoError(t, err, "some error")
	assert.Equal(t, Pod, rs, "wrong type")
	consumer := func(*api_v1.Pod) *api_v1.Pod { return nil }
	rs, err = getResourceTypeOfConsumer(consumer)
	assert.NoError(t, err, "some error")
	assert.Equal(t, Pod, rs, "wrong type")
	function := func(*api_v1.Pod) {}
	rs, err = getResourceTypeOfFunction(function)
	assert.NoError(t, err, "some error")
	assert.Equal(t, Pod, rs, "wrong type")
}
