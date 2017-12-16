package lambda

// NotEmpty checks if any element remains in lambda collection
// Returns true if the lambda collection is not empty and error if upstream lambda fails
func (lambda *Lambda) NotEmpty() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	for item := range lambda.val {
		if item != nil {
			return true, nil
		}
	}
	return false, lambda.Error
}

// Create creates every element remains in lambda collection
// Returns true if every element is successfully created and lambda error chain
// Fails if any element already exists
func (lambda *Lambda) Create() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	allCreated := true
	lambda.op.opListInterface()
	for item := range lambda.val {
		if _, err := lambda.op.opCreateInterface(item); err != nil {
			allCreated = false
			lambda.Error = err
		}
	}
	return allCreated, lambda.Error
}

// CreateIfNotExist creates element in the lambda collection
// Will not return false if any element fails to be created
func (lambda *Lambda) CreateIfNotExist() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	created := false
	searchHit := false
	for item := range lambda.val {
		if _, err := lambda.op.opGetInterface(getNameOfResource(item)); err != nil {
			if _, err := lambda.op.opCreateInterface(item); err != nil {
				lambda.Error = err
			} else {
				created = true
			}
		} else {
			searchHit = true
		}
	}
	return searchHit || created, lambda.Error
}

// Delete remove every element in the lambda collection
func (lambda *Lambda) Delete() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	deleted := false
	for item := range lambda.val {
		if err := lambda.op.opDeleteInterface(getNameOfResource(item)); err != nil {
			lambda.Error = err
		} else {
			deleted = true
		}
	}
	return deleted, lambda.Error
}

// DeleteIfExist delete elements in the lambda collection if it exists
func (lambda *Lambda) DeleteIfExist() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	deleted := false
	for item := range lambda.val {
		if _, err := lambda.op.opGetInterface(getNameOfResource(item)); err == nil {
			if err := lambda.op.opDeleteInterface(getNameOfResource(item)); err != nil {
				lambda.Error = err
			} else {
				deleted = true
			}
		}
	}
	return deleted, lambda.Error
}

// Update updates elements to kuberentes resources
func (lambda *Lambda) Update() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	updated := false
	for item := range lambda.val {
		if _, err := lambda.op.opUpdateInterface(item); err != nil {
			lambda.Error = err
		} else {
			updated = true
		}
	}
	return updated, lambda.Error
}

// UpdateIfExist checks if the element exists and update it value
func (lambda *Lambda) UpdateIfExist() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	deleted := false
	for item := range lambda.val {
		if _, err := lambda.op.opGetInterface(getNameOfResource(item)); err == nil {
			if err := lambda.op.opDeleteInterface(getNameOfResource(item)); err != nil {
				lambda.Error = err
			} else {
				deleted = true
			}
		}
	}
	return deleted, lambda.Error
}

// Sync automatically decides to create / update a resource
func (lambda *Lambda) Sync() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	updated := false
	created := false
	for item := range lambda.val {
		if rs, err := lambda.op.opGetInterface(getNameOfResource(item)); err != nil {
			if _, err := lambda.op.opCreateInterface(item); err != nil {
				lambda.Error = err
			} else {
				created = true
			}
		} else {
			if _, err := lambda.op.opUpdateInterface(rs); err != nil {
				lambda.Error = err
			} else {
				updated = true
			}
		}
	}
	return updated || created, lambda.Error
}
