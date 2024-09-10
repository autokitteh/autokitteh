# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: autokitteh/runner_manager/v1/svc.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n&autokitteh/runner_manager/v1/svc.proto\x12\x1c\x61utokitteh.runner_manager.v1\"\'\n\x0f\x43ontainerConfig\x12\x14\n\x05image\x18\x01 \x01(\tR\x05image\"\xb9\x02\n\x0cStartRequest\x12X\n\x10\x63ontainer_config\x18\x01 \x01(\x0b\x32-.autokitteh.runner_manager.v1.ContainerConfigR\x0f\x63ontainerConfig\x12%\n\x0e\x62uild_artifact\x18\x02 \x01(\x0cR\rbuildArtifact\x12H\n\x04vars\x18\x03 \x03(\x0b\x32\x34.autokitteh.runner_manager.v1.StartRequest.VarsEntryR\x04vars\x12%\n\x0eworker_address\x18\x04 \x01(\tR\rworkerAddress\x1a\x37\n\tVarsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"i\n\rStartResponse\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12%\n\x0erunner_address\x18\x02 \x01(\tR\rrunnerAddress\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\"2\n\x13RunnerHealthRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"F\n\x14RunnerHealthResponse\x12\x18\n\x07healthy\x18\x01 \x01(\x08R\x07healthy\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"*\n\x0bStopRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"$\n\x0cStopResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"\x0f\n\rHealthRequest\"&\n\x0eHealthResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror2\xbb\x03\n\x14RunnerManagerService\x12\x62\n\x05Start\x12*.autokitteh.runner_manager.v1.StartRequest\x1a+.autokitteh.runner_manager.v1.StartResponse\"\x00\x12w\n\x0cRunnerHealth\x12\x31.autokitteh.runner_manager.v1.RunnerHealthRequest\x1a\x32.autokitteh.runner_manager.v1.RunnerHealthResponse\"\x00\x12_\n\x04Stop\x12).autokitteh.runner_manager.v1.StopRequest\x1a*.autokitteh.runner_manager.v1.StopResponse\"\x00\x12\x65\n\x06Health\x12+.autokitteh.runner_manager.v1.HealthRequest\x1a,.autokitteh.runner_manager.v1.HealthResponse\"\x00\x42\x93\x02\n com.autokitteh.runner_manager.v1B\x08SvcProtoP\x01ZWgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runner_manager/v1;runner_managerv1\xa2\x02\x03\x41RX\xaa\x02\x1b\x41utokitteh.RunnerManager.V1\xca\x02\x1b\x41utokitteh\\RunnerManager\\V1\xe2\x02\'Autokitteh\\RunnerManager\\V1\\GPBMetadata\xea\x02\x1d\x41utokitteh::RunnerManager::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'autokitteh.runner_manager.v1.svc_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n com.autokitteh.runner_manager.v1B\010SvcProtoP\001ZWgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runner_manager/v1;runner_managerv1\242\002\003ARX\252\002\033Autokitteh.RunnerManager.V1\312\002\033Autokitteh\\RunnerManager\\V1\342\002\'Autokitteh\\RunnerManager\\V1\\GPBMetadata\352\002\035Autokitteh::RunnerManager::V1'
  _STARTREQUEST_VARSENTRY._options = None
  _STARTREQUEST_VARSENTRY._serialized_options = b'8\001'
  _globals['_CONTAINERCONFIG']._serialized_start=72
  _globals['_CONTAINERCONFIG']._serialized_end=111
  _globals['_STARTREQUEST']._serialized_start=114
  _globals['_STARTREQUEST']._serialized_end=427
  _globals['_STARTREQUEST_VARSENTRY']._serialized_start=372
  _globals['_STARTREQUEST_VARSENTRY']._serialized_end=427
  _globals['_STARTRESPONSE']._serialized_start=429
  _globals['_STARTRESPONSE']._serialized_end=534
  _globals['_RUNNERHEALTHREQUEST']._serialized_start=536
  _globals['_RUNNERHEALTHREQUEST']._serialized_end=586
  _globals['_RUNNERHEALTHRESPONSE']._serialized_start=588
  _globals['_RUNNERHEALTHRESPONSE']._serialized_end=658
  _globals['_STOPREQUEST']._serialized_start=660
  _globals['_STOPREQUEST']._serialized_end=702
  _globals['_STOPRESPONSE']._serialized_start=704
  _globals['_STOPRESPONSE']._serialized_end=740
  _globals['_HEALTHREQUEST']._serialized_start=742
  _globals['_HEALTHREQUEST']._serialized_end=757
  _globals['_HEALTHRESPONSE']._serialized_start=759
  _globals['_HEALTHRESPONSE']._serialized_end=797
  _globals['_RUNNERMANAGERSERVICE']._serialized_start=800
  _globals['_RUNNERMANAGERSERVICE']._serialized_end=1243
# @@protoc_insertion_point(module_scope)