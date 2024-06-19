// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: autokitteh/triggers/v1/svc.proto

package triggersv1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Trigger *Trigger `protobuf:"bytes,1,opt,name=trigger,proto3" json:"trigger,omitempty"`
}

func (x *CreateRequest) Reset() {
	*x = CreateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateRequest) ProtoMessage() {}

func (x *CreateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateRequest.ProtoReflect.Descriptor instead.
func (*CreateRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{0}
}

func (x *CreateRequest) GetTrigger() *Trigger {
	if x != nil {
		return x.Trigger
	}
	return nil
}

type CreateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TriggerId string `protobuf:"bytes,1,opt,name=trigger_id,json=triggerId,proto3" json:"trigger_id,omitempty"`
}

func (x *CreateResponse) Reset() {
	*x = CreateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateResponse) ProtoMessage() {}

func (x *CreateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateResponse.ProtoReflect.Descriptor instead.
func (*CreateResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{1}
}

func (x *CreateResponse) GetTriggerId() string {
	if x != nil {
		return x.TriggerId
	}
	return ""
}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Trigger *Trigger `protobuf:"bytes,1,opt,name=trigger,proto3" json:"trigger,omitempty"`
}

func (x *UpdateRequest) Reset() {
	*x = UpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRequest) ProtoMessage() {}

func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRequest.ProtoReflect.Descriptor instead.
func (*UpdateRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateRequest) GetTrigger() *Trigger {
	if x != nil {
		return x.Trigger
	}
	return nil
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UpdateResponse) Reset() {
	*x = UpdateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateResponse) ProtoMessage() {}

func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateResponse.ProtoReflect.Descriptor instead.
func (*UpdateResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{3}
}

type DeleteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TriggerId string `protobuf:"bytes,1,opt,name=trigger_id,json=triggerId,proto3" json:"trigger_id,omitempty"`
}

func (x *DeleteRequest) Reset() {
	*x = DeleteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteRequest) ProtoMessage() {}

func (x *DeleteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteRequest.ProtoReflect.Descriptor instead.
func (*DeleteRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteRequest) GetTriggerId() string {
	if x != nil {
		return x.TriggerId
	}
	return ""
}

type DeleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteResponse) Reset() {
	*x = DeleteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteResponse) ProtoMessage() {}

func (x *DeleteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteResponse.ProtoReflect.Descriptor instead.
func (*DeleteResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{5}
}

type GetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TriggerId string `protobuf:"bytes,1,opt,name=trigger_id,json=triggerId,proto3" json:"trigger_id,omitempty"`
}

func (x *GetRequest) Reset() {
	*x = GetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRequest) ProtoMessage() {}

func (x *GetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRequest.ProtoReflect.Descriptor instead.
func (*GetRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{6}
}

func (x *GetRequest) GetTriggerId() string {
	if x != nil {
		return x.TriggerId
	}
	return ""
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Trigger *Trigger `protobuf:"bytes,1,opt,name=trigger,proto3" json:"trigger,omitempty"`
}

func (x *GetResponse) Reset() {
	*x = GetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetResponse) ProtoMessage() {}

func (x *GetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetResponse.ProtoReflect.Descriptor instead.
func (*GetResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{7}
}

func (x *GetResponse) GetTrigger() *Trigger {
	if x != nil {
		return x.Trigger
	}
	return nil
}

type ListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EnvId        string `protobuf:"bytes,1,opt,name=env_id,json=envId,proto3" json:"env_id,omitempty"`
	ConnectionId string `protobuf:"bytes,2,opt,name=connection_id,json=connectionId,proto3" json:"connection_id,omitempty"`
	ProjectId    string `protobuf:"bytes,3,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
}

func (x *ListRequest) Reset() {
	*x = ListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRequest) ProtoMessage() {}

func (x *ListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRequest.ProtoReflect.Descriptor instead.
func (*ListRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{8}
}

func (x *ListRequest) GetEnvId() string {
	if x != nil {
		return x.EnvId
	}
	return ""
}

func (x *ListRequest) GetConnectionId() string {
	if x != nil {
		return x.ConnectionId
	}
	return ""
}

func (x *ListRequest) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

type ListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Triggers []*Trigger `protobuf:"bytes,1,rep,name=triggers,proto3" json:"triggers,omitempty"`
}

func (x *ListResponse) Reset() {
	*x = ListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListResponse) ProtoMessage() {}

func (x *ListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_triggers_v1_svc_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListResponse.ProtoReflect.Descriptor instead.
func (*ListResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_triggers_v1_svc_proto_rawDescGZIP(), []int{9}
}

func (x *ListResponse) GetTriggers() []*Trigger {
	if x != nil {
		return x.Triggers
	}
	return nil
}

var File_autokitteh_triggers_v1_svc_proto protoreflect.FileDescriptor

var file_autokitteh_triggers_v1_svc_proto_rawDesc = []byte{
	0x0a, 0x20, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x74, 0x72, 0x69,
	0x67, 0x67, 0x65, 0x72, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x76, 0x63, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x16, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x74,
	0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x24, 0x61, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2f,
	0x76, 0x31, 0x2f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x95, 0x03,
	0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x39, 0x0a, 0x07, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1f, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x74, 0x72,
	0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65,
	0x72, 0x52, 0x07, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x3a, 0xc8, 0x02, 0xfa, 0xf7, 0x18,
	0xc3, 0x02, 0x1a, 0x78, 0x0a, 0x20, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x74, 0x72,
	0x69, 0x67, 0x67, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x5f, 0x6d, 0x75, 0x73, 0x74, 0x5f, 0x62, 0x65,
	0x5f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x20, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x20, 0x6d, 0x75, 0x73, 0x74, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x62, 0x65, 0x20, 0x73,
	0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x65, 0x64, 0x1a, 0x32, 0x68, 0x61, 0x73, 0x28, 0x74, 0x68,
	0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x29, 0x20, 0x26, 0x26, 0x20, 0x74,
	0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x74, 0x72, 0x69, 0x67,
	0x67, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x20, 0x3d, 0x3d, 0x20, 0x27, 0x27, 0x1a, 0xc6, 0x01, 0x0a,
	0x23, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x5f, 0x71, 0x75, 0x61, 0x6c, 0x69, 0x66, 0x69, 0x65, 0x72, 0x73, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x69, 0x72, 0x65, 0x64, 0x12, 0x32, 0x61, 0x74, 0x20, 0x6c, 0x65, 0x61, 0x73, 0x74, 0x20, 0x6f,
	0x6e, 0x65, 0x20, 0x69, 0x73, 0x20, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x3a, 0x20,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x2c, 0x20, 0x66, 0x69, 0x6c, 0x74,
	0x65, 0x72, 0x2c, 0x20, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x6b, 0x68, 0x61, 0x73, 0x28, 0x74, 0x68,
	0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x29, 0x20, 0x26, 0x26, 0x20, 0x28,
	0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x20, 0x21, 0x3d, 0x20, 0x27, 0x27, 0x20, 0x7c, 0x7c,
	0x20, 0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x20, 0x21, 0x3d, 0x20, 0x27, 0x27, 0x20, 0x7c, 0x7c, 0x20, 0x68, 0x61,
	0x73, 0x28, 0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x64,
	0x61, 0x74, 0x61, 0x29, 0x29, 0x22, 0x39, 0x0a, 0x0e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x0a, 0x74, 0x72, 0x69, 0x67, 0x67,
	0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0xf7, 0x18,
	0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x09, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x49, 0x64,
	0x22, 0x86, 0x03, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x39, 0x0a, 0x07, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68,
	0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x69,
	0x67, 0x67, 0x65, 0x72, 0x52, 0x07, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x3a, 0xb9, 0x02,
	0xfa, 0xf7, 0x18, 0xb4, 0x02, 0x1a, 0x69, 0x0a, 0x1b, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x69, 0x72, 0x65, 0x64, 0x12, 0x16, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x5f, 0x69, 0x64,
	0x20, 0x69, 0x73, 0x20, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x1a, 0x32, 0x68, 0x61,
	0x73, 0x28, 0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x29, 0x20,
	0x26, 0x26, 0x20, 0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e,
	0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x20, 0x21, 0x3d, 0x20, 0x27, 0x27,
	0x1a, 0xc6, 0x01, 0x0a, 0x23, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x74, 0x72, 0x69,
	0x67, 0x67, 0x65, 0x72, 0x5f, 0x71, 0x75, 0x61, 0x6c, 0x69, 0x66, 0x69, 0x65, 0x72, 0x73, 0x5f,
	0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x12, 0x32, 0x61, 0x74, 0x20, 0x6c, 0x65, 0x61,
	0x73, 0x74, 0x20, 0x6f, 0x6e, 0x65, 0x20, 0x69, 0x73, 0x20, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72,
	0x65, 0x64, 0x3a, 0x20, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x2c, 0x20,
	0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x2c, 0x20, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x6b, 0x68, 0x61,
	0x73, 0x28, 0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x29, 0x20,
	0x26, 0x26, 0x20, 0x28, 0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x20, 0x21, 0x3d, 0x20, 0x27,
	0x27, 0x20, 0x7c, 0x7c, 0x20, 0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65,
	0x72, 0x2e, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x20, 0x21, 0x3d, 0x20, 0x27, 0x27, 0x20, 0x7c,
	0x7c, 0x20, 0x68, 0x61, 0x73, 0x28, 0x74, 0x68, 0x69, 0x73, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67,
	0x65, 0x72, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x29, 0x29, 0x22, 0x10, 0x0a, 0x0e, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x38, 0x0a, 0x0d, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0a,
	0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x08, 0xfa, 0xf7, 0x18, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x09, 0x74, 0x72, 0x69, 0x67,
	0x67, 0x65, 0x72, 0x49, 0x64, 0x22, 0x10, 0x0a, 0x0e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x35, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0a, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0xf7, 0x18, 0x04, 0x72,
	0x02, 0x10, 0x01, 0x52, 0x09, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x49, 0x64, 0x22, 0x48,
	0x0a, 0x0b, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x39, 0x0a,
	0x07, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f,
	0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x74, 0x72, 0x69, 0x67,
	0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x52,
	0x07, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x22, 0x68, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x65, 0x6e, 0x76, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6e, 0x76, 0x49, 0x64, 0x12, 0x23,
	0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x49, 0x64, 0x22, 0x59, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x49, 0x0a, 0x08, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72,
	0x69, 0x67, 0x67, 0x65, 0x72, 0x42, 0x0c, 0xfa, 0xf7, 0x18, 0x08, 0x92, 0x01, 0x05, 0x22, 0x03,
	0xc8, 0x01, 0x01, 0x52, 0x08, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x32, 0xbf, 0x03,
	0x0a, 0x0f, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x57, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x25, 0x2e, 0x61, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x73, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x26, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x57, 0x0a, 0x06, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x12, 0x25, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x61, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72,
	0x73, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x57, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x25, 0x2e,
	0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67,
	0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4e, 0x0a, 0x03,
	0x47, 0x65, 0x74, 0x12, 0x22, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68,
	0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69,
	0x74, 0x74, 0x65, 0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x51, 0x0a, 0x04,
	0x4c, 0x69, 0x73, 0x74, 0x12, 0x23, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x61, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e,
	0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42,
	0xed, 0x01, 0x0a, 0x1a, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74,
	0x65, 0x68, 0x2e, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x42, 0x08,
	0x53, 0x76, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x4b, 0x67, 0x6f, 0x2e, 0x61,
	0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x61, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67,
	0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68,
	0x2f, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x74, 0x72, 0x69,
	0x67, 0x67, 0x65, 0x72, 0x73, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x41, 0x54, 0x58, 0xaa, 0x02, 0x16,
	0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67,
	0x65, 0x72, 0x73, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x16, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x5c, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x5c, 0x56, 0x31, 0xe2,
	0x02, 0x22, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x5c, 0x54, 0x72, 0x69,
	0x67, 0x67, 0x65, 0x72, 0x73, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x18, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x3a, 0x3a, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x73, 0x3a, 0x3a, 0x56, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_autokitteh_triggers_v1_svc_proto_rawDescOnce sync.Once
	file_autokitteh_triggers_v1_svc_proto_rawDescData = file_autokitteh_triggers_v1_svc_proto_rawDesc
)

func file_autokitteh_triggers_v1_svc_proto_rawDescGZIP() []byte {
	file_autokitteh_triggers_v1_svc_proto_rawDescOnce.Do(func() {
		file_autokitteh_triggers_v1_svc_proto_rawDescData = protoimpl.X.CompressGZIP(file_autokitteh_triggers_v1_svc_proto_rawDescData)
	})
	return file_autokitteh_triggers_v1_svc_proto_rawDescData
}

var file_autokitteh_triggers_v1_svc_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_autokitteh_triggers_v1_svc_proto_goTypes = []interface{}{
	(*CreateRequest)(nil),  // 0: autokitteh.triggers.v1.CreateRequest
	(*CreateResponse)(nil), // 1: autokitteh.triggers.v1.CreateResponse
	(*UpdateRequest)(nil),  // 2: autokitteh.triggers.v1.UpdateRequest
	(*UpdateResponse)(nil), // 3: autokitteh.triggers.v1.UpdateResponse
	(*DeleteRequest)(nil),  // 4: autokitteh.triggers.v1.DeleteRequest
	(*DeleteResponse)(nil), // 5: autokitteh.triggers.v1.DeleteResponse
	(*GetRequest)(nil),     // 6: autokitteh.triggers.v1.GetRequest
	(*GetResponse)(nil),    // 7: autokitteh.triggers.v1.GetResponse
	(*ListRequest)(nil),    // 8: autokitteh.triggers.v1.ListRequest
	(*ListResponse)(nil),   // 9: autokitteh.triggers.v1.ListResponse
	(*Trigger)(nil),        // 10: autokitteh.triggers.v1.Trigger
}
var file_autokitteh_triggers_v1_svc_proto_depIdxs = []int32{
	10, // 0: autokitteh.triggers.v1.CreateRequest.trigger:type_name -> autokitteh.triggers.v1.Trigger
	10, // 1: autokitteh.triggers.v1.UpdateRequest.trigger:type_name -> autokitteh.triggers.v1.Trigger
	10, // 2: autokitteh.triggers.v1.GetResponse.trigger:type_name -> autokitteh.triggers.v1.Trigger
	10, // 3: autokitteh.triggers.v1.ListResponse.triggers:type_name -> autokitteh.triggers.v1.Trigger
	0,  // 4: autokitteh.triggers.v1.TriggersService.Create:input_type -> autokitteh.triggers.v1.CreateRequest
	2,  // 5: autokitteh.triggers.v1.TriggersService.Update:input_type -> autokitteh.triggers.v1.UpdateRequest
	4,  // 6: autokitteh.triggers.v1.TriggersService.Delete:input_type -> autokitteh.triggers.v1.DeleteRequest
	6,  // 7: autokitteh.triggers.v1.TriggersService.Get:input_type -> autokitteh.triggers.v1.GetRequest
	8,  // 8: autokitteh.triggers.v1.TriggersService.List:input_type -> autokitteh.triggers.v1.ListRequest
	1,  // 9: autokitteh.triggers.v1.TriggersService.Create:output_type -> autokitteh.triggers.v1.CreateResponse
	3,  // 10: autokitteh.triggers.v1.TriggersService.Update:output_type -> autokitteh.triggers.v1.UpdateResponse
	5,  // 11: autokitteh.triggers.v1.TriggersService.Delete:output_type -> autokitteh.triggers.v1.DeleteResponse
	7,  // 12: autokitteh.triggers.v1.TriggersService.Get:output_type -> autokitteh.triggers.v1.GetResponse
	9,  // 13: autokitteh.triggers.v1.TriggersService.List:output_type -> autokitteh.triggers.v1.ListResponse
	9,  // [9:14] is the sub-list for method output_type
	4,  // [4:9] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_autokitteh_triggers_v1_svc_proto_init() }
func file_autokitteh_triggers_v1_svc_proto_init() {
	if File_autokitteh_triggers_v1_svc_proto != nil {
		return
	}
	file_autokitteh_triggers_v1_trigger_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_autokitteh_triggers_v1_svc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_autokitteh_triggers_v1_svc_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_autokitteh_triggers_v1_svc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_autokitteh_triggers_v1_svc_proto_goTypes,
		DependencyIndexes: file_autokitteh_triggers_v1_svc_proto_depIdxs,
		MessageInfos:      file_autokitteh_triggers_v1_svc_proto_msgTypes,
	}.Build()
	File_autokitteh_triggers_v1_svc_proto = out.File
	file_autokitteh_triggers_v1_svc_proto_rawDesc = nil
	file_autokitteh_triggers_v1_svc_proto_goTypes = nil
	file_autokitteh_triggers_v1_svc_proto_depIdxs = nil
}
