package common

import "errors"

type nonEmptyStringValue string

func NewNonEmptyString(val string, p *string) *nonEmptyStringValue {
	*p = val
	return (*nonEmptyStringValue)(p)
}

func (nes *nonEmptyStringValue) Set(val string) error {
	if val == "" {
		return errors.New("value cannot be empty")
	}
	*nes = nonEmptyStringValue(val)
	return nil
}

func (nes *nonEmptyStringValue) Type() string {
	return "non-empty string"
}

func (nes *nonEmptyStringValue) String() string { return string(*nes) }
