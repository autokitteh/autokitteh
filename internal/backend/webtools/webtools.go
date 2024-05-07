package webtools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/backend/webtools/web"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const defaultAddr = "default"

type Svc interface {
	Init(muxes *muxes.Muxes)
	Setup(ctx context.Context) error
}

type svc struct {
	cfg *Config
	db  *terminalDB
}

func New(cfg *Config, db db.DB) (Svc, error) {
	if !cfg.Enabled {
		return &svc{cfg: cfg}, nil
	}

	return &svc{db: newDB(db), cfg: cfg}, nil
}

func (svc *svc) Setup(ctx context.Context) error {
	if !svc.cfg.Enabled {
		return nil
	}

	return svc.db.Setup(ctx)
}

func (svc *svc) Init(muxes *muxes.Muxes) {
	if !svc.cfg.Enabled {
		return
	}

	// messages api paths.
	muxes.Auth.HandleFunc("GET /webtools/api/msgs/{addr}", svc.getMessages)
	muxes.Auth.HandleFunc("GET /webtools/api/msgs", svc.getMessages)
	muxes.Auth.HandleFunc("POST /webtools/api/msgs/{addr}", svc.postMessage)
	muxes.Auth.HandleFunc("POST /webtools/api/msgs", svc.postMessage)
	muxes.Auth.HandleFunc("DELETE /webtools/api/msgs/{addr}/{id}", svc.deleteMessage)
	muxes.Auth.HandleFunc("DELETE /webtools/api/msgs/{addr}", svc.deleteMessage)

	// messages app paths.
	muxes.Auth.Handle("/webtools/msgs", http.RedirectHandler("/webtools/msgs/"+defaultAddr, http.StatusFound))
	muxes.Auth.Handle("/webtools/msgs/{addr}/*", kittehs.StripWildcardPrefix("/webtools/msgs/{addr}", http.FileServer(http.FS(web.Messages))))

	muxes.Auth.HandleFunc("GET /webtools/msgs/{addr}/msgs", svc.getMessages)
	muxes.Auth.HandleFunc("POST /webtools/msgs/{addr}/msgs", svc.postMessage)
	muxes.Auth.HandleFunc("DELETE /webtools/msgs/{addr}/msgs/{id}", svc.deleteMessage)
	muxes.Auth.HandleFunc("DELETE /webtools/msgs/{addr}/msgs", svc.deleteMessage)

	// terminal paths.
	muxes.Auth.Handle("/webtools/terminal", http.RedirectHandler("/webtools/terminal/"+defaultAddr, http.StatusFound))
	muxes.Auth.Handle("/webtools/terminal/{addr}/*", kittehs.StripWildcardPrefix("/webtools/terminal/{addr}", http.FileServer(http.FS(web.Terminal))))

	muxes.Auth.HandleFunc("GET /webtools/terminal/{addr}/msgs", svc.getMessages)
	muxes.Auth.HandleFunc("POST /webtools/terminal/{addr}/msgs", svc.postMessage)
	muxes.Auth.HandleFunc("DELETE /webtools/terminal/{addr}/msgs/{id}", svc.deleteMessage)
	muxes.Auth.HandleFunc("DELETE /webtools/terminal/{addr}/msgs", svc.deleteMessage)
}

func (s *svc) getMessages(w http.ResponseWriter, r *http.Request) {
	addr := r.PathValue("addr")
	if addr == "" {
		addr = defaultAddr
	}

	msgs, err := s.db.GetMessages(r.Context(), addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(msgs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *svc) postMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	addr := r.PathValue("addr")
	if addr == "" {
		addr = defaultAddr
	}

	id, err := s.db.AddMessage(r.Context(), addr, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	resp := struct {
		ID uint `json:"id"`
	}{ID: id}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *svc) deleteMessage(w http.ResponseWriter, r *http.Request) {
	var id int

	if idStr := r.PathValue("id"); idStr != "" && idStr != "all" {
		var err error
		if id, err = strconv.Atoi(idStr); err != nil || id < 0 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if err := s.db.DeleteMessage(r.Context(), r.PathValue("addr"), uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
