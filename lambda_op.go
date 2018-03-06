package lambda

import (
	"bytes"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

//********************************************************
// Basic Operation
//********************************************************

// NotEmpty checks if any element remains
// Returns true if the lambda collection is not empty and error if upstream lambda fails
func (lambda *Lambda) NotEmpty() (noempty bool, err error) {
	noempty = true
	err = lambda.run(
		func() {
			for item := range lambda.val {
				if item == nil {
					noempty = false
				}
			}
		},
	)
	return
}

// Every checks if every element get a true from predicate
func (lambda *Lambda) Every(predicate Predicate) (every bool, err error) {
	every = true
	err = lambda.run(
		func() {
			for item := range lambda.val {
				if !callPredicate(predicate, item) {
					every = false
				}
			}
		},
	)
	return
}

// Any checks if any element get a true from predicate
func (lambda *Lambda) Any(predicate Predicate) (any bool, err error) {
	err = lambda.run(
		func() {
			for item := range lambda.val {
				if callPredicate(predicate, item) {
					any = true
				}
			}
		},
	)
	return
}

// Each applies function to every element
func (lambda *Lambda) Each(function Function) error {
	return lambda.run(
		func() {
			for item := range lambda.val {
				callFunction(function, item)
			}
		},
	)
}

//********************************************************
// Kubernetes Operation
//********************************************************

// Create creates every element remains in lambda collection
// Returns true if every element is successfully created and lambda error chain
// Fails if any element already exists
func (lambda *Lambda) Create() (created bool, err error) {
	err = lambda.run(func() {
		for item := range lambda.val {
			if err := create(lambda.clientInterface, lambda.rs, item); err != nil {
				lambda.addError(err)
				continue
			} else {
				created = true
			}
		}
	})
	return
}

// CreateIfNotExist creates element in the lambda collection
// Will not return false if any element fails to be created
func (lambda *Lambda) CreateIfNotExist() (created, existed bool, err error) {
	err = lambda.run(
		func() {
			for item := range lambda.val {
				accessor, err := meta.Accessor(item)
				if err != nil {
					lambda.addError(err)
					continue
				}
				if _, err := lambda.getFunc(accessor.GetNamespace(), accessor.GetName()); err != nil {
					// TODO: judge if the error is errors.StatusError
					if err := create(lambda.clientInterface, lambda.rs, item); err != nil {
						lambda.addError(err)
					} else {
						created = true
					}
				} else {
					existed = true
				}
			}
		},
	)
	return
}

// Delete remove every element in the lambda collection
func (lambda *Lambda) Delete() (deleted bool, err error) {
	err = lambda.run(
		func() {
			for item := range lambda.val {
				if err != nil {
					lambda.addError(err)
					continue
				}
				if err := delete(lambda.clientInterface, lambda.rs, item); err != nil {
					lambda.addError(err)
				} else {
					deleted = true
				}
			}
		},
	)
	return
}

// DeleteIfExist delete elements in the lambda collection if it exists
func (lambda *Lambda) DeleteIfExist() (deleted, existed bool, err error) {
	err = lambda.run(
		func() {
			for item := range lambda.val {
				accessor, err := meta.Accessor(item)
				if err != nil {
					lambda.addError(err)
					continue
				}
				if _, err := lambda.getFunc(accessor.GetNamespace(), accessor.GetName()); err == nil {
					if err := delete(lambda.clientInterface, lambda.rs, item); err != nil {
						lambda.addError(err)
					} else {
						deleted = true
						existed = true
					}
				} else {
					deleted = true
				}
			}
		},
	)
	return
}

// Update updates elements to kuberentes resources
func (lambda *Lambda) Update() (updated bool, err error) {
	err = lambda.run(
		func() {
			for item := range lambda.val {
				if err := update(lambda.clientInterface, lambda.rs, item); err != nil {
					lambda.addError(err)
				} else {
					updated = true
				}
			}
		},
	)
	return
}

// UpdateIfExist checks if the element exists and update it value
func (lambda *Lambda) UpdateIfExist() (updated, existed bool, err error) {
	lambda.run(
		func() {
			for item := range lambda.val {
				accessor, err := meta.Accessor(item)
				if err != nil {
					lambda.addError(err)
					continue
				}
				if _, err := lambda.getFunc(accessor.GetNamespace(), accessor.GetName()); err == nil {
					if err := update(lambda.clientInterface, lambda.rs, item); err != nil {
						lambda.addError(err)
					} else {
						updated = true
					}
				}
			}
		},
	)
	return
}

func (lambda *Lambda) UpdateOrCreate() (updated, created bool, err error) {
	err = lambda.run(
		func() {
			for item := range lambda.val {
				accessor, err := meta.Accessor(item)
				if err != nil {
					lambda.addError(err)
					continue
				}
				if _, err := lambda.getFunc(accessor.GetNamespace(), accessor.GetName()); err == nil {
					if err := update(lambda.clientInterface, lambda.rs, item); err != nil {
						lambda.addError(err)
					} else {
						updated = true
					}
				} else {
					if err := create(lambda.clientInterface, lambda.rs, item); err != nil {
						lambda.addError(err)
					} else {
						created = true
					}
				}
			}
		},
	)
	return
}

func castObjectToUnstructured(object runtime.Object) (*unstructured.Unstructured, error) {
	buffer := new(bytes.Buffer)
	err := unstructured.UnstructuredJSONScheme.Encode(object, buffer)
	if err != nil {
		return nil, err
	}
	obj, _, err := unstructured.UnstructuredJSONScheme.Decode(buffer.Bytes(), nil, nil)
	if err != nil {
		return nil, err
	}
	return obj.(*unstructured.Unstructured), nil
}

func create(i dynamic.Interface, rs Resource, object runtime.Object) error {
	api := GetResouceIndexerInstance().GetAPIResource(rs)
	accessor, err := meta.Accessor(object)
	if err != nil {
		return err
	}
	tmpObj, err := castObjectToUnstructured(object)
	if err != nil {
		return err
	}
	if _, err := i.Resource(api, accessor.GetNamespace()).Create(tmpObj); err != nil {
		return err
	}
	return nil
}

func delete(i dynamic.Interface, rs Resource, object runtime.Object) error {
	api := GetResouceIndexerInstance().GetAPIResource(rs)
	accessor, err := meta.Accessor(object)
	if err != nil {
		return err
	}
	if err := i.Resource(api, accessor.GetNamespace()).Delete(accessor.GetName(), &metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func update(i dynamic.Interface, rs Resource, object runtime.Object) error {
	api := GetResouceIndexerInstance().GetAPIResource(rs)
	accessor, err := meta.Accessor(object)
	if err != nil {
		return err
	}
	tmpObj, err := castObjectToUnstructured(object)
	if err != nil {
		return err
	}
	if _, err := i.Resource(api, accessor.GetNamespace()).Update(tmpObj); err != nil {
		return err
	}
	return nil
}
