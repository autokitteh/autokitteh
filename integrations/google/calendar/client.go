package calendar

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type api struct {
	logger *zap.Logger
	vars   sdkservices.Vars
	cid    sdktypes.ConnectionID
}

var IntegrationID = sdktypes.NewIntegrationIDFromName("googlecalendar")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: IntegrationID.String(),
	UniqueName:    "googlecalendar",
	DisplayName:   "Google Calendar",
	Description:   "Google Calendar is a time-management and scheduling calendar service developed by Google.",
	LogoUrl:       "/static/images/google_calendar.svg",
	ConnectionUrl: "/googlecalendar/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.Empty,
		connections.ConnStatus(cvars),
		connections.ConnTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// Extract the calendar ID from the connection's vars.
// Return an empty string if the calendar ID wasn't set (i.e. do nothing).
func (a api) calendarID(ctx context.Context) (string, error) {
	data, err := a.connectionData(ctx)
	if err != nil {
		return "", err
	}

	return data.CalendarID, nil
}

func (a api) calendarClient(ctx context.Context) (*calendar.Service, error) {
	data, err := a.connectionData(ctx)
	if err != nil {
		return nil, err
	}

	var src oauth2.TokenSource
	if data.OAuthData != "" {
		if src, err = oauthTokenSource(ctx, data.OAuthData); err != nil {
			return nil, err
		}
	} else {
		src, err = jwtTokenSource(ctx, data.JSON)
		if err != nil {
			return nil, err
		}
	}

	svc, err := calendar.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func oauthTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	d, err := sdkintegrations.DecodeOAuthData(data)
	if err != nil {
		return nil, err
	}

	return oauthConfig().TokenSource(ctx, d.Token), nil
}

// TODO(ENG-112): Use OAuth().Get() instead of calling this function.
func oauthConfig() *oauth2.Config {
	addr := os.Getenv("WEBHOOK_ADDRESS")
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("https://%s/oauth/redirect/google", addr),
		// https://developers.google.com/calendar/api/auth
		Scopes: []string{
			// Non-sensitive.
			googleoauth2.OpenIDScope,
			googleoauth2.UserinfoEmailScope,
			googleoauth2.UserinfoProfileScope,
			// Sensitive.
			calendar.CalendarScope,
			calendar.CalendarEventsScope,
		},
	}
}

func jwtTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	scopes := oauthConfig().Scopes

	cfg, err := google.JWTConfigFromJSON([]byte(data), scopes...)
	if err != nil {
		return nil, err
	}

	return cfg.TokenSource(ctx), nil
}

func (a api) connectionData(ctx context.Context) (*vars.Vars, error) {
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !cid.IsValid() {
		cid = a.cid // Fallback during authentication flows.
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}

	var decoded vars.Vars
	vs.Decode(&decoded)
	return &decoded, nil
}

// https://developers.google.com/calendar/api/guides/push
// https://developers.google.com/calendar/api/v3/reference/events/watch
func (a api) watchEvents(ctx context.Context, connID sdktypes.ConnectionID, userEmail, calID string) (*calendar.Channel, error) {
	client, err := a.calendarClient(ctx)
	if err != nil {
		return nil, err
	}

	addr := os.Getenv("WEBHOOK_ADDRESS")
	req := client.Events.Watch(calID, &calendar.Channel{
		Id:      connID.String() + "/events",
		Token:   fmt.Sprintf("%s/%s/events", userEmail, calID),
		Address: fmt.Sprintf("https://%s/googlecalendar/notif", addr),
		Type:    "web_hook",
	})

	resp, err := req.Do()
	if err == nil {
		return resp, nil
	}

	gerr, ok := err.(*googleapi.Error)
	if !ok || gerr.Code != 400 || len(gerr.Errors) != 1 {
		return nil, err
	}
	if gerr.Errors[0].Reason != "channelIdNotUnique" {
		return nil, err
	}

	// If the channel already exists, stop and recreate it.
	a.logger.Info("Google Calendar watch channel already exists - stopping and recreating")
	if err := a.stopWatch(ctx, connID); err != nil {
		return nil, err
	}

	resp, err = req.Do()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// https://developers.google.com/calendar/api/v3/reference/channels/stop
func (a api) stopWatch(ctx context.Context, cid sdktypes.ConnectionID) error {
	client, err := a.calendarClient(ctx)
	if err != nil {
		return err
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.CalendarEventsWatchResID)
	if err != nil {
		return err
	}

	err = client.Channels.Stop(&calendar.Channel{
		Id:         cid.String() + "/events",
		ResourceId: vs.Get(vars.CalendarEventsWatchResID).Value(),
	}).Do()
	if err != nil {
		return err
	}
	return nil
}

// https://developers.google.com/calendar/api/guides/sync
// TODO(ENG-1499): Store event data in DB, to support event start/end notifs.
func (a api) syncEvents(ctx context.Context, calID string) error {
	client, err := a.calendarClient(ctx)
	if err != nil {
		return err
	}

	// Initial request.
	req := client.Events.List(calID).ShowHiddenInvitations(true).SingleEvents(true)
	req = req.TimeMin(time.Now().UTC().Format(time.RFC3339))
	resp, err := req.Do()
	if err != nil {
		return err
	}

	// https://developers.google.com/calendar/api/guides/pagination
	for resp.NextPageToken != "" {
		a.logger.Debug("Requesting next page of Google Calendar events",
			zap.String("pageToken", resp.NextPageToken),
		)

		req = req.PageToken(resp.NextPageToken)
		resp, err = req.Do()
		if err != nil {
			return err
		}
	}

	// Update the sync token for the next request.
	v := sdktypes.NewVar(vars.CalendarEventsSyncToken).SetValue(resp.NextSyncToken)
	if err = a.vars.Set(ctx, v.WithScopeID(sdktypes.NewVarScopeID(a.cid))); err != nil {
		return err
	}
	a.logger.Debug("Google Calendar connection's new events sync token",
		zap.String("cid", a.cid.String()),
		zap.String("syncToken", resp.NextSyncToken),
	)

	return nil
}

// https://developers.google.com/calendar/api/v3/reference/events/list
func (a api) listEvents(ctx context.Context) ([]*calendar.Event, error) {
	client, err := a.calendarClient(ctx)
	if err != nil {
		return nil, err
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(a.cid))
	if err != nil {
		return nil, err
	}

	// https://developers.google.com/calendar/api/guides/sync
	// TODO: Does this fail too because there's no AK user ID?
	syncToken := vs.Get(vars.CalendarEventsSyncToken).Value()
	a.logger.Debug("Google Calendar connection's existing events sync token",
		zap.String("cid", a.cid.String()),
		zap.String("syncToken", syncToken),
	)

	var events []*calendar.Event

	// TODO: Does this fail too because there's no AK user ID?
	req := client.Events.List(vs.Get(vars.CalendarID).Value())
	req = req.ShowHiddenInvitations(true).SingleEvents(true)
	// TODO: Does this fail too because there's no AK user ID?
	resp, err := req.SyncToken(vs.Get(vars.CalendarEventsSyncToken).Value()).Do()
	if err != nil {
		return nil, err
	}
	events = append(events, resp.Items...)

	// https://developers.google.com/calendar/api/guides/pagination
	for resp.NextPageToken != "" {
		a.logger.Debug("Requesting next page of Google Calendar events",
			zap.String("pageToken", resp.NextPageToken),
		)

		req = req.PageToken(resp.NextPageToken)
		resp, err = req.Do()
		if err != nil {
			return nil, err
		}

		events = append(events, resp.Items...)
	}

	// TODO(ENG-952): Update the sync token for the next request.
	// (Currently blocked because this handler can't have an authenticated AK user ID).
	// v := sdktypes.NewVar(vars.CalendarEventsSyncToken).SetValue(resp.NextSyncToken)
	// if err = a.vars.Set(ctx, v.WithScopeID(sdktypes.NewVarScopeID(a.cid))); err != nil {
	// 	return nil, err
	// }
	// a.logger.Debug("Google Calendar connection's new events sync token",
	// 	zap.String("cid", a.cid.String()),
	// 	zap.String("syncToken", resp.NextSyncToken),
	// )

	return events, nil
}
