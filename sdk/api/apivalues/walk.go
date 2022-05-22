package apivalues

import (
	"errors"
)

var (
	// Ignore current value's children.
	ErrSkip = errors.New("skip")

	// Abort scanning of current fields of parent and continue to parent sibling, if exists.
	ErrUp = errors.New("up")
)

type Role string

const (
	RoleNone             Role = ""
	RoleListItem         Role = "list_item"
	RoleSetItem          Role = "set_item"
	RoleStructCtor       Role = "struct_ctor"
	RoleDictKey          Role = "dict_key"
	RoleDictValue        Role = "dict_value"
	RoleStructFieldValue Role = "struct_field_value"
	RoleModuleMember     Role = "module_member"
)

// NOTE: f can mutate the values it's given if it chooses to.
func Walk(v *Value, f func(curr, parent *Value, role Role) error) error {
	return walk(v, nil, RoleNone, f)
}

func walk(curr, parent *Value, role Role, f func(curr, parent *Value, role Role) error) error {
	if err := f(curr, parent, role); err != nil {
		if errors.Is(err, ErrSkip) {
			return nil
		}

		return err
	}

	ignoreUpError := func(err error) error {
		if errors.Is(err, ErrUp) {
			return nil
		}
		return err
	}

	vv := curr.Get()

	// this allows the callee to mutate the value.
	defer curr.set(vv)

	switch vv := vv.(type) {
	case ListValue:
		for _, v := range vv {
			return ignoreUpError(walk(v, curr, RoleListItem, f))
		}
	case SetValue:
		for _, v := range vv {
			return ignoreUpError(walk(v, curr, RoleSetItem, f))
		}
	case DictValue:
		for _, kv := range vv {
			if err := walk(kv.K, curr, RoleDictKey, f); err != nil {
				return ignoreUpError(err)
			}

			if err := walk(kv.V, curr, RoleDictValue, f); err != nil {
				return ignoreUpError(err)
			}
		}
	case StructValue:
		if err := walk(vv.Ctor, curr, RoleStructCtor, f); err != nil {
			return ignoreUpError(err)
		}

		for _, v := range vv.Fields {
			if err := walk(v, curr, RoleStructFieldValue, f); err != nil {
				return ignoreUpError(err)
			}
		}
	case ModuleValue:
		for _, v := range vv.Members {
			if err := walk(v, curr, RoleModuleMember, f); err != nil {
				return ignoreUpError(err)
			}
		}
	}

	return nil
}
