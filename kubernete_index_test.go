package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstLetterCaptalization(t *testing.T) {
	assert.Equal(t, "Abc", capitalizeFirstLetter("abc"), "nope")
	assert.Equal(t, "Abc", capitalizeFirstLetter("Abc"), "nope")
	assert.Equal(t, "", capitalizeFirstLetter(""), "nope")
	assert.Equal(t, "A", capitalizeFirstLetter("a"), "nope")
}

func TestIndexerInitialization(t *testing.T) {
	initIndexer()
}
