# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from autokitteh_pb.integrations.v1 import svc_pb2 as autokitteh_dot_integrations_dot_v1_dot_svc__pb2


class IntegrationsServiceStub(object):
    """Implemented by integration providers.
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Get = channel.unary_unary(
                '/autokitteh.integrations.v1.IntegrationsService/Get',
                request_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetRequest.SerializeToString,
                response_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetResponse.FromString,
                )
        self.List = channel.unary_unary(
                '/autokitteh.integrations.v1.IntegrationsService/List',
                request_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ListRequest.SerializeToString,
                response_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ListResponse.FromString,
                )
        self.Configure = channel.unary_unary(
                '/autokitteh.integrations.v1.IntegrationsService/Configure',
                request_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ConfigureRequest.SerializeToString,
                response_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ConfigureResponse.FromString,
                )
        self.TestConnection = channel.unary_unary(
                '/autokitteh.integrations.v1.IntegrationsService/TestConnection',
                request_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.TestConnectionRequest.SerializeToString,
                response_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.TestConnectionResponse.FromString,
                )
        self.GetConnectionStatus = channel.unary_unary(
                '/autokitteh.integrations.v1.IntegrationsService/GetConnectionStatus',
                request_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionStatusRequest.SerializeToString,
                response_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionStatusResponse.FromString,
                )
        self.GetConnectionConfig = channel.unary_unary(
                '/autokitteh.integrations.v1.IntegrationsService/GetConnectionConfig',
                request_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionConfigRequest.SerializeToString,
                response_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionConfigResponse.FromString,
                )
        self.Call = channel.unary_unary(
                '/autokitteh.integrations.v1.IntegrationsService/Call',
                request_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.CallRequest.SerializeToString,
                response_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.CallResponse.FromString,
                )


class IntegrationsServiceServicer(object):
    """Implemented by integration providers.
    """

    def Get(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def List(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Configure(self, request, context):
        """Get all values for a specific configuration of the integration.
        The returned values ExecutorIDs will be the integration id.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def TestConnection(self, request, context):
        """Actively test the connection to the integration.
        requires supports_connection_test.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetConnectionStatus(self, request, context):
        """If connection_id is not provided, will return the status of a new connection as
        set in `Integration.initial_connection_status`.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetConnectionConfig(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Call(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_IntegrationsServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Get': grpc.unary_unary_rpc_method_handler(
                    servicer.Get,
                    request_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetRequest.FromString,
                    response_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetResponse.SerializeToString,
            ),
            'List': grpc.unary_unary_rpc_method_handler(
                    servicer.List,
                    request_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ListRequest.FromString,
                    response_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ListResponse.SerializeToString,
            ),
            'Configure': grpc.unary_unary_rpc_method_handler(
                    servicer.Configure,
                    request_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ConfigureRequest.FromString,
                    response_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ConfigureResponse.SerializeToString,
            ),
            'TestConnection': grpc.unary_unary_rpc_method_handler(
                    servicer.TestConnection,
                    request_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.TestConnectionRequest.FromString,
                    response_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.TestConnectionResponse.SerializeToString,
            ),
            'GetConnectionStatus': grpc.unary_unary_rpc_method_handler(
                    servicer.GetConnectionStatus,
                    request_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionStatusRequest.FromString,
                    response_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionStatusResponse.SerializeToString,
            ),
            'GetConnectionConfig': grpc.unary_unary_rpc_method_handler(
                    servicer.GetConnectionConfig,
                    request_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionConfigRequest.FromString,
                    response_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionConfigResponse.SerializeToString,
            ),
            'Call': grpc.unary_unary_rpc_method_handler(
                    servicer.Call,
                    request_deserializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.CallRequest.FromString,
                    response_serializer=autokitteh_dot_integrations_dot_v1_dot_svc__pb2.CallResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'autokitteh.integrations.v1.IntegrationsService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class IntegrationsService(object):
    """Implemented by integration providers.
    """

    @staticmethod
    def Get(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.integrations.v1.IntegrationsService/Get',
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetRequest.SerializeToString,
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def List(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.integrations.v1.IntegrationsService/List',
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ListRequest.SerializeToString,
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ListResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Configure(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.integrations.v1.IntegrationsService/Configure',
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ConfigureRequest.SerializeToString,
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.ConfigureResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def TestConnection(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.integrations.v1.IntegrationsService/TestConnection',
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.TestConnectionRequest.SerializeToString,
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.TestConnectionResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetConnectionStatus(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.integrations.v1.IntegrationsService/GetConnectionStatus',
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionStatusRequest.SerializeToString,
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionStatusResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetConnectionConfig(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.integrations.v1.IntegrationsService/GetConnectionConfig',
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionConfigRequest.SerializeToString,
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.GetConnectionConfigResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Call(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.integrations.v1.IntegrationsService/Call',
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.CallRequest.SerializeToString,
            autokitteh_dot_integrations_dot_v1_dot_svc__pb2.CallResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
