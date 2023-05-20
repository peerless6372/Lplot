package utils

import "reflect"

func StructToMap(obj interface{}) map[string]interface{} {
	typ := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < typ.NumField(); i++ {
		data[typ.Field(i).Name] = val.Field(i).Interface()
	}
	return data
}
