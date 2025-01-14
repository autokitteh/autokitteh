# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from autokitteh_pb.orgs.v1 import svc_pb2 as autokitteh_dot_orgs_dot_v1_dot_svc__pb2


class OrgsServiceStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.Create = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/Create',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.CreateRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.CreateResponse.FromString,
                )
        self.Get = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/Get',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetResponse.FromString,
                )
        self.Update = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/Update',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateResponse.FromString,
                )
        self.Delete = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/Delete',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.DeleteRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.DeleteResponse.FromString,
                )
        self.AddMember = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/AddMember',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.AddMemberRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.AddMemberResponse.FromString,
                )
        self.UpdateMember = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/UpdateMember',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateMemberRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateMemberResponse.FromString,
                )
        self.RemoveMember = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/RemoveMember',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.RemoveMemberRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.RemoveMemberResponse.FromString,
                )
        self.ListMembers = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/ListMembers',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.ListMembersRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.ListMembersResponse.FromString,
                )
        self.GetMember = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/GetMember',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetMemberRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetMemberResponse.FromString,
                )
        self.GetOrgsForUser = channel.unary_unary(
                '/autokitteh.orgs.v1.OrgsService/GetOrgsForUser',
                request_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetOrgsForUserRequest.SerializeToString,
                response_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetOrgsForUserResponse.FromString,
                )


class OrgsServiceServicer(object):
    """Missing associated documentation comment in .proto file."""

    def Create(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Get(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Update(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Delete(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def AddMember(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def UpdateMember(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def RemoveMember(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def ListMembers(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetMember(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetOrgsForUser(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_OrgsServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'Create': grpc.unary_unary_rpc_method_handler(
                    servicer.Create,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.CreateRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.CreateResponse.SerializeToString,
            ),
            'Get': grpc.unary_unary_rpc_method_handler(
                    servicer.Get,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetResponse.SerializeToString,
            ),
            'Update': grpc.unary_unary_rpc_method_handler(
                    servicer.Update,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateResponse.SerializeToString,
            ),
            'Delete': grpc.unary_unary_rpc_method_handler(
                    servicer.Delete,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.DeleteRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.DeleteResponse.SerializeToString,
            ),
            'AddMember': grpc.unary_unary_rpc_method_handler(
                    servicer.AddMember,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.AddMemberRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.AddMemberResponse.SerializeToString,
            ),
            'UpdateMember': grpc.unary_unary_rpc_method_handler(
                    servicer.UpdateMember,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateMemberRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateMemberResponse.SerializeToString,
            ),
            'RemoveMember': grpc.unary_unary_rpc_method_handler(
                    servicer.RemoveMember,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.RemoveMemberRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.RemoveMemberResponse.SerializeToString,
            ),
            'ListMembers': grpc.unary_unary_rpc_method_handler(
                    servicer.ListMembers,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.ListMembersRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.ListMembersResponse.SerializeToString,
            ),
            'GetMember': grpc.unary_unary_rpc_method_handler(
                    servicer.GetMember,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetMemberRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetMemberResponse.SerializeToString,
            ),
            'GetOrgsForUser': grpc.unary_unary_rpc_method_handler(
                    servicer.GetOrgsForUser,
                    request_deserializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetOrgsForUserRequest.FromString,
                    response_serializer=autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetOrgsForUserResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'autokitteh.orgs.v1.OrgsService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class OrgsService(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def Create(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/Create',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.CreateRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.CreateResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

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
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/Get',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Update(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/Update',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Delete(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/Delete',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.DeleteRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.DeleteResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def AddMember(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/AddMember',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.AddMemberRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.AddMemberResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def UpdateMember(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/UpdateMember',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateMemberRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.UpdateMemberResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def RemoveMember(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/RemoveMember',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.RemoveMemberRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.RemoveMemberResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def ListMembers(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/ListMembers',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.ListMembersRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.ListMembersResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetMember(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/GetMember',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetMemberRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetMemberResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetOrgsForUser(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/autokitteh.orgs.v1.OrgsService/GetOrgsForUser',
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetOrgsForUserRequest.SerializeToString,
            autokitteh_dot_orgs_dot_v1_dot_svc__pb2.GetOrgsForUserResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
