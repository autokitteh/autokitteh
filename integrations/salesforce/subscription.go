package salesforce

import (
	"context"
	"sync"

	pb "github.com/developerforce/pub-sub-api/go/proto"
	"github.com/linkedin/goavro/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	// Key = Salesforce instance URL (to ensure one gRPC client per app).
	// Writes happen during server startup, and when a user
	// creates a new Salesforce connection in the UI.
	pubSubClients = make(map[string]*pb.PubSubClient)

	mu = &sync.Mutex{}
)

// subscribe creates a new gRPC client and subscribes to a generic Salesforce Change Data Capture channel.
// https://developer.salesforce.com/docs/platform/pub-sub-api/references/methods/subscribe-rpc.html
func (h handler) subscribe(instanceURL, orgID string, cid sdktypes.ConnectionID) {
	// Prevent duplication due to race conditions.
	mu.Lock()
	defer mu.Unlock()

	if _, ok := pubSubClients[instanceURL]; ok {
		return
	}

	l := h.logger.With(zap.String("connection_id", cid.String()))
	ctx := context.Background()
	conn, err := initConn(l, h.bearerToken(ctx, l, cid), instanceURL, orgID)
	if err != nil {
		return
	}

	client := pb.NewPubSubClient(conn)
	pubSubClients[instanceURL] = &client

	// https://developer.salesforce.com/docs/atlas.en-us.platform_events.meta/platform_events/platform_events_objects_change_data_capture.htm
	go h.eventLoop(ctx, client, "/data/ChangeEvents", cid)
}

// eventLoop processes incoming messages, and automatically renews the subscription.
// https://developer.salesforce.com/docs/platform/pub-sub-api/guide/pub-sub-features.html
func (h handler) eventLoop(ctx context.Context, client pb.PubSubClient, topicName string, cid sdktypes.ConnectionID) {
	l := h.logger.With(zap.String("connection_id", cid.String()))
	const defaultBatchSize = int32(100)
	numLeftToReceive := 0

	stream, err := client.Subscribe(ctx)
	if err != nil {
		l.Error("failed to create gRPC stream for Salesforce events", zap.Error(err))
		return
	}

	// Start receiving messages.
	for {
		if numLeftToReceive <= 0 {
			n, err := renewSubscription(l, stream, defaultBatchSize, topicName)
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
			if gstatus.Code(err) != codes.Unauthenticated {
				l.Error("authentication error receiving Salesforce event", zap.Error(err))
			}
			l.Error("error receiving Salesforce event", zap.Error(err))
			return
		}

		// TODO(INT-314): Save the latest replay ID.
		// latestReplayId = msg.GetLatestReplayId()

		// Process the received message.
		numLeftToReceive -= len(msg.Events)
		for _, event := range msg.Events {
			schema, err := client.GetSchema(ctx, &pb.SchemaRequest{SchemaId: event.Event.SchemaId})
			if err != nil {
				l.Error("failed to get Salesforce event schema", zap.String("schema_id", event.Event.SchemaId), zap.Error(err))
				continue
			}
			data, err := decodePayload(l, schema, event.Event.Payload)
			if err != nil {
				continue
			}

			h.dispatchEvent(data, topicName)
		}
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
		l.Error("failed to create codec for decoding Salesforce event", zap.Error(err))
		return nil, err
	}

	var data map[string]any
	native, remaining, err := codec.NativeFromBinary(payload)
	if len(remaining) > 0 {
		l.Warn("remaining bytes after decoding Salesforce event", zap.Int("len", len(remaining)), zap.Any("remaining", remaining))
	}
	if err != nil {
		l.Error("failed to decode event", zap.Error(err))
		return nil, err
	}
	data, ok := native.(map[string]any)
	if !ok {
		l.Error("failed to cast decoded event to map[string]any")
		return nil, err
	}
	return data, nil
}
