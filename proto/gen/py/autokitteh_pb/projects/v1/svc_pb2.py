# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: autokitteh/projects/v1/svc.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from autokitteh_pb.program.v1 import program_pb2 as autokitteh_dot_program_dot_v1_dot_program__pb2
from autokitteh_pb.projects.v1 import project_pb2 as autokitteh_dot_projects_dot_v1_dot_project__pb2
from buf.validate import validate_pb2 as buf_dot_validate_dot_validate__pb2
from google.protobuf import field_mask_pb2 as google_dot_protobuf_dot_field__mask__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n autokitteh/projects/v1/svc.proto\x12\x16\x61utokitteh.projects.v1\x1a#autokitteh/program/v1/program.proto\x1a$autokitteh/projects/v1/project.proto\x1a\x1b\x62uf/validate/validate.proto\x1a google/protobuf/field_mask.proto\"\xd3\x01\n\rCreateRequest\x12\x42\n\x07project\x18\x01 \x01(\x0b\x32\x1f.autokitteh.projects.v1.ProjectB\x07\xfa\xf7\x18\x03\xc8\x01\x01R\x07project:~\xfa\xf7\x18z\x1ax\n project.project_id_must_be_empty\x12 project_id must not be specified\x1a\x32has(this.project) && this.project.project_id == \'\'\"9\n\x0e\x43reateResponse\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"8\n\rDeleteRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"\x10\n\x0e\x44\x65leteResponse\"\xf8\x02\n\nGetRequest\x12\x1d\n\nproject_id\x18\x01 \x01(\tR\tprojectId\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x19\n\x08owner_id\x18\x03 \x01(\tR\x07ownerId:\x9b\x02\xfa\xf7\x18\x96\x02\x1a\x9b\x01\n\x13project_id_xor_name\x12*project_id and name are mutually exclusive\x1aX(this.project_id == \'\' && this.name != \'\') || (this.project_id != \'\' && this.name == \'\')\x1av\n\x17owner_id_with_name_only\x12\x33owner_id can be specified only if name is specified\x1a&this.owner_id == \'\' || this.name != \'\'\"H\n\x0bGetResponse\x12\x39\n\x07project\x18\x01 \x01(\x0b\x32\x1f.autokitteh.projects.v1.ProjectR\x07project\"\x8e\x02\n\rUpdateRequest\x12\x42\n\x07project\x18\x01 \x01(\x0b\x32\x1f.autokitteh.projects.v1.ProjectB\x07\xfa\xf7\x18\x03\xc8\x01\x01R\x07project\x12\x42\n\nfield_mask\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.FieldMaskB\x07\xfa\xf7\x18\x03\xc8\x01\x01R\tfieldMask:u\xfa\xf7\x18q\x1ao\n\x1bproject.project_id_required\x12\x1cproject_id must be specified\x1a\x32has(this.project) && this.project.project_id != \'\'\"\x10\n\x0eUpdateResponse\"\r\n\x0bListRequest\"Y\n\x0cListResponse\x12I\n\x08projects\x18\x01 \x03(\x0b\x32\x1f.autokitteh.projects.v1.ProjectB\x0c\xfa\xf7\x18\x08\x92\x01\x05\"\x03\xc8\x01\x01R\x08projects\":\n\x13ListForOwnerRequest\x12#\n\x08owner_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x18\x00R\x07ownerId\"a\n\x14ListForOwnerResponse\x12I\n\x08projects\x18\x01 \x03(\x0b\x32\x1f.autokitteh.projects.v1.ProjectB\x0c\xfa\xf7\x18\x08\x92\x01\x05\"\x03\xc8\x01\x01R\x08projects\"7\n\x0c\x42uildRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"^\n\rBuildResponse\x12\x19\n\x08\x62uild_id\x18\x01 \x01(\tR\x07\x62uildId\x12\x32\n\x05\x65rror\x18\x02 \x01(\x0b\x32\x1c.autokitteh.program.v1.ErrorR\x05\x65rror\"\xd6\x01\n\x13SetResourcesRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\x12X\n\tresources\x18\x02 \x03(\x0b\x32:.autokitteh.projects.v1.SetResourcesRequest.ResourcesEntryR\tresources\x1a<\n\x0eResourcesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"\x16\n\x14SetResourcesResponse\"C\n\x18\x44ownloadResourcesRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"\xb9\x01\n\x19\x44ownloadResourcesResponse\x12^\n\tresources\x18\x02 \x03(\x0b\x32@.autokitteh.projects.v1.DownloadResourcesResponse.ResourcesEntryR\tresources\x1a<\n\x0eResourcesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"8\n\rExportRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"4\n\x0e\x45xportResponse\x12\"\n\x07project\x18\x01 \x01(\x0c\x42\x08\xfa\xf7\x18\x04z\x02\x10\nR\x07project2\xbe\x07\n\x0fProjectsService\x12W\n\x06\x43reate\x12%.autokitteh.projects.v1.CreateRequest\x1a&.autokitteh.projects.v1.CreateResponse\x12W\n\x06\x44\x65lete\x12%.autokitteh.projects.v1.DeleteRequest\x1a&.autokitteh.projects.v1.DeleteResponse\x12N\n\x03Get\x12\".autokitteh.projects.v1.GetRequest\x1a#.autokitteh.projects.v1.GetResponse\x12W\n\x06Update\x12%.autokitteh.projects.v1.UpdateRequest\x1a&.autokitteh.projects.v1.UpdateResponse\x12Q\n\x04List\x12#.autokitteh.projects.v1.ListRequest\x1a$.autokitteh.projects.v1.ListResponse\x12T\n\x05\x42uild\x12$.autokitteh.projects.v1.BuildRequest\x1a%.autokitteh.projects.v1.BuildResponse\x12i\n\x0cListForOwner\x12+.autokitteh.projects.v1.ListForOwnerRequest\x1a,.autokitteh.projects.v1.ListForOwnerResponse\x12i\n\x0cSetResources\x12+.autokitteh.projects.v1.SetResourcesRequest\x1a,.autokitteh.projects.v1.SetResourcesResponse\x12x\n\x11\x44ownloadResources\x12\x30.autokitteh.projects.v1.DownloadResourcesRequest\x1a\x31.autokitteh.projects.v1.DownloadResourcesResponse\x12W\n\x06\x45xport\x12%.autokitteh.projects.v1.ExportRequest\x1a&.autokitteh.projects.v1.ExportResponseB\xed\x01\n\x1a\x63om.autokitteh.projects.v1B\x08SvcProtoP\x01ZKgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1;projectsv1\xa2\x02\x03\x41PX\xaa\x02\x16\x41utokitteh.Projects.V1\xca\x02\x16\x41utokitteh\\Projects\\V1\xe2\x02\"Autokitteh\\Projects\\V1\\GPBMetadata\xea\x02\x18\x41utokitteh::Projects::V1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'autokitteh.projects.v1.svc_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\032com.autokitteh.projects.v1B\010SvcProtoP\001ZKgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1;projectsv1\242\002\003APX\252\002\026Autokitteh.Projects.V1\312\002\026Autokitteh\\Projects\\V1\342\002\"Autokitteh\\Projects\\V1\\GPBMetadata\352\002\030Autokitteh::Projects::V1'
  _CREATEREQUEST.fields_by_name['project']._options = None
  _CREATEREQUEST.fields_by_name['project']._serialized_options = b'\372\367\030\003\310\001\001'
  _CREATEREQUEST._options = None
  _CREATEREQUEST._serialized_options = b'\372\367\030z\032x\n project.project_id_must_be_empty\022 project_id must not be specified\0322has(this.project) && this.project.project_id == \'\''
  _CREATERESPONSE.fields_by_name['project_id']._options = None
  _CREATERESPONSE.fields_by_name['project_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _DELETEREQUEST.fields_by_name['project_id']._options = None
  _DELETEREQUEST.fields_by_name['project_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _GETREQUEST._options = None
  _GETREQUEST._serialized_options = b'\372\367\030\226\002\032\233\001\n\023project_id_xor_name\022*project_id and name are mutually exclusive\032X(this.project_id == \'\' && this.name != \'\') || (this.project_id != \'\' && this.name == \'\')\032v\n\027owner_id_with_name_only\0223owner_id can be specified only if name is specified\032&this.owner_id == \'\' || this.name != \'\''
  _UPDATEREQUEST.fields_by_name['project']._options = None
  _UPDATEREQUEST.fields_by_name['project']._serialized_options = b'\372\367\030\003\310\001\001'
  _UPDATEREQUEST.fields_by_name['field_mask']._options = None
  _UPDATEREQUEST.fields_by_name['field_mask']._serialized_options = b'\372\367\030\003\310\001\001'
  _UPDATEREQUEST._options = None
  _UPDATEREQUEST._serialized_options = b'\372\367\030q\032o\n\033project.project_id_required\022\034project_id must be specified\0322has(this.project) && this.project.project_id != \'\''
  _LISTRESPONSE.fields_by_name['projects']._options = None
  _LISTRESPONSE.fields_by_name['projects']._serialized_options = b'\372\367\030\010\222\001\005\"\003\310\001\001'
  _LISTFOROWNERREQUEST.fields_by_name['owner_id']._options = None
  _LISTFOROWNERREQUEST.fields_by_name['owner_id']._serialized_options = b'\372\367\030\004r\002\030\000'
  _LISTFOROWNERRESPONSE.fields_by_name['projects']._options = None
  _LISTFOROWNERRESPONSE.fields_by_name['projects']._serialized_options = b'\372\367\030\010\222\001\005\"\003\310\001\001'
  _BUILDREQUEST.fields_by_name['project_id']._options = None
  _BUILDREQUEST.fields_by_name['project_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _SETRESOURCESREQUEST_RESOURCESENTRY._options = None
  _SETRESOURCESREQUEST_RESOURCESENTRY._serialized_options = b'8\001'
  _SETRESOURCESREQUEST.fields_by_name['project_id']._options = None
  _SETRESOURCESREQUEST.fields_by_name['project_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _DOWNLOADRESOURCESREQUEST.fields_by_name['project_id']._options = None
  _DOWNLOADRESOURCESREQUEST.fields_by_name['project_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _DOWNLOADRESOURCESRESPONSE_RESOURCESENTRY._options = None
  _DOWNLOADRESOURCESRESPONSE_RESOURCESENTRY._serialized_options = b'8\001'
  _EXPORTREQUEST.fields_by_name['project_id']._options = None
  _EXPORTREQUEST.fields_by_name['project_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _EXPORTRESPONSE.fields_by_name['project']._options = None
  _EXPORTRESPONSE.fields_by_name['project']._serialized_options = b'\372\367\030\004z\002\020\n'
  _globals['_CREATEREQUEST']._serialized_start=199
  _globals['_CREATEREQUEST']._serialized_end=410
  _globals['_CREATERESPONSE']._serialized_start=412
  _globals['_CREATERESPONSE']._serialized_end=469
  _globals['_DELETEREQUEST']._serialized_start=471
  _globals['_DELETEREQUEST']._serialized_end=527
  _globals['_DELETERESPONSE']._serialized_start=529
  _globals['_DELETERESPONSE']._serialized_end=545
  _globals['_GETREQUEST']._serialized_start=548
  _globals['_GETREQUEST']._serialized_end=924
  _globals['_GETRESPONSE']._serialized_start=926
  _globals['_GETRESPONSE']._serialized_end=998
  _globals['_UPDATEREQUEST']._serialized_start=1001
  _globals['_UPDATEREQUEST']._serialized_end=1271
  _globals['_UPDATERESPONSE']._serialized_start=1273
  _globals['_UPDATERESPONSE']._serialized_end=1289
  _globals['_LISTREQUEST']._serialized_start=1291
  _globals['_LISTREQUEST']._serialized_end=1304
  _globals['_LISTRESPONSE']._serialized_start=1306
  _globals['_LISTRESPONSE']._serialized_end=1395
  _globals['_LISTFOROWNERREQUEST']._serialized_start=1397
  _globals['_LISTFOROWNERREQUEST']._serialized_end=1455
  _globals['_LISTFOROWNERRESPONSE']._serialized_start=1457
  _globals['_LISTFOROWNERRESPONSE']._serialized_end=1554
  _globals['_BUILDREQUEST']._serialized_start=1556
  _globals['_BUILDREQUEST']._serialized_end=1611
  _globals['_BUILDRESPONSE']._serialized_start=1613
  _globals['_BUILDRESPONSE']._serialized_end=1707
  _globals['_SETRESOURCESREQUEST']._serialized_start=1710
  _globals['_SETRESOURCESREQUEST']._serialized_end=1924
  _globals['_SETRESOURCESREQUEST_RESOURCESENTRY']._serialized_start=1864
  _globals['_SETRESOURCESREQUEST_RESOURCESENTRY']._serialized_end=1924
  _globals['_SETRESOURCESRESPONSE']._serialized_start=1926
  _globals['_SETRESOURCESRESPONSE']._serialized_end=1948
  _globals['_DOWNLOADRESOURCESREQUEST']._serialized_start=1950
  _globals['_DOWNLOADRESOURCESREQUEST']._serialized_end=2017
  _globals['_DOWNLOADRESOURCESRESPONSE']._serialized_start=2020
  _globals['_DOWNLOADRESOURCESRESPONSE']._serialized_end=2205
  _globals['_DOWNLOADRESOURCESRESPONSE_RESOURCESENTRY']._serialized_start=1864
  _globals['_DOWNLOADRESOURCESRESPONSE_RESOURCESENTRY']._serialized_end=1924
  _globals['_EXPORTREQUEST']._serialized_start=2207
  _globals['_EXPORTREQUEST']._serialized_end=2263
  _globals['_EXPORTRESPONSE']._serialized_start=2265
  _globals['_EXPORTRESPONSE']._serialized_end=2317
  _globals['_PROJECTSSERVICE']._serialized_start=2320
  _globals['_PROJECTSSERVICE']._serialized_end=3278
# @@protoc_insertion_point(module_scope)
