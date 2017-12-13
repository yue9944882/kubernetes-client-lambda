package lambda

func (lambda *Lambda) MustNoError() bool {
	if lambda.Error != nil {
		panic(lambda.Error)
	}
	return true
}

func (lambda *Lambda) NoError() bool {
	return lambda.Error == nil
}
