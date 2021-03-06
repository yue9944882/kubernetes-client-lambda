package lambda

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

// Predicate is a function has only one parameter and return boolean.
// When return value type is not boolean, panic will occur.
// The parameter can be of any type but nil and predicates is always used to test an element.
type Predicate interface{}

// Consumer is a function has one parameter and returns one value.
// Nil is not allowed to be used as parameter or return value, or panic will occur.
// Consumer is always used to apply some transformation to elements.
type Consumer interface{}

// Function is a function has one parameter and has no return value.
// The parameter must not be a nil
type Function interface{}

// Producer is a function takes no parameter and returns a value.
// Producer is recommeneded to be a closure so that the returning value can be controlled outside lambda.
type Producer interface{}

func callPredicate(f interface{}, arg interface{}) bool {
	if isZeroOfUnderlyingType(arg) {
		panic(fmt.Sprintf("nil argument detected when calling predicate %#v with arg %#v", f, arg))
	}
	ret := reflect.ValueOf(f).Call([]reflect.Value{
		reflect.ValueOf(arg),
	})
	return ret[0].Bool()
}

func callFunction(f interface{}, arg interface{}) {
	if isZeroOfUnderlyingType(arg) {
		panic(fmt.Sprintf("nil argument detected when calling function %#v with arg %#v", f, arg))
	}
	reflect.ValueOf(f).Call([]reflect.Value{
		reflect.ValueOf(arg),
	})
}

func callConsumer(f interface{}, arg interface{}) interface{} {
	if isZeroOfUnderlyingType(arg) {
		panic(fmt.Sprintf("nil argument detected when calling consumer %#v with arg %#v", f, arg))
	}
	ret := reflect.ValueOf(f).Call([]reflect.Value{
		reflect.ValueOf(arg),
	})
	if isZeroOfUnderlyingType(ret[0].Interface()) {
		panic(fmt.Sprintf("nil return value detected when calling consumer %#v with arg %#v", f, arg))
	}
	return ret[0].Interface()
}

func callProducer(f interface{}) interface{} {
	ret := reflect.ValueOf(f).Call([]reflect.Value{})
	if isZeroOfUnderlyingType(ret[0].Interface()) {
		panic(fmt.Sprintf("nil return value detected when calling producer %#v", f))
	}
	return ret[0].Interface()
}

// Lambda is a basic and core type of KCL. It holds a channel for receiving elements from previous
// lambda or kubernetes resource fetcher. Error is assigned if any error occured during lambda
// pipelining. The error will be recorded but the lambda pipelining will continue on, and forcing it
// fail-hard needs call MustNoError method. The error can be also be returned at the end of a pipeline
// via lambda operation method which is defined in lambda_operation.go
type Lambda struct {
	getFunc    func(namespace, name string) (runtime.Object, error)
	listFunc   func(namespace string, selector labels.Selector) ([]runtime.Object, error)
	createFunc func(runtime.Object) error
	updateFunc func(runtime.Object) error
	deleteFunc func(runtime.Object) error

	clientInterface dynamic.Interface
	rs              Resource
	namespaces      []string
	val             <-chan runtime.Object
	Errors          []error
}

func (lambda *Lambda) run(f func()) error {
	if !lambda.NoError() {
		return &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	f()
	if len(lambda.Errors) != 0 {
		return &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return nil
}

func (lambda *Lambda) addError(err error) {
	if lambda.Errors == nil {
		lambda.Errors = []error{err}
	}
	lambda.Errors = append(lambda.Errors, err)
}

func (lambda *Lambda) clone() (*Lambda, chan runtime.Object) {
	ch := make(chan runtime.Object)
	l := &Lambda{
		rs:              lambda.rs,
		val:             ch,
		Errors:          lambda.Errors,
		getFunc:         lambda.getFunc,
		listFunc:        lambda.listFunc,
		createFunc:      lambda.createFunc,
		updateFunc:      lambda.updateFunc,
		deleteFunc:      lambda.deleteFunc,
		clientInterface: lambda.clientInterface,
	}
	return l, ch
}

//********************************************************
// Lambda with no parameter
//********************************************************

// Collect deep copies every element in the lambda
func (lambda *Lambda) Collect() *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		for item := range lambda.val {
			if obj, ok := item.(runtime.Object); !ok {
				l.addError(fmt.Errorf("Invalid object type of %#v", obj))
			} else {
				ch <- obj.DeepCopyObject()
			}
		}
	}()
	return l
}

// LatestCreated filters out the latest created object
func (lambda *Lambda) LatestCreated() *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		var latestObj runtime.Object
		var latestTimestamp *time.Time
		for item := range lambda.val {
			item := item
			accessor, _ := meta.Accessor(item)
			if latestTimestamp == nil || accessor.GetCreationTimestamp().Time.After(*latestTimestamp) {
				t := accessor.GetCreationTimestamp().Time
				latestTimestamp = &t
				latestObj = item
			}
		}
		ch <- latestObj
	}()
	return l
}

// Element returns a single element. Note that if multiple elements
// exists in the lambda pipeline, a random object may be returned.
func (lambda *Lambda) Element() (runtime.Object, error) {
	var element runtime.Object
	for item := range lambda.val {
		element = item
	}
	if element == nil {
		return nil, fmt.Errorf("no element found")
	}
	return element, nil
}

// Elements returns all elements from the lambda pipeline.
func (lambda *Lambda) Elements() ([]runtime.Object, error) {
	var elements []runtime.Object
	for item := range lambda.val {
		item := item
		elements = append(elements, item)
	}
	if len(elements) == 0 {
		return nil, fmt.Errorf("no elements found")
	}
	return elements, nil
}

//********************************************************
// Lambda using Predicate
//********************************************************

// First returnes the first element matches the predicate
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

// Grep returnes the elements matches the predicate
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

// Map transforms and replace the elements and put them to the next lambda
func (lambda *Lambda) Map(consumer Consumer) *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		for item := range lambda.val {
			if v := callConsumer(consumer, item); v != nil {
				ch <- v.(runtime.Object)
			}
		}
	}()
	return l
}

//********************************************************
// Lambda using Function
//********************************************************

// Iter iterates the elements and apply function to them
// Note that modifying is not recommened in Iter, use Map to modify elements instead
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

// Add calls the producer and put the returned value into elements
func (lambda *Lambda) Add(producer Producer) *Lambda {
	l, ch := lambda.clone()
	go func() {
		defer close(ch)
		for item := range lambda.val {
			ch <- item
		}
		if v := callProducer(producer); v != nil {
			ch <- v.(runtime.Object)
		}
	}()
	return l
}

//********************************************************
// Snippet Kubernetes Lambda Functions
// Utility Lambda Function
//********************************************************

// NameEqual filter the elements out if its name mismatches with the argument name
func (lambda *Lambda) NameEqual(name string) *Lambda {
	return lambda.Grep(func(object runtime.Object) bool {
		accessor, err := meta.Accessor(object)
		if err != nil {
			return false
		}
		return accessor.GetName() == name
	})
}

// NamePrefix filter the elements out if its name doesn't have the prefix
func (lambda *Lambda) NamePrefix(prefix string) *Lambda {
	return lambda.Grep(func(object runtime.Object) bool {
		accessor, err := meta.Accessor(object)
		if err != nil {
			return false
		}
		return strings.HasPrefix(accessor.GetName(), prefix)
	})
}

// NameRegex filter the elements out if its name fails to matches the regexp
func (lambda *Lambda) NameRegex(regex string) *Lambda {
	return lambda.Grep(func(object runtime.Object) bool {
		accessor, err := meta.Accessor(object)
		if err != nil {
			return false
		}
		name := accessor.GetName()
		matched, err := regexMatch(name, regex)
		if err != nil {
			lambda.addError(err)
			return false
		}
		return matched
	})
}

// HasAnnotation filter the elements out if it doesn't have the arugument annotation
func (lambda *Lambda) HasAnnotation(key, value string) *Lambda {
	return lambda.Grep(func(object runtime.Object) bool {
		accessor, err := meta.Accessor(object)
		if err != nil {
			return false
		}
		return accessor.GetAnnotations()[key] == value
	})
}

// HasAnnotationKey filter the elements out if it doesn't have the arugument annotation key
func (lambda *Lambda) HasAnnotationKey(key string) *Lambda {
	return lambda.Grep(func(object runtime.Object) bool {
		accessor, err := meta.Accessor(object)
		if err != nil {
			return false
		}
		return accessor.GetAnnotations()[key] != ""
	})
}

// HasLabel filter the elements out if it doesn't have the arugument label
func (lambda *Lambda) HasLabel(key, value string) *Lambda {
	return lambda.Grep(func(object runtime.Object) bool {
		accessor, err := meta.Accessor(object)
		if err != nil {
			return false
		}
		return accessor.GetLabels()[key] != value
	})
}

// HasLabelKey filter the elements out if it doesn't have the arugument label key
func (lambda *Lambda) HasLabelKey(key string) *Lambda {
	return lambda.Grep(func(object runtime.Object) bool {
		accessor, err := meta.Accessor(object)
		if err != nil {
			return false
		}
		return accessor.GetLabels()[key] != ""
	})
}
