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
	any := lambda.Any(func(s *simple) bool {
		return s.v == 50
	})
	assert.Equal(t, true, any, "given v doesn't exists")
	assert.NoError(t, lambda.Error, "fail to apply lambda")
}

func TestSimpleLambdaEvery(t *testing.T) {
	lambda := simpleLambda()
	every := lambda.Every(func(s *simple) bool {
		return s.v >= 0
	})
	assert.Equal(t, true, every, "given v tainted")
	assert.NoError(t, lambda.Error, "fail to apply lambda")
}

func TestSimpleLambdaMap(t *testing.T) {
	lambda := simpleLambda()
	mapped := lambda.Map(func(s *simple) *simple {
		m := &simple{}
		return m
	}).Every(func(s *simple) bool {
		return s.v == 0
	})
	assert.Equal(t, true, mapped, "given v mapping failure")
	assert.NoErrorf(t, lambda.Error, "fail to apply lambda")
}

func TestSimpleLambdaFirst(t *testing.T) {
	lambda := simpleLambda()
	ok := lambda.First(func(s *simple) bool {
		return s.v > 0 && s.v%33 == 0
	}).Every(func(s *simple) bool {
		return s.v == 33
	})
	assert.Equal(t, true, ok, "given v first failure")
	assert.NoErrorf(t, lambda.Error, "fail to apply lambda")
}

func TestSimpleLambdaGrep(t *testing.T) {
	lambda := simpleLambda()
	ok := lambda.Grep(func(s *simple) bool {
		return s.v%3 == 0
	}).Every(func(s *simple) bool {
		return s.v%3 == 0
	})
	assert.Equal(t, true, ok, "given v grep failure")
	assert.NoErrorf(t, lambda.Error, "fail to apply lambda")
}

func TestSimpleLambdaAdd(t *testing.T) {
	lambda := simpleLambda()
	cnt := 0
	ok := lambda.Add(func() *simple {
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
	assert.NoErrorf(t, lambda.Error, "fail to apply lambda")
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
	Pod.Mock(true).InNamespace("test").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "test1"
		pod.Annotations = make(map[string]string)
		pod.Annotations["a1"] = "v1"
		pod.Labels = make(map[string]string)
		pod.Labels["l1"] = "b1"
		return &pod
	}).Create()
	exist, err := Pod.Mock(true).InNamespace("test").NameEqual("test1").NotEmpty()
	assert.Equal(t, true, exist, "name snippet failure")
	assert.NoError(t, err, "some error")
	exist, err = Pod.Mock(true).InNamespace("test").HasAnnotation("a1", "v1").NotEmpty()
	assert.Equal(t, true, exist, "annotation snippet failure")
	assert.NoError(t, err, "some error")
	exist, err = Pod.Mock(true).InNamespace("test").HasLabel("l1", "b1").NotEmpty()
	assert.Equal(t, true, exist, "label snippet failure")
	assert.NoError(t, err, "some error")
	exist, err = Pod.Mock(true).InNamespace("test").HasAnnotationKey("a1").NotEmpty()
	assert.Equal(t, true, exist, "annotation key snippet failure")
	assert.NoError(t, err, "some error")
	exist, err = Pod.Mock(true).InNamespace("test").HasAnnotationKey("a1").NotEmpty()
	assert.Equal(t, true, exist, "annotation key snippet failure")
	assert.NoError(t, err, "some error")
}

func TestDummyLamba(t *testing.T) {
	Pod.Mock(true).InNamespace("test").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "default"
		pod.Annotations = make(map[string]string)
		pod.Annotations["a1"] = "v1"
		pod.Labels = make(map[string]string)
		pod.Labels["l1"] = "b1"
		return &pod
	}).Create()
	exist, err := Pod.Mock(true).InNamespace("test").NameEqual("test1").NotEmpty()
	assert.Equal(t, true, exist, "name snippet failure")
	assert.NoError(t, err, "some error")

}

func TestIterLambda(t *testing.T) {
	Pod.Mock(true).InNamespace("test").Add(func() *api_v1.Pod {
		var pod api_v1.Pod
		pod.Name = "default"
		pod.Annotations = make(map[string]string)
		pod.Annotations["a1"] = "v1"
		pod.Labels = make(map[string]string)
		pod.Labels["l1"] = "b1"
		return &pod
	}).Create()
	count := 0
	exist, err := Pod.Mock(true).InNamespace("test").NameEqual("default").Iter(func(pod *api_v1.Pod) {
		if pod != nil {
			count++
		}
	}).NotEmpty()
	assert.Equal(t, true, exist, "name snippet failure")
	assert.Equal(t, 1, count, "count mismatch")
	assert.NoError(t, err, "some error")
}
