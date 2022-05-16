package accountsstoregorm

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiaccount"
)

type account struct {
	Name      string `gorm:"primaryKey"`
	Enabled   bool   `gorm:"index"`
	Memo      datatypes.JSON
	CreatedAt time.Time // only set in initial creation, then carried over.
	UpdatedAt time.Time // set on updates.
}

func marshalMemo(memo map[string]string) (datatypes.JSON, error) {
	if memo == nil {
		return nil, nil
	}

	bs, err := json.Marshal(memo)
	if err != nil {
		return nil, fmt.Errorf("memo marshal: %w", err)
	}

	return datatypes.JSON(bs), nil
}

func unmarshalMemo(f datatypes.JSON) (map[string]string, error) {
	if len(f) == 0 {
		return nil, nil
	}

	var m map[string]string
	if err := json.Unmarshal(f, &m); err != nil {
		return nil, fmt.Errorf("memo unmarshal: %w", err)
	}

	return m, nil
}

func decodeAccount(a *account) (*apiaccount.Account, error) {
	memo, err := unmarshalMemo(a.Memo)
	if err != nil {
		return nil, err
	}

	var updatedAt *time.Time

	if !a.UpdatedAt.IsZero() {
		updatedAt = &a.UpdatedAt
	}

	aa, err := apiaccount.NewAccount(
		apiaccount.AccountName(a.Name),
		(&apiaccount.AccountSettings{}).
			SetEnabled(a.Enabled).
			SetMemo(memo),
		a.CreatedAt,
		updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid account record: %w", err)
	}

	return aa, nil
}
