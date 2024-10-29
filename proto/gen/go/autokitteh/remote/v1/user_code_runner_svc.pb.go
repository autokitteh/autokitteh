// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: autokitteh/remote/v1/user_code_runner_svc.proto

package remotev1

import (
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

type ExportsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FileName string `protobuf:"bytes,1,opt,name=file_name,json=fileName,proto3" json:"file_name,omitempty"`
}

func (x *ExportsRequest) Reset() {
	*x = ExportsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExportsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportsRequest) ProtoMessage() {}

func (x *ExportsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportsRequest.ProtoReflect.Descriptor instead.
func (*ExportsRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{0}
}

func (x *ExportsRequest) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

type ExportsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Exports []string `protobuf:"bytes,1,rep,name=exports,proto3" json:"exports,omitempty"`
	Error   string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ExportsResponse) Reset() {
	*x = ExportsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExportsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportsResponse) ProtoMessage() {}

func (x *ExportsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportsResponse.ProtoReflect.Descriptor instead.
func (*ExportsResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{1}
}

func (x *ExportsResponse) GetExports() []string {
	if x != nil {
		return x.Exports
	}
	return nil
}

func (x *ExportsResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type StartRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EntryPoint string `protobuf:"bytes,1,opt,name=entry_point,json=entryPoint,proto3" json:"entry_point,omitempty"` // "main.py:on_event"
	Event      *Event `protobuf:"bytes,2,opt,name=event,proto3" json:"event,omitempty"`
}

func (x *StartRequest) Reset() {
	*x = StartRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartRequest) ProtoMessage() {}

func (x *StartRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartRequest.ProtoReflect.Descriptor instead.
func (*StartRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{2}
}

func (x *StartRequest) GetEntryPoint() string {
	if x != nil {
		return x.EntryPoint
	}
	return ""
}

func (x *StartRequest) GetEvent() *Event {
	if x != nil {
		return x.Event
	}
	return nil
}

type ExecuteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ExecuteRequest) Reset() {
	*x = ExecuteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecuteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteRequest) ProtoMessage() {}

func (x *ExecuteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteRequest.ProtoReflect.Descriptor instead.
func (*ExecuteRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{3}
}

func (x *ExecuteRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type ExecuteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result    []byte   `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
	Error     string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	Traceback []*Frame `protobuf:"bytes,3,rep,name=traceback,proto3" json:"traceback,omitempty"`
}

func (x *ExecuteResponse) Reset() {
	*x = ExecuteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecuteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteResponse) ProtoMessage() {}

func (x *ExecuteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteResponse.ProtoReflect.Descriptor instead.
func (*ExecuteResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{4}
}

func (x *ExecuteResponse) GetResult() []byte {
	if x != nil {
		return x.Result
	}
	return nil
}

func (x *ExecuteResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *ExecuteResponse) GetTraceback() []*Frame {
	if x != nil {
		return x.Traceback
	}
	return nil
}

type StartResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error     string   `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Traceback []*Frame `protobuf:"bytes,2,rep,name=traceback,proto3" json:"traceback,omitempty"`
}

func (x *StartResponse) Reset() {
	*x = StartResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartResponse) ProtoMessage() {}

func (x *StartResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartResponse.ProtoReflect.Descriptor instead.
func (*StartResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{5}
}

func (x *StartResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *StartResponse) GetTraceback() []*Frame {
	if x != nil {
		return x.Traceback
	}
	return nil
}

type ActivityReplyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data   []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Result []byte `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"`
	Error  string `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ActivityReplyRequest) Reset() {
	*x = ActivityReplyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ActivityReplyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActivityReplyRequest) ProtoMessage() {}

func (x *ActivityReplyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActivityReplyRequest.ProtoReflect.Descriptor instead.
func (*ActivityReplyRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{6}
}

func (x *ActivityReplyRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *ActivityReplyRequest) GetResult() []byte {
	if x != nil {
		return x.Result
	}
	return nil
}

func (x *ActivityReplyRequest) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type ActivityReplyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ActivityReplyResponse) Reset() {
	*x = ActivityReplyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ActivityReplyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActivityReplyResponse) ProtoMessage() {}

func (x *ActivityReplyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActivityReplyResponse.ProtoReflect.Descriptor instead.
func (*ActivityReplyResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{7}
}

func (x *ActivityReplyResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type UserCodeRunnerHealthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UserCodeRunnerHealthRequest) Reset() {
	*x = UserCodeRunnerHealthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserCodeRunnerHealthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserCodeRunnerHealthRequest) ProtoMessage() {}

func (x *UserCodeRunnerHealthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserCodeRunnerHealthRequest.ProtoReflect.Descriptor instead.
func (*UserCodeRunnerHealthRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{8}
}

type UserCodeRunnerHealthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *UserCodeRunnerHealthResponse) Reset() {
	*x = UserCodeRunnerHealthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserCodeRunnerHealthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserCodeRunnerHealthResponse) ProtoMessage() {}

func (x *UserCodeRunnerHealthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserCodeRunnerHealthResponse.ProtoReflect.Descriptor instead.
func (*UserCodeRunnerHealthResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP(), []int{9}
}

func (x *UserCodeRunnerHealthResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_autokitteh_remote_v1_user_code_runner_svc_proto protoreflect.FileDescriptor

var file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDesc = []byte{
	0x0a, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x72, 0x65, 0x6d,
	0x6f, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65,
	0x5f, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x73, 0x76, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x14, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65,
	0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x20, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x2f, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2d, 0x0a, 0x0e, 0x45, 0x78, 0x70,
	0x6f, 0x72, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x66,
	0x69, 0x6c, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x41, 0x0a, 0x0f, 0x45, 0x78, 0x70, 0x6f,
	0x72, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x65,
	0x78, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x65, 0x78,
	0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x62, 0x0a, 0x0c, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x65,
	0x6e, 0x74, 0x72, 0x79, 0x5f, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x31, 0x0a, 0x05,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x61, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22,
	0x24, 0x0a, 0x0e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x7a, 0x0a, 0x0f, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x39, 0x0a, 0x09, 0x74, 0x72, 0x61, 0x63, 0x65, 0x62,
	0x61, 0x63, 0x6b, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x61, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x52, 0x09, 0x74, 0x72, 0x61, 0x63, 0x65, 0x62, 0x61, 0x63,
	0x6b, 0x22, 0x60, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x39, 0x0a, 0x09, 0x74, 0x72, 0x61, 0x63,
	0x65, 0x62, 0x61, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x61, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x52, 0x09, 0x74, 0x72, 0x61, 0x63, 0x65, 0x62,
	0x61, 0x63, 0x6b, 0x22, 0x58, 0x0a, 0x14, 0x41, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x2d, 0x0a,
	0x15, 0x41, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x1d, 0x0a, 0x1b,
	0x55, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x34, 0x0a, 0x1c, 0x55,
	0x73, 0x65, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x48, 0x65, 0x61,
	0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x32, 0xfe, 0x03, 0x0a, 0x15, 0x55, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x75,
	0x6e, 0x6e, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x58, 0x0a, 0x07, 0x45,
	0x78, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x24, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78,
	0x70, 0x6f, 0x72, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x61,
	0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x52, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x72, 0x74, 0x12, 0x22,
	0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f,
	0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x23, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x58, 0x0a, 0x07, 0x45, 0x78, 0x65,
	0x63, 0x75, 0x74, 0x65, 0x12, 0x24, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x61, 0x75, 0x74,
	0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x6a, 0x0a, 0x0d, 0x41, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x12, 0x2a, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x63, 0x74, 0x69,
	0x76, 0x69, 0x74, 0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x2b, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65,
	0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x71, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12, 0x31, 0x2e, 0x61, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x55, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x32, 0x2e, 0x61,
	0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x75, 0x6e, 0x6e,
	0x65, 0x72, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x42, 0xed, 0x01, 0x0a, 0x18, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b,
	0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x42,
	0x16, 0x55, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x53,
	0x76, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x47, 0x67, 0x6f, 0x2e, 0x61, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x61, 0x75, 0x74,
	0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65,
	0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f,
	0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65,
	0x76, 0x31, 0xa2, 0x02, 0x03, 0x41, 0x52, 0x58, 0xaa, 0x02, 0x14, 0x41, 0x75, 0x74, 0x6f, 0x6b,
	0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x56, 0x31, 0xca,
	0x02, 0x14, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x5c, 0x52, 0x65, 0x6d,
	0x6f, 0x74, 0x65, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x20, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x5c, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50,
	0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x16, 0x41, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x3a, 0x3a, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x3a, 0x3a,
	0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescOnce sync.Once
	file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescData = file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDesc
)

func file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescGZIP() []byte {
	file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescOnce.Do(func() {
		file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescData = protoimpl.X.CompressGZIP(file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescData)
	})
	return file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDescData
}

var file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_autokitteh_remote_v1_user_code_runner_svc_proto_goTypes = []interface{}{
	(*ExportsRequest)(nil),               // 0: autokitteh.remote.v1.ExportsRequest
	(*ExportsResponse)(nil),              // 1: autokitteh.remote.v1.ExportsResponse
	(*StartRequest)(nil),                 // 2: autokitteh.remote.v1.StartRequest
	(*ExecuteRequest)(nil),               // 3: autokitteh.remote.v1.ExecuteRequest
	(*ExecuteResponse)(nil),              // 4: autokitteh.remote.v1.ExecuteResponse
	(*StartResponse)(nil),                // 5: autokitteh.remote.v1.StartResponse
	(*ActivityReplyRequest)(nil),         // 6: autokitteh.remote.v1.ActivityReplyRequest
	(*ActivityReplyResponse)(nil),        // 7: autokitteh.remote.v1.ActivityReplyResponse
	(*UserCodeRunnerHealthRequest)(nil),  // 8: autokitteh.remote.v1.UserCodeRunnerHealthRequest
	(*UserCodeRunnerHealthResponse)(nil), // 9: autokitteh.remote.v1.UserCodeRunnerHealthResponse
	(*Event)(nil),                        // 10: autokitteh.remote.v1.Event
	(*Frame)(nil),                        // 11: autokitteh.remote.v1.Frame
}
var file_autokitteh_remote_v1_user_code_runner_svc_proto_depIdxs = []int32{
	10, // 0: autokitteh.remote.v1.StartRequest.event:type_name -> autokitteh.remote.v1.Event
	11, // 1: autokitteh.remote.v1.ExecuteResponse.traceback:type_name -> autokitteh.remote.v1.Frame
	11, // 2: autokitteh.remote.v1.StartResponse.traceback:type_name -> autokitteh.remote.v1.Frame
	0,  // 3: autokitteh.remote.v1.UserCodeRunnerService.Exports:input_type -> autokitteh.remote.v1.ExportsRequest
	2,  // 4: autokitteh.remote.v1.UserCodeRunnerService.Start:input_type -> autokitteh.remote.v1.StartRequest
	3,  // 5: autokitteh.remote.v1.UserCodeRunnerService.Execute:input_type -> autokitteh.remote.v1.ExecuteRequest
	6,  // 6: autokitteh.remote.v1.UserCodeRunnerService.ActivityReply:input_type -> autokitteh.remote.v1.ActivityReplyRequest
	8,  // 7: autokitteh.remote.v1.UserCodeRunnerService.Health:input_type -> autokitteh.remote.v1.UserCodeRunnerHealthRequest
	1,  // 8: autokitteh.remote.v1.UserCodeRunnerService.Exports:output_type -> autokitteh.remote.v1.ExportsResponse
	5,  // 9: autokitteh.remote.v1.UserCodeRunnerService.Start:output_type -> autokitteh.remote.v1.StartResponse
	4,  // 10: autokitteh.remote.v1.UserCodeRunnerService.Execute:output_type -> autokitteh.remote.v1.ExecuteResponse
	7,  // 11: autokitteh.remote.v1.UserCodeRunnerService.ActivityReply:output_type -> autokitteh.remote.v1.ActivityReplyResponse
	9,  // 12: autokitteh.remote.v1.UserCodeRunnerService.Health:output_type -> autokitteh.remote.v1.UserCodeRunnerHealthResponse
	8,  // [8:13] is the sub-list for method output_type
	3,  // [3:8] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_autokitteh_remote_v1_user_code_runner_svc_proto_init() }
func file_autokitteh_remote_v1_user_code_runner_svc_proto_init() {
	if File_autokitteh_remote_v1_user_code_runner_svc_proto != nil {
		return
	}
	file_autokitteh_remote_v1_types_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExportsRequest); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExportsResponse); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartRequest); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecuteRequest); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecuteResponse); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartResponse); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ActivityReplyRequest); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ActivityReplyResponse); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserCodeRunnerHealthRequest); i {
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
		file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserCodeRunnerHealthResponse); i {
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
			RawDescriptor: file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_autokitteh_remote_v1_user_code_runner_svc_proto_goTypes,
		DependencyIndexes: file_autokitteh_remote_v1_user_code_runner_svc_proto_depIdxs,
		MessageInfos:      file_autokitteh_remote_v1_user_code_runner_svc_proto_msgTypes,
	}.Build()
	File_autokitteh_remote_v1_user_code_runner_svc_proto = out.File
	file_autokitteh_remote_v1_user_code_runner_svc_proto_rawDesc = nil
	file_autokitteh_remote_v1_user_code_runner_svc_proto_goTypes = nil
	file_autokitteh_remote_v1_user_code_runner_svc_proto_depIdxs = nil
}
