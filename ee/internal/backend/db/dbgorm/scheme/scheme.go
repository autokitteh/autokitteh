package scheme

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(ENG-192): use proper foreign keys and normalize model.

// TODO: keep some log of actions performed. Something that
// can be used for recovery from unintended/malicious actions.

var Tables = []any{
	&Org{},
	&OrgMember{},
	&User{},
	&UserExternalIdentitiy{},
}

type Org struct {
	OrgID string `gorm:"primaryKey"`
	Name  string `gorm:"uniqueIndex"`
}

func ParseOrg(r Org) (sdktypes.Org, error) {
	org, err := sdktypes.StrictOrgFromProto(&sdktypes.OrgPB{
		OrgId: r.OrgID,
		Name:  r.Name,
	})
	if err != nil {
		return sdktypes.InvalidOrg, fmt.Errorf("invalid record: %w", err)
	}

	return org, nil
}

type OrgMember struct {
	// {oid.uuid}/{uid.uuid}. easier to detect dups.
	// For some reason gorm refuses to translate dup errors
	// for separate orgid and userid as a combined primary key.
	MembershipID string `gorm:"uniqueIndex"`

	OrgID  string
	Org    Org `gorm:"references:OrgID"`
	UserID string
	User   User `gorm:"references:UserID"`
}

type User struct {
	UserID      string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	DisplayName string
}

func ParseUser(r User) (sdktypes.User, error) {
	return sdktypes.StrictUserFromProto(&sdktypes.UserPB{
		UserId: r.UserID,
		Name:   r.Name,
	})
}

type UserExternalIdentitiy struct {
	UserExternalIdentitiyID string `gorm:"primaryKey"`
	ExternalID              string `gorm:"uniqueIndex"`
	IdentityType            string
	Email                   string
	Name                    string
	UserID                  string
	User                    User `gorm:"references:UserID"`
}
