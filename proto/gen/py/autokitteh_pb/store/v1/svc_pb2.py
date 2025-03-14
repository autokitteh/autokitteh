# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: autokitteh/store/v1/svc.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from autokitteh_pb.values.v1 import values_pb2 as autokitteh_dot_values_dot_v1_dot_values__pb2
from buf.validate import validate_pb2 as buf_dot_validate_dot_validate__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1d\x61utokitteh/store/v1/svc.proto\x12\x13\x61utokitteh.store.v1\x1a!autokitteh/values/v1/values.proto\x1a\x1b\x62uf/validate/validate.proto\"N\n\nGetRequest\x12\x1d\n\nproject_id\x18\x01 \x01(\tR\tprojectId\x12!\n\x04keys\x18\x02 \x03(\tB\r\xfa\xf7\x18\t\x92\x01\x06\"\x04r\x02\x10\x01R\x04keys\"\xbf\x01\n\x0bGetResponse\x12X\n\x06values\x18\x01 \x03(\x0b\x32,.autokitteh.store.v1.GetResponse.ValuesEntryB\x12\xfa\xf7\x18\x0e\x9a\x01\x0b\"\x04r\x02\x10\x01*\x03\xc8\x01\x01R\x06values\x1aV\n\x0bValuesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x31\n\x05value\x18\x02 \x01(\x0b\x32\x1b.autokitteh.values.v1.ValueR\x05value:\x02\x38\x01\",\n\x0bListRequest\x12\x1d\n\nproject_id\x18\x01 \x01(\tR\tprojectId\"1\n\x0cListResponse\x12!\n\x04keys\x18\x01 \x03(\tB\r\xfa\xf7\x18\t\x92\x01\x06\"\x04r\x02\x10\x01R\x04keys2\xa5\x01\n\x0cStoreService\x12H\n\x03Get\x12\x1f.autokitteh.store.v1.GetRequest\x1a .autokitteh.store.v1.GetResponse\x12K\n\x04List\x12 .autokitteh.store.v1.ListRequest\x1a!.autokitteh.store.v1.ListResponseB\xd8\x01\n\x17\x63om.autokitteh.store.v1B\x08SvcProtoP\x01ZEgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1;storev1\xa2\x02\x03\x41SX\xaa\x02\x13\x41utokitteh.Store.V1\xca\x02\x13\x41utokitteh\\Store\\V1\xe2\x02\x1f\x41utokitteh\\Store\\V1\\GPBMetadata\xea\x02\x15\x41utokitteh::Store::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'autokitteh.store.v1.svc_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\027com.autokitteh.store.v1B\010SvcProtoP\001ZEgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1;storev1\242\002\003ASX\252\002\023Autokitteh.Store.V1\312\002\023Autokitteh\\Store\\V1\342\002\037Autokitteh\\Store\\V1\\GPBMetadata\352\002\025Autokitteh::Store::V1'
  _GETREQUEST.fields_by_name['keys']._options = None
  _GETREQUEST.fields_by_name['keys']._serialized_options = b'\372\367\030\t\222\001\006\"\004r\002\020\001'
  _GETRESPONSE_VALUESENTRY._options = None
  _GETRESPONSE_VALUESENTRY._serialized_options = b'8\001'
  _GETRESPONSE.fields_by_name['values']._options = None
  _GETRESPONSE.fields_by_name['values']._serialized_options = b'\372\367\030\016\232\001\013\"\004r\002\020\001*\003\310\001\001'
  _LISTRESPONSE.fields_by_name['keys']._options = None
  _LISTRESPONSE.fields_by_name['keys']._serialized_options = b'\372\367\030\t\222\001\006\"\004r\002\020\001'
  _globals['_GETREQUEST']._serialized_start=118
  _globals['_GETREQUEST']._serialized_end=196
  _globals['_GETRESPONSE']._serialized_start=199
  _globals['_GETRESPONSE']._serialized_end=390
  _globals['_GETRESPONSE_VALUESENTRY']._serialized_start=304
  _globals['_GETRESPONSE_VALUESENTRY']._serialized_end=390
  _globals['_LISTREQUEST']._serialized_start=392
  _globals['_LISTREQUEST']._serialized_end=436
  _globals['_LISTRESPONSE']._serialized_start=438
  _globals['_LISTRESPONSE']._serialized_end=487
  _globals['_STORESERVICE']._serialized_start=490
  _globals['_STORESERVICE']._serialized_end=655
# @@protoc_insertion_point(module_scope)
