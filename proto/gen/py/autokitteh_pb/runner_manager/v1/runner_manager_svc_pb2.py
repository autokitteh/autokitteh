# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: autokitteh/runner_manager/v1/runner_manager_svc.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n5autokitteh/runner_manager/v1/runner_manager_svc.proto\x12\x1c\x61utokitteh.runner_manager.v1\"\'\n\x0f\x43ontainerConfig\x12\x14\n\x05image\x18\x01 \x01(\tR\x05image\"\xc5\x02\n\x12StartRunnerRequest\x12X\n\x10\x63ontainer_config\x18\x01 \x01(\x0b\x32-.autokitteh.runner_manager.v1.ContainerConfigR\x0f\x63ontainerConfig\x12%\n\x0e\x62uild_artifact\x18\x02 \x01(\x0cR\rbuildArtifact\x12N\n\x04vars\x18\x03 \x03(\x0b\x32:.autokitteh.runner_manager.v1.StartRunnerRequest.VarsEntryR\x04vars\x12%\n\x0eworker_address\x18\x04 \x01(\tR\rworkerAddress\x1a\x37\n\tVarsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"o\n\x13StartRunnerResponse\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12%\n\x0erunner_address\x18\x02 \x01(\tR\rrunnerAddress\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\"2\n\x13RunnerHealthRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"F\n\x14RunnerHealthResponse\x12\x18\n\x07healthy\x18\x01 \x01(\x08R\x07healthy\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"0\n\x11StopRunnerRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"*\n\x12StopRunnerResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"\x0f\n\rHealthRequest\"&\n\x0eHealthResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror2\xdf\x03\n\x14RunnerManagerService\x12t\n\x0bStartRunner\x12\x30.autokitteh.runner_manager.v1.StartRunnerRequest\x1a\x31.autokitteh.runner_manager.v1.StartRunnerResponse\"\x00\x12w\n\x0cRunnerHealth\x12\x31.autokitteh.runner_manager.v1.RunnerHealthRequest\x1a\x32.autokitteh.runner_manager.v1.RunnerHealthResponse\"\x00\x12q\n\nStopRunner\x12/.autokitteh.runner_manager.v1.StopRunnerRequest\x1a\x30.autokitteh.runner_manager.v1.StopRunnerResponse\"\x00\x12\x65\n\x06Health\x12+.autokitteh.runner_manager.v1.HealthRequest\x1a,.autokitteh.runner_manager.v1.HealthResponse\"\x00\x42\xa0\x02\n com.autokitteh.runner_manager.v1B\x15RunnerManagerSvcProtoP\x01ZWgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runner_manager/v1;runner_managerv1\xa2\x02\x03\x41RX\xaa\x02\x1b\x41utokitteh.RunnerManager.V1\xca\x02\x1b\x41utokitteh\\RunnerManager\\V1\xe2\x02\'Autokitteh\\RunnerManager\\V1\\GPBMetadata\xea\x02\x1d\x41utokitteh::RunnerManager::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'autokitteh.runner_manager.v1.runner_manager_svc_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n com.autokitteh.runner_manager.v1B\025RunnerManagerSvcProtoP\001ZWgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runner_manager/v1;runner_managerv1\242\002\003ARX\252\002\033Autokitteh.RunnerManager.V1\312\002\033Autokitteh\\RunnerManager\\V1\342\002\'Autokitteh\\RunnerManager\\V1\\GPBMetadata\352\002\035Autokitteh::RunnerManager::V1'
  _STARTRUNNERREQUEST_VARSENTRY._options = None
  _STARTRUNNERREQUEST_VARSENTRY._serialized_options = b'8\001'
  _globals['_CONTAINERCONFIG']._serialized_start=87
  _globals['_CONTAINERCONFIG']._serialized_end=126
  _globals['_STARTRUNNERREQUEST']._serialized_start=129
  _globals['_STARTRUNNERREQUEST']._serialized_end=454
  _globals['_STARTRUNNERREQUEST_VARSENTRY']._serialized_start=399
  _globals['_STARTRUNNERREQUEST_VARSENTRY']._serialized_end=454
  _globals['_STARTRUNNERRESPONSE']._serialized_start=456
  _globals['_STARTRUNNERRESPONSE']._serialized_end=567
  _globals['_RUNNERHEALTHREQUEST']._serialized_start=569
  _globals['_RUNNERHEALTHREQUEST']._serialized_end=619
  _globals['_RUNNERHEALTHRESPONSE']._serialized_start=621
  _globals['_RUNNERHEALTHRESPONSE']._serialized_end=691
  _globals['_STOPRUNNERREQUEST']._serialized_start=693
  _globals['_STOPRUNNERREQUEST']._serialized_end=741
  _globals['_STOPRUNNERRESPONSE']._serialized_start=743
  _globals['_STOPRUNNERRESPONSE']._serialized_end=785
  _globals['_HEALTHREQUEST']._serialized_start=787
  _globals['_HEALTHREQUEST']._serialized_end=802
  _globals['_HEALTHRESPONSE']._serialized_start=804
  _globals['_HEALTHRESPONSE']._serialized_end=842
  _globals['_RUNNERMANAGERSERVICE']._serialized_start=845
  _globals['_RUNNERMANAGERSERVICE']._serialized_end=1324
# @@protoc_insertion_point(module_scope)
