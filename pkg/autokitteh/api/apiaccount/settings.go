package apiaccount

import (
	"google.golang.org/protobuf/proto"

	pbaccount "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/account"
)

type AccountSettingsPB = pbaccount.AccountSettings

type AccountSettings struct{ pb *pbaccount.AccountSettings }

func (a *AccountSettings) PB() *pbaccount.AccountSettings {
	if a == nil {
		return nil
	}
	return proto.Clone(a.pb).(*pbaccount.AccountSettings)
}

func (a *AccountSettings) Clone() *AccountSettings { return &AccountSettings{pb: a.PB()} }

func (a *AccountSettings) prep() *AccountSettings {
	if a == nil || a.pb == nil {
		return &AccountSettings{pb: &pbaccount.AccountSettings{}}
	}

	return a.Clone()
}

func (a *AccountSettings) Enabled() bool {
	if a == nil || a.pb == nil {
		return false
	}
	return a.pb.Enabled
}

func (a *AccountSettings) SetEnabled(e bool) *AccountSettings {
	a = a.prep()
	a.pb.Enabled = e
	return a
}

func (a *AccountSettings) Memo() map[string]string {
	if a == nil || a.pb == nil {
		return nil
	}
	return a.pb.Memo
}

func (a *AccountSettings) SetMemo(memo map[string]string) *AccountSettings {
	a = a.prep()
	a.pb.Memo = memo
	return a
}

func MustAccountSettingsFromProto(pb *pbaccount.AccountSettings) *AccountSettings {
	d, err := AccountSettingsFromProto(pb)
	if err != nil {
		panic(err)
	}
	return d
}

func AccountSettingsFromProto(pb *pbaccount.AccountSettings) (*AccountSettings, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&AccountSettings{pb: pb}).Clone(), nil
}
