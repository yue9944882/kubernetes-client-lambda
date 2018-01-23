package lambda

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
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

func getFirstArgInType(f interface{}) (Resource, error) {
	outType := reflect.TypeOf(f).In(0)
	return getResourceFromObject(outType)
}

func getFirstArgOutType(f interface{}) (Resource, error) {
	outType := reflect.TypeOf(f).Out(0)
	return getResourceFromObject(outType)
}

func getResourceTypeOfPredicate(predicate Predicate) (Resource, error) {
	return getFirstArgInType(predicate)
}

func getResourceTypeOfConsumer(consumer Consumer) (Resource, error) {
	return getFirstArgInType(consumer)
}

func getResourceTypeOfFunction(function Function) (Resource, error) {
	return getFirstArgInType(function)
}

func getResourceTypeOfProducer(producer Producer) (Resource, error) {
	return getFirstArgOutType(producer)
}

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
	op     *kubernetesExecutable
	val    <-chan kubernetesResource
	Errors []error
}

func (lambda *Lambda) addError(err error) {
	if lambda.Errors == nil {
		lambda.Errors = []error{err}
	}
	lambda.Errors = append(lambda.Errors, err)
}

func (lambda *Lambda) clone() (*Lambda, chan kubernetesResource) {
	ch := make(chan kubernetesResource)
	l := &Lambda{
		op:     lambda.op,
		val:    ch,
		Errors: lambda.Errors,
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
				ch <- v
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
			ch <- v
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
	nameEqual := func(kr kubernetesResource) bool {
		return reflect.Indirect(reflect.ValueOf(kr)).FieldByName("Name").String() == name
	}
	return lambda.Grep(nameEqual)
}

// NamePrefix filter the elements out if its name doesn't have the prefix
func (lambda *Lambda) NamePrefix(prefix string) *Lambda {
	namePrefix := func(kr kubernetesResource) bool {
		name := reflect.Indirect(reflect.ValueOf(kr)).FieldByName("Name").String()
		return strings.HasPrefix(name, prefix)
	}
	return lambda.Grep(namePrefix)
}

// NameRegex filter the elements out if its name fails to matches the regexp
func (lambda *Lambda) NameRegex(regex string) *Lambda {
	nameRegex := func(kr kubernetesResource) bool {
		name := reflect.Indirect(reflect.ValueOf(kr)).FieldByName("Name").String()
		matched, err := regexMatch(name, regex)
		if err != nil {
			lambda.addError(err)
			return false
		}
		return matched
	}
	return lambda.Grep(nameRegex)
}

// HasAnnotation filter the elements out if it doesn't have the arugument annotation
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

// HasAnnotationKey filter the elements out if it doesn't have the arugument annotation key
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

// HasLabel filter the elements out if it doesn't have the arugument label
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

// HasLabelKey filter the elements out if it doesn't have the arugument label key
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
