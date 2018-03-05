package lambda

import (
	"fmt"
	"strings"
)

// Dummy do nothing and passing the elements next
func (lambda *Lambda) Dummy() *Lambda {
	l, ch := lambda.clone()
	go func() {
		for item := range lambda.val {
			ch <- item
		}
	}()
	return l
}

// MustNoError panics if any error occured
func (lambda *Lambda) MustNoError() *Lambda {
	if !lambda.NoError() {
		panic(lambda.Errors)
	}
	return lambda.Dummy()
}

// NoError checks if any error occured before
func (lambda *Lambda) NoError() bool {
	return lambda.Errors == nil || len(lambda.Errors) == 0
}

// ErrMultiLambdaFailure contains one or more error occured from
// lambda invocation chain
type ErrMultiLambdaFailure struct {
	errors []error
}

func (e ErrMultiLambdaFailure) Error() string {
	msgs := []string{}
	for _, err := range e.errors {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("%d error occured: %s", len(e.errors), strings.Join(msgs, ", "))
}
