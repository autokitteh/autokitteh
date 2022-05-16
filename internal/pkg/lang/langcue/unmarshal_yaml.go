package langcue

import (
	"fmt"

	"go.dagger.io/dagger/compiler"
)

func UnmarshalYAML(src []byte, dst interface{}) error {
	v, err := compiler.DecodeYAML("", src)
	if err != nil {
		return fmt.Errorf("src decode: %w", err)
	}

	if err := v.Decode(&dst); err != nil {
		return fmt.Errorf("value decode: %w", err)
	}

	return nil
}
