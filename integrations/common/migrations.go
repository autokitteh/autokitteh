package common

import (
	"context"
	"regexp"
	"time"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// RenameVar renames a variable in the given connection scope. It does nothing if
// the variable doesn't already exist. This is useful for non-trivial data migrations.
func RenameVar(ctx context.Context, v sdkservices.Vars, vsid sdktypes.VarScopeID, old, new sdktypes.Symbol) error {
	vs, err := v.Get(ctx, vsid, old)
	if err != nil {
		return err
	}

	o := vs.Get(old)
	if !o.IsValid() {
		return nil
	}

	n := sdktypes.NewVar(new).SetValue(o.Value()).SetSecret(o.IsSecret())
	if err := v.Set(ctx, n.WithScopeID(vsid)); err != nil {
		return err
	}

	return v.Delete(ctx, vsid, old)
}

// MigrateAuthType migrates a connection's "auth_type" variable from the old
// "oauth" value to the new "oauthDefault". It is assumed that the variable exists.
func MigrateAuthType(ctx context.Context, v sdkservices.Vars, vsid sdktypes.VarScopeID) error {
	vs, err := v.Get(ctx, vsid, AuthTypeVar)
	if err != nil {
		return err
	}

	o := vs.Get(AuthTypeVar)
	if !o.IsValid() {
		return nil
	}

	if o.Value() == integrations.OAuth {
		o = o.SetValue(integrations.OAuthDefault)
		return v.Set(ctx, o.WithScopeID(vsid))
	}

	return nil
}

// MigrateDateTimeToRFC3339 migrates a connection's timestamp variable
// from the Time.String() format to RFC-3339 format. It is assumed that
// the variable exists. If the variable is already in RFC-3339 format, it is
// left unchanged. An otherwise unrecognized format is reported as an error.
func MigrateDateTimeToRFC3339(ctx context.Context, v sdkservices.Vars, vsid sdktypes.VarScopeID, varName sdktypes.Symbol) error {
	vs, err := v.Get(ctx, vsid, varName)
	if err != nil {
		return err
	}

	o := vs.Get(varName)
	if !o.IsValid() {
		return nil
	}

	s := o.Value()
	if _, err := time.Parse(time.RFC3339, s); err == nil {
		return nil
	}

	s = regexp.MustCompile(` [A-Z].*`).ReplaceAllString(s, "")
	t, err := time.Parse("2006-01-02 15:04:05 -0700", s)
	if err != nil {
		return err
	}

	o = o.SetValue(t.UTC().Format(time.RFC3339))
	return v.Set(ctx, o.WithScopeID(vsid))
}
