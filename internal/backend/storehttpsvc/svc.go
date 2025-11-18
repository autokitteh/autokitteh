package storehttpsvc

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// An unwrapper that is always safe to serialize to string afterwards.
var unwrapper = sdktypes.ValueWrapper{
	SafeForJSON: true,
}

const storePathPrefix = "store/"

type svc struct {
	l  *zap.Logger
	db db.DB
}

func Init(muxes *muxes.Muxes, db db.DB, l *zap.Logger) {
	s := &svc{db: db, l: l}
	muxes.NoAuth.Handle("/"+storePathPrefix+"{pid}/{key}", s)
}

func (s *svc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pid, err := sdktypes.StrictParseProjectID(r.PathValue("pid"))
	if err != nil {
		http.Error(w, "invalid project ID", http.StatusBadRequest)
		return
	}

	key := r.PathValue("key")

	l := s.l.With(
		zap.String("pid", pid.String()),
		zap.String("key", key),
	)

	published, err := s.db.IsStoreValuePublished(r.Context(), pid, key)
	if errors.Is(err, sdkerrors.ErrNotFound) || !published {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err != nil {
		l.Error("failed to check if store value is published", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	v, err := s.db.GetStoreValue(r.Context(), pid, key)
	if err != nil {
		l.Error("failed to get store value", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	u, err := unwrapper.Unwrap(v)
	if err != nil {
		l.Error("failed to unwrap store value", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(u); err != nil {
		l.Error("failed to encode store value", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
