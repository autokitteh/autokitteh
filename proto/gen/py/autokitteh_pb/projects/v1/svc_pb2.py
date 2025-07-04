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


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n autokitteh/projects/v1/svc.proto\x12\x16\x61utokitteh.projects.v1\x1a#autokitteh/program/v1/program.proto\x1a$autokitteh/projects/v1/project.proto\x1a\x1b\x62uf/validate/validate.proto\"\xd3\x01\n\rCreateRequest\x12\x42\n\x07project\x18\x01 \x01(\x0b\x32\x1f.autokitteh.projects.v1.ProjectB\x07\xfa\xf7\x18\x03\xc8\x01\x01R\x07project:~\xfa\xf7\x18z\x1ax\n project.project_id_must_be_empty\x12 project_id must not be specified\x1a\x32has(this.project) && this.project.project_id == \'\'\"9\n\x0e\x43reateResponse\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"8\n\rDeleteRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"\x10\n\x0e\x44\x65leteResponse\"\xee\x02\n\nGetRequest\x12\x1d\n\nproject_id\x18\x01 \x01(\tR\tprojectId\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x15\n\x06org_id\x18\x03 \x01(\tR\x05orgId:\x95\x02\xfa\xf7\x18\x90\x02\x1a\x9b\x01\n\x13project_id_xor_name\x12*project_id and name are mutually exclusive\x1aX(this.project_id == \'\' && this.name != \'\') || (this.project_id != \'\' && this.name == \'\')\x1ap\n\x15org_id_with_name_only\x12\x31org_id can be specified only if name is specified\x1a$this.org_id == \'\' || this.name != \'\'\"H\n\x0bGetResponse\x12\x39\n\x07project\x18\x01 \x01(\x0b\x32\x1f.autokitteh.projects.v1.ProjectR\x07project\"\xca\x01\n\rUpdateRequest\x12\x42\n\x07project\x18\x01 \x01(\x0b\x32\x1f.autokitteh.projects.v1.ProjectB\x07\xfa\xf7\x18\x03\xc8\x01\x01R\x07project:u\xfa\xf7\x18q\x1ao\n\x1bproject.project_id_required\x12\x1cproject_id must be specified\x1a\x32has(this.project) && this.project.project_id != \'\'\"\x10\n\x0eUpdateResponse\"$\n\x0bListRequest\x12\x15\n\x06org_id\x18\x01 \x01(\tR\x05orgId\"Y\n\x0cListResponse\x12I\n\x08projects\x18\x01 \x03(\x0b\x32\x1f.autokitteh.projects.v1.ProjectB\x0c\xfa\xf7\x18\x08\x92\x01\x05\"\x03\xc8\x01\x01R\x08projects\"7\n\x0c\x42uildRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"^\n\rBuildResponse\x12\x19\n\x08\x62uild_id\x18\x01 \x01(\tR\x07\x62uildId\x12\x32\n\x05\x65rror\x18\x02 \x01(\x0b\x32\x1c.autokitteh.program.v1.ErrorR\x05\x65rror\"\xd6\x01\n\x13SetResourcesRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\x12X\n\tresources\x18\x02 \x03(\x0b\x32:.autokitteh.projects.v1.SetResourcesRequest.ResourcesEntryR\tresources\x1a<\n\x0eResourcesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"\x16\n\x14SetResourcesResponse\"C\n\x18\x44ownloadResourcesRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\"\xb9\x01\n\x19\x44ownloadResourcesResponse\x12^\n\tresources\x18\x02 \x03(\x0b\x32@.autokitteh.projects.v1.DownloadResourcesResponse.ResourcesEntryR\tresources\x1a<\n\x0eResourcesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"l\n\rExportRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\x12\x32\n\x15include_vars_contents\x18\x02 \x01(\x08R\x13includeVarsContents\";\n\x0e\x45xportResponse\x12)\n\x0bzip_archive\x18\x01 \x01(\x0c\x42\x08\xfa\xf7\x18\x04z\x02\x10\nR\nzipArchive\"\xeb\x01\n\x0bLintRequest\x12\'\n\nproject_id\x18\x01 \x01(\tB\x08\xfa\xf7\x18\x04r\x02\x10\x01R\tprojectId\x12P\n\tresources\x18\x02 \x03(\x0b\x32\x32.autokitteh.projects.v1.LintRequest.ResourcesEntryR\tresources\x12#\n\rmanifest_file\x18\x03 \x01(\tR\x0cmanifestFile\x1a<\n\x0eResourcesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"\x8c\x02\n\x0e\x43heckViolation\x12?\n\x08location\x18\x01 \x01(\x0b\x32#.autokitteh.program.v1.CodeLocationR\x08location\x12\x42\n\x05level\x18\x02 \x01(\x0e\x32,.autokitteh.projects.v1.CheckViolation.LevelR\x05level\x12\x18\n\x07message\x18\x03 \x01(\tR\x07message\x12\x17\n\x07rule_id\x18\x04 \x01(\tR\x06ruleId\"B\n\x05Level\x12\x15\n\x11LEVEL_UNSPECIFIED\x10\x00\x12\x11\n\rLEVEL_WARNING\x10\x01\x12\x0f\n\x0bLEVEL_ERROR\x10\x02\"V\n\x0cLintResponse\x12\x46\n\nviolations\x18\x01 \x03(\x0b\x32&.autokitteh.projects.v1.CheckViolationR\nviolations2\xa6\x07\n\x0fProjectsService\x12W\n\x06\x43reate\x12%.autokitteh.projects.v1.CreateRequest\x1a&.autokitteh.projects.v1.CreateResponse\x12W\n\x06\x44\x65lete\x12%.autokitteh.projects.v1.DeleteRequest\x1a&.autokitteh.projects.v1.DeleteResponse\x12N\n\x03Get\x12\".autokitteh.projects.v1.GetRequest\x1a#.autokitteh.projects.v1.GetResponse\x12W\n\x06Update\x12%.autokitteh.projects.v1.UpdateRequest\x1a&.autokitteh.projects.v1.UpdateResponse\x12Q\n\x04List\x12#.autokitteh.projects.v1.ListRequest\x1a$.autokitteh.projects.v1.ListResponse\x12T\n\x05\x42uild\x12$.autokitteh.projects.v1.BuildRequest\x1a%.autokitteh.projects.v1.BuildResponse\x12i\n\x0cSetResources\x12+.autokitteh.projects.v1.SetResourcesRequest\x1a,.autokitteh.projects.v1.SetResourcesResponse\x12x\n\x11\x44ownloadResources\x12\x30.autokitteh.projects.v1.DownloadResourcesRequest\x1a\x31.autokitteh.projects.v1.DownloadResourcesResponse\x12W\n\x06\x45xport\x12%.autokitteh.projects.v1.ExportRequest\x1a&.autokitteh.projects.v1.ExportResponse\x12Q\n\x04Lint\x12#.autokitteh.projects.v1.LintRequest\x1a$.autokitteh.projects.v1.LintResponseB\xed\x01\n\x1a\x63om.autokitteh.projects.v1B\x08SvcProtoP\x01ZKgo.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1;projectsv1\xa2\x02\x03\x41PX\xaa\x02\x16\x41utokitteh.Projects.V1\xca\x02\x16\x41utokitteh\\Projects\\V1\xe2\x02\"Autokitteh\\Projects\\V1\\GPBMetadata\xea\x02\x18\x41utokitteh::Projects::V1b\x06proto3')

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
  _GETREQUEST._serialized_options = b'\372\367\030\220\002\032\233\001\n\023project_id_xor_name\022*project_id and name are mutually exclusive\032X(this.project_id == \'\' && this.name != \'\') || (this.project_id != \'\' && this.name == \'\')\032p\n\025org_id_with_name_only\0221org_id can be specified only if name is specified\032$this.org_id == \'\' || this.name != \'\''
  _UPDATEREQUEST.fields_by_name['project']._options = None
  _UPDATEREQUEST.fields_by_name['project']._serialized_options = b'\372\367\030\003\310\001\001'
  _UPDATEREQUEST._options = None
  _UPDATEREQUEST._serialized_options = b'\372\367\030q\032o\n\033project.project_id_required\022\034project_id must be specified\0322has(this.project) && this.project.project_id != \'\''
  _LISTRESPONSE.fields_by_name['projects']._options = None
  _LISTRESPONSE.fields_by_name['projects']._serialized_options = b'\372\367\030\010\222\001\005\"\003\310\001\001'
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
  _EXPORTRESPONSE.fields_by_name['zip_archive']._options = None
  _EXPORTRESPONSE.fields_by_name['zip_archive']._serialized_options = b'\372\367\030\004z\002\020\n'
  _LINTREQUEST_RESOURCESENTRY._options = None
  _LINTREQUEST_RESOURCESENTRY._serialized_options = b'8\001'
  _LINTREQUEST.fields_by_name['project_id']._options = None
  _LINTREQUEST.fields_by_name['project_id']._serialized_options = b'\372\367\030\004r\002\020\001'
  _globals['_CREATEREQUEST']._serialized_start=165
  _globals['_CREATEREQUEST']._serialized_end=376
  _globals['_CREATERESPONSE']._serialized_start=378
  _globals['_CREATERESPONSE']._serialized_end=435
  _globals['_DELETEREQUEST']._serialized_start=437
  _globals['_DELETEREQUEST']._serialized_end=493
  _globals['_DELETERESPONSE']._serialized_start=495
  _globals['_DELETERESPONSE']._serialized_end=511
  _globals['_GETREQUEST']._serialized_start=514
  _globals['_GETREQUEST']._serialized_end=880
  _globals['_GETRESPONSE']._serialized_start=882
  _globals['_GETRESPONSE']._serialized_end=954
  _globals['_UPDATEREQUEST']._serialized_start=957
  _globals['_UPDATEREQUEST']._serialized_end=1159
  _globals['_UPDATERESPONSE']._serialized_start=1161
  _globals['_UPDATERESPONSE']._serialized_end=1177
  _globals['_LISTREQUEST']._serialized_start=1179
  _globals['_LISTREQUEST']._serialized_end=1215
  _globals['_LISTRESPONSE']._serialized_start=1217
  _globals['_LISTRESPONSE']._serialized_end=1306
  _globals['_BUILDREQUEST']._serialized_start=1308
  _globals['_BUILDREQUEST']._serialized_end=1363
  _globals['_BUILDRESPONSE']._serialized_start=1365
  _globals['_BUILDRESPONSE']._serialized_end=1459
  _globals['_SETRESOURCESREQUEST']._serialized_start=1462
  _globals['_SETRESOURCESREQUEST']._serialized_end=1676
  _globals['_SETRESOURCESREQUEST_RESOURCESENTRY']._serialized_start=1616
  _globals['_SETRESOURCESREQUEST_RESOURCESENTRY']._serialized_end=1676
  _globals['_SETRESOURCESRESPONSE']._serialized_start=1678
  _globals['_SETRESOURCESRESPONSE']._serialized_end=1700
  _globals['_DOWNLOADRESOURCESREQUEST']._serialized_start=1702
  _globals['_DOWNLOADRESOURCESREQUEST']._serialized_end=1769
  _globals['_DOWNLOADRESOURCESRESPONSE']._serialized_start=1772
  _globals['_DOWNLOADRESOURCESRESPONSE']._serialized_end=1957
  _globals['_DOWNLOADRESOURCESRESPONSE_RESOURCESENTRY']._serialized_start=1616
  _globals['_DOWNLOADRESOURCESRESPONSE_RESOURCESENTRY']._serialized_end=1676
  _globals['_EXPORTREQUEST']._serialized_start=1959
  _globals['_EXPORTREQUEST']._serialized_end=2067
  _globals['_EXPORTRESPONSE']._serialized_start=2069
  _globals['_EXPORTRESPONSE']._serialized_end=2128
  _globals['_LINTREQUEST']._serialized_start=2131
  _globals['_LINTREQUEST']._serialized_end=2366
  _globals['_LINTREQUEST_RESOURCESENTRY']._serialized_start=1616
  _globals['_LINTREQUEST_RESOURCESENTRY']._serialized_end=1676
  _globals['_CHECKVIOLATION']._serialized_start=2369
  _globals['_CHECKVIOLATION']._serialized_end=2637
  _globals['_CHECKVIOLATION_LEVEL']._serialized_start=2571
  _globals['_CHECKVIOLATION_LEVEL']._serialized_end=2637
  _globals['_LINTRESPONSE']._serialized_start=2639
  _globals['_LINTRESPONSE']._serialized_end=2725
  _globals['_PROJECTSSERVICE']._serialized_start=2728
  _globals['_PROJECTSSERVICE']._serialized_end=3662
# @@protoc_insertion_point(module_scope)
