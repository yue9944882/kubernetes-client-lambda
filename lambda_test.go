package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type simple struct {
	v int
}

func simpleLambda() *Lambda {
	ch := make(chan interface{})
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
