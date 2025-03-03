package checks

import (
	"errors"
)

// CheckKeyAndValue returns an error if k == "" or if v == nil
func CheckKeyAndValue(k string, v any) error {
	return errors.Join(CheckKey(k), CheckVal(v))
}

// CheckKey returns an error if k == ""
func CheckKey(k string) error {
	if k == "" {
		return errors.New("The passed key is an empty string, which is invalid")
	}
	return nil
}

// CheckVal returns an error if v == nil
func CheckVal(v any) error {
	if v == nil {
		return errors.New("The passed value is nil, which is not allowed")
	}
	return nil
}
