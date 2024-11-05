package vars

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Vars struct {
	OAuthData string `var:"secret"`
	JSON      string `var:"secret"`

	CalendarID           string
	DriveID              string
	FormID               string
	FormResponsesWatchID string
	FormSchemaWatchID    string
}

var (
	AuthType = sdktypes.NewSymbol("authType")

	OAuthData = sdktypes.NewSymbol("OAuthData")
	JSON      = sdktypes.NewSymbol("JSON")

	CalendarID               = sdktypes.NewSymbol("CalendarID")
	CalendarEventsWatchID    = sdktypes.NewSymbol("CalendarEventsWatchID")
	CalendarEventsWatchResID = sdktypes.NewSymbol("CalendarEventsWatchResID")
	CalendarEventsWatchExp   = sdktypes.NewSymbol("CalendarEventsWatchExp")
	CalendarEventsSyncToken  = sdktypes.NewSymbol("CalendarEventsSyncToken")

	DriveID                    = sdktypes.NewSymbol("DriveID")
	DriveEventsWatchID         = sdktypes.NewSymbol("DriveEventsWatchID")
	DriveEventsWatchResID      = sdktypes.NewSymbol("DriveEventsWatchResID")
	DriveEventsWatchExp        = sdktypes.NewSymbol("DriveEventsWatchExp")
	DriveChangesStartPageToken = sdktypes.NewSymbol("DriveChangesStartPageToken")

	FormID               = sdktypes.NewSymbol("FormID")
	FormResponsesWatchID = sdktypes.NewSymbol("FormResponsesWatchID")
	FormSchemaWatchID    = sdktypes.NewSymbol("FormSchemaWatchID")

	UserEmail = sdktypes.NewSymbol("user_email")
	UserScope = sdktypes.NewSymbol("user_scope")
)
