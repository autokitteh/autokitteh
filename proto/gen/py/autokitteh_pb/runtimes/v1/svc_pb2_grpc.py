# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from autokitteh_pb.runtimes.v1 import svc_pb2 as autokitteh_dot_runtimes_dot_v1_dot_svc__pb2


class RuntimesServiceStub(object):
    """Runtimes are expected to be registered during deploy (from configuration).
    Dynamic registration of runtimes will not be supported.
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Describe = channel.unary_unary(
                '/autokitteh.runtimes.v1.RuntimesService/Describe',
                request_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.DescribeRequest.SerializeToString,
                response_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.DescribeResponse.FromString,
                )
        self.List = channel.unary_unary(
                '/autokitteh.runtimes.v1.RuntimesService/List',
                request_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.ListRequest.SerializeToString,
                response_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.ListResponse.FromString,
                )
        self.Build = channel.unary_unary(
                '/autokitteh.runtimes.v1.RuntimesService/Build',
                request_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BuildRequest.SerializeToString,
                response_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BuildResponse.FromString,
                )
        self.Run = channel.unary_stream(
                '/autokitteh.runtimes.v1.RuntimesService/Run',
                request_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.RunRequest.SerializeToString,
                response_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.RunResponse.FromString,
                )
        self.BidiRun = channel.stream_stream(
                '/autokitteh.runtimes.v1.RuntimesService/BidiRun',
                request_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BidiRunRequest.SerializeToString,
                response_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BidiRunResponse.FromString,
                )
        self.Build1 = channel.unary_unary(
                '/autokitteh.runtimes.v1.RuntimesService/Build1',
                request_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.Build1Request.SerializeToString,
                response_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.Build1Response.FromString,
                )


class RuntimesServiceServicer(object):
    """Runtimes are expected to be registered during deploy (from configuration).
    Dynamic registration of runtimes will not be supported.
    """

    def Describe(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def List(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Build(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Run(self, request, context):
        """This is a simplified version that should be used
        for testing and local runs only.
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def BidiRun(self, request_iterator, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Build1(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_RuntimesServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Describe': grpc.unary_unary_rpc_method_handler(
                    servicer.Describe,
                    request_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.DescribeRequest.FromString,
                    response_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.DescribeResponse.SerializeToString,
            ),
            'List': grpc.unary_unary_rpc_method_handler(
                    servicer.List,
                    request_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.ListRequest.FromString,
                    response_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.ListResponse.SerializeToString,
            ),
            'Build': grpc.unary_unary_rpc_method_handler(
                    servicer.Build,
                    request_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BuildRequest.FromString,
                    response_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BuildResponse.SerializeToString,
            ),
            'Run': grpc.unary_stream_rpc_method_handler(
                    servicer.Run,
                    request_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.RunRequest.FromString,
                    response_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.RunResponse.SerializeToString,
            ),
            'BidiRun': grpc.stream_stream_rpc_method_handler(
                    servicer.BidiRun,
                    request_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BidiRunRequest.FromString,
                    response_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BidiRunResponse.SerializeToString,
            ),
            'Build1': grpc.unary_unary_rpc_method_handler(
                    servicer.Build1,
                    request_deserializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.Build1Request.FromString,
                    response_serializer=autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.Build1Response.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'autokitteh.runtimes.v1.RuntimesService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class RuntimesService(object):
    """Runtimes are expected to be registered during deploy (from configuration).
    Dynamic registration of runtimes will not be supported.
    """

    @staticmethod
    def Describe(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.runtimes.v1.RuntimesService/Describe',
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.DescribeRequest.SerializeToString,
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.DescribeResponse.FromString,
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
        return grpc.experimental.unary_unary(request, target, '/autokitteh.runtimes.v1.RuntimesService/List',
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.ListRequest.SerializeToString,
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.ListResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Build(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.runtimes.v1.RuntimesService/Build',
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BuildRequest.SerializeToString,
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BuildResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Run(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_stream(request, target, '/autokitteh.runtimes.v1.RuntimesService/Run',
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.RunRequest.SerializeToString,
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.RunResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def BidiRun(request_iterator,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.stream_stream(request_iterator, target, '/autokitteh.runtimes.v1.RuntimesService/BidiRun',
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BidiRunRequest.SerializeToString,
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.BidiRunResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Build1(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.runtimes.v1.RuntimesService/Build1',
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.Build1Request.SerializeToString,
            autokitteh_dot_runtimes_dot_v1_dot_svc__pb2.Build1Response.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
