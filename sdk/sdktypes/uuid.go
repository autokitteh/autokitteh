package sdktypes

import (
	"encoding/hex"
	"errors"
)

func validateUUID(raw string) error {
	bs, err := hex.DecodeString(raw)
	if err != nil {
		return err
	}

	if len(bs) != 16 {
		return errors.New("unexpected length")
	}

	return nil
}
