package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	OutOfClusterDefault()
}

func TestSimpleClientSetDiscovery(t *testing.T) {
	for _, resource := range GetResources() {
		_, exist := indexerInstance.(*resourceIndexerImpl).store[resource]
		assert.NotEmpty(t, exist, "resource %s index missing", resource)
	}
}
