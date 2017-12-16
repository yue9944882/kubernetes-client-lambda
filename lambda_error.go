package lambda

func (lambda *Lambda) Dummy() *Lambda {

	l, ch := lambda.clone()
	go func() {
		for item := range lambda.val {
			ch <- item
		}
	}()
	return l
}

func (lambda *Lambda) MustNoError() *Lambda {
	if lambda.Error != nil {
		panic(lambda.Error)
	}
	return lambda.Dummy()
}

func (lambda *Lambda) NoError() bool {
	return lambda.Error == nil
}
