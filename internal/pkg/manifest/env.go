package manifest

import (
	"context"
	"fmt"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/pluginsreg"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
)

type Env struct {
	EventSources eventsrcsstore.Store
	Plugins      *pluginsreg.Registry
	Projects     projectsstore.Store
	Accounts     accountsstore.Store
}

func (e *Env) Apply(ctx context.Context, actions []*Action) (log []string, _ error) {
	for _, a := range actions {
		result, err := a.Run(ctx, e)
		if err != nil {
			err := fmt.Errorf("%s: %s", a.Desc, err)
			return append(log, err.Error()), err
		}

		if result != "" {
			result = ": " + result
		}

		log = append(log, fmt.Sprintf("%s%s", a.Desc, result))
	}

	return
}
