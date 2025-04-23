package salesforce

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	pb "github.com/developerforce/pub-sub-api/go/proto"
	"github.com/linkedin/goavro/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const defaultBatchSize = 100

var (
	// Key = Salesforce instance URL (to ensure one gRPC client per app).
	// Writes happen during server startup, and when a user
	// creates a new Salesforce connection in the UI.
	pubSubClients = make(map[string]pb.PubSubClient)
	// Store connections so we can close them properly.
	pubSubConnections = make(map[string]*grpc.ClientConn)

	mu = &sync.Mutex{}
)

// subscribe creates a new gRPC client and subscribes to a generic Salesforce Change Data Capture channel.
// https://developer.salesforce.com/docs/platform/pub-sub-api/references/methods/subscribe-rpc.html
func (h handler) subscribe(l *zap.Logger, clientID string, cid sdktypes.ConnectionID) {
	mu.Lock()
	_, ok := pubSubClients[clientID]
	mu.Unlock()

	if ok {
		return
	}

	ctx, _, err := h.initPubSubClient(l, cid, clientID)
	if err != nil {
		l.Error("failed to create Salesforce client", zap.String("client_id", clientID), zap.Error(err))
		return
	}

	// https://developer.salesforce.com/docs/atlas.en-us.platform_events.meta/platform_events/platform_events_objects_change_data_capture.htm
	go h.eventLoop(ctx, l, clientID, cid)
}

// eventLoop is a goroutine that processes incoming messages, and automatically renews the subscription.
// https://developer.salesforce.com/docs/platform/pub-sub-api/guide/pub-sub-features.html
func (h handler) eventLoop(ctx context.Context, l *zap.Logger, clientID string, cid sdktypes.ConnectionID) {
	client := pubSubClients[clientID]
	stream := h.initStream(ctx, l, cid, clientID)

	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid), userIDVar)
	if err != nil {
		l.Error("failed to get connection vars", zap.Error(err))
		return
	}
	userID := vs.GetValue(userIDVar)
	if userID == "" {
		l.Error("user_id is not set in Salesforce connection vars")
		return
	}

	// Start receiving messages.
	for {
		eventsLeftToReceive := defaultBatchSize
		if err := h.renewSubscription(ctx, l, stream, cid); err != nil {
			l.Error("failed to renew Salesforce events subscription", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for eventsLeftToReceive > 0 {
			// Assumes that the stream will abort after 270 seconds of inactivity.
			// https://developer.salesforce.com/docs/platform/pub-sub-api/references/methods/subscribe-rpc.html#subscribe-keepalive-behavior
			msg, err := stream.Recv()
			if err != nil {
				l.Error("error receiving Salesforce event", zap.Error(err))
				cleanupClient(l, clientID)
				ctx, client, err = h.initPubSubClient(l, cid, clientID)
				if err != nil {
					l.Error("failed to reinitialize Salesforce client", zap.Error(err))
					time.Sleep(time.Second)
					break
				}
				stream = h.initStream(ctx, l, cid, clientID)
				break
			}

			l.Debug("received Salesforce event", zap.Any("event", msg))

			// TODO(INT-314): Save the latest replay ID.
			// latestReplayId = msg.GetLatestReplayId()

			// Process the received message.
			eventsLeftToReceive -= len(msg.Events)
			for _, event := range msg.Events {
				schema, err := client.GetSchema(ctx, &pb.SchemaRequest{SchemaId: event.Event.SchemaId})
				if err != nil {
					l.Error("failed to get Salesforce event schema", zap.String("schema_id", event.Event.SchemaId), zap.Error(err))
					continue // TODO: handle authentication error differently?
				}

				data, err := decodePayload(l, schema, event.Event.Payload)
				if err != nil {
					l.Error("failed to decode Salesforce event", zap.Error(err))
					continue
				}

				header, ok := data["ChangeEventHeader"]
				if !ok {
					l.Error("ChangeEventHeader is not present in event data")
					continue
				}
				m, ok := header.(map[string]any)
				if !ok {
					l.Error("ChangeEventHeader is not a map in event data")
					continue
				}

				commitUser, ok := m["commitUser"]
				if !ok {
					l.Error("commitUser is not present in ChangeEventHeader")
					continue
				}
				s, ok := commitUser.(string)
				if !ok {
					l.Error("commitUser is not a string in ChangeEventHeader")
					continue
				}

				// Ignore self-triggered events.
				if s == userID {
					l.Debug("ignoring Salesforce event", zap.String("commitUser", s))
					continue
				}

				// Extract changed entity name for the event type.
				entityName, ok := m["entityName"]
				if !ok {
					l.Error("entityName is not present in ChangeEventHeader")
					continue
				}
				s, ok = entityName.(string)
				if !ok {
					l.Error("entityName is not a string in ChangeEventHeader")
					continue
				}

				h.dispatchEvent(data, strings.ToLower(s))

				// Save the latest replay ID.
				replayID := base64.StdEncoding.EncodeToString(msg.LatestReplayId)
				l.Debug("saving Salesforce replay ID", zap.String("replay_id", replayID))

				vsid := sdktypes.NewVarScopeID(cid)
				v := sdktypes.NewVar(replayIDVar).SetValue(replayID)
				if err := h.vars.Set(ctx, v.WithScopeID(vsid)); err != nil {
					l.Error("failed to save Salesforce replay ID", zap.Error(err))
				}
			}
		}
	}
}

// initPubSubClient initializes or reinitializes a PubSub client
// If errorMessage is provided, errors will be logged with this message
// Returns the context and client, which may be nil on error
func (h handler) initPubSubClient(l *zap.Logger, cid sdktypes.ConnectionID, clientID string) (context.Context, pb.PubSubClient, error) {
	conn, err := h.initConn(l, cid)
	if err != nil {
		return nil, nil, err
	}

	client := pb.NewPubSubClient(conn)

	mu.Lock()
	pubSubConnections[clientID] = conn
	pubSubClients[clientID] = client
	mu.Unlock()

	ctx := authcontext.SetAuthnSystemUser(context.Background())
	return ctx, client, nil
}

// cleanupClient removes a client from the pubSubClients map.
// This allows the underlying connection to be garbage collected.
func cleanupClient(l *zap.Logger, clientID string) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := pubSubClients[clientID]; !ok {
		l.Error("Salesforce pubSubClient already deleted", zap.Any("clientID", clientID))
		return
	}

	conn := pubSubConnections[clientID]
	if err := conn.Close(); err != nil {
		l.Error("error closing Salesforce connection",
			zap.String("client_id", clientID),
			zap.Error(err))
	}

	delete(pubSubClients, clientID)
	delete(pubSubConnections, clientID)
	l.Debug("cleaned up Salesforce client", zap.String("client_id", clientID))
}

func (h handler) initStream(ctx context.Context, l *zap.Logger, cid sdktypes.ConnectionID, clientID string) pb.PubSub_SubscribeClient {
	for {
		mu.Lock()
		client := pubSubClients[clientID]
		mu.Unlock()

		stream, err := client.Subscribe(ctx)
		if err != nil {
			l.Error("failed to subscribe to Salesforce topic", zap.Error(err))
			time.Sleep(time.Second)
			cleanupClient(l, clientID)
			ctx, _, err = h.initPubSubClient(l, cid, clientID)
			if err != nil {
				l.Error("failed to reinitialize Salesforce client", zap.Error(err))
			}
			continue
		}

		return stream
	}
}

// https://developer.salesforce.com/docs/atlas.en-us.platform_events.meta/platform_events/platform_events_objects_change_data_capture.htm
func (h handler) renewSubscription(ctx context.Context, l *zap.Logger, stream pb.PubSub_SubscribeClient, cid sdktypes.ConnectionID) error {
	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid), replayIDVar)
	if err != nil {
		return fmt.Errorf("failed to get connection vars: %w", err)
	}
	replayID := vs.GetValue(replayIDVar)

	fetchReq := &pb.FetchRequest{
		ReplayPreset: pb.ReplayPreset_LATEST,
		TopicName:    "/data/ChangeEvents",
		NumRequested: defaultBatchSize,
	}

	// https://developer.salesforce.com/docs/platform/pub-sub-api/references/methods/subscribe-rpc.html?q=replay#replaying-an-event-stream
	if replayID != "" {
		fetchReq.ReplayPreset = pb.ReplayPreset_CUSTOM
		decodedReplayID, err := base64.StdEncoding.DecodeString(replayID)
		if err != nil {
			return fmt.Errorf("failed to decode replay ID: %w", err)
		}
		fetchReq.ReplayId = decodedReplayID
	}

	if err := stream.Send(fetchReq); err != nil {
		return fmt.Errorf("failed to request more Salesforce events: %w", err)
	}

	return nil
}

func decodePayload(l *zap.Logger, schema *pb.SchemaInfo, payload []byte) (map[string]any, error) {
	codec, err := goavro.NewCodec(schema.SchemaJson)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	native, remaining, err := codec.NativeFromBinary(payload)
	if len(remaining) > 0 {
		l.Warn("remaining bytes after decoding Salesforce event", zap.Int("len", len(remaining)), zap.Any("remaining", remaining))
	}
	if err != nil {
		return nil, err
	}

	data, ok := native.(map[string]any)
	if !ok {
		return nil, err
	}

	return data, nil
}
