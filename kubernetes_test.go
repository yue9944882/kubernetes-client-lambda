package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionParse(t *testing.T) {
	testCases := []struct {
		versionString string
		version       string
		suffix        string
	}{
		{"v1", "1", ""},
		{"v1alpha1", "1", "alpha1"},
		{"v2beta1", "2", "beta1"},
		{"v2beta", "2", ""},
		{"v", "", ""},
		{"v1ss", "1", ""},
	}
	for _, testCase := range testCases {
		v := Version(testCase.versionString)
		assert.Equal(t, testCase.version, v.GetNumericVersion(), "version wrong")
		assert.Equal(t, testCase.suffix, v.GetSuffix(), "suffix wrong")
	}
}
