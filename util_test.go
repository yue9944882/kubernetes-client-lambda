package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
	api_v1 "k8s.io/api/core/v1"
)

type simpleRs struct {
	Name string
}

func TestNameExtracting(t *testing.T) {
	srs := &simpleRs{
		Name: "foo",
	}
	assert.Equal(t, "foo", getNameOfResource(srs), "Name extrating failed")
}

func TestNamespaced(t *testing.T) {
	var ns api_v1.Namespace
	ok := isNamedspaced(&ns)
	assert.Equal(t, true, ok, "ns has no namespace field")
}
