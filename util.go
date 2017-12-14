package lambda

import (
	"reflect"
)

func getNameOfResource(kr kubernetesResource) string {
	return reflect.Indirect(
		reflect.ValueOf(kr),
	).FieldByName("Name").String()
}

func isNamedspaced(kr kubernetesResource) bool {
	_, ok := reflect.Indirect(reflect.ValueOf(kr)).Type().FieldByName("Namespace")
	return ok
}
