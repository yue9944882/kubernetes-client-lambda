package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
	api_v1 "k8s.io/api/core/v1"
)

type simple struct {
	v int
}

func simpleLambda() *Lambda {
	ch := make(chan kubernetesResource)
	lambda := &Lambda{
		val: ch,
	}
	go func() {
		for i := 0; i < 100; i++ {
			ch <- &simple{
				v: i,
			}
		}
		close(ch)
	}()
	return lambda
}

func TestSimpleLambdaAny(t *testing.T) {
	lambda := simpleLambda()
	any, err := lambda.Any(func(s *simple) bool {
		return s.v == 50
	})
	assert.Equal(t, true, any, "given v doesn't exists")
	assert.NoError(t, err, "fail to apply lambda")
}

func TestSimpleLambdaEvery(t *testing.T) {
	lambda := simpleLambda()
	every, err := lambda.Every(func(s *simple) bool {
		return s.v >= 0
	})
	assert.Equal(t, true, every, "given v tainted")
	assert.NoError(t, err, "fail to apply lambda")
}

func TestSimpleLambdaMap(t *testing.T) {
	lambda := simpleLambda()
	mapped, err := lambda.Map(func(s *simple) *simple {
		m := &simple{}
		return m
	}).Every(func(s *simple) bool {
		return s.v == 0
	})
	assert.Equal(t, true, mapped, "given v mapping failure")
	assert.NoErrorf(t, err, "fail to apply lambda")
}

func TestSimpleLambdaFirst(t *testing.T) {
	lambda := simpleLambda()
	ok, err := lambda.First(func(s *simple) bool {
		return s.v > 0 && s.v%33 == 0
	}).Every(func(s *simple) bool {
		return s.v == 33
	})
	assert.Equal(t, true, ok, "given v first failure")
	assert.NoErrorf(t, err, "fail to apply lambda")
}

func TestSimpleLambdaGrep(t *testing.T) {
	lambda := simpleLambda()
	ok, err := lambda.Grep(func(s *simple) bool {
		return s.v%3 == 0
	}).Every(func(s *simple) bool {
		return s.v%3 == 0
	})
	assert.Equal(t, true, ok, "given v grep failure")
	assert.NoErrorf(t, err, "fail to apply lambda")
}

func TestSimpleLambdaAdd(t *testing.T) {
	lambda := simpleLambda()
	cnt := 0
	ok, err := lambda.Add(func() *simple {
		return &simple{
			v: -1,
		}
	}).Any(func(s *simple) bool {
		if s.v < 0 {
			cnt++
			return true
		}
		return false
	})
	assert.Equal(t, true, ok, "given v prepend failure")
	assert.NoErrorf(t, err, "fail to apply lambda")
	assert.Equal(t, 1, cnt, "more than element added")
}

func TestSimpleLambdaEach(t *testing.T) {
	lambda := simpleLambda()
	cnt := 0
	lambda.Each(func(s *simple) {
		if s != nil {
			cnt++
		}
	})
	assert.Equal(t, 100, cnt, "more element iterated")
}

func TestSimpleKRName(t *testing.T) {
	Pod.Mock().InNamespace("test").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "test1"
		pod.Annotations = make(map[string]string)
		pod.Annotations["a1"] = "v1"
		pod.Labels = make(map[string]string)
		pod.Labels["l1"] = "b1"
		return &pod
	}).Create()
	exist, err := Pod.Mock().InNamespace("test").NameEqual("test1").NotEmpty()
	assert.Equal(t, true, exist, "name snippet failure")
	assert.NoError(t, err, "some error")
	exist, err = Pod.Mock().InNamespace("test").HasAnnotation("a1", "v1").NotEmpty()
	assert.Equal(t, true, exist, "annotation snippet failure")
	assert.NoError(t, err, "some error")
	exist, err = Pod.Mock().InNamespace("test").HasLabel("l1", "b1").NotEmpty()
	assert.Equal(t, true, exist, "label snippet failure")
	assert.NoError(t, err, "some error")
	exist, err = Pod.Mock().InNamespace("test").HasAnnotationKey("a1").NotEmpty()
	assert.Equal(t, true, exist, "annotation key snippet failure")
	assert.NoError(t, err, "some error")
	exist, err = Pod.Mock().InNamespace("test").HasAnnotationKey("a1").NotEmpty()
	assert.Equal(t, true, exist, "annotation key snippet failure")
	assert.NoError(t, err, "some error")
}

func TestDummyLamba(t *testing.T) {
	Pod.Mock().InNamespace("test").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "default"
		pod.Annotations = make(map[string]string)
		pod.Annotations["a1"] = "v1"
		pod.Labels = make(map[string]string)
		pod.Labels["l1"] = "b1"
		return &pod
	}).Create()
	exist, err := Pod.Mock().InNamespace("test").NameEqual("test1").NotEmpty()
	assert.Equal(t, true, exist, "name snippet failure")
	assert.NoError(t, err, "some error")

}

func TestIterLambda(t *testing.T) {
	Pod.Mock().InNamespace("test").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "default"
		pod.Annotations = make(map[string]string)
		pod.Annotations["a1"] = "v1"
		pod.Labels = make(map[string]string)
		pod.Labels["l1"] = "b1"
		return &pod
	}).Create()
	count := 0
	exist, err := Pod.Mock().InNamespace("test").NameEqual("default").Iter(func(pod *api_v1.Pod) {
		if pod != nil {
			count++
		}
	}).NotEmpty()
	assert.Equal(t, true, exist, "name snippet failure")
	assert.Equal(t, 1, count, "count mismatch")
	assert.NoError(t, err, "some error")
}

func TestKubernetesOperation(t *testing.T) {
	var tmppod *api_v1.Pod
	ok, err := Pod.Mock().InNamespace("test").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "default"
		pod.Annotations = make(map[string]string)
		pod.Annotations["a1"] = "v1"
		pod.Labels = make(map[string]string)
		pod.Labels["l1"] = "b1"
		tmppod = &pod
		return &pod
	}).UpdateOrCreate()
	assert.Equal(t, true, ok, "Failed to create pod")
	assert.Equal(t, "b1", tmppod.Labels["l1"], "Label mismatched")
	assert.NoError(t, err, "Some error")
	ok, err = Pod.Mock().InNamespace("test").NameEqual("default").Map(func(pod *api_v1.Pod) *api_v1.Pod {
		pod.Labels["l2"] = "b2"
		tmppod = pod
		return pod
	}).Update()
	assert.Equal(t, "b2", tmppod.Labels["l2"], "Label mismatched")
	assert.Equal(t, true, ok, "Failed to update pod")
	assert.NoError(t, err, "Some error")
	ok, err = Pod.Mock().InNamespace("test").NameEqual("default").Map(func(pod *api_v1.Pod) *api_v1.Pod {
		pod.Labels["l3"] = "b3"
		tmppod = pod
		return pod
	}).UpdateIfExist()
	assert.Equal(t, "b3", tmppod.Labels["l3"], "Label mismatched")
	assert.Equal(t, true, ok, "Failed to update-if-exist pod")
	assert.NoError(t, err, "Some error")
	ok, err = Pod.Mock().InNamespace("test").NameEqual("default").Delete()
	assert.Equal(t, true, ok, "Failed to delete pod")
	assert.NoError(t, err, "Some error")
}

func TestLambdaCollect(t *testing.T) {
	Pod.ResetMock()
	ok, err := Pod.Mock().InNamespace("test").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "foo"
		return &pod
	}).Create()
	assert.Equal(t, true, ok, "Failed to create pod")
	assert.NoError(t, err, "Some error")
	count := 0
	Pod.Mock().InNamespace("test").Collect().Each(func(pod *api_v1.Pod) {
		count++
		assert.Equal(t, "foo", pod.Name, "Deep copied pod name mismatch")
	})
	assert.Equal(t, 1, count, "Failed to iter over pods")
}
