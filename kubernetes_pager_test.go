package lambda

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestKubernetesPager(t *testing.T) {
	clientset := fake.NewSimpleClientset(getAllRuntimeObject()...)
	count := 0
	for i := 0; i < 50; i++ {
		var cm api_v1.ConfigMap
		cm.Name = "cm" + strconv.Itoa(count)
		cm.Data = make(map[string]string)
		cm.Data["k"] = "v"
		clientset.CoreV1().ConfigMaps("default").Create(&cm)
		count++
	}
	pgr := &ListPager{
		PageSize: 5,
		PageFn:   clientset.CoreV1().ConfigMaps("default").List,
	}
	ch, err := pgr.List(meta_v1.ListOptions{})
	assert.NoError(t, err, "Some error")
	assert.NotNil(t, ch, "Nil list channel")
	chcount := 0
	for _ = range ch {
		chcount++
	}
	assert.Equal(t, 50, chcount, "pod count mismatch")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
