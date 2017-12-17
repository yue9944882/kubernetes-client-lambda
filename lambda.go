package lambda

import (
	"reflect"
)

type Predicate interface{}
type Consumer interface{}
type Function interface{}
type Producer interface{}

func callPredicate(f interface{}, arg interface{}) bool {
	// TODO: ArgIn check
	ret := reflect.ValueOf(f).Call([]reflect.Value{
		reflect.ValueOf(arg),
	})
	return ret[0].Bool()
}

func callFunction(f interface{}, arg interface{}) {
	reflect.ValueOf(f).Call([]reflect.Value{
		reflect.ValueOf(arg),
	})
}

func callConsumer(f interface{}, arg interface{}) interface{} {
	ret := reflect.ValueOf(f).Call([]reflect.Value{
		reflect.ValueOf(arg),
	})
	return ret[0].Interface()
}

func callProducer(f interface{}) interface{} {
	ret := reflect.ValueOf(f).Call([]reflect.Value{})
	return ret[0].Interface()
}

type Lambda struct {
	op    KubernetesOperation
	val   <-chan kubernetesResource
	Error error
}

func (lambda *Lambda) clone() (*Lambda, chan kubernetesResource) {
	ch := make(chan kubernetesResource)
	l := &Lambda{
		op:  lambda.op,
		val: ch,
	}
	return l, ch
}

//********************************************************
// Lambda using Predicate
//********************************************************

func (lambda *Lambda) First(predicate Predicate) *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		for item := range lambda.val {
			if callPredicate(predicate, item) {
				ch <- item
				break
			}
		}
	}()
	return l
}

func (lambda *Lambda) Grep(predicate Predicate) *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		for item := range lambda.val {
			if callPredicate(predicate, item) {
				ch <- item
			}
		}
	}()
	return l
}

//********************************************************
// Lambda using Consumer
//********************************************************

func (lambda *Lambda) Map(consumer Consumer) *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		for item := range lambda.val {
			if v := callConsumer(consumer, item); v != nil {
				ch <- v
			}
		}
	}()
	return l
}

//********************************************************
// Lambda using Function
//********************************************************

func (lambda *Lambda) Iter(function Function) *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		for item := range lambda.val {
			callFunction(function, item)
			ch <- item
		}
	}()
	return l
}

//********************************************************
// Lambda using Producer
//********************************************************

func (lambda *Lambda) Add(producer Producer) *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		for item := range lambda.val {
			ch <- item
		}
		if v := callProducer(producer); v != nil {
			ch <- v
		}
	}()
	return l
}

//********************************************************
// Snippet Kubernetes Lambda Functions
// Utility Lambda Function
//********************************************************

func (lambda *Lambda) NameEqual(name string) *Lambda {
	nameEqual := func(kr kubernetesResource) bool {
		return reflect.Indirect(reflect.ValueOf(kr)).FieldByName("Name").String() == name
	}
	return lambda.Grep(nameEqual)
}

func (lambda *Lambda) HasAnnotation(key, value string) *Lambda {
	annotationEqual := func(kr kubernetesResource) bool {
		annotations := reflect.Indirect(reflect.ValueOf(kr)).FieldByName("Annotations")
		m, err := annotationMap(annotations.Interface())
		if err != nil {
			return false
		}
		if aValue, exist := m[key]; exist && aValue == value {
			return true
		}
		return false
	}
	return lambda.Grep(annotationEqual)
}

func (lambda *Lambda) HasAnnotationKey(key string) *Lambda {
	annotationKeyEqual := func(kr kubernetesResource) bool {
		annotations := reflect.Indirect(reflect.ValueOf(kr)).FieldByName("Annotations")
		m, err := annotationMap(annotations.Interface())
		if err != nil {
			return false
		}
		if _, exist := m[key]; exist {
			return true
		}
		return false
	}
	return lambda.Grep(annotationKeyEqual)
}

func (lambda *Lambda) HasLabel(key, value string) *Lambda {
	labelEqual := func(kr kubernetesResource) bool {
		labels := reflect.Indirect(reflect.ValueOf(kr)).FieldByName("Labels")
		m, err := annotationMap(labels.Interface())
		if err != nil {
			return false
		}
		if aValue, exist := m[key]; exist && aValue == value {
			return true
		}
		return false
	}
	return lambda.Grep(labelEqual)
}

func (lambda *Lambda) HasLabelKey(key string) *Lambda {
	labelEqual := func(kr kubernetesResource) bool {
		labels := reflect.Indirect(reflect.ValueOf(kr)).FieldByName("Labels")
		m, err := annotationMap(labels.Interface())
		if err != nil {
			return false
		}
		if _, exist := m[key]; exist {
			return true
		}
		return false
	}
	return lambda.Grep(labelEqual)
}
