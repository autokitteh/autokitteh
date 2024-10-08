# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: autokitteh/remote/v1/remote.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n!autokitteh/remote/v1/remote.proto\x12\x14\x61utokitteh.remote.v1\"\'\n\x0f\x43ontainerConfig\x12\x14\n\x05image\x18\x01 \x01(\tR\x05image\"\x1b\n\x05\x45vent\x12\x12\n\x04\x64\x61ta\x18\x01 \x01(\x0cR\x04\x64\x61ta\"\x0f\n\rHealthRequest\"&\n\x0eHealthResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"\xb5\x02\n\x12StartRunnerRequest\x12P\n\x10\x63ontainer_config\x18\x01 \x01(\x0b\x32%.autokitteh.remote.v1.ContainerConfigR\x0f\x63ontainerConfig\x12%\n\x0e\x62uild_artifact\x18\x02 \x01(\x0cR\rbuildArtifact\x12\x46\n\x04vars\x18\x03 \x03(\x0b\x32\x32.autokitteh.remote.v1.StartRunnerRequest.VarsEntryR\x04vars\x12%\n\x0eworker_address\x18\x04 \x01(\tR\rworkerAddress\x1a\x37\n\tVarsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"o\n\x13StartRunnerResponse\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12%\n\x0erunner_address\x18\x02 \x01(\tR\rrunnerAddress\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\"2\n\x13RunnerHealthRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"F\n\x14RunnerHealthResponse\x12\x18\n\x07healthy\x18\x01 \x01(\x08R\x07healthy\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"*\n\x0bStopRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"$\n\x0cStopResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"b\n\x0cStartRequest\x12\x1f\n\x0b\x65ntry_point\x18\x01 \x01(\tR\nentryPoint\x12\x31\n\x05\x65vent\x18\x02 \x01(\x0b\x32\x1b.autokitteh.remote.v1.EventR\x05\x65vent\"`\n\rStartResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\x12\x39\n\ttraceback\x18\x02 \x03(\x0b\x32\x1b.autokitteh.remote.v1.FrameR\ttraceback\"c\n\x05\x46rame\x12\x1a\n\x08\x66ilename\x18\x01 \x01(\tR\x08\x66ilename\x12\x16\n\x06lineno\x18\x02 \x01(\rR\x06lineno\x12\x12\n\x04\x63ode\x18\x03 \x01(\tR\x04\x63ode\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\"$\n\x0e\x45xecuteRequest\x12\x12\n\x04\x64\x61ta\x18\x01 \x01(\x0cR\x04\x64\x61ta\"z\n\x0f\x45xecuteResponse\x12\x16\n\x06result\x18\x01 \x01(\x0cR\x06result\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\x12\x39\n\ttraceback\x18\x03 \x03(\x0b\x32\x1b.autokitteh.remote.v1.FrameR\ttraceback\"X\n\x14\x41\x63tivityReplyRequest\x12\x12\n\x04\x64\x61ta\x18\x01 \x01(\x0cR\x04\x64\x61ta\x12\x16\n\x06result\x18\x02 \x01(\x0cR\x06result\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\"-\n\x15\x41\x63tivityReplyResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"-\n\x0e\x45xportsRequest\x12\x1b\n\tfile_name\x18\x01 \x01(\tR\x08\x66ileName\"A\n\x0f\x45xportsResponse\x12\x18\n\x07\x65xports\x18\x01 \x03(\tR\x07\x65xports\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"\xb9\x01\n\x08\x43\x61llInfo\x12\x1a\n\x08\x66unction\x18\x01 \x01(\tR\x08\x66unction\x12\x12\n\x04\x61rgs\x18\x02 \x03(\tR\x04\x61rgs\x12\x42\n\x06kwargs\x18\x03 \x03(\x0b\x32*.autokitteh.remote.v1.CallInfo.KwargsEntryR\x06kwargs\x1a\x39\n\x0bKwargsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"\x7f\n\x0f\x41\x63tivityRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x12\n\x04\x64\x61ta\x18\x02 \x01(\x0cR\x04\x64\x61ta\x12;\n\tcall_info\x18\x03 \x01(\x0b\x32\x1e.autokitteh.remote.v1.CallInfoR\x08\x63\x61llInfo\"(\n\x10\x41\x63tivityResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"\x93\x01\n\x0b\x44oneRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x16\n\x06result\x18\x02 \x01(\x0cR\x06result\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\x12\x39\n\ttraceback\x18\x04 \x03(\x0b\x32\x1b.autokitteh.remote.v1.FrameR\ttraceback\"\x0e\n\x0c\x44oneResponse\"L\n\x0cSleepRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x1f\n\x0b\x64uration_ms\x18\x02 \x01(\x03R\ndurationMs\"%\n\rSleepResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"g\n\x10SubscribeRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x1e\n\nconnection\x18\x02 \x01(\tR\nconnection\x12\x16\n\x06\x66ilter\x18\x03 \x01(\tR\x06\x66ilter\"F\n\x11SubscribeResponse\x12\x1b\n\tsignal_id\x18\x01 \x01(\tR\x08signalId\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"m\n\x10NextEventRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x1d\n\nsignal_ids\x18\x02 \x03(\tR\tsignalIds\x12\x1d\n\ntimeout_ms\x18\x03 \x01(\x03R\ttimeoutMs\"\\\n\x11NextEventResponse\x12\x31\n\x05\x65vent\x18\x01 \x01(\x0b\x32\x1b.autokitteh.remote.v1.EventR\x05\x65vent\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\"N\n\x12UnsubscribeRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x1b\n\tsignal_id\x18\x02 \x01(\tR\x08signalId\"+\n\x13UnsubscribeResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"Y\n\nLogRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x14\n\x05level\x18\x02 \x01(\tR\x05level\x12\x18\n\x07message\x18\x03 \x01(\tR\x07message\"#\n\x0bLogResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"E\n\x0cPrintRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x18\n\x07message\x18\x02 \x01(\tR\x07message\"%\n\rPrintResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"4\n\x15IsActiveRunnerRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"K\n\x16IsActiveRunnerResponse\x12\x1b\n\tis_active\x18\x01 \x01(\x08R\x08isActive\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror2\x80\x03\n\rRunnerManager\x12^\n\x05Start\x12(.autokitteh.remote.v1.StartRunnerRequest\x1a).autokitteh.remote.v1.StartRunnerResponse\"\x00\x12g\n\x0cRunnerHealth\x12).autokitteh.remote.v1.RunnerHealthRequest\x1a*.autokitteh.remote.v1.RunnerHealthResponse\"\x00\x12O\n\x04Stop\x12!.autokitteh.remote.v1.StopRequest\x1a\".autokitteh.remote.v1.StopResponse\"\x00\x12U\n\x06Health\x12#.autokitteh.remote.v1.HealthRequest\x1a$.autokitteh.remote.v1.HealthResponse\"\x00\x32\xd3\x03\n\x06Runner\x12X\n\x07\x45xports\x12$.autokitteh.remote.v1.ExportsRequest\x1a%.autokitteh.remote.v1.ExportsResponse\"\x00\x12R\n\x05Start\x12\".autokitteh.remote.v1.StartRequest\x1a#.autokitteh.remote.v1.StartResponse\"\x00\x12X\n\x07\x45xecute\x12$.autokitteh.remote.v1.ExecuteRequest\x1a%.autokitteh.remote.v1.ExecuteResponse\"\x00\x12j\n\rActivityReply\x12*.autokitteh.remote.v1.ActivityReplyRequest\x1a+.autokitteh.remote.v1.ActivityReplyResponse\"\x00\x12U\n\x06Health\x12#.autokitteh.remote.v1.HealthRequest\x1a$.autokitteh.remote.v1.HealthResponse\"\x00\x32\x98\x07\n\x06Worker\x12[\n\x08\x41\x63tivity\x12%.autokitteh.remote.v1.ActivityRequest\x1a&.autokitteh.remote.v1.ActivityResponse\"\x00\x12O\n\x04\x44one\x12!.autokitteh.remote.v1.DoneRequest\x1a\".autokitteh.remote.v1.DoneResponse\"\x00\x12L\n\x03Log\x12 .autokitteh.remote.v1.LogRequest\x1a!.autokitteh.remote.v1.LogResponse\"\x00\x12R\n\x05Print\x12\".autokitteh.remote.v1.PrintRequest\x1a#.autokitteh.remote.v1.PrintResponse\"\x00\x12R\n\x05Sleep\x12\".autokitteh.remote.v1.SleepRequest\x1a#.autokitteh.remote.v1.SleepResponse\"\x00\x12^\n\tSubscribe\x12&.autokitteh.remote.v1.SubscribeRequest\x1a\'.autokitteh.remote.v1.SubscribeResponse\"\x00\x12^\n\tNextEvent\x12&.autokitteh.remote.v1.NextEventRequest\x1a\'.autokitteh.remote.v1.NextEventResponse\"\x00\x12\x64\n\x0bUnsubscribe\x12(.autokitteh.remote.v1.UnsubscribeRequest\x1a).autokitteh.remote.v1.UnsubscribeResponse\"\x00\x12U\n\x06Health\x12#.autokitteh.remote.v1.HealthRequest\x1a$.autokitteh.remote.v1.HealthResponse\"\x00\x12m\n\x0eIsActiveRunner\x12+.autokitteh.remote.v1.IsActiveRunnerRequest\x1a,.autokitteh.remote.v1.IsActiveRunnerResponse\"\x00\x42\xe2\x01\n\x18\x63om.autokitteh.remote.v1B\x0bRemoteProtoP\x01ZGgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1;remotev1\xa2\x02\x03\x41RX\xaa\x02\x14\x41utokitteh.Remote.V1\xca\x02\x14\x41utokitteh\\Remote\\V1\xe2\x02 Autokitteh\\Remote\\V1\\GPBMetadata\xea\x02\x16\x41utokitteh::Remote::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'autokitteh.remote.v1.remote_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\030com.autokitteh.remote.v1B\013RemoteProtoP\001ZGgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1;remotev1\242\002\003ARX\252\002\024Autokitteh.Remote.V1\312\002\024Autokitteh\\Remote\\V1\342\002 Autokitteh\\Remote\\V1\\GPBMetadata\352\002\026Autokitteh::Remote::V1'
  _STARTRUNNERREQUEST_VARSENTRY._options = None
  _STARTRUNNERREQUEST_VARSENTRY._serialized_options = b'8\001'
  _CALLINFO_KWARGSENTRY._options = None
  _CALLINFO_KWARGSENTRY._serialized_options = b'8\001'
  _globals['_CONTAINERCONFIG']._serialized_start=59
  _globals['_CONTAINERCONFIG']._serialized_end=98
  _globals['_EVENT']._serialized_start=100
  _globals['_EVENT']._serialized_end=127
  _globals['_HEALTHREQUEST']._serialized_start=129
  _globals['_HEALTHREQUEST']._serialized_end=144
  _globals['_HEALTHRESPONSE']._serialized_start=146
  _globals['_HEALTHRESPONSE']._serialized_end=184
  _globals['_STARTRUNNERREQUEST']._serialized_start=187
  _globals['_STARTRUNNERREQUEST']._serialized_end=496
  _globals['_STARTRUNNERREQUEST_VARSENTRY']._serialized_start=441
  _globals['_STARTRUNNERREQUEST_VARSENTRY']._serialized_end=496
  _globals['_STARTRUNNERRESPONSE']._serialized_start=498
  _globals['_STARTRUNNERRESPONSE']._serialized_end=609
  _globals['_RUNNERHEALTHREQUEST']._serialized_start=611
  _globals['_RUNNERHEALTHREQUEST']._serialized_end=661
  _globals['_RUNNERHEALTHRESPONSE']._serialized_start=663
  _globals['_RUNNERHEALTHRESPONSE']._serialized_end=733
  _globals['_STOPREQUEST']._serialized_start=735
  _globals['_STOPREQUEST']._serialized_end=777
  _globals['_STOPRESPONSE']._serialized_start=779
  _globals['_STOPRESPONSE']._serialized_end=815
  _globals['_STARTREQUEST']._serialized_start=817
  _globals['_STARTREQUEST']._serialized_end=915
  _globals['_STARTRESPONSE']._serialized_start=917
  _globals['_STARTRESPONSE']._serialized_end=1013
  _globals['_FRAME']._serialized_start=1015
  _globals['_FRAME']._serialized_end=1114
  _globals['_EXECUTEREQUEST']._serialized_start=1116
  _globals['_EXECUTEREQUEST']._serialized_end=1152
  _globals['_EXECUTERESPONSE']._serialized_start=1154
  _globals['_EXECUTERESPONSE']._serialized_end=1276
  _globals['_ACTIVITYREPLYREQUEST']._serialized_start=1278
  _globals['_ACTIVITYREPLYREQUEST']._serialized_end=1366
  _globals['_ACTIVITYREPLYRESPONSE']._serialized_start=1368
  _globals['_ACTIVITYREPLYRESPONSE']._serialized_end=1413
  _globals['_EXPORTSREQUEST']._serialized_start=1415
  _globals['_EXPORTSREQUEST']._serialized_end=1460
  _globals['_EXPORTSRESPONSE']._serialized_start=1462
  _globals['_EXPORTSRESPONSE']._serialized_end=1527
  _globals['_CALLINFO']._serialized_start=1530
  _globals['_CALLINFO']._serialized_end=1715
  _globals['_CALLINFO_KWARGSENTRY']._serialized_start=1658
  _globals['_CALLINFO_KWARGSENTRY']._serialized_end=1715
  _globals['_ACTIVITYREQUEST']._serialized_start=1717
  _globals['_ACTIVITYREQUEST']._serialized_end=1844
  _globals['_ACTIVITYRESPONSE']._serialized_start=1846
  _globals['_ACTIVITYRESPONSE']._serialized_end=1886
  _globals['_DONEREQUEST']._serialized_start=1889
  _globals['_DONEREQUEST']._serialized_end=2036
  _globals['_DONERESPONSE']._serialized_start=2038
  _globals['_DONERESPONSE']._serialized_end=2052
  _globals['_SLEEPREQUEST']._serialized_start=2054
  _globals['_SLEEPREQUEST']._serialized_end=2130
  _globals['_SLEEPRESPONSE']._serialized_start=2132
  _globals['_SLEEPRESPONSE']._serialized_end=2169
  _globals['_SUBSCRIBEREQUEST']._serialized_start=2171
  _globals['_SUBSCRIBEREQUEST']._serialized_end=2274
  _globals['_SUBSCRIBERESPONSE']._serialized_start=2276
  _globals['_SUBSCRIBERESPONSE']._serialized_end=2346
  _globals['_NEXTEVENTREQUEST']._serialized_start=2348
  _globals['_NEXTEVENTREQUEST']._serialized_end=2457
  _globals['_NEXTEVENTRESPONSE']._serialized_start=2459
  _globals['_NEXTEVENTRESPONSE']._serialized_end=2551
  _globals['_UNSUBSCRIBEREQUEST']._serialized_start=2553
  _globals['_UNSUBSCRIBEREQUEST']._serialized_end=2631
  _globals['_UNSUBSCRIBERESPONSE']._serialized_start=2633
  _globals['_UNSUBSCRIBERESPONSE']._serialized_end=2676
  _globals['_LOGREQUEST']._serialized_start=2678
  _globals['_LOGREQUEST']._serialized_end=2767
  _globals['_LOGRESPONSE']._serialized_start=2769
  _globals['_LOGRESPONSE']._serialized_end=2804
  _globals['_PRINTREQUEST']._serialized_start=2806
  _globals['_PRINTREQUEST']._serialized_end=2875
  _globals['_PRINTRESPONSE']._serialized_start=2877
  _globals['_PRINTRESPONSE']._serialized_end=2914
  _globals['_ISACTIVERUNNERREQUEST']._serialized_start=2916
  _globals['_ISACTIVERUNNERREQUEST']._serialized_end=2968
  _globals['_ISACTIVERUNNERRESPONSE']._serialized_start=2970
  _globals['_ISACTIVERUNNERRESPONSE']._serialized_end=3045
  _globals['_RUNNERMANAGER']._serialized_start=3048
  _globals['_RUNNERMANAGER']._serialized_end=3432
  _globals['_RUNNER']._serialized_start=3435
  _globals['_RUNNER']._serialized_end=3902
  _globals['_WORKER']._serialized_start=3905
  _globals['_WORKER']._serialized_end=4825
# @@protoc_insertion_point(module_scope)
