package lambda

func (lambda *Lambda) Exists() (bool, error) {
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

func (lambda *Lambda) Create() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	allCreated := true
	lambda.op.opListInterface()
	for item := range lambda.val {
		if _, err := lambda.op.opCreateInterface(item); err != nil {
			allCreated = false
		}
	}
	return allCreated, lambda.Error
}

func (lambda *Lambda) CreateIfNotExist() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	created := false
	for item := range lambda.val {
		if _, err := lambda.op.opGetInterface(getNameOfResource(item)); err != nil {
			if _, err := lambda.op.opCreateInterface(item); err != nil {
				lambda.Error = err
			} else {
				created = true
			}
		}
	}
	return created, lambda.Error
}

func (lambda *Lambda) Update() (bool, error) {
	if !lambda.NoError() {
		return false, lambda.Error
	}
	updated := false
	for item := range lambda.val {
		if rs, err := lambda.op.opGetInterface(getNameOfResource(item)); err == nil {
			if _, err := lambda.op.opUpdateInterface(rs); err != nil {
				lambda.Error = err
			} else {
				updated = true
			}
		}
	}
	return updated, lambda.Error
}
