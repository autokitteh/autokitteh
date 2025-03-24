package salesforce

import (
	"context"
	"strings"
	"sync"

	pb "github.com/developerforce/pub-sub-api/go/proto"
	"github.com/linkedin/goavro/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

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
func (h handler) subscribe(clientID, orgID, instanceURL string, cid sdktypes.ConnectionID) {
	// Prevent duplication due to race conditions.
	mu.Lock()
	defer mu.Unlock()

	if _, ok := pubSubClients[clientID]; ok {
		return
	}

	ctx, client := h.initPubSubClient(cid, clientID, nil, "")
	if ctx == nil || client == nil {
		h.logger.Error("failed to create Salesforce client", zap.String("client_id", clientID))
		return
	}

	// https://developer.salesforce.com/docs/atlas.en-us.platform_events.meta/platform_events/platform_events_objects_change_data_capture.htm
	go h.eventLoop(ctx, clientID, "/data/ChangeEvents", cid)
}

// eventLoop processes incoming messages, and automatically renews the subscription.
// https://developer.salesforce.com/docs/platform/pub-sub-api/guide/pub-sub-features.html
func (h handler) eventLoop(ctx context.Context, clientID string, subscribeTopic string, cid sdktypes.ConnectionID) {
	l := h.logger.With(zap.String("connection_id", cid.String()))
	const defaultBatchSize = int32(100)
	numLeftToReceive := 0

	client := pubSubClients[clientID]
	stream := h.initStream(ctx, l, client, cid, clientID)

	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid), userIDVar)
	if err != nil {
		l.Error("failed to get connection vars", zap.Error(err))
		return
	}
	userID := vs.GetValue(userIDVar)
	if userID == "" {
		l.Error("user_id is not set in connection vars")
		return
	}

	// Start receiving messages.
	for {
		if numLeftToReceive <= 0 {
			n, err := renewSubscription(l, stream, defaultBatchSize, subscribeTopic)
			if err != nil {
				l.Error("failed to renew Salesforce events subscription", zap.Error(err))
				continue
			}
			numLeftToReceive = n
		}

		// Assumes that the stream will abort after 270 seconds of inactivity.
		// https://developer.salesforce.com/docs/platform/pub-sub-api/references/methods/subscribe-rpc.html#subscribe-keepalive-behavior
		msg, err := stream.Recv()
		if err != nil {
			switch {
			case gstatus.Code(err) == codes.Unauthenticated:
				l.Error("authentication error receiving Salesforce event", zap.Error(err))
				cleanupClient(l, clientID)
				ctx, client = h.initPubSubClient(cid, clientID, l, "failed to reinitialize Salesforce client")
			case err.Error() == "EOF":
				l.Warn("Salesforce stream connection closed (EOF), reconnecting...", zap.Error(err))
				cleanupClient(l, clientID)
				ctx, client = h.initPubSubClient(cid, clientID, l, "failed to reinitialize Salesforce client after EOF")
			default:
				l.Error("error receiving Salesforce event", zap.Error(err))
			}
			stream = h.initStream(ctx, l, client, cid, clientID)
			continue
		}
		l.Debug("received Salesforce event", zap.Any("event", msg))

		// TODO(INT-314): Save the latest replay ID.
		// latestReplayId = msg.GetLatestReplayId()

		// Process the received message.
		numLeftToReceive -= len(msg.Events)
		for _, event := range msg.Events {
			schema, err := client.GetSchema(ctx, &pb.SchemaRequest{SchemaId: event.Event.SchemaId})
			if err != nil {
				l.Error("failed to get Salesforce event schema", zap.String("schema_id", event.Event.SchemaId), zap.Error(err))
				if gstatus.Code(err) != codes.Unauthenticated {
					continue
				}

				// If it's an authentication error, try to reinitialize the client.
				cleanupClient(l, clientID)
				ctx, client = h.initPubSubClient(cid, clientID, l, "failed to reinitialize Salesforce client after schema auth error")
				if ctx == nil || client == nil {
					continue
				}

				// Try to get the schema again with the new client.
				schema, err = client.GetSchema(ctx, &pb.SchemaRequest{SchemaId: event.Event.SchemaId})
				if err != nil {
					l.Error("still failed to get Salesforce schema after client reinitialization", zap.Error(err))
					continue
				}
			}

			data, err := decodePayload(l, schema, event.Event.Payload)
			if err != nil {
				l.Error("failed to decode Salesforce event", zap.Error(err))
				continue
			}

			header, ok := data["ChangeEventHeader"].(map[string]any)
			if !ok {
				l.Error("ChangeEventHeader is not a map in event data")
				continue
			}
			commitUser, ok := header["commitUser"].(string)
			if !ok {
				l.Error("commitUser is not a string in ChangeEventHeader")
				continue
			}

			// Ignore self-triggered events.
			if commitUser == userID {
				l.Debug("ignoring Salesforce event", zap.String("commitUser", commitUser))
				continue
			}

			// Extract changed entity name for the event type.
			entityName, ok := header["entityName"].(string)
			if !ok {
				l.Error("entityName is not a string in ChangeEventHeader")
				continue
			}

			h.dispatchEvent(data, strings.ToLower(entityName))
		}
	}
}

// initPubSubClient initializes or reinitializes a PubSub client
// If errorMessage is provided, errors will be logged with this message
// Returns the context and client, which may be nil on error
func (h handler) initPubSubClient(cid sdktypes.ConnectionID, clientID string, l *zap.Logger, errorMessage string) (context.Context, pb.PubSubClient) {
	// TODO: return error to handle error outside this function
	ctx := authcontext.SetAuthnSystemUser(context.Background())

	// Use the provided logger if available, otherwise create one
	if l == nil {
		l = h.logger.With(zap.String("connection_id", cid.String()))
	}

	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Error("failed to read connection vars",
			zap.String("connection_id", cid.String()), zap.Error(err),
		)
		if errorMessage != "" {
			l.Error(errorMessage, zap.Error(err))
		}
		return nil, nil
	}

	t := common.FreshOAuthToken(ctx, l, h.oauth, h.vars, desc, vs)
	cfg, _, err := h.oauth.Get(ctx, desc.UniqueName().String())
	if err != nil {
		l.Error("failed to get Salesforce OAuth config", zap.Error(err))
		if errorMessage != "" {
			l.Error(errorMessage, zap.Error(err))
		}
		return nil, nil
	}

	conn, err := initConn(l, cfg, t, vs.GetValue(instanceURLVar), vs.GetValue(orgIDVar))
	if err != nil {
		if errorMessage != "" {
			l.Error(errorMessage, zap.Error(err))
		}
		return nil, nil
	}

	client := pb.NewPubSubClient(conn)
	pubSubClients[clientID] = client
	pubSubConnections[clientID] = conn

	return ctx, client
}

// cleanupClient removes a client from the pubSubClients map.
// This allows the underlying connection to be garbage collected.
func cleanupClient(l *zap.Logger, clientID string) {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := pubSubClients[clientID]; ok {
		if conn, ok := pubSubConnections[clientID]; ok {
			if err := conn.Close(); err != nil {
				l.Error("error closing Salesforce connection",
					zap.String("client_id", clientID),
					zap.Error(err))
			}
		}
		delete(pubSubClients, clientID)
		delete(pubSubConnections, clientID)
		l.Debug("cleaned up Salesforce client", zap.String("client_id", clientID))
	}
}

func (h handler) initStream(ctx context.Context, l *zap.Logger, client pb.PubSubClient, cid sdktypes.ConnectionID, clientID string) pb.PubSub_SubscribeClient {
	for {
		stream, err := client.Subscribe(ctx)
		if err != nil {
			cleanupClient(l, clientID)
			h.initPubSubClient(cid, clientID, l, "failed to create gRPC stream for Salesforce events")
			// TODO(INT-352): error handling
			continue
		}

		return stream
	}
}

func renewSubscription(l *zap.Logger, stream pb.PubSub_SubscribeClient, defaultBatchSize int32, topicName string) (int, error) {
	fetchReq := &pb.FetchRequest{
		TopicName:    topicName,
		NumRequested: defaultBatchSize,
	}

	// TODO(INT-314): Use the latest replay ID if available for resumption.

	err := stream.Send(fetchReq)
	if err != nil {
		l.Error("failed to request more Salesforce events", zap.Error(err))
		return 0, err
	}
	return int(defaultBatchSize), nil
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
