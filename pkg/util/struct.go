package util

import (
	"fmt"
	"reflect"
	"strings"
)

func CheckStruct(o interface{}) error {
	obj := reflect.ValueOf(o)
	typ := reflect.TypeOf(o)
	for i := 0; i < obj.Elem().NumField(); i++ {
		field := obj.Elem().Field(i)
		//fieldName := obj.Elem().Type().Field(i).Name
		fieldName := typ.Elem().Field(i).Tag.Get("json")
		names := strings.Split(fieldName, ",")
		if len(names) > 1 && names[1] == "omitempty" {
			continue
		}
		if field.Type().String() == "string" {
			if value, ok := field.Interface().(string); ok {
				if len(value) == 0 {
					return fmt.Errorf("%s is null", fieldName)
				}
			} else {
				return fmt.Errorf("cannot cast %+v to string", field.Interface())
			}
		}
	}
	return nil
}
