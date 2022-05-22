package githubeventsrcsvc

import (
	"net/http"

	"github.com/google/go-github/v42/github"
	"github.com/iancoleman/strcase"

	"github.com/autokitteh/autokitteh/sdk/api/apivalues"

	H "github.com/autokitteh/autokitteh/pkg/h"
)

type parsedEvent struct {
	Installation *github.Installation        // installation and its id are always set
	Owner, Repo  string                      // owner is always set. repo not always.
	Data         map[string]*apivalues.Value // always set.
}

func parseEvent(event interface{}) (*parsedEvent, error) {
	pev := parsedEvent{
		Data: make(map[string]*apivalues.Value, 8),
	}

	populate := func(inst *github.Installation, repo *github.Repository) {
		pev.Installation = inst
		if repo != nil {
			if n := repo.Name; n != nil {
				pev.Repo = *n
			}

			if o := repo.Owner; o != nil {
				if l := o.Login; l != nil {
					pev.Owner = *l
				}
			}
		}
	}

	switch event := event.(type) {
	case *github.IssueCommentEvent:
		populate(event.Installation, event.Repo)

	case *github.IssuesEvent:
		populate(event.Installation, event.Repo)

	case *github.PullRequestEvent:
		populate(event.Installation, event.Repo)

	case *github.PullRequestReviewCommentEvent:
		populate(event.Installation, event.Repo)

	case *github.PullRequestReviewEvent:
		populate(event.Installation, event.Repo)

	case *github.PushEvent:
		populate(event.Installation, nil)

		// PushEvent has PushEventRepo which is special :-|.
		if repo := event.Repo; repo != nil {
			if n := repo.Name; n != nil {
				pev.Repo = *n
			}

			if o := repo.Owner; o != nil {
				if l := o.Login; l != nil {
					pev.Owner = *l
				}
			}
		}

	case *github.CheckRunEvent:
		populate(event.Installation, event.Repo)

	case *github.CheckSuiteEvent:
		populate(event.Installation, event.Repo)

	case *github.StatusEvent:
		populate(event.Installation, event.Repo)

	default:
		// OK instead of CREATED.
		return nil, H.NewError(http.StatusOK, "unhandled event")
	}

	if pev.Installation == nil || pev.Installation.ID == nil {
		return nil, H.NewError(http.StatusBadRequest, "no installation or installation id in event")
	}

	if pev.Owner == "" {
		return nil, H.NewError(http.StatusBadRequest, "no owner in event")
	}

	if err := apivalues.WrapIntoValuesMap(
		pev.Data,
		event,
		apivalues.WithStructFieldNameConverter(strcase.ToSnake),
	); err != nil {
		return nil, H.NewError(http.StatusInternalServerError, "wrap event values error", "err", err)
	}

	return &pev, nil
}
