# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc
import warnings

import remote_pb2 as remote__pb2

GRPC_GENERATED_VERSION = '1.65.1'
GRPC_VERSION = grpc.__version__
EXPECTED_ERROR_RELEASE = '1.66.0'
SCHEDULED_RELEASE_DATE = 'August 6, 2024'
_version_not_supported = False

try:
    from grpc._utilities import first_version_is_lower
    _version_not_supported = first_version_is_lower(GRPC_VERSION, GRPC_GENERATED_VERSION)
except ImportError:
    _version_not_supported = True

if _version_not_supported:
    warnings.warn(
        f'The grpc package installed is at version {GRPC_VERSION},'
        + f' but the generated code in remote_pb2_grpc.py depends on'
        + f' grpcio>={GRPC_GENERATED_VERSION}.'
        + f' Please upgrade your grpc module to grpcio>={GRPC_GENERATED_VERSION}'
        + f' or downgrade your generated code using grpcio-tools<={GRPC_VERSION}.'
        + f' This warning will become an error in {EXPECTED_ERROR_RELEASE},'
        + f' scheduled for release on {SCHEDULED_RELEASE_DATE}.',
        RuntimeWarning
    )


class RunnerManagerStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Start = channel.unary_unary(
                '/RunnerManager/Start',
                request_serializer=remote__pb2.StartRunnerRequest.SerializeToString,
                response_deserializer=remote__pb2.StartRunnerResponse.FromString,
                _registered_method=True)
        self.RunnerHealth = channel.unary_unary(
                '/RunnerManager/RunnerHealth',
                request_serializer=remote__pb2.RunnerHealthRequest.SerializeToString,
                response_deserializer=remote__pb2.RunnerHealthResponse.FromString,
                _registered_method=True)
        self.Stop = channel.unary_unary(
                '/RunnerManager/Stop',
                request_serializer=remote__pb2.StopRequest.SerializeToString,
                response_deserializer=remote__pb2.StopResponse.FromString,
                _registered_method=True)
        self.Health = channel.unary_unary(
                '/RunnerManager/Health',
                request_serializer=remote__pb2.HealthRequest.SerializeToString,
                response_deserializer=remote__pb2.HealthResponse.FromString,
                _registered_method=True)


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
                    request_deserializer=remote__pb2.StartRunnerRequest.FromString,
                    response_serializer=remote__pb2.StartRunnerResponse.SerializeToString,
            ),
            'RunnerHealth': grpc.unary_unary_rpc_method_handler(
                    servicer.RunnerHealth,
                    request_deserializer=remote__pb2.RunnerHealthRequest.FromString,
                    response_serializer=remote__pb2.RunnerHealthResponse.SerializeToString,
            ),
            'Stop': grpc.unary_unary_rpc_method_handler(
                    servicer.Stop,
                    request_deserializer=remote__pb2.StopRequest.FromString,
                    response_serializer=remote__pb2.StopResponse.SerializeToString,
            ),
            'Health': grpc.unary_unary_rpc_method_handler(
                    servicer.Health,
                    request_deserializer=remote__pb2.HealthRequest.FromString,
                    response_serializer=remote__pb2.HealthResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'RunnerManager', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('RunnerManager', rpc_method_handlers)


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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/RunnerManager/Start',
            remote__pb2.StartRunnerRequest.SerializeToString,
            remote__pb2.StartRunnerResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/RunnerManager/RunnerHealth',
            remote__pb2.RunnerHealthRequest.SerializeToString,
            remote__pb2.RunnerHealthResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/RunnerManager/Stop',
            remote__pb2.StopRequest.SerializeToString,
            remote__pb2.StopResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/RunnerManager/Health',
            remote__pb2.HealthRequest.SerializeToString,
            remote__pb2.HealthResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)


class RunnerStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Exports = channel.unary_unary(
                '/Runner/Exports',
                request_serializer=remote__pb2.ExportsRequest.SerializeToString,
                response_deserializer=remote__pb2.ExportsResponse.FromString,
                _registered_method=True)
        self.Start = channel.unary_unary(
                '/Runner/Start',
                request_serializer=remote__pb2.StartRequest.SerializeToString,
                response_deserializer=remote__pb2.StartResponse.FromString,
                _registered_method=True)
        self.Execute = channel.unary_unary(
                '/Runner/Execute',
                request_serializer=remote__pb2.ExecuteRequest.SerializeToString,
                response_deserializer=remote__pb2.ExecuteResponse.FromString,
                _registered_method=True)
        self.ActivityReply = channel.unary_unary(
                '/Runner/ActivityReply',
                request_serializer=remote__pb2.ActivityReplyRequest.SerializeToString,
                response_deserializer=remote__pb2.ActivityReplyResponse.FromString,
                _registered_method=True)
        self.Health = channel.unary_unary(
                '/Runner/Health',
                request_serializer=remote__pb2.HealthRequest.SerializeToString,
                response_deserializer=remote__pb2.HealthResponse.FromString,
                _registered_method=True)


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
                    request_deserializer=remote__pb2.ExportsRequest.FromString,
                    response_serializer=remote__pb2.ExportsResponse.SerializeToString,
            ),
            'Start': grpc.unary_unary_rpc_method_handler(
                    servicer.Start,
                    request_deserializer=remote__pb2.StartRequest.FromString,
                    response_serializer=remote__pb2.StartResponse.SerializeToString,
            ),
            'Execute': grpc.unary_unary_rpc_method_handler(
                    servicer.Execute,
                    request_deserializer=remote__pb2.ExecuteRequest.FromString,
                    response_serializer=remote__pb2.ExecuteResponse.SerializeToString,
            ),
            'ActivityReply': grpc.unary_unary_rpc_method_handler(
                    servicer.ActivityReply,
                    request_deserializer=remote__pb2.ActivityReplyRequest.FromString,
                    response_serializer=remote__pb2.ActivityReplyResponse.SerializeToString,
            ),
            'Health': grpc.unary_unary_rpc_method_handler(
                    servicer.Health,
                    request_deserializer=remote__pb2.HealthRequest.FromString,
                    response_serializer=remote__pb2.HealthResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'Runner', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('Runner', rpc_method_handlers)


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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Runner/Exports',
            remote__pb2.ExportsRequest.SerializeToString,
            remote__pb2.ExportsResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Runner/Start',
            remote__pb2.StartRequest.SerializeToString,
            remote__pb2.StartResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Runner/Execute',
            remote__pb2.ExecuteRequest.SerializeToString,
            remote__pb2.ExecuteResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Runner/ActivityReply',
            remote__pb2.ActivityReplyRequest.SerializeToString,
            remote__pb2.ActivityReplyResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Runner/Health',
            remote__pb2.HealthRequest.SerializeToString,
            remote__pb2.HealthResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)


class WorkerStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Activity = channel.unary_unary(
                '/Worker/Activity',
                request_serializer=remote__pb2.ActivityRequest.SerializeToString,
                response_deserializer=remote__pb2.ActivityResponse.FromString,
                _registered_method=True)
        self.Done = channel.unary_unary(
                '/Worker/Done',
                request_serializer=remote__pb2.DoneRequest.SerializeToString,
                response_deserializer=remote__pb2.DoneResponse.FromString,
                _registered_method=True)
        self.Log = channel.unary_unary(
                '/Worker/Log',
                request_serializer=remote__pb2.LogRequest.SerializeToString,
                response_deserializer=remote__pb2.LogResponse.FromString,
                _registered_method=True)
        self.Print = channel.unary_unary(
                '/Worker/Print',
                request_serializer=remote__pb2.PrintRequest.SerializeToString,
                response_deserializer=remote__pb2.PrintResponse.FromString,
                _registered_method=True)
        self.Sleep = channel.unary_unary(
                '/Worker/Sleep',
                request_serializer=remote__pb2.SleepRequest.SerializeToString,
                response_deserializer=remote__pb2.SleepResponse.FromString,
                _registered_method=True)
        self.Subscribe = channel.unary_unary(
                '/Worker/Subscribe',
                request_serializer=remote__pb2.SubscribeRequest.SerializeToString,
                response_deserializer=remote__pb2.SubscribeResponse.FromString,
                _registered_method=True)
        self.NextEvent = channel.unary_unary(
                '/Worker/NextEvent',
                request_serializer=remote__pb2.NextEventRequest.SerializeToString,
                response_deserializer=remote__pb2.NextEventResponse.FromString,
                _registered_method=True)
        self.Unsubscribe = channel.unary_unary(
                '/Worker/Unsubscribe',
                request_serializer=remote__pb2.UnsubscribeRequest.SerializeToString,
                response_deserializer=remote__pb2.UnsubscribeResponse.FromString,
                _registered_method=True)
        self.Health = channel.unary_unary(
                '/Worker/Health',
                request_serializer=remote__pb2.HealthRequest.SerializeToString,
                response_deserializer=remote__pb2.HealthResponse.FromString,
                _registered_method=True)


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


def add_WorkerServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Activity': grpc.unary_unary_rpc_method_handler(
                    servicer.Activity,
                    request_deserializer=remote__pb2.ActivityRequest.FromString,
                    response_serializer=remote__pb2.ActivityResponse.SerializeToString,
            ),
            'Done': grpc.unary_unary_rpc_method_handler(
                    servicer.Done,
                    request_deserializer=remote__pb2.DoneRequest.FromString,
                    response_serializer=remote__pb2.DoneResponse.SerializeToString,
            ),
            'Log': grpc.unary_unary_rpc_method_handler(
                    servicer.Log,
                    request_deserializer=remote__pb2.LogRequest.FromString,
                    response_serializer=remote__pb2.LogResponse.SerializeToString,
            ),
            'Print': grpc.unary_unary_rpc_method_handler(
                    servicer.Print,
                    request_deserializer=remote__pb2.PrintRequest.FromString,
                    response_serializer=remote__pb2.PrintResponse.SerializeToString,
            ),
            'Sleep': grpc.unary_unary_rpc_method_handler(
                    servicer.Sleep,
                    request_deserializer=remote__pb2.SleepRequest.FromString,
                    response_serializer=remote__pb2.SleepResponse.SerializeToString,
            ),
            'Subscribe': grpc.unary_unary_rpc_method_handler(
                    servicer.Subscribe,
                    request_deserializer=remote__pb2.SubscribeRequest.FromString,
                    response_serializer=remote__pb2.SubscribeResponse.SerializeToString,
            ),
            'NextEvent': grpc.unary_unary_rpc_method_handler(
                    servicer.NextEvent,
                    request_deserializer=remote__pb2.NextEventRequest.FromString,
                    response_serializer=remote__pb2.NextEventResponse.SerializeToString,
            ),
            'Unsubscribe': grpc.unary_unary_rpc_method_handler(
                    servicer.Unsubscribe,
                    request_deserializer=remote__pb2.UnsubscribeRequest.FromString,
                    response_serializer=remote__pb2.UnsubscribeResponse.SerializeToString,
            ),
            'Health': grpc.unary_unary_rpc_method_handler(
                    servicer.Health,
                    request_deserializer=remote__pb2.HealthRequest.FromString,
                    response_serializer=remote__pb2.HealthResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'Worker', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('Worker', rpc_method_handlers)


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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/Activity',
            remote__pb2.ActivityRequest.SerializeToString,
            remote__pb2.ActivityResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/Done',
            remote__pb2.DoneRequest.SerializeToString,
            remote__pb2.DoneResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/Log',
            remote__pb2.LogRequest.SerializeToString,
            remote__pb2.LogResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/Print',
            remote__pb2.PrintRequest.SerializeToString,
            remote__pb2.PrintResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/Sleep',
            remote__pb2.SleepRequest.SerializeToString,
            remote__pb2.SleepResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/Subscribe',
            remote__pb2.SubscribeRequest.SerializeToString,
            remote__pb2.SubscribeResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/NextEvent',
            remote__pb2.NextEventRequest.SerializeToString,
            remote__pb2.NextEventResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/Unsubscribe',
            remote__pb2.UnsubscribeRequest.SerializeToString,
            remote__pb2.UnsubscribeResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

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
        return grpc.experimental.unary_unary(
            request,
            target,
            '/Worker/Health',
            remote__pb2.HealthRequest.SerializeToString,
            remote__pb2.HealthResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)
