package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	kubernetes "github.com/yue9944882/kubernetes-client-lambda"
	api_v1 "k8s.io/api/core/v1"
)

func TestNamespaceOperation(t *testing.T) {
	ok, err := kubernetes.OutOfCluster(restconfig).Type(kubernetes.Namespace).All().Any(
		func(ns *api_v1.Namespace) bool {
			return ns.Name == "default"
		},
	)
	assert.Equal(t, true, ok, "namespace doesnt' exist")
	assert.NoError(t, err, "some error")

	success, err := kubernetes.OutOfCluster(restconfig).Type(kubernetes.Namespace).All().Add(func() *api_v1.Namespace {
		var ns api_v1.Namespace
		ns.APIVersion = "v1"
		ns.Name = "new"
		return &ns
	}).CreateIfNotExist()
	assert.Equal(t, true, success, "ns creation failure")
	assert.NoError(t, err, "Some error")

}
