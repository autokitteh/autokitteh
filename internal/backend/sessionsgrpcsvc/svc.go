package sessionsgrpcsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1/sessionsv1connect"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	sessions sdkservices.Sessions

	sessionsv1connect.UnimplementedSessionsServiceHandler
}

var _ sessionsv1connect.SessionsServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, sessions sdkservices.Sessions) {
	srv := server{sessions: sessions}

	path, handler := sessionsv1connect.NewSessionsServiceHandler(&srv)
	muxes.Auth.Handle(path, handler)
}

type testStruct struct {
	Body struct {
		JSON json.RawMessage `json:"json"` // Preserve JSON for further unmarshaling
	} `json:"body"`
}

func (s *server) Start(ctx context.Context, req *connect.Request[sessionsv1.StartRequest]) (*connect.Response[sessionsv1.StartResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	wrappedValue := &valuesv1.Value{
		Struct: &valuesv1.Struct{
			Ctor: &valuesv1.Value{
				Symbol: &valuesv1.Symbol{
					Name: "data",
				},
			},
			Fields: make(map[string]*valuesv1.Value),
		},
	}

	// Process each JsonInputs field
	for key, jsonStr := range msg.JsonInputs {
		// Try parsing as JSON object directly
		var jsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err == nil {
			// Convert JSON object to value map
			valueMap, err := convertToValueMap(jsonObj)
			if err != nil {
				return nil, sdkerrors.AsConnectError(err)
			}

			wrappedValue.Struct.Fields[key] = &valuesv1.Value{
				Dict: &valuesv1.Dict{
					Items: convertMapToItems(valueMap),
				},
			}
		} else {
			// Handle as simple string if not valid JSON object
			wrappedValue.Struct.Fields[key] = &valuesv1.Value{
				String_: &valuesv1.String{V: jsonStr},
			}
		}
	}

	msg.Session.Inputs = map[string]*valuesv1.Value{
		"data": wrappedValue,
	}

	session, err := sdktypes.SessionFromProto(msg.Session)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := s.sessions.Start(ctx, session)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&sessionsv1.StartResponse{SessionId: uid.String()}), nil
}

func convertMapToItems(m map[string]*valuesv1.Value) []*valuesv1.Dict_Item {
	items := make([]*valuesv1.Dict_Item, 0, len(m))
	for k, v := range m {
		items = append(items, &valuesv1.Dict_Item{
			K: &valuesv1.Value{String_: &valuesv1.String{V: k}},
			V: v,
		})
	}
	return items
}

func convertToValueMap(input map[string]interface{}) (map[string]*valuesv1.Value, error) {
	result := make(map[string]*valuesv1.Value)

	for k, v := range input {
		value, err := interfaceToValue(v)
		if err != nil {
			return nil, fmt.Errorf("error converting key %s: %w", k, err)
		}
		result[k] = value
	}

	return result, nil
}

func interfaceToValue(v interface{}) (*valuesv1.Value, error) {
	switch x := v.(type) {
	case string:
		return &valuesv1.Value{
			String_: &valuesv1.String{
				V: x,
			},
		}, nil
	case float64:
		return &valuesv1.Value{
			Float: &valuesv1.Float{
				V: x,
			},
		}, nil
	case bool:
		return &valuesv1.Value{
			Boolean: &valuesv1.Boolean{
				V: x,
			},
		}, nil
	case nil:
		return &valuesv1.Value{
			Nothing: &valuesv1.Nothing{},
		}, nil
	case map[string]interface{}:
		items := make([]*valuesv1.Dict_Item, 0, len(x))
		for k, v := range x {
			val, err := interfaceToValue(v)
			if err != nil {
				return nil, err
			}
			kval, err := interfaceToValue(k)
			if err != nil {
				return nil, err
			}
			items = append(items, &valuesv1.Dict_Item{
				K: kval,
				V: val,
			})
		}
		return &valuesv1.Value{
			Dict: &valuesv1.Dict{
				Items: items,
			},
		}, nil
	case []interface{}:
		values := make([]*valuesv1.Value, len(x))
		for i, v := range x {
			val, err := interfaceToValue(v)
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
		return &valuesv1.Value{
			List: &valuesv1.List{
				Vs: values,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func (s *server) Get(ctx context.Context, req *connect.Request[sessionsv1.GetRequest]) (*connect.Response[sessionsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.ParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	session, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pb := session.ToProto()

	if req.Msg.JsonValues {
		if pb.Inputs, err = kittehs.TransformMapValuesError(pb.Inputs, sdktypes.ValueProtoToJSONStringValue); err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&sessionsv1.GetResponse{Session: pb}), nil
}

func (s *server) Stop(ctx context.Context, req *connect.Request[sessionsv1.StopRequest]) (*connect.Response[sessionsv1.StopResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.ParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	var forceTimeout time.Duration
	if msg.TerminationDelay != nil {
		forceTimeout = msg.TerminationDelay.AsDuration()
	}

	if err := s.sessions.Stop(ctx, sessionID, msg.Reason, msg.Terminate, forceTimeout); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&sessionsv1.StopResponse{}), nil
}

func (s *server) GetLog(ctx context.Context, req *connect.Request[sessionsv1.GetLogRequest]) (*connect.Response[sessionsv1.GetLogResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.ParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	filter := sdkservices.ListSessionLogRecordsFilter{
		PaginationRequest: sdktypes.PaginationRequest{
			Skip:      msg.Skip,
			PageToken: msg.PageToken,
			PageSize:  msg.PageSize,
			Ascending: msg.Ascending,
		},
		Types:     msg.Types,
		SessionID: sessionID,
	}

	hist, err := s.sessions.GetLog(ctx, filter)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			return connect.NewResponse(&sessionsv1.GetLogResponse{}), nil
		}
		return nil, sdkerrors.AsConnectError(err)
	}

	logpb := hist.Log.ToProto()
	if req.Msg.JsonValues {
		for _, r := range logpb.Records {
			if c := r.CallAttemptComplete; c != nil {
				if c.Result.Value, err = sdktypes.ValueProtoToJSONStringValue(c.Result.Value); err != nil {
					return nil, err
				}
			} else if c := r.CallSpec; c != nil {
				if c.Args, err = kittehs.TransformError(c.Args, sdktypes.ValueProtoToJSONStringValue); err != nil {
					return nil, err
				}

				if c.Kwargs, err = kittehs.TransformMapValuesError(c.Kwargs, sdktypes.ValueProtoToJSONStringValue); err != nil {
					return nil, err
				}
			}
		}
	}

	return connect.NewResponse(&sessionsv1.GetLogResponse{Log: logpb, Count: hist.TotalCount, NextPageToken: hist.NextPageToken}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[sessionsv1.ListRequest]) (*connect.Response[sessionsv1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	stateType, err := sdktypes.SessionStateTypeFromProto(msg.StateType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("state_type: %w", err))
	}

	filter := sdkservices.ListSessionsFilter{
		StateType: stateType,
		CountOnly: msg.CountOnly,
		PaginationRequest: sdktypes.PaginationRequest{
			Skip:      msg.Skip,
			PageToken: msg.PageToken,
		},
	}

	filter.PageSize = msg.PageSize
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	if filter.PageSize < 10 {
		filter.PageSize = 10
	}

	if filter.DeploymentID, err = sdktypes.ParseDeploymentID(req.Msg.DeploymentId); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if filter.EventID, err = sdktypes.ParseEventID(req.Msg.EventId); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if filter.ProjectID, err = sdktypes.ParseProjectID(req.Msg.ProjectId); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if filter.OrgID, err = sdktypes.ParseOrgID(req.Msg.OrgId); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	result, err := s.sessions.List(ctx, filter)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pbsessions := kittehs.Transform(result.Sessions, sdktypes.ToProto)

	return connect.NewResponse(&sessionsv1.ListResponse{Sessions: pbsessions, Count: result.TotalCount, NextPageToken: result.NextPageToken}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[sessionsv1.DeleteRequest]) (*connect.Response[sessionsv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.ParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err = s.sessions.Delete(ctx, sessionID); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	return connect.NewResponse(&sessionsv1.DeleteResponse{}), nil
}
