package salesforce

import (
	"context"
	"sync"

	pb "github.com/developerforce/pub-sub-api/go/proto"
	"github.com/linkedin/goavro/v2"
	"go.uber.org/zap"

	// "go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	// Key = Salesforce instance URL (to ensure one gRPC client per app).
	// Writes happen during server startup, and when a user
	// creates a new Salesforce connection in the UI.
	pubSubClients = make(map[string]*pb.PubSubClient)

	mu = &sync.Mutex{}
)

func (h handler) Subscribe(instanceURL, orgID, accessToken string, cid sdktypes.ConnectionID) {
	// Ensure multiple users don't reference the same app at the same time.
	mu.Lock()
	defer mu.Unlock()

	if _, ok := pubSubClients[instanceURL]; ok {
		return
	}

	ctx := context.Background()
	conn, err := initConn(accessToken, instanceURL, orgID, h.logger)
	if err != nil {
		h.logger.Error("failed to create gRPC connection", zap.Error(err))
		return
	}

	client := pb.NewPubSubClient(conn)
	pubSubClients[instanceURL] = &client

	for _, eventType := range supportedEventTypes() {
		go h.handleSalesForceSubscription(ctx, client, eventType, cid)
	}
}

func (h handler) handleSalesForceSubscription(ctx context.Context, client pb.PubSubClient, topicName string, cid sdktypes.ConnectionID) {
	const defaultBatchSize = int32(100)
	// var latestReplayId []byte
	numLeftToReceive := 0

	stream, err := client.Subscribe(ctx)
	if err != nil {
		h.logger.Error("failed to create stream", zap.Error(err))
		return
	}

	// Start receiving messages
	for {
		if numLeftToReceive <= 0 {
			n, err := renewSubscription(stream, defaultBatchSize, topicName, h.logger)
			if err != nil {
				h.logger.Error("failed to renew subscription", zap.Error(err))
				continue
			}
			numLeftToReceive = n
		}

		// Assumes that the stream will abort after 270 seconds of inactivity.
		// https://developer.salesforce.com/docs/platform/pub-sub-api/references/methods/subscribe-rpc.html#subscribe-keepalive-behavior
		msg, err := stream.Recv()
		if err != nil {
			h.logger.Error("error receiving message", zap.Error(err))
			continue
		}

		// TODO(INT-314): Save the latest replay ID.
		// latestReplayId = msg.GetLatestReplayId()

		// Process the received message.
		eventCount := len(msg.Events)
		if eventCount > 0 {
			numLeftToReceive -= eventCount
		}

		for _, event := range msg.Events {
			schema, err := client.GetSchema(ctx, &pb.SchemaRequest{
				SchemaId: event.Event.SchemaId,
			})
			if err != nil {
				h.logger.Error("failed to get schema", zap.Error(err))
				continue
			}
			data, err := decodePayload(schema, event.Event.Payload, h.logger)
			if err != nil {
				h.logger.Error("failed to decode event", zap.Error(err))
				continue
			}

			h.handleSalesForceEvent(data, topicName)
		}
	}
}

func renewSubscription(stream pb.PubSub_SubscribeClient, defaultBatchSize int32, topicName string, l *zap.Logger) (int, error) {
	l.Info("requesting more messages", zap.Int32("batchSize", defaultBatchSize))
	fetchReq := &pb.FetchRequest{
		TopicName:    topicName,
		NumRequested: defaultBatchSize,
	}

	// TODO(INT-314): Use the latest replay ID if available for resumption

	err := stream.Send(fetchReq)
	if err != nil {
		l.Error("failed to request more messages", zap.Error(err))
		return 0, err
	}
	return int(defaultBatchSize), nil
}

func decodePayload(schema *pb.SchemaInfo, payload []byte, l *zap.Logger) (map[string]any, error) {
	codec, err := goavro.NewCodec(schema.SchemaJson)
	if err != nil {
		l.Error("failed to create codec", zap.Error(err))
		return nil, err
	}

	var data map[string]any
	// Payload may contain multiple records
	native, _, err := codec.NativeFromBinary(payload)
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
