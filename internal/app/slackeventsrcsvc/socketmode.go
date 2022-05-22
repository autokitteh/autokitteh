package slackeventsrcsvc

import (
	"context"
	"time"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	L "github.com/autokitteh/L"
)

func (s *Svc) startSocketMode() {
	l := s.L.Named("socketmode")

	sm := socketmode.New(
		s.client,
		socketmode.OptionDebug(s.Config.Debug),
		socketmode.OptionLog(wrapLogger(l)),
	)

	go s.socketModeHandler(l, sm)

	go func() {
		l := l.Named("sockermoderun")

		for {
			l.Debug("running")

			if err := sm.Run(); err != nil {
				l.Error("run error", "err", err)
				time.Sleep(time.Second)
				continue
			}

			l.Warn("run returned with no error - assuming graceful termination")
			return
		}
	}()
}

func (s *Svc) socketModeHandler(l L.L, sm *socketmode.Client) {
	l.Debug("consuming events")

	for evt := range sm.Events {
		l.Debug("received event", "event", evt)

		l.Debug("outer", "outer", evt.Data)

		switch outer := evt.Data.(type) {
		case slackevents.EventsAPIEvent:
			inner := outer.InnerEvent

			data := outer.Data.(*slackevents.EventsAPICallbackEvent)

			l.Debug("inner", "inner", inner)

			s.handleCallbackEvent(context.Background(), data, &inner)

			sm.Ack(*evt.Request)
		}
	}
}
