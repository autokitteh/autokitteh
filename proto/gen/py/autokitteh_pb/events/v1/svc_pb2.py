# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: autokitteh/events/v1/svc.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from autokitteh_pb.events.v1 import event_pb2 as autokitteh_dot_events_dot_v1_dot_event__pb2
from buf.validate import validate_pb2 as buf_dot_validate_dot_validate__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1e\x61utokitteh/events/v1/svc.proto\x12\x14\x61utokitteh.events.v1\x1a autokitteh/events/v1/event.proto\x1a\x1b\x62uf/validate/validate.proto\"\xa4\x02\n\x0bSaveRequest\x12\x31\n\x05\x65vent\x18\x01 \x01(\x0b\x32\x1b.autokitteh.events.v1.EventR\x05\x65vent:\xe1\x01\xfa\xf7\x18\xdc\x01\x1ak\n\x1d\x65vents.missing_destination_id\x12\x16missing destination_id\x1a\x32has(this.event) && this.event.destination_id != \'\'\x1am\n\x1d\x65vents.event_id_must_be_empty\x12\x1e\x65vent_id must not be specified\x1a,has(this.event) && this.event.event_id == \'\'\"3\n\x0cSaveResponse\x12#\n\x08\x65vent_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\x07\x65ventId\"R\n\nGetRequest\x12#\n\x08\x65vent_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\x07\x65ventId\x12\x1f\n\x0bjson_values\x18\x02 \x01(\x08R\njsonValues\"@\n\x0bGetResponse\x12\x31\n\x05\x65vent\x18\x01 \x01(\x0b\x32\x1b.autokitteh.events.v1.EventR\x05\x65vent\"\x88\x02\n\x0bListRequest\x12%\n\x0eintegration_id\x18\x01 \x01(\tR\rintegrationId\x12%\n\x0e\x64\x65stination_id\x18\x02 \x01(\tR\rdestinationId\x12\x1d\n\nevent_type\x18\x03 \x01(\tR\teventType\x12\x1f\n\x0bmax_results\x18\x04 \x01(\rR\nmaxResults\x12\x14\n\x05order\x18\x05 \x01(\tR\x05order\x12\x1d\n\nproject_id\x18\x07 \x01(\tR\tprojectId\x12\x15\n\x06org_id\x18\x08 \x01(\tR\x05orgId\x12\x1f\n\x0bjson_values\x18\x06 \x01(\x08R\njsonValues\"C\n\x0cListResponse\x12\x33\n\x06\x65vents\x18\x01 \x03(\x0b\x32\x1b.autokitteh.events.v1.EventR\x06\x65vents2\xf9\x01\n\rEventsService\x12M\n\x04Save\x12!.autokitteh.events.v1.SaveRequest\x1a\".autokitteh.events.v1.SaveResponse\x12J\n\x03Get\x12 .autokitteh.events.v1.GetRequest\x1a!.autokitteh.events.v1.GetResponse\x12M\n\x04List\x12!.autokitteh.events.v1.ListRequest\x1a\".autokitteh.events.v1.ListResponseB\xdf\x01\n\x18\x63om.autokitteh.events.v1B\x08SvcProtoP\x01ZGgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1;eventsv1\xa2\x02\x03\x41\x45X\xaa\x02\x14\x41utokitteh.Events.V1\xca\x02\x14\x41utokitteh\\Events\\V1\xe2\x02 Autokitteh\\Events\\V1\\GPBMetadata\xea\x02\x16\x41utokitteh::Events::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'autokitteh.events.v1.svc_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\030com.autokitteh.events.v1B\010SvcProtoP\001ZGgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1;eventsv1\242\002\003AEX\252\002\024Autokitteh.Events.V1\312\002\024Autokitteh\\Events\\V1\342\002 Autokitteh\\Events\\V1\\GPBMetadata\352\002\026Autokitteh::Events::V1'
  _SAVEREQUEST._options = None
  _SAVEREQUEST._serialized_options = b'\372\367\030\334\001\032k\n\035events.missing_destination_id\022\026missing destination_id\0322has(this.event) && this.event.destination_id != \'\'\032m\n\035events.event_id_must_be_empty\022\036event_id must not be specified\032,has(this.event) && this.event.event_id == \'\''
  _SAVERESPONSE.fields_by_name['event_id']._options = None
  _SAVERESPONSE.fields_by_name['event_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _GETREQUEST.fields_by_name['event_id']._options = None
  _GETREQUEST.fields_by_name['event_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _globals['_SAVEREQUEST']._serialized_start=120
  _globals['_SAVEREQUEST']._serialized_end=412
  _globals['_SAVERESPONSE']._serialized_start=414
  _globals['_SAVERESPONSE']._serialized_end=465
  _globals['_GETREQUEST']._serialized_start=467
  _globals['_GETREQUEST']._serialized_end=549
  _globals['_GETRESPONSE']._serialized_start=551
  _globals['_GETRESPONSE']._serialized_end=615
  _globals['_LISTREQUEST']._serialized_start=618
  _globals['_LISTREQUEST']._serialized_end=882
  _globals['_LISTRESPONSE']._serialized_start=884
  _globals['_LISTRESPONSE']._serialized_end=951
  _globals['_EVENTSSERVICE']._serialized_start=954
  _globals['_EVENTSSERVICE']._serialized_end=1203
# @@protoc_insertion_point(module_scope)
