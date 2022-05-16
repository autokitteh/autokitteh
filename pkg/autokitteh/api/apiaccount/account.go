package apiaccount

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbaccount "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/account"
)

type AccountPB = pbaccount.Account

type Account struct{ pb *pbaccount.Account }

func (a *Account) PB() *pbaccount.Account {
	if a == nil || a.pb == nil {
		return nil
	}

	return proto.Clone(a.pb).(*pbaccount.Account)
}

func (a *Account) Clone() *Account {
	if a == nil || a.pb == nil {
		return nil
	}

	return &Account{pb: a.PB()}
}

func (a *Account) Settings() *AccountSettings { return MustAccountSettingsFromProto(a.pb.Settings) }

func AccountFromProto(pb *pbaccount.Account) (*Account, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&Account{pb: pb}).Clone(), nil
}

func (a *Account) Name() AccountName { return AccountName(a.pb.Name) }

func NewAccount(name AccountName, d *AccountSettings, createdAt time.Time, updatedAt *time.Time) (*Account, error) {
	var pbupdatedat *timestamppb.Timestamp
	if updatedAt != nil {
		pbupdatedat = timestamppb.New(*updatedAt)
	}

	return AccountFromProto(
		&pbaccount.Account{
			Name:      string(name),
			Settings:  d.PB(),
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: pbupdatedat,
		},
	)
}
