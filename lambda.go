package lambda

import (
	"reflect"
)

type Lambda struct {
	op    KubernetesOperation
	val   <-chan interface{}
	Error error
}

func (lambda *Lambda) Every(fn interface{}) bool {
	for item := range lambda.val {
		ret := reflect.ValueOf(fn).Call([]reflect.Value{
			reflect.ValueOf(item),
		})
		if !ret[0].Bool() {
			return false
		}
	}
	return true
}

func (lambda *Lambda) Any(fn interface{}) bool {
	for item := range lambda.val {
		ret := reflect.ValueOf(fn).Call([]reflect.Value{
			reflect.ValueOf(item),
		})
		if ret[0].Bool() {
			return true
		}
	}
	return false
}

func (lambda *Lambda) Each(fn interface{}) {
	for item := range lambda.val {
		reflect.ValueOf(fn).Call([]reflect.Value{
			reflect.ValueOf(item),
		})
	}
}

func (lambda *Lambda) First(fn interface{}) *Lambda {
	ch := make(chan interface{})
	l := &Lambda{
		op:  lambda.op,
		val: ch,
	}
	go func() {
		defer close(ch)
		for item := range lambda.val {
			ret := reflect.ValueOf(fn).Call([]reflect.Value{
				reflect.ValueOf(item),
			})
			if ret[0].Bool() {
				ch <- item
				break
			}
		}
	}()
	return l
}

func (lambda *Lambda) Grep(fn interface{}) *Lambda {
	ch := make(chan interface{})
	l := &Lambda{
		op:  lambda.op,
		val: ch,
	}
	go func() {
		defer close(ch)
		for item := range lambda.val {
			ret := reflect.ValueOf(fn).Call([]reflect.Value{
				reflect.ValueOf(item),
			})
			if ret[0].Bool() {
				ch <- item
			}
		}
	}()
	return l
}

func (lambda *Lambda) Map(fn interface{}) *Lambda {
	ch := make(chan interface{})
	l := &Lambda{
		op:  lambda.op,
		val: ch,
	}
	go func() {
		defer close(ch)
		for item := range lambda.val {
			ret := reflect.ValueOf(fn).Call([]reflect.Value{
				reflect.ValueOf(item),
			})
			ch <- ret[0].Interface()
		}
	}()
	return l
}

func (lambda *Lambda) Add(fn interface{}) *Lambda {
	ch := make(chan interface{})
	l := &Lambda{
		op:  lambda.op,
		val: ch,
	}
	go func() {
		defer close(ch)
		for item := range lambda.val {
			ch <- item
		}
		ret := reflect.ValueOf(fn).Call([]reflect.Value{})
		ch <- ret[0].Interface()
	}()
	return l
}
