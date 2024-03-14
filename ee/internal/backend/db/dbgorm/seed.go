package dbgorm

import (
	"context"
	"errors"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var seeds = []func(ctx context.Context, db *gormdb) error{
	// Create Autokitteh org
	func(ctx context.Context, db *gormdb) error {
		if _, err := db.GetOrgByName(ctx, fixtures.AutokittehOrgName); err != nil {
			if !errors.Is(err, sdkerrors.ErrNotFound) {
				return err
			}
			if err = db.CreateOrg(ctx, kittehs.Must1(sdktypes.OrgFromProto(
				&sdktypes.OrgPB{
					Name:  fixtures.AutokittehOrgName.String(),
					OrgId: fixtures.AutokittehOrgID.String(),
				},
			))); err != nil {
				return err
			}
		}
		return nil
	},
	// Create anonymous user
	func(ctx context.Context, db *gormdb) error {
		if _, err := db.GetUserByID(ctx, fixtures.AutokittehAnonymousUserID); err != nil {
			if !errors.Is(err, sdkerrors.ErrNotFound) {
				return err
			}
			if err = db.CreateUser(ctx, kittehs.Must1(sdktypes.UserFromProto(
				&sdktypes.UserPB{
					Name:   fixtures.AutokittehAnonymousUserName,
					UserId: fixtures.AutokittehAnonymousUserID.String(),
				}))); err != nil {
				return err
			}

		}
		return nil
	},
}
