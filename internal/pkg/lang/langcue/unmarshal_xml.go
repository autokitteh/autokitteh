package langcue

import (
	"fmt"
)

func UnmarshalXML(src []byte, dst interface{}) error {
	// TODO: Needs apivalues to support XML unmarshaling.
	return fmt.Errorf("not implemented yet")
}
