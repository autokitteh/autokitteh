package salesforce

import (
	"context"
	"crypto/tls"
	"sync"

	pb "github.com/developerforce/pub-sub-api/go/proto"
	"github.com/linkedin/goavro/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	// Key = Salesforce instance URL (to ensure one gRPC client per app).
	// Writes happen during server startup, and when a user
	// creates a new Salesforce connection in the UI.
	pubSubClients = make(map[string]*pb.PubSubClient)

	mu = &sync.Mutex{}
)

type salesforceAuth struct {
	accessToken string
	instanceURL string
	tenantID    string
}

func (a *salesforceAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"accesstoken": a.accessToken,
		"instanceurl": a.instanceURL,
		"tenantid":    a.tenantID, // Optional
	}, nil
}

func (a *salesforceAuth) RequireTransportSecurity() bool {
	return true
}

func (h handler) Subscribe(instanceURL, orgID, accessToken string) {
	// Ensure multiple users don't reference the same app at the same time.
	mu.Lock()
	defer mu.Unlock()

	if _, ok := pubSubClients[instanceURL]; ok {
		// TODO: bug – if an error shuts down the connection, the only way to retry is to restart the server
		// i.e. reauthenticating does not help because there is a key in this map for that instance URL
		return
	}

	ctx := context.Background()

	conn, err := initConn(h.logger, accessToken, instanceURL, orgID)
	if err != nil {
		h.logger.Error("failed to create gRPC connection", zap.Error(err))
		return
	}

	client := pb.NewPubSubClient(conn)
	pubSubClients[instanceURL] = &client

	// TODO: handle multiple topics + support for custom topics
	go h.handleSalesForceSubscription(ctx, client, "/event/Test__e")
}

func (h handler) handleSalesForceSubscription(ctx context.Context, client pb.PubSubClient, topicName string) {
	topic, err := client.GetTopic(ctx, &pb.TopicRequest{
		TopicName: topicName,
	})
	if err != nil {
		h.logger.Error("failed to get topic", zap.Error(err))
		return
	}

	// TODO: Update schema periodically when changes are made to the topic.
	schema, err := client.GetSchema(ctx, &pb.SchemaRequest{
		SchemaId: topic.SchemaId,
	})
	if err != nil {
		h.logger.Error("failed to get schema", zap.Error(err))
		return
	}

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
			n, err := renewSubscription(defaultBatchSize, stream, h.logger)
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

		// TODO: Save the latest replay ID.
		// latestReplayId = msg.GetLatestReplayId()

		// Process the received message.
		eventCount := len(msg.Events)
		if eventCount > 0 {
			numLeftToReceive -= eventCount
		}

		for _, event := range msg.Events {
			data, err := decodePayload(schema, event.Event.Payload, h.logger)
			if err != nil {
				h.logger.Error("failed to decode event", zap.Error(err))
				continue
			}
			// TODO: dispatch event to autokitteh connections
			h.logger.Info("decoded event", zap.Any("decoded", data))
		}
	}
}

func initConn(l *zap.Logger, accessToken, instanceURL, orgID string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"api.pubsub.salesforce.com:443",
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&salesforceAuth{
			accessToken: accessToken,
			instanceURL: instanceURL,
			tenantID:    orgID,
		}),
	)
	if err != nil {
		l.Error("failed to create gRPC connection", zap.Error(err))
		return nil, err
	}
	return conn, nil
}

func renewSubscription(defaultBatchSize int32, stream pb.PubSub_SubscribeClient, l *zap.Logger) (int, error) {
	l.Info("requesting more messages", zap.Int32("batchSize", defaultBatchSize))
	fetchReq := &pb.FetchRequest{
		TopicName:    "/event/Test__e",
		NumRequested: defaultBatchSize,
	}

	// TODO: Use the latest replay ID if available for resumption

	err := stream.Send(fetchReq)
	if err != nil {
		l.Error("failed to request more messages", zap.Error(err))
		return 0, err
	}
	return int(defaultBatchSize), nil
}

func decodePayload(schema *pb.SchemaInfo, payload []byte, l *zap.Logger) (map[string]interface{}, error) {
	codec, err := goavro.NewCodec(schema.SchemaJson)
	if err != nil {
		l.Error("failed to create codec", zap.Error(err))
		return nil, err
	}
	native, _, err := codec.NativeFromBinary(payload)
	if err != nil {
		l.Error("failed to decode event", zap.Error(err))
		return nil, err
	}
	data, ok := native.(map[string]interface{})
	if !ok {
		l.Error("failed to cast decoded event to map[string]interface{}")
		return nil, err
	}

	return data, nil
}

func (h handler) dispatchAsyncEventsToConnections(cids []sdktypes.ConnectionID, e sdktypes.Event) {
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	for _, cid := range cids {
		eid, err := h.dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
		l := h.logger.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			l.Error("Event dispatch failed", zap.Error(err))
			return
		}
		l.Debug("Event dispatched")
	}
}
