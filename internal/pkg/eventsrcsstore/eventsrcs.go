package eventsrcsstore

import (
	"context"
	"errors"

	"github.com/autokitteh/autokitteh/sdk/api/apiaccount"
	"github.com/autokitteh/autokitteh/sdk/api/apieventsrc"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store interface {
	Add(context.Context, apieventsrc.EventSourceID, *apieventsrc.EventSourceSettings) error
	Update(context.Context, apieventsrc.EventSourceID, *apieventsrc.EventSourceSettings) error
	Get(context.Context, apieventsrc.EventSourceID) (*apieventsrc.EventSource, error)
	List(context.Context, *apiaccount.AccountName) ([]apieventsrc.EventSourceID, error)

	// TODO: remove.
	AddProjectBinding(
		_ context.Context,
		_ apieventsrc.EventSourceID,
		_ apiproject.ProjectID,
		name string,
		assoc string,
		cfg string,
		approved bool,
		_ *apieventsrc.EventSourceProjectBindingSettings,
	) error

	UpdateProjectBinding(context.Context, apieventsrc.EventSourceID, apiproject.ProjectID, string, bool, *apieventsrc.EventSourceProjectBindingSettings) error
	GetProjectBindings(context.Context, *apieventsrc.EventSourceID, *apiproject.ProjectID, string, string, bool) ([]*apieventsrc.EventSourceProjectBinding, error)

	Setup(context.Context) error
	Teardown(context.Context) error
}
