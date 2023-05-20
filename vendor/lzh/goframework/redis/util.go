package redis

import (
	"reflect"

	json "github.com/json-iterator/go"
)

func parseToString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	default:
		b, e := json.Marshal(value)
		if e != nil {
			return ""
		}
		return string(b)
	}
}

func packArgs(items ...interface{}) (args []interface{}) {
	for _, item := range items {
		v := reflect.ValueOf(item)
		switch v.Kind() {
		case reflect.Slice:
			if v.IsNil() {
				continue
			}
			for i := 0; i < v.Len(); i++ {
				args = append(args, v.Index(i).Interface())
			}
		case reflect.Map:
			if v.IsNil() {
				continue
			}
			for _, key := range v.MapKeys() {
				args = append(args, key.Interface(), v.MapIndex(key).Interface())
			}
		default:
			args = append(args, v.Interface())
		}
	}
	return args
}
