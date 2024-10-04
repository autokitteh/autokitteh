# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from . import remote_pb2 as autokitteh_dot_remote_dot_v1_dot_remote__pb2


class RunnerManagerStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Start = channel.unary_unary(
                '/autokitteh.remote.v1.RunnerManager/Start',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRunnerRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRunnerResponse.FromString,
                )
        self.RunnerHealth = channel.unary_unary(
                '/autokitteh.remote.v1.RunnerManager/RunnerHealth',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.RunnerHealthRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.RunnerHealthResponse.FromString,
                )
        self.Stop = channel.unary_unary(
                '/autokitteh.remote.v1.RunnerManager/Stop',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StopRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StopResponse.FromString,
                )
        self.Health = channel.unary_unary(
                '/autokitteh.remote.v1.RunnerManager/Health',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.FromString,
                )


class RunnerManagerServicer(object):
    """Missing associated documentation comment in .proto file."""

    def Start(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def RunnerHealth(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Stop(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Health(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_RunnerManagerServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Start': grpc.unary_unary_rpc_method_handler(
                    servicer.Start,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRunnerRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRunnerResponse.SerializeToString,
            ),
            'RunnerHealth': grpc.unary_unary_rpc_method_handler(
                    servicer.RunnerHealth,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.RunnerHealthRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.RunnerHealthResponse.SerializeToString,
            ),
            'Stop': grpc.unary_unary_rpc_method_handler(
                    servicer.Stop,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StopRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StopResponse.SerializeToString,
            ),
            'Health': grpc.unary_unary_rpc_method_handler(
                    servicer.Health,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'autokitteh.remote.v1.RunnerManager', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class RunnerManager(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def Start(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.RunnerManager/Start',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRunnerRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRunnerResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def RunnerHealth(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.RunnerManager/RunnerHealth',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.RunnerHealthRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.RunnerHealthResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Stop(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.RunnerManager/Stop',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.StopRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.StopResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Health(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.RunnerManager/Health',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)


class RunnerStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Exports = channel.unary_unary(
                '/autokitteh.remote.v1.Runner/Exports',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExportsRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExportsResponse.FromString,
                )
        self.Start = channel.unary_unary(
                '/autokitteh.remote.v1.Runner/Start',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartResponse.FromString,
                )
        self.Execute = channel.unary_unary(
                '/autokitteh.remote.v1.Runner/Execute',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExecuteRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExecuteResponse.FromString,
                )
        self.ActivityReply = channel.unary_unary(
                '/autokitteh.remote.v1.Runner/ActivityReply',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityReplyRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityReplyResponse.FromString,
                )
        self.Health = channel.unary_unary(
                '/autokitteh.remote.v1.Runner/Health',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.FromString,
                )


class RunnerServicer(object):
    """Missing associated documentation comment in .proto file."""

    def Exports(self, request, context):
        """Get exports
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Start(self, request, context):
        """Called at start of session
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Execute(self, request, context):
        """Execute a function in the runtime (skipped if it's a reply)
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def ActivityReply(self, request, context):
        """Reply from activity
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Health(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_RunnerServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Exports': grpc.unary_unary_rpc_method_handler(
                    servicer.Exports,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExportsRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExportsResponse.SerializeToString,
            ),
            'Start': grpc.unary_unary_rpc_method_handler(
                    servicer.Start,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartResponse.SerializeToString,
            ),
            'Execute': grpc.unary_unary_rpc_method_handler(
                    servicer.Execute,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExecuteRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExecuteResponse.SerializeToString,
            ),
            'ActivityReply': grpc.unary_unary_rpc_method_handler(
                    servicer.ActivityReply,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityReplyRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityReplyResponse.SerializeToString,
            ),
            'Health': grpc.unary_unary_rpc_method_handler(
                    servicer.Health,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'autokitteh.remote.v1.Runner', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class Runner(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def Exports(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Runner/Exports',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExportsRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExportsResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Start(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Runner/Start',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.StartResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Execute(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Runner/Execute',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExecuteRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.ExecuteResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def ActivityReply(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Runner/ActivityReply',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityReplyRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityReplyResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Health(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Runner/Health',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)


class WorkerStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Activity = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/Activity',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityResponse.FromString,
                )
        self.Done = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/Done',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.DoneRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.DoneResponse.FromString,
                )
        self.Log = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/Log',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.LogRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.LogResponse.FromString,
                )
        self.Print = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/Print',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.PrintRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.PrintResponse.FromString,
                )
        self.Sleep = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/Sleep',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.SleepRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.SleepResponse.FromString,
                )
        self.Subscribe = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/Subscribe',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.SubscribeRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.SubscribeResponse.FromString,
                )
        self.NextEvent = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/NextEvent',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.NextEventRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.NextEventResponse.FromString,
                )
        self.Unsubscribe = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/Unsubscribe',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.UnsubscribeRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.UnsubscribeResponse.FromString,
                )
        self.Health = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/Health',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.FromString,
                )
        self.IsActiveRunner = channel.unary_unary(
                '/autokitteh.remote.v1.Worker/IsActiveRunner',
                request_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.IsActiveRunnerRequest.SerializeToString,
                response_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.IsActiveRunnerResponse.FromString,
                )


class WorkerServicer(object):
    """Missing associated documentation comment in .proto file."""

    def Activity(self, request, context):
        """Runner starting activity
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Done(self, request, context):
        """Runner done with activity
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Log(self, request, context):
        """Session logs
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Print(self, request, context):
        """Print to session log
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Sleep(self, request, context):
        """ak functions
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Subscribe(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def NextEvent(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Unsubscribe(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Health(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def IsActiveRunner(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_WorkerServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Activity': grpc.unary_unary_rpc_method_handler(
                    servicer.Activity,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityResponse.SerializeToString,
            ),
            'Done': grpc.unary_unary_rpc_method_handler(
                    servicer.Done,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.DoneRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.DoneResponse.SerializeToString,
            ),
            'Log': grpc.unary_unary_rpc_method_handler(
                    servicer.Log,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.LogRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.LogResponse.SerializeToString,
            ),
            'Print': grpc.unary_unary_rpc_method_handler(
                    servicer.Print,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.PrintRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.PrintResponse.SerializeToString,
            ),
            'Sleep': grpc.unary_unary_rpc_method_handler(
                    servicer.Sleep,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.SleepRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.SleepResponse.SerializeToString,
            ),
            'Subscribe': grpc.unary_unary_rpc_method_handler(
                    servicer.Subscribe,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.SubscribeRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.SubscribeResponse.SerializeToString,
            ),
            'NextEvent': grpc.unary_unary_rpc_method_handler(
                    servicer.NextEvent,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.NextEventRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.NextEventResponse.SerializeToString,
            ),
            'Unsubscribe': grpc.unary_unary_rpc_method_handler(
                    servicer.Unsubscribe,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.UnsubscribeRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.UnsubscribeResponse.SerializeToString,
            ),
            'Health': grpc.unary_unary_rpc_method_handler(
                    servicer.Health,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.SerializeToString,
            ),
            'IsActiveRunner': grpc.unary_unary_rpc_method_handler(
                    servicer.IsActiveRunner,
                    request_deserializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.IsActiveRunnerRequest.FromString,
                    response_serializer=autokitteh_dot_remote_dot_v1_dot_remote__pb2.IsActiveRunnerResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'autokitteh.remote.v1.Worker', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class Worker(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def Activity(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/Activity',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.ActivityResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Done(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/Done',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.DoneRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.DoneResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Log(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/Log',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.LogRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.LogResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Print(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/Print',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.PrintRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.PrintResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Sleep(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/Sleep',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.SleepRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.SleepResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Subscribe(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/Subscribe',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.SubscribeRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.SubscribeResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def NextEvent(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/NextEvent',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.NextEventRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.NextEventResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Unsubscribe(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/Unsubscribe',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.UnsubscribeRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.UnsubscribeResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Health(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/Health',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.HealthResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def IsActiveRunner(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.remote.v1.Worker/IsActiveRunner',
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.IsActiveRunnerRequest.SerializeToString,
            autokitteh_dot_remote_dot_v1_dot_remote__pb2.IsActiveRunnerResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)