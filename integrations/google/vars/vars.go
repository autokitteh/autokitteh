package vars

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Vars struct {
	OAuthData string `var:"secret"`
	JSON      string `var:"secret"`

	CalendarID string

	FormID               string
	FormResponsesWatchID string
	FormSchemaWatchID    string
}

var (
	AuthType = sdktypes.NewSymbol("auth_type")

	OAuthData = sdktypes.NewSymbol("OAuthData")
	JSON      = sdktypes.NewSymbol("JSON")

	CalendarID               = sdktypes.NewSymbol("CalendarID")
	CalendarEventsWatchID    = sdktypes.NewSymbol("CalendarEventsWatchID")
	CalendarEventsWatchResID = sdktypes.NewSymbol("CalendarEventsWatchResID")
	CalendarEventsWatchExp   = sdktypes.NewSymbol("CalendarEventsWatchExp")
	CalendarEventsSyncToken  = sdktypes.NewSymbol("CalendarEventsSyncToken")

	DriveChangesStartPageToken = sdktypes.NewSymbol("DriveChangesStartPageToken")
	DriveEventsWatchExp        = sdktypes.NewSymbol("DriveEventsWatchExp")
	DriveEventsWatchID         = sdktypes.NewSymbol("DriveEventsWatchID")
	DriveEventsWatchResID      = sdktypes.NewSymbol("DriveEventsWatchResID")

	FormID                = sdktypes.NewSymbol("FormID")
	FormResponsesWatchID  = sdktypes.NewSymbol("FormResponsesWatchID")
	FormSchemaWatchID     = sdktypes.NewSymbol("FormSchemaWatchID")
	FormWatchesExpiration = sdktypes.NewSymbol("FormWatchesExpiration")

	UserEmail = sdktypes.NewSymbol("user_email")
	UserScope = sdktypes.NewSymbol("user_scope")
)
