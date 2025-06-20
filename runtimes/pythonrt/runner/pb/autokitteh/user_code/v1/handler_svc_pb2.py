# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: autokitteh/user_code/v1/handler_svc.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from pb.autokitteh.user_code.v1 import user_code_pb2 as autokitteh_dot_user__code_dot_v1_dot_user__code__pb2
from pb.autokitteh.values.v1 import values_pb2 as autokitteh_dot_values_dot_v1_dot_values__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n)autokitteh/user_code/v1/handler_svc.proto\x12\x17\x61utokitteh.user_code.v1\x1a\'autokitteh/user_code/v1/user_code.proto\x1a!autokitteh/values/v1/values.proto\x1a\x1fgoogle/protobuf/timestamp.proto\"\xf6\x01\n\x08\x43\x61llInfo\x12\x1a\n\x08\x66unction\x18\x01 \x01(\tR\x08\x66unction\x12/\n\x04\x61rgs\x18\x02 \x03(\x0b\x32\x1b.autokitteh.values.v1.ValueR\x04\x61rgs\x12\x45\n\x06kwargs\x18\x03 \x03(\x0b\x32-.autokitteh.user_code.v1.CallInfo.KwargsEntryR\x06kwargs\x1aV\n\x0bKwargsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x31\n\x05value\x18\x02 \x01(\x0b\x32\x1b.autokitteh.values.v1.ValueR\x05value:\x02\x38\x01\"\x82\x01\n\x0f\x41\x63tivityRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x12\n\x04\x64\x61ta\x18\x02 \x01(\x0cR\x04\x64\x61ta\x12>\n\tcall_info\x18\x03 \x01(\x0b\x32!.autokitteh.user_code.v1.CallInfoR\x08\x63\x61llInfo\"(\n\x10\x41\x63tivityResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"\xb3\x01\n\x0b\x44oneRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x33\n\x06result\x18\x02 \x01(\x0b\x32\x1b.autokitteh.values.v1.ValueR\x06result\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\x12<\n\ttraceback\x18\x04 \x03(\x0b\x32\x1e.autokitteh.user_code.v1.FrameR\ttraceback\"\x0e\n\x0c\x44oneResponse\"L\n\x0cSleepRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x1f\n\x0b\x64uration_ms\x18\x02 \x01(\x03R\ndurationMs\"%\n\rSleepResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"g\n\x10SubscribeRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x1e\n\nconnection\x18\x02 \x01(\tR\nconnection\x12\x16\n\x06\x66ilter\x18\x03 \x01(\tR\x06\x66ilter\"F\n\x11SubscribeResponse\x12\x1b\n\tsignal_id\x18\x01 \x01(\tR\x08signalId\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"m\n\x10NextEventRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x1d\n\nsignal_ids\x18\x02 \x03(\tR\tsignalIds\x12\x1d\n\ntimeout_ms\x18\x03 \x01(\x03R\ttimeoutMs\"_\n\x11NextEventResponse\x12\x34\n\x05\x65vent\x18\x01 \x01(\x0b\x32\x1e.autokitteh.user_code.v1.EventR\x05\x65vent\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\"N\n\x12UnsubscribeRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x1b\n\tsignal_id\x18\x02 \x01(\tR\x08signalId\"+\n\x13UnsubscribeResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"S\n\x06Signal\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x35\n\x07payload\x18\x03 \x01(\x0b\x32\x1b.autokitteh.values.v1.ValueR\x07payload\"\x84\x01\n\rSignalRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x37\n\x06signal\x18\x02 \x01(\x0b\x32\x1f.autokitteh.user_code.v1.SignalR\x06signal\x12\x1d\n\nsession_id\x18\x03 \x01(\tR\tsessionId\"&\n\x0eSignalResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"e\n\x11NextSignalRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x14\n\x05names\x18\x02 \x03(\tR\x05names\x12\x1d\n\ntimeout_ms\x18\x03 \x01(\x03R\ttimeoutMs\"c\n\x12NextSignalResponse\x12\x37\n\x06signal\x18\x01 \x01(\x0b\x32\x1f.autokitteh.user_code.v1.SignalR\x06signal\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"Y\n\nLogRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x14\n\x05level\x18\x02 \x01(\tR\x05level\x12\x18\n\x07message\x18\x03 \x01(\tR\x07message\"#\n\x0bLogResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"E\n\x0cPrintRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x18\n\x07message\x18\x02 \x01(\tR\x07message\"%\n\rPrintResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"/\n\x10StoreListRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"=\n\x11StoreListResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\x12\x12\n\x04keys\x18\x02 \x03(\tR\x04keys\"\x9a\x01\n\x12StoreMutateRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x10\n\x03key\x18\x02 \x01(\tR\x03key\x12\x1c\n\toperation\x18\x03 \x01(\tR\toperation\x12\x37\n\x08operands\x18\x04 \x03(\x0b\x32\x1b.autokitteh.values.v1.ValueR\x08operands\"`\n\x13StoreMutateResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\x12\x33\n\x06result\x18\x02 \x01(\x0b\x32\x1b.autokitteh.values.v1.ValueR\x06result\"\x86\x01\n\x13StartSessionRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x10\n\x03loc\x18\x02 \x01(\tR\x03loc\x12\x12\n\x04\x64\x61ta\x18\x03 \x01(\x0cR\x04\x64\x61ta\x12\x12\n\x04memo\x18\x04 \x01(\x0cR\x04memo\x12\x18\n\x07project\x18\x05 \x01(\tR\x07project\"K\n\x14StartSessionResponse\x12\x1d\n\nsession_id\x18\x01 \x01(\tR\tsessionId\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"\xfb\x01\n\x10\x45ncodeJWTRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12P\n\x07payload\x18\x02 \x03(\x0b\x32\x36.autokitteh.user_code.v1.EncodeJWTRequest.PayloadEntryR\x07payload\x12\x1e\n\nconnection\x18\x03 \x01(\tR\nconnection\x12\x1c\n\talgorithm\x18\x04 \x01(\tR\talgorithm\x1a:\n\x0cPayloadEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x03R\x05value:\x02\x38\x01\";\n\x11\x45ncodeJWTResponse\x12\x10\n\x03jwt\x18\x01 \x01(\tR\x03jwt\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"o\n\x0eRefreshRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12 \n\x0bintegration\x18\x02 \x01(\tR\x0bintegration\x12\x1e\n\nconnection\x18\x03 \x01(\tR\nconnection\"s\n\x0fRefreshResponse\x12\x14\n\x05token\x18\x01 \x01(\tR\x05token\x12\x34\n\x07\x65xpires\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\x07\x65xpires\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\"4\n\x15IsActiveRunnerRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\"K\n\x16IsActiveRunnerResponse\x12\x1b\n\tis_active\x18\x01 \x01(\x08R\x08isActive\x12\x14\n\x05\x65rror\x18\x02 \x01(\tR\x05\x65rror\"\x16\n\x14HandlerHealthRequest\"-\n\x15HandlerHealthResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror\"\xbb\x01\n\x13\x45xecuteReplyRequest\x12\x1b\n\trunner_id\x18\x01 \x01(\tR\x08runnerId\x12\x33\n\x06result\x18\x02 \x01(\x0b\x32\x1b.autokitteh.values.v1.ValueR\x06result\x12\x14\n\x05\x65rror\x18\x03 \x01(\tR\x05\x65rror\x12<\n\ttraceback\x18\x04 \x03(\x0b\x32\x1e.autokitteh.user_code.v1.FrameR\ttraceback\",\n\x14\x45xecuteReplyResponse\x12\x14\n\x05\x65rror\x18\x01 \x01(\tR\x05\x65rror2\xb0\x0e\n\x0eHandlerService\x12\x61\n\x08\x41\x63tivity\x12(.autokitteh.user_code.v1.ActivityRequest\x1a).autokitteh.user_code.v1.ActivityResponse\"\x00\x12m\n\x0c\x45xecuteReply\x12,.autokitteh.user_code.v1.ExecuteReplyRequest\x1a-.autokitteh.user_code.v1.ExecuteReplyResponse\"\x00\x12U\n\x04\x44one\x12$.autokitteh.user_code.v1.DoneRequest\x1a%.autokitteh.user_code.v1.DoneResponse\"\x00\x12R\n\x03Log\x12#.autokitteh.user_code.v1.LogRequest\x1a$.autokitteh.user_code.v1.LogResponse\"\x00\x12X\n\x05Print\x12%.autokitteh.user_code.v1.PrintRequest\x1a&.autokitteh.user_code.v1.PrintResponse\"\x00\x12X\n\x05Sleep\x12%.autokitteh.user_code.v1.SleepRequest\x1a&.autokitteh.user_code.v1.SleepResponse\"\x00\x12\x64\n\tSubscribe\x12).autokitteh.user_code.v1.SubscribeRequest\x1a*.autokitteh.user_code.v1.SubscribeResponse\"\x00\x12\x64\n\tNextEvent\x12).autokitteh.user_code.v1.NextEventRequest\x1a*.autokitteh.user_code.v1.NextEventResponse\"\x00\x12j\n\x0bUnsubscribe\x12+.autokitteh.user_code.v1.UnsubscribeRequest\x1a,.autokitteh.user_code.v1.UnsubscribeResponse\"\x00\x12m\n\x0cStartSession\x12,.autokitteh.user_code.v1.StartSessionRequest\x1a-.autokitteh.user_code.v1.StartSessionResponse\"\x00\x12[\n\x06Signal\x12&.autokitteh.user_code.v1.SignalRequest\x1a\'.autokitteh.user_code.v1.SignalResponse\"\x00\x12g\n\nNextSignal\x12*.autokitteh.user_code.v1.NextSignalRequest\x1a+.autokitteh.user_code.v1.NextSignalResponse\"\x00\x12\x64\n\tStoreList\x12).autokitteh.user_code.v1.StoreListRequest\x1a*.autokitteh.user_code.v1.StoreListResponse\"\x00\x12j\n\x0bStoreMutate\x12+.autokitteh.user_code.v1.StoreMutateRequest\x1a,.autokitteh.user_code.v1.StoreMutateResponse\"\x00\x12\x64\n\tEncodeJWT\x12).autokitteh.user_code.v1.EncodeJWTRequest\x1a*.autokitteh.user_code.v1.EncodeJWTResponse\"\x00\x12h\n\x11RefreshOAuthToken\x12\'.autokitteh.user_code.v1.RefreshRequest\x1a(.autokitteh.user_code.v1.RefreshResponse\"\x00\x12i\n\x06Health\x12-.autokitteh.user_code.v1.HandlerHealthRequest\x1a..autokitteh.user_code.v1.HandlerHealthResponse\"\x00\x12s\n\x0eIsActiveRunner\x12..autokitteh.user_code.v1.IsActiveRunnerRequest\x1a/.autokitteh.user_code.v1.IsActiveRunnerResponse\"\x00\x42\xf7\x01\n\x1b\x63om.autokitteh.user_code.v1B\x0fHandlerSvcProtoP\x01ZMgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/user_code/v1;user_codev1\xa2\x02\x03\x41UX\xaa\x02\x16\x41utokitteh.UserCode.V1\xca\x02\x16\x41utokitteh\\UserCode\\V1\xe2\x02\"Autokitteh\\UserCode\\V1\\GPBMetadata\xea\x02\x18\x41utokitteh::UserCode::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'autokitteh.user_code.v1.handler_svc_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\033com.autokitteh.user_code.v1B\017HandlerSvcProtoP\001ZMgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/user_code/v1;user_codev1\242\002\003AUX\252\002\026Autokitteh.UserCode.V1\312\002\026Autokitteh\\UserCode\\V1\342\002\"Autokitteh\\UserCode\\V1\\GPBMetadata\352\002\030Autokitteh::UserCode::V1'
  _CALLINFO_KWARGSENTRY._options = None
  _CALLINFO_KWARGSENTRY._serialized_options = b'8\001'
  _ENCODEJWTREQUEST_PAYLOADENTRY._options = None
  _ENCODEJWTREQUEST_PAYLOADENTRY._serialized_options = b'8\001'
  _globals['_CALLINFO']._serialized_start=180
  _globals['_CALLINFO']._serialized_end=426
  _globals['_CALLINFO_KWARGSENTRY']._serialized_start=340
  _globals['_CALLINFO_KWARGSENTRY']._serialized_end=426
  _globals['_ACTIVITYREQUEST']._serialized_start=429
  _globals['_ACTIVITYREQUEST']._serialized_end=559
  _globals['_ACTIVITYRESPONSE']._serialized_start=561
  _globals['_ACTIVITYRESPONSE']._serialized_end=601
  _globals['_DONEREQUEST']._serialized_start=604
  _globals['_DONEREQUEST']._serialized_end=783
  _globals['_DONERESPONSE']._serialized_start=785
  _globals['_DONERESPONSE']._serialized_end=799
  _globals['_SLEEPREQUEST']._serialized_start=801
  _globals['_SLEEPREQUEST']._serialized_end=877
  _globals['_SLEEPRESPONSE']._serialized_start=879
  _globals['_SLEEPRESPONSE']._serialized_end=916
  _globals['_SUBSCRIBEREQUEST']._serialized_start=918
  _globals['_SUBSCRIBEREQUEST']._serialized_end=1021
  _globals['_SUBSCRIBERESPONSE']._serialized_start=1023
  _globals['_SUBSCRIBERESPONSE']._serialized_end=1093
  _globals['_NEXTEVENTREQUEST']._serialized_start=1095
  _globals['_NEXTEVENTREQUEST']._serialized_end=1204
  _globals['_NEXTEVENTRESPONSE']._serialized_start=1206
  _globals['_NEXTEVENTRESPONSE']._serialized_end=1301
  _globals['_UNSUBSCRIBEREQUEST']._serialized_start=1303
  _globals['_UNSUBSCRIBEREQUEST']._serialized_end=1381
  _globals['_UNSUBSCRIBERESPONSE']._serialized_start=1383
  _globals['_UNSUBSCRIBERESPONSE']._serialized_end=1426
  _globals['_SIGNAL']._serialized_start=1428
  _globals['_SIGNAL']._serialized_end=1511
  _globals['_SIGNALREQUEST']._serialized_start=1514
  _globals['_SIGNALREQUEST']._serialized_end=1646
  _globals['_SIGNALRESPONSE']._serialized_start=1648
  _globals['_SIGNALRESPONSE']._serialized_end=1686
  _globals['_NEXTSIGNALREQUEST']._serialized_start=1688
  _globals['_NEXTSIGNALREQUEST']._serialized_end=1789
  _globals['_NEXTSIGNALRESPONSE']._serialized_start=1791
  _globals['_NEXTSIGNALRESPONSE']._serialized_end=1890
  _globals['_LOGREQUEST']._serialized_start=1892
  _globals['_LOGREQUEST']._serialized_end=1981
  _globals['_LOGRESPONSE']._serialized_start=1983
  _globals['_LOGRESPONSE']._serialized_end=2018
  _globals['_PRINTREQUEST']._serialized_start=2020
  _globals['_PRINTREQUEST']._serialized_end=2089
  _globals['_PRINTRESPONSE']._serialized_start=2091
  _globals['_PRINTRESPONSE']._serialized_end=2128
  _globals['_STORELISTREQUEST']._serialized_start=2130
  _globals['_STORELISTREQUEST']._serialized_end=2177
  _globals['_STORELISTRESPONSE']._serialized_start=2179
  _globals['_STORELISTRESPONSE']._serialized_end=2240
  _globals['_STOREMUTATEREQUEST']._serialized_start=2243
  _globals['_STOREMUTATEREQUEST']._serialized_end=2397
  _globals['_STOREMUTATERESPONSE']._serialized_start=2399
  _globals['_STOREMUTATERESPONSE']._serialized_end=2495
  _globals['_STARTSESSIONREQUEST']._serialized_start=2498
  _globals['_STARTSESSIONREQUEST']._serialized_end=2632
  _globals['_STARTSESSIONRESPONSE']._serialized_start=2634
  _globals['_STARTSESSIONRESPONSE']._serialized_end=2709
  _globals['_ENCODEJWTREQUEST']._serialized_start=2712
  _globals['_ENCODEJWTREQUEST']._serialized_end=2963
  _globals['_ENCODEJWTREQUEST_PAYLOADENTRY']._serialized_start=2905
  _globals['_ENCODEJWTREQUEST_PAYLOADENTRY']._serialized_end=2963
  _globals['_ENCODEJWTRESPONSE']._serialized_start=2965
  _globals['_ENCODEJWTRESPONSE']._serialized_end=3024
  _globals['_REFRESHREQUEST']._serialized_start=3026
  _globals['_REFRESHREQUEST']._serialized_end=3137
  _globals['_REFRESHRESPONSE']._serialized_start=3139
  _globals['_REFRESHRESPONSE']._serialized_end=3254
  _globals['_ISACTIVERUNNERREQUEST']._serialized_start=3256
  _globals['_ISACTIVERUNNERREQUEST']._serialized_end=3308
  _globals['_ISACTIVERUNNERRESPONSE']._serialized_start=3310
  _globals['_ISACTIVERUNNERRESPONSE']._serialized_end=3385
  _globals['_HANDLERHEALTHREQUEST']._serialized_start=3387
  _globals['_HANDLERHEALTHREQUEST']._serialized_end=3409
  _globals['_HANDLERHEALTHRESPONSE']._serialized_start=3411
  _globals['_HANDLERHEALTHRESPONSE']._serialized_end=3456
  _globals['_EXECUTEREPLYREQUEST']._serialized_start=3459
  _globals['_EXECUTEREPLYREQUEST']._serialized_end=3646
  _globals['_EXECUTEREPLYRESPONSE']._serialized_start=3648
  _globals['_EXECUTEREPLYRESPONSE']._serialized_end=3692
  _globals['_HANDLERSERVICE']._serialized_start=3695
  _globals['_HANDLERSERVICE']._serialized_end=5535
# @@protoc_insertion_point(module_scope)
