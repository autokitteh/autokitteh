package dashboardsvc

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initEvents() {
	s.Muxes.Auth.HandleFunc("/events", s.events)
	s.Muxes.Auth.HandleFunc("/events/{eid}", s.event)
}

type event struct{ sdktypes.Event }

func (p event) FieldsOrder() []string {
	return []string{"created_at", "event_id", "connection_id", "integration_id"}
}

func (p event) HideFields() []string        { return nil }
func (p event) ExtraFields() map[string]any { return nil }

func toEvent(sdkP sdktypes.Event) event { return event{sdkP} }

func (s Svc) listEvents(w http.ResponseWriter, r *http.Request, f sdkservices.ListEventsFilter) (list, error) {
	if f.Limit <= 0 {
		f.Limit = 50

		if l := r.URL.Query().Get("max_events"); l != "" {
			l64, err := strconv.ParseInt(l, 10, 32)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return list{}, err

			}

			f.Limit = int(l64)
		}
	}

	if f.MinSequenceNumber == 0 {
		if l := r.URL.Query().Get("min_event_seq"); l != "" {
			l64, err := strconv.ParseInt(l, 10, 32)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return list{}, err

			}

			f.MinSequenceNumber = uint64(l64)
		}
	}

	sdkCs, err := s.Svcs.Events().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	var drops []string
	if f.IntegrationID.IsValid() {
		drops = append(drops, "integration_id")
	}

	if f.ConnectionID.IsValid() {
		drops = append(drops, "connection_id")
	}

	return genListData(f, kittehs.Transform(sdkCs, toEvent), drops...), nil
}

func (s Svc) events(w http.ResponseWriter, r *http.Request) {
	ts, err := s.listEvents(w, r, sdkservices.ListEventsFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "events", ts)
}

func (s Svc) event(w http.ResponseWriter, r *http.Request) {
	eid, err := sdktypes.ParseEventID(r.PathValue("eid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkE, err := s.Svcs.Events().Get(r.Context(), eid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rs, err := s.Svcs.Events().ListEventRecords(r.Context(), sdkservices.ListEventRecordsFilter{EventID: eid})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vw := sdktypes.DefaultValueWrapper
	vw.SafeForJSON = true

	data, err := vw.UnwrapMap(sdkE.Data())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "event.html", struct {
		Title     string
		ID        string
		EventJSON template.HTML
		LogJSON   template.HTML
		DataJSON  template.HTML
	}{
		Title:     "Event: " + sdkE.ID().String(),
		ID:        sdkE.ID().String(),
		EventJSON: marshalObject(sdkE.WithData(nil).ToProto()),
		LogJSON:   template.HTML(kittehs.Must1(kittehs.MarshalProtoSliceJSON(kittehs.Transform(rs, sdktypes.ToProto)))),
		DataJSON:  template.HTML(jsonData),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
