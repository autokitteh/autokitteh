package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fullstorydev/grpcurl"
	"github.com/iancoleman/strcase"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/descriptorpb"
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

// proto.names should be a list of proto files
// proto.paths should be a list of paths where proto files could be found
// both has to be provided, otherwise an error is returned
// for example:
// proto.names = []string{"mysvc.proto", "mysvc2.proto"}
// proto.paths = []string{"/actual/path/to/protos/basedir"}
//
//lint:ignore U1000 Ignore unused to keep this code until we decide if we need it
func newGRPCClientWithProtos(conn *grpc.ClientConn, protos protoFiles) (*grpcclient, error) {
	if len(protos.names) == 0 || len(protos.paths) == 0 {
		return nil, errors.New("protos has to have both names and paths")
	}

	c, err := newGRPCClient(conn)
	if err != nil {
		return nil, err
	}

	fileDesc, err := grpcurl.DescriptorSourceFromProtoFiles(protos.paths, protos.names...)
	if err != nil {
		return nil, err
	}

	c.descSource = compositeSource{
		reflection: c.descSource,
		file:       fileDesc,
	}

	return c, nil
}

//lint:ignore U1000 Ignore unused to keep this code until we decide if we need it
func (r *grpcclient) getServiceDescriptor(service string) (*desc.ServiceDescriptor, error) {
	d, err := r.descSource.FindSymbol(service)
	if err != nil {
		return nil, err
	}

	svc, ok := d.(*desc.ServiceDescriptor)
	if !ok {
		return nil, errors.New("supplied desc is not a service descriptor")
	}

	return svc, nil
}

//lint:ignore U1000 Ignore unused to keep this code until we decide if we need it
func (r *grpcclient) listMethods(service string) ([]method, error) {
	svc, err := r.getServiceDescriptor(service)
	if err != nil {
		return nil, err
	}

	var methods []method

	for _, m := range svc.GetMethods() {
		var (
			inputs, constants []string
		)

		for _, fd := range m.GetInputType().GetFields() {
			if fd.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
				for _, v := range fd.GetEnumType().GetValues() {
					constants = append(constants, v.GetName())
				}
			}
			inputs = append(inputs, fd.GetJSONName())
		}

		name := strcase.ToSnake(fmt.Sprintf("%s%s", svc.GetName(), m.GetName()))
		methods = append(methods, method{
			Name:      name,
			Fullname:  m.GetFullyQualifiedName(),
			Inputs:    inputs,
			Constants: constants,
		})
	}

	return methods, nil
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
