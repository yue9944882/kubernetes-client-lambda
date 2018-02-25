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
func (lambda *Lambda) NotEmpty() (bool, error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	for item := range lambda.val {
		if item != nil {
			return true, nil
		}
	}
	return false, nil
}

// Every checks if every element get a true from predicate
func (lambda *Lambda) Every(predicate Predicate) (bool, error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	for item := range lambda.val {
		if !predicate(item) {
			return false, nil
		}
	}
	return true, nil
}

// Any checks if any element get a true from predicate
func (lambda *Lambda) Any(predicate Predicate) (bool, error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	for item := range lambda.val {
		if predicate(item) {
			return true, nil
		}
	}
	return false, nil
}

// Each applies function to every element
func (lambda *Lambda) Each(function Function) error {
	if !lambda.NoError() {
		return &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	for item := range lambda.val {
		function(item)
	}
	return nil
}

//********************************************************
// Kubernetes Operation
//********************************************************

// Create creates every element remains in lambda collection
// Returns true if every element is successfully created and lambda error chain
// Fails if any element already exists
func (lambda *Lambda) Create() (bool, error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	allCreated := true
	for item := range lambda.val {
		if err := create(lambda.clientInterface, lambda.rs, item); err != nil {
			lambda.addError(err)
			allCreated = false
		}
	}
	if len(lambda.Errors) != 0 {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return allCreated, nil
}

// CreateIfNotExist creates element in the lambda collection
// Will not return false if any element fails to be created
func (lambda *Lambda) CreateIfNotExist() (bool, error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	created := false
	searchHit := false
	for item := range lambda.val {
		accessor, err := meta.Accessor(item)
		if err != nil {
			return false, err
		}
		if obj, err := lambda.getFunc(accessor.GetNamespace(), accessor.GetName()); err != nil {
			if err := create(lambda.clientInterface, lambda.rs, obj); err != nil {
				lambda.addError(err)
			} else {
				created = true
			}
		} else {
			searchHit = true
		}
	}
	if len(lambda.Errors) != 0 {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return searchHit || created, nil
}

// Delete remove every element in the lambda collection
func (lambda *Lambda) Delete() (bool, error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	deleted := false
	for item := range lambda.val {
		if err := delete(lambda.clientInterface, lambda.rs, item); err != nil {
			lambda.addError(err)
		} else {
			deleted = true
		}
	}
	if len(lambda.Errors) != 0 {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return deleted, nil
}

// DeleteIfExist delete elements in the lambda collection if it exists
func (lambda *Lambda) DeleteIfExist() (bool, error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	deleted := false
	for item := range lambda.val {
		accessor, err := meta.Accessor(item)
		if err != nil {
			return false, err
		}
		if _, err := lambda.getFunc(accessor.GetNamespace(), accessor.GetName()); err == nil {
			if err := delete(lambda.clientInterface, lambda.rs, item); err != nil {
				lambda.addError(err)
			} else {
				deleted = true
			}
		}
	}
	if len(lambda.Errors) != 0 {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return deleted, nil
}

// Update updates elements to kuberentes resources
func (lambda *Lambda) Update() (updated bool, err error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	updated = false
	for item := range lambda.val {
		if err := update(lambda.clientInterface, lambda.rs, item); err != nil {
			lambda.addError(err)
		} else {
			updated = true
		}
	}
	if len(lambda.Errors) != 0 {
		err = &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return
}

// UpdateIfExist checks if the element exists and update it value
func (lambda *Lambda) UpdateIfExist() (updated bool, err error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	updated = false
	for item := range lambda.val {
		accessor, err := meta.Accessor(item)
		if err != nil {
			return false, err
		}
		if _, err := lambda.getFunc(accessor.GetNamespace(), accessor.GetName()); err == nil {
			if err := update(lambda.clientInterface, lambda.rs, item); err != nil {
				lambda.addError(err)
			} else {
				updated = true
			}
		}
	}
	if len(lambda.Errors) != 0 {
		err = &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return
}

// Sync automatically decides to create / update a resource
// If the resource doesn't exist,
// DO NOT USE THIS: TODO: resolve critical BUG!!
/*
func (lambda *Lambda) UpdateOrCreate() (success bool, err error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	updated := false
	created := false
	for item := range lambda.val {
		if rs, err := lambda.op.opGetInterface(getNameOfResource(item)); err != nil {
			if _, err := lambda.op.opCreateInterface(item); err != nil {
				lambda.addError(err)
			} else {
				created = true
			}
		} else {
			if _, err := lambda.op.opUpdateInterface(rs); err != nil {
				lambda.addError(err)
			} else {
				updated = true
			}
		}
	}
	if len(lambda.Errors) != 0 {
		err = &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	success = updated || created
	return
}
*/

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
	api := getResouceIndexerInstance().GetAPIResource(rs)
	accessor, err := meta.Accessor(object)
	if err != nil {
		return err
	}
	tmpObj, err := castObjectToUnstructured(object)
	if err != nil {
		return err
	}
	if _, err := i.Resource(&api, accessor.GetNamespace()).Create(tmpObj); err != nil {
		return err
	}
	return nil
}

func delete(i dynamic.Interface, rs Resource, object runtime.Object) error {
	api := getResouceIndexerInstance().GetAPIResource(rs)
	accessor, err := meta.Accessor(object)
	if err != nil {
		return err
	}
	if err := i.Resource(&api, accessor.GetNamespace()).Delete(accessor.GetName(), &metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func update(i dynamic.Interface, rs Resource, object runtime.Object) error {
	api := getResouceIndexerInstance().GetAPIResource(rs)
	accessor, err := meta.Accessor(object)
	if err != nil {
		return err
	}
	tmpObj, err := castObjectToUnstructured(object)
	if err != nil {
		return err
	}
	if _, err := i.Resource(&api, accessor.GetNamespace()).Update(tmpObj); err != nil {
		return err
	}
	return nil
}
