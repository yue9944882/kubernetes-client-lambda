package lambda

import (
	"errors"
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

func annotationMap(i interface{}) (map[string]string, error) {
	// Get type
	t := reflect.TypeOf(i)

	switch t.Kind() {
	case reflect.Map:
		// Get the value of the provided map
		v := reflect.ValueOf(i)
		it := reflect.TypeOf("")
		// Create the map of the specific type. Key type is t.Key(), and element type is it
		m := reflect.MakeMap(reflect.MapOf(t.Key(), it))
		// Copy values to new map
		for _, mk := range v.MapKeys() {
			m.SetMapIndex(mk, v.MapIndex(mk))
		}
		return m.Interface().(map[string]string), nil
	}

	return nil, errors.New("Unsupported type")
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
