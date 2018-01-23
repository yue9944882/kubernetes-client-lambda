package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	kubernetes "github.com/yue9944882/kubernetes-client-lambda"
	api_ext_v1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestIngressperation(t *testing.T) {
	ns := "default"

	success, err := kubernetes.OutOfCluster(restconfig).Type(kubernetes.Ingress).InNamespace(ns).Add(
		func() *api_ext_v1.Ingress {
			var ing api_ext_v1.Ingress
			ing.Name = "foo"
			ing.Namespace = ns
			ing.Spec.Rules = []api_ext_v1.IngressRule{
				api_ext_v1.IngressRule{
					Host: "xxx.com",
					IngressRuleValue: api_ext_v1.IngressRuleValue{
						HTTP: &api_ext_v1.HTTPIngressRuleValue{
							Paths: []api_ext_v1.HTTPIngressPath{
								api_ext_v1.HTTPIngressPath{
									Path: "/",
									Backend: api_ext_v1.IngressBackend{
										ServiceName: "foo",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			}
			return &ing
		},
	).CreateIfNotExist()
	assert.Equal(t, true, success, "failed to create ing")
	assert.NoError(t, err, "some error happened")
	success, err = kubernetes.OutOfCluster(restconfig).Type(kubernetes.Ingress).InNamespace(ns).Grep(func(ing *api_ext_v1.Ingress) bool {
		return ing.Name == "foo"
	}).Map(func(ing *api_ext_v1.Ingress) *api_ext_v1.Ingress {
		if ing.Labels == nil {
			ing.Labels = make(map[string]string)
		}
		ing.Labels["test"] = "yes"
		return ing
	}).UpdateIfExist()
	assert.Equal(t, true, success, "failed to update ing")
	assert.NoError(t, err, "some error happened")
	success, err = kubernetes.OutOfCluster(restconfig).Type(kubernetes.Ingress).InNamespace(ns).Grep(func(ing *api_ext_v1.Ingress) bool {
		return ing.Name == "foo"
	}).Delete()
	assert.Equal(t, true, success, "failed to delete ing")
	assert.NoError(t, err, "some error happened")
}
