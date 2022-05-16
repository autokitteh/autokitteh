package httpeventsrc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/ucarion/urlpath"

	pb "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/httpeventsrc"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apievent"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/events"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsrcsstore"
	L "gitlab.com/softkitteh/autokitteh/pkg/l"
)

var EventTypes = []string{"get", "put", "post", "delete", "patch", "head"}

type Route = pb.Route

type Config struct {
	EventSourceID apieventsrc.EventSourceID `envconfig:"EVENT_SOURCE_ID" json:"event_source_id"`
}

type bindingConfig struct {
	Name   string      `json:"name"`
	Routes []*pb.Route `json:"routes"`
}

type HTTPEventSource struct {
	Config       Config
	Events       *events.Events
	EventSources eventsrcsstore.Store
	Prefix       string
	L            L.Nullable
}

func (s *HTTPEventSource) Remove(
	ctx context.Context,
	pid apiproject.ProjectID,
	name string,
) error {
	// nop
	return nil
}

func (s *HTTPEventSource) Add(
	ctx context.Context,
	pid apiproject.ProjectID,
	name string,
	routes []*pb.Route,
) error {
	b := bindingConfig{Name: name, Routes: routes}

	cfg, err := json.Marshal(&b)
	if err != nil {
		s.L.Panic("binding marshal error", "err", err)
	}

	if err := s.EventSources.AddProjectBinding(
		ctx,
		s.Config.EventSourceID,
		pid,
		name,
		fmt.Sprintf("%v.%s", pid, name),
		string(cfg),
		true,
		(&apieventsrc.EventSourceProjectBindingSettings{}).SetEnabled(true),
	); err != nil {
		return fmt.Errorf("add binding: %w", err)
	}

	return nil
}

func match(path string, method string, cfgs map[string]*bindingConfig) (string, string, map[string]string, string, bool) {
	method = strings.ToUpper(method)

	for name, cfg := range cfgs {
		for _, r := range cfg.Routes {
			p := urlpath.New(r.Path)
			if m, ok := p.Match(path); ok {
				if len(r.Methods) != 0 {
					found := false
					for _, meth := range r.Methods {
						if method == strings.ToUpper(meth) {
							found = true
							break
						}
					}

					if !found {
						continue
					}
				}

				return name, r.Name, m.Params, m.Trailing, true
			}
		}
	}

	return "", "", nil, "", false
}

func values(vs url.Values) *apivalues.Value {
	var d apivalues.DictValue

	for k, vs := range vs {
		di := apivalues.DictItem{
			K: apivalues.String(k),
		}

		ivs := make(apivalues.ListValue, len(vs))
		for i, v := range vs {
			ivs[i] = apivalues.String(v)
		}

		di.V = apivalues.MustNewValue(ivs)

		d = append(d, &di)
	}

	return apivalues.MustNewValue(d)
}

// returns "", nil for not found.
func (s *HTTPEventSource) Handle(req *http.Request) (apievent.EventID, error) {
	if s.Config.EventSourceID.Empty() {
		return "", httpError(http.StatusNotImplemented, "event source not configured")
	}

	// TODO: rate limit.

	path := strings.TrimPrefix(req.URL.Path, s.Prefix)
	if path[0] == '/' {
		path = path[1:]
	}

	pathParts := strings.SplitN(path, "/", 2)

	pid := apiproject.ProjectID(pathParts[0])

	var rest string

	if len(pathParts) > 1 {
		rest = pathParts[1]
	}

	l := s.L.With("pid", pid)

	l.Debug("got http request", "path", rest, "method", req.Method)

	bs, err := s.EventSources.GetProjectBindings(req.Context(), &s.Config.EventSourceID, &pid, "", "", true)
	if err != nil {
		if errors.Is(err, eventsrcsstore.ErrNotFound) {
			return "", httpError(http.StatusNotFound, "project not found")
		}

		return "", httpError(http.StatusInternalServerError, "%v", err)
	}

	cfgs := make(map[string]*bindingConfig, len(bs))
	for _, b := range bs {
		var cfg bindingConfig

		if err := json.Unmarshal([]byte(b.SourceConfig()), &cfg); err != nil {
			l.Error("unmarshal config error", "name", b.Name(), "err", err)
			continue
		}

		cfgs[b.Name()] = &cfg
	}

	bindingName, routeName, params, trailing, found := match(rest, req.Method, cfgs)

	if !found {
		l.Debug("no matching route")

		return "", httpError(http.StatusNotFound, "no matching route")
	}

	paramsValue, err := apivalues.Wrap(params)
	if err != nil {
		return "", httpError(http.StatusInternalServerError, "params error: %v", err)
	}

	l.Debug("matched route", "route_name", routeName, "params", params)

	// TODO: limit body size.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", httpError(http.StatusInternalServerError, "body read error: %v", err)
	}

	data := map[string]*apivalues.Value{
		// TODO: form values (probably need to be setup as form in the source db).
		"method":       apivalues.String(req.Method),
		"path":         apivalues.String(rest),
		"trailing":     apivalues.String(trailing),
		"body":         apivalues.String(string(body)), // TODO: any issues with string cast here? security?
		"route_name":   apivalues.String(routeName),
		"binding_name": apivalues.String(bindingName),
		"params":       paramsValue,
		"raw_query":    apivalues.String(req.URL.RawQuery),
		"query_values": values(req.URL.Query()),
	}

	id, err := s.Events.IngestEvent(
		req.Context(),
		s.Config.EventSourceID,
		fmt.Sprintf("%v.%s", pid, bindingName),
		/* originalID */ path, // TODO: get from header? If so, need to configure how.
		/* type */ req.Method,
		data,
		map[string]string{
			"description": fmt.Sprintf("%s %v/%s (%s)", req.Method, pid, rest, routeName),
		},
	)
	if err != nil {
		return "", httpError(http.StatusInternalServerError, "ingestion error: %v", err)
	}

	return id, nil
}
