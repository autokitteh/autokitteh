package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type method struct {
	Name      string
	Inputs    []string
	Fullname  string
	Constants []string
}

type grpcclient struct {
	conn             *grpc.ClientConn
	reflectionClient *grpcreflect.Client
	descSource       grpcurl.DescriptorSource
}

type protoFiles struct {
	paths []string
	names []string
}

func newGRPCClient(conn *grpc.ClientConn) (*grpcclient, error) {
	reflectionClient := grpcreflect.NewClientAuto(context.Background(), conn)
	var descSource grpcurl.DescriptorSource = grpcurl.DescriptorSourceFromServer(context.Background(), reflectionClient)
	return &grpcclient{
		conn:             conn,
		reflectionClient: reflectionClient,
		descSource:       descSource,
	}, nil
}

func (r *grpcclient) invoke(method string, payload map[string]any) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("grpc integration invoke data invalid json: %w", err)
	}

	dataProvider := bytes.NewReader(jsonBytes)
	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.FormatJSON, r.descSource, dataProvider, grpcurl.FormatOptions{})
	if err != nil {
		return nil, fmt.Errorf("grpc integration: request parser: %w", err)
	}

	var output bytes.Buffer
	h := &grpcurl.DefaultEventHandler{
		Out:       &output,
		Formatter: formatter,
	}

	if err := grpcurl.InvokeRPC(context.Background(), r.descSource, r.conn, method, nil, h, rf.Next); err != nil {
		return nil, fmt.Errorf("grpc integration: invoke: %w", err)
	}
	if h.Status.Code() != codes.OK {
		return nil, fmt.Errorf("grpc integration status code not ok: %w", h.Status.Err())
	}

	var t map[string]interface{}
	if err := json.Unmarshal(output.Bytes(), &t); err != nil {
		return nil, fmt.Errorf("grpc integration: unmarshal: %w", err)
	}

	return t, nil
}
