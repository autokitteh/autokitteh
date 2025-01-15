package kittehs

import (
	"fmt"
	"reflect"
)

func GetStructField(x any, name string) (any, error) {
	v := reflect.ValueOf(x)

	if v.Kind() == reflect.Ptr {
		return GetStructField(v.Elem().Interface(), name)
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %T", x)
	}

	f := v.FieldByName(name)
	if !f.IsValid() {
		return nil, fmt.Errorf("field %q not found", name)
	}

	return f.Interface(), nil
}
