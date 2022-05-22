package projectsstore

import (
	"context"
	"errors"

	"github.com/autokitteh/autokitteh/sdk/api/apiaccount"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrInvalidAccount = errors.New("account invalid or missing")
)

const AutoProjectID = apiproject.ProjectID("")

type Store interface {
	Create(context.Context, apiaccount.AccountName, apiproject.ProjectID, *apiproject.ProjectSettings) (apiproject.ProjectID, error)
	Update(context.Context, apiproject.ProjectID, *apiproject.ProjectSettings) error
	Get(context.Context, apiproject.ProjectID) (*apiproject.Project, error)
	BatchGet(context.Context, []apiproject.ProjectID) (map[apiproject.ProjectID]*apiproject.Project, error)
	Setup(context.Context) error
	Teardown(context.Context) error
}
