package apiaccount

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	pbaccount "github.com/autokitteh/autokitteh/gen/proto/stubs/go/account"
)

var (
	_ json.Marshaler   = &Account{}
	_ json.Unmarshaler = &Account{}
)

func (a *Account) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(a.pb)
}

func (a *Account) UnmarshalJSON(bs []byte) error {
	if a.pb == nil {
		a.pb = &pbaccount.Account{}
	}

	if err := protojson.Unmarshal(bs, a.pb); err != nil {
		return err
	}

	return a.pb.Validate()
}

//--

var (
	_ json.Marshaler   = &AccountSettings{}
	_ json.Unmarshaler = &AccountSettings{}
)

func (a *AccountSettings) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(a.pb)
}

func (a *AccountSettings) UnmarshalJSON(bs []byte) error {
	if a.pb == nil {
		a.pb = &pbaccount.AccountSettings{}
	}

	if err := protojson.Unmarshal(bs, a.pb); err != nil {
		return err
	}

	return a.pb.Validate()
}
