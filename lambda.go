package lambda

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

// Predicate is a function has only one parameter and return boolean.
// When return value type is not boolean, panic will occur.
// The parameter can be of any type but nil and predicates is always used to test an element.
type Predicate func(runtime.Object) bool

// Consumer is a function has one parameter and returns one value.
// Nil is not allowed to be used as parameter or return value, or panic will occur.
// Consumer is always used to apply some transformation to elements.
type Consumer func(runtime.Object) runtime.Object

// Function is a function has one parameter and has no return value.
// The parameter must not be a nil
type Function func(runtime.Object)

// Producer is a function takes no parameter and returns a value.
// Producer is recommeneded to be a closure so that the returning value can be controlled outside lambda.
type Producer func() runtime.Object

// Lambda is a basic and core type of KCL. It holds a channel for receiving elements from previous
// lambda or kubernetes resource fetcher. Error is assigned if any error occured during lambda
// pipelining. The error will be recorded but the lambda pipelining will continue on, and forcing it
// fail-hard needs call MustNoError method. The error can be also be returned at the end of a pipeline
// via lambda operation method which is defined in lambda_operation.go
type Lambda struct {
	getFunc         func(namespace, name string) (runtime.Object, error)
	clientInterface dynamic.Interface
	op              *kubernetesExecutable
	rs              Resource
	val             <-chan runtime.Object
	Errors          []error
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
			if predicate(item) {
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
			if predicate(item) {
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
			if v := consumer(item); v != nil {
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
			function(item)
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
		if v := producer(); v != nil {
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
