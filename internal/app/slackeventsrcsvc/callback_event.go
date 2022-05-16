package slackeventsrcsvc

import (
	"context"
	"fmt"

	"github.com/slack-go/slack/slackevents"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
)

func (s *Svc) handleCallbackEvent(ctx context.Context, data *slackevents.EventsAPICallbackEvent, inner *slackevents.EventsAPIInnerEvent) {
	l := s.L.With("team_id", data.TeamID, "type", data.Type)

	oauthToken, err := s.getOAuthToken(ctx, apiproject.EmptyProjectID, data.TeamID)
	if err != nil {
		l.Error("get oauth tokens error", "err", err)
		return
	}

	evtData := map[string]*apivalues.Value{
		// TODO: make sure this is hidden in any ui
		"secret_access_token": apivalues.Bytes(oauthToken),
		"team_id":             apivalues.String(data.TeamID),
	}

	if err := apivalues.WrapIntoValuesMap(evtData, inner.Data); err != nil {
		l.Error("event data wrap error", "err", err)
		return
	}

	id, err := s.Events.IngestEvent(
		ctx,
		s.Config.EventSourceID,
		/* assoc */ data.TeamID,
		/* originalID */ data.EventID,
		/* type */ inner.Type,
		evtData,
		map[string]string{
			"description": fmt.Sprintf("%s: %s", data.TeamID, inner.Type),
		},
	)

	if err != nil {
		l.Error("ingest error", "err", err)
		return
	}

	l.Debug("ingested", "id", id)
}
