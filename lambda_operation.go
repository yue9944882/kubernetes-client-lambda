package lambda

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
		if !callPredicate(predicate, item) {
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
		if callPredicate(predicate, item) {
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
		callFunction(function, item)
	}
	return nil
}

//********************************************************
// Kubernetes Operation
//********************************************************

// Create creates every element remains in lambda collection
// Returns true if every element is successfully created and lambda error chain
// Fails if any element already exists
func (lambda *Lambda) Create() (allCreated bool, err error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	allCreated = true
	lambda.op.opListInterface()
	for item := range lambda.val {
		if _, err := lambda.op.opCreateInterface(item); err != nil {
			allCreated = false
			lambda.addError(err)
		}
	}
	if len(lambda.Errors) != 0 {
		err = &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return
}

// CreateIfNotExist creates element in the lambda collection
// Will not return false if any element fails to be created
func (lambda *Lambda) CreateIfNotExist() (success bool, err error) {
	if !lambda.NoError() {
		err = &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
		return false, err
	}
	created := false
	searchHit := false
	for item := range lambda.val {
		if _, err := lambda.op.opGetInterface(getNameOfResource(item)); err != nil {
			if _, err := lambda.op.opCreateInterface(item); err != nil {
				lambda.addError(err)
			} else {
				created = true
			}
		} else {
			searchHit = true
		}
	}
	if len(lambda.Errors) != 0 {
		err = &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	success = searchHit || created
	return
}

// Delete remove every element in the lambda collection
func (lambda *Lambda) Delete() (deleted bool, err error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	deleted = false
	for item := range lambda.val {
		if err := lambda.op.opDeleteInterface(getNameOfResource(item)); err != nil {
			lambda.addError(err)
		} else {
			deleted = true
		}
	}
	if len(lambda.Errors) != 0 {
		err = &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	return
}

// DeleteIfExist delete elements in the lambda collection if it exists
func (lambda *Lambda) DeleteIfExist() (deleted bool, err error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	deleted = false
	for item := range lambda.val {
		if _, err := lambda.op.opGetInterface(getNameOfResource(item)); err == nil {
			if err := lambda.op.opDeleteInterface(getNameOfResource(item)); err != nil {
				lambda.addError(err)
			} else {
				deleted = true
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

// Update updates elements to kuberentes resources
func (lambda *Lambda) Update() (updated bool, err error) {
	if !lambda.NoError() {
		return false, &ErrMultiLambdaFailure{
			errors: lambda.Errors,
		}
	}
	updated = false
	for item := range lambda.val {
		if _, err := lambda.op.opUpdateInterface(item); err != nil {
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
		if _, err := lambda.op.opGetInterface(getNameOfResource(item)); err == nil {
			if _, err := lambda.op.opUpdateInterface(item); err != nil {
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
