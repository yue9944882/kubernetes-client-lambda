package lambda

import (
	"reflect"
	"regexp"
)

func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func regexMatch(str, regex string) (bool, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return false, err
	}
	return r.MatchString(str), nil
}
