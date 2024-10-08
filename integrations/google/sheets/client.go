package sheets

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	googleScope = "google"
)

type api struct {
	vars sdkservices.Vars
}

var integrationID = sdktypes.NewIntegrationIDFromName("googlesheets")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "googlesheets",
	DisplayName:   "Google Sheets",
	Description:   "Google Sheets is a web-based spreadsheet application that is part of the Google Workspace office suite.",
	LogoUrl:       "/static/images/google_sheets.svg",
	ConnectionUrl: "/googlesheets/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	scope := googleScope

	opts := ExportedFunctions(cvars, scope, false)

	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(opts...),
		connections.ConnStatus(cvars),
		connections.ConnTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars))
}

func ExportedFunctions(cvars sdkservices.Vars, scope string, prefix bool) []sdkmodule.Optfn {
	a := api{vars: cvars}
	return []sdkmodule.Optfn{
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "a1_range"),
			a.a1Range,
			sdkmodule.WithFuncDoc("https://developers.google.com/sheets/api/guides/concepts#expandable-1"),
			sdkmodule.WithArgs("sheet_name?", "from?", "to?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "read_cell"),
			a.readCell,
			sdkmodule.WithFuncDesc("Read a single cell"),
			sdkmodule.WithArgs("spreadsheet_id", "sheet_name?", "row_index", "col_index", "value_render_option?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "read_range"),
			a.readRange,
			sdkmodule.WithFuncDesc("Read a range of cells"),
			sdkmodule.WithArgs("spreadsheet_id", "a1_range", "value_render_option?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "set_background_color"),
			a.setBackgroundColor,
			sdkmodule.WithFuncDesc("Set the background color in a range of cells"),
			sdkmodule.WithArgs("spreadsheet_id", "a1_range", "color")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "set_text_format"),
			a.setTextFormat,
			sdkmodule.WithFuncDesc("Set the text format in a range of cells"),
			sdkmodule.WithArgs("spreadsheet_id", "a1_range", "color?", "bold?", "italic?", "strikethrough?", "underline?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "write_cell"),
			a.writeCell,
			sdkmodule.WithFuncDesc("Write a single of cell"),
			sdkmodule.WithArgs("spreadsheet_id", "sheet_name?", "row_index", "col_index", "value")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "write_range"),
			a.writeRange,
			sdkmodule.WithFuncDesc("Write a range of cells"),
			sdkmodule.WithArgs("spreadsheet_id", "a1_range", "data")),
	}
}

func withOrWithout(prefix bool, s string) string {
	if prefix {
		return "sheets_" + s
	}
	return s
}

func (a api) sheetsClient(ctx context.Context) (*sheets.Service, error) {
	data, err := a.connectionData(ctx)
	if err != nil {
		return nil, err
	}

	var src oauth2.TokenSource
	if data.OAuthData != "" {
		if src, err = a.oauthTokenSource(ctx, data.OAuthData); err != nil {
			return nil, err
		}
	} else {
		src, err = a.jwtTokenSource(ctx, data.JSON)
		if err != nil {
			return nil, err
		}
	}

	svc, err := sheets.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (a api) connectionData(ctx context.Context) (*vars.Vars, error) {
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}

	var vars vars.Vars
	vs.Decode(&vars)
	return &vars, nil
}

func (a api) oauthTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	tok, err := sdkintegrations.DecodeOAuthData(data)
	if err != nil {
		return nil, err
	}

	return oauthConfig().TokenSource(ctx, tok.Token), nil
}

// TODO(ENG-112): Use OAuth().Get() instead of calling this function.
func oauthConfig() *oauth2.Config {
	addr := os.Getenv("WEBHOOK_ADDRESS")
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("https://%s/oauth/redirect/google", addr),
		Scopes: []string{
			googleoauth2.OpenIDScope,
			googleoauth2.UserinfoEmailScope,
			googleoauth2.UserinfoProfileScope,
			sheets.SpreadsheetsScope,
		},
	}
}

func (a api) jwtTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	scopes := oauthConfig().Scopes

	cfg, err := google.JWTConfigFromJSON([]byte(data), scopes...)
	if err != nil {
		return nil, err
	}

	return cfg.TokenSource(ctx), nil
}
