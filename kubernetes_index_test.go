package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleClientSetDiscovery(t *testing.T) {
	OutOfClusterDefault()
	for _, resource := range GetResources() {
		_, exist := indexerInstance.(*resourceIndexerImpl).store[resource]
		assert.NotEmpty(t, exist, "resource index missing")
	}
}
