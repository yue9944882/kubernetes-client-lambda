package lambda

import (
	"reflect"
)

func getNameOfResource(kr kubernetesResource) string {
	return reflect.Indirect(
		reflect.ValueOf(kr),
	).FieldByName("Name").String()
}
