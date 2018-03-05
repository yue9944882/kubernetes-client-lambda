package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	kcl "github.com/yue9944882/kubernetes-client-lambda"
)

func init() {
	kcl.OutOfClusterDefault()
}

func TestSimpleClientSetDiscovery(t *testing.T) {
	for _, resource := range kcl.GetResources() {
		apiResource := kcl.GetResouceIndexerInstance().GetAPIResource(resource)
		assert.NotEmpty(t, apiResource, "resource %s index missing", resource)
	}
}
