# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: autokitteh/notifications/v1/svc.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from autokitteh_pb.notifications.v1 import notification_pb2 as autokitteh_dot_notifications_dot_v1_dot_notification__pb2
from buf.validate import validate_pb2 as buf_dot_validate_dot_validate__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n%autokitteh/notifications/v1/svc.proto\x12\x1b\x61utokitteh.notifications.v1\x1a.autokitteh/notifications/v1/notification.proto\x1a\x1b\x62uf/validate/validate.proto\"e\n\x0bSendRequest\x12V\n\x0cnotification\x18\x01 \x01(\x0b\x32).autokitteh.notifications.v1.NotificationB\x07\xfa\xf7\x18\x03\xc8\x01\x01R\x0cnotification\"A\n\x0cSendResponse\x12\x31\n\x0fnotification_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\x0enotificationId\"\xab\x01\n\x0bListRequest\x12+\n\x0crecipient_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\x0brecipientId\x12\x12\n\x04type\x18\x02 \x01(\tR\x04type\x12\x1f\n\x0bunread_only\x18\x03 \x01(\x08R\nunreadOnly\x12\x1d\n\ncount_only\x18\x04 \x01(\x08R\tcountOnly\x12\x1b\n\tpage_size\x18\x05 \x01(\x05R\x08pageSize\"m\n\x0cListResponse\x12]\n\rnotifications\x18\x01 \x03(\x0b\x32).autokitteh.notifications.v1.NotificationB\x0c\xfa\xf7\x18\x08\x92\x01\x05\"\x03\xc8\x01\x01R\rnotifications\"<\n\x11MarkAsReadRequest\x12\'\n\x0fnotification_id\x18\x01 \x01(\tR\x0enotificationId\"\x14\n\x12MarkAsReadResponse2\xbf\x02\n\x14NotificationsService\x12[\n\x04Send\x12(.autokitteh.notifications.v1.SendRequest\x1a).autokitteh.notifications.v1.SendResponse\x12[\n\x04List\x12(.autokitteh.notifications.v1.ListRequest\x1a).autokitteh.notifications.v1.ListResponse\x12m\n\nMarkAsRead\x12..autokitteh.notifications.v1.MarkAsReadRequest\x1a/.autokitteh.notifications.v1.MarkAsReadResponseB\x90\x02\n\x1f\x63om.autokitteh.notifications.v1B\x08SvcProtoP\x01ZUgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/notifications/v1;notificationsv1\xa2\x02\x03\x41NX\xaa\x02\x1b\x41utokitteh.Notifications.V1\xca\x02\x1b\x41utokitteh\\Notifications\\V1\xe2\x02\'Autokitteh\\Notifications\\V1\\GPBMetadata\xea\x02\x1d\x41utokitteh::Notifications::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'autokitteh.notifications.v1.svc_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\037com.autokitteh.notifications.v1B\010SvcProtoP\001ZUgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/notifications/v1;notificationsv1\242\002\003ANX\252\002\033Autokitteh.Notifications.V1\312\002\033Autokitteh\\Notifications\\V1\342\002\'Autokitteh\\Notifications\\V1\\GPBMetadata\352\002\035Autokitteh::Notifications::V1'
  _SENDREQUEST.fields_by_name['notification']._options = None
  _SENDREQUEST.fields_by_name['notification']._serialized_options = b'\372\367\030\003\310\001\001'
  _SENDRESPONSE.fields_by_name['notification_id']._options = None
  _SENDRESPONSE.fields_by_name['notification_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _LISTREQUEST.fields_by_name['recipient_id']._options = None
  _LISTREQUEST.fields_by_name['recipient_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _LISTRESPONSE.fields_by_name['notifications']._options = None
  _LISTRESPONSE.fields_by_name['notifications']._serialized_options = b'\372\367\030\010\222\001\005\"\003\310\001\001'
  _globals['_SENDREQUEST']._serialized_start=147
  _globals['_SENDREQUEST']._serialized_end=248
  _globals['_SENDRESPONSE']._serialized_start=250
  _globals['_SENDRESPONSE']._serialized_end=315
  _globals['_LISTREQUEST']._serialized_start=318
  _globals['_LISTREQUEST']._serialized_end=489
  _globals['_LISTRESPONSE']._serialized_start=491
  _globals['_LISTRESPONSE']._serialized_end=600
  _globals['_MARKASREADREQUEST']._serialized_start=602
  _globals['_MARKASREADREQUEST']._serialized_end=662
  _globals['_MARKASREADRESPONSE']._serialized_start=664
  _globals['_MARKASREADRESPONSE']._serialized_end=684
  _globals['_NOTIFICATIONSSERVICE']._serialized_start=687
  _globals['_NOTIFICATIONSSERVICE']._serialized_end=1006
# @@protoc_insertion_point(module_scope)
