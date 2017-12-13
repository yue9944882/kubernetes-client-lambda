package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
