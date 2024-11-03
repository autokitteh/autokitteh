// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: autokitteh/runner_manager/v1/runner_manager_svc.proto

package runner_managerv1

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

type ContainerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Image string `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"` // TBD by @efiShtain
}

func (x *ContainerConfig) Reset() {
	*x = ContainerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContainerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContainerConfig) ProtoMessage() {}

func (x *ContainerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContainerConfig.ProtoReflect.Descriptor instead.
func (*ContainerConfig) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{0}
}

func (x *ContainerConfig) GetImage() string {
	if x != nil {
		return x.Image
	}
	return ""
}

// TODO: Will become Start once we split to files
// Tell runner manager to start a runner
type StartRunnerRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ContainerConfig *ContainerConfig `protobuf:"bytes,1,opt,name=container_config,json=containerConfig,proto3" json:"container_config,omitempty"`
	// user code as tar archive
	BuildArtifact []byte `protobuf:"bytes,2,opt,name=build_artifact,json=buildArtifact,proto3" json:"build_artifact,omitempty"`
	// vars from manifest, secrets and connections
	Vars          map[string]string `protobuf:"bytes,3,rep,name=vars,proto3" json:"vars,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	WorkerAddress string            `protobuf:"bytes,4,opt,name=worker_address,json=workerAddress,proto3" json:"worker_address,omitempty"`
}

func (x *StartRunnerRequest) Reset() {
	*x = StartRunnerRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartRunnerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartRunnerRequest) ProtoMessage() {}

func (x *StartRunnerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartRunnerRequest.ProtoReflect.Descriptor instead.
func (*StartRunnerRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{1}
}

func (x *StartRunnerRequest) GetContainerConfig() *ContainerConfig {
	if x != nil {
		return x.ContainerConfig
	}
	return nil
}

func (x *StartRunnerRequest) GetBuildArtifact() []byte {
	if x != nil {
		return x.BuildArtifact
	}
	return nil
}

func (x *StartRunnerRequest) GetVars() map[string]string {
	if x != nil {
		return x.Vars
	}
	return nil
}

func (x *StartRunnerRequest) GetWorkerAddress() string {
	if x != nil {
		return x.WorkerAddress
	}
	return ""
}

type StartRunnerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunnerId      string `protobuf:"bytes,1,opt,name=runner_id,json=runnerId,proto3" json:"runner_id,omitempty"`
	RunnerAddress string `protobuf:"bytes,2,opt,name=runner_address,json=runnerAddress,proto3" json:"runner_address,omitempty"`
	Error         string `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *StartRunnerResponse) Reset() {
	*x = StartRunnerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartRunnerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartRunnerResponse) ProtoMessage() {}

func (x *StartRunnerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartRunnerResponse.ProtoReflect.Descriptor instead.
func (*StartRunnerResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{2}
}

func (x *StartRunnerResponse) GetRunnerId() string {
	if x != nil {
		return x.RunnerId
	}
	return ""
}

func (x *StartRunnerResponse) GetRunnerAddress() string {
	if x != nil {
		return x.RunnerAddress
	}
	return ""
}

func (x *StartRunnerResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type RunnerHealthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunnerId string `protobuf:"bytes,1,opt,name=runner_id,json=runnerId,proto3" json:"runner_id,omitempty"`
}

func (x *RunnerHealthRequest) Reset() {
	*x = RunnerHealthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunnerHealthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunnerHealthRequest) ProtoMessage() {}

func (x *RunnerHealthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunnerHealthRequest.ProtoReflect.Descriptor instead.
func (*RunnerHealthRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{3}
}

func (x *RunnerHealthRequest) GetRunnerId() string {
	if x != nil {
		return x.RunnerId
	}
	return ""
}

type RunnerHealthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Healthy bool   `protobuf:"varint,1,opt,name=healthy,proto3" json:"healthy,omitempty"`
	Error   string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *RunnerHealthResponse) Reset() {
	*x = RunnerHealthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunnerHealthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunnerHealthResponse) ProtoMessage() {}

func (x *RunnerHealthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunnerHealthResponse.ProtoReflect.Descriptor instead.
func (*RunnerHealthResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{4}
}

func (x *RunnerHealthResponse) GetHealthy() bool {
	if x != nil {
		return x.Healthy
	}
	return false
}

func (x *RunnerHealthResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type StopRunnerRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunnerId string `protobuf:"bytes,1,opt,name=runner_id,json=runnerId,proto3" json:"runner_id,omitempty"`
}

func (x *StopRunnerRequest) Reset() {
	*x = StopRunnerRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopRunnerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopRunnerRequest) ProtoMessage() {}

func (x *StopRunnerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopRunnerRequest.ProtoReflect.Descriptor instead.
func (*StopRunnerRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{5}
}

func (x *StopRunnerRequest) GetRunnerId() string {
	if x != nil {
		return x.RunnerId
	}
	return ""
}

type StopRunnerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *StopRunnerResponse) Reset() {
	*x = StopRunnerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopRunnerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopRunnerResponse) ProtoMessage() {}

func (x *StopRunnerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopRunnerResponse.ProtoReflect.Descriptor instead.
func (*StopRunnerResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{6}
}

func (x *StopRunnerResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type HealthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HealthRequest) Reset() {
	*x = HealthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthRequest) ProtoMessage() {}

func (x *HealthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthRequest.ProtoReflect.Descriptor instead.
func (*HealthRequest) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{7}
}

type HealthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *HealthResponse) Reset() {
	*x = HealthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthResponse) ProtoMessage() {}

func (x *HealthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthResponse.ProtoReflect.Descriptor instead.
func (*HealthResponse) Descriptor() ([]byte, []int) {
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP(), []int{8}
}

func (x *HealthResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_autokitteh_runner_manager_v1_runner_manager_svc_proto protoreflect.FileDescriptor

var file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDesc = []byte{
	0x0a, 0x35, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x72, 0x75, 0x6e,
	0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x72,
	0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x5f, 0x73, 0x76,
	0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1c, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x22, 0x27, 0x0a, 0x0f, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e,
	0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6d, 0x61, 0x67,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x22, 0xc5,
	0x02, 0x0a, 0x12, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x58, 0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e,
	0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x2d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x75, 0x6e,
	0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43,
	0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0f,
	0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x25, 0x0a, 0x0e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x41, 0x72,
	0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x12, 0x4e, 0x0a, 0x04, 0x76, 0x61, 0x72, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x3a, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x56, 0x61, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x04, 0x76, 0x61, 0x72, 0x73, 0x12, 0x25, 0x0a, 0x0e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72,
	0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d,
	0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x1a, 0x37, 0x0a,
	0x09, 0x56, 0x61, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x6f, 0x0a, 0x13, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52,
	0x75, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1b, 0x0a,
	0x09, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x75,
	0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0d, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x32, 0x0a, 0x13, 0x52, 0x75, 0x6e, 0x6e, 0x65,
	0x72, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b,
	0x0a, 0x09, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x22, 0x46, 0x0a, 0x14, 0x52,
	0x75, 0x6e, 0x6e, 0x65, 0x72, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79, 0x12, 0x14, 0x0a,
	0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x22, 0x30, 0x0a, 0x11, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x75, 0x6e, 0x6e, 0x65,
	0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x72, 0x75, 0x6e, 0x6e,
	0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x75, 0x6e,
	0x6e, 0x65, 0x72, 0x49, 0x64, 0x22, 0x2a, 0x0a, 0x12, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x75, 0x6e,
	0x6e, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x22, 0x0f, 0x0a, 0x0d, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x26, 0x0a, 0x0e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x32, 0xdf, 0x03, 0x0a, 0x14, 0x52,
	0x75, 0x6e, 0x6e, 0x65, 0x72, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x74, 0x0a, 0x0b, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x75, 0x6e, 0x6e,
	0x65, 0x72, 0x12, 0x30, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x31, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x77, 0x0a, 0x0c, 0x52, 0x75, 0x6e,
	0x6e, 0x65, 0x72, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12, 0x31, 0x2e, 0x61, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61,
	0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x32, 0x2e, 0x61,
	0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72,
	0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x75, 0x6e, 0x6e,
	0x65, 0x72, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x71, 0x0a, 0x0a, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72,
	0x12, 0x2f, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x75,
	0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x53, 0x74, 0x6f, 0x70, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x30, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72,
	0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x65, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12,
	0x2b, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x75, 0x6e,
	0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x61,
	0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72,
	0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0xa0, 0x02, 0x0a,
	0x20, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x42, 0x15, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x53, 0x76, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x57, 0x67, 0x6f, 0x2e, 0x61,
	0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x61, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67,
	0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68,
	0x2f, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f,
	0x76, 0x31, 0x3b, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65,
	0x72, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x41, 0x52, 0x58, 0xaa, 0x02, 0x1b, 0x41, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x4d, 0x61, 0x6e,
	0x61, 0x67, 0x65, 0x72, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x1b, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69,
	0x74, 0x74, 0x65, 0x68, 0x5c, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x4d, 0x61, 0x6e, 0x61, 0x67,
	0x65, 0x72, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x27, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74,
	0x65, 0x68, 0x5c, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x1d, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x3a, 0x3a, 0x52, 0x75,
	0x6e, 0x6e, 0x65, 0x72, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x3a, 0x3a, 0x56, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescOnce sync.Once
	file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescData = file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDesc
)

func file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescGZIP() []byte {
	file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescOnce.Do(func() {
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescData = protoimpl.X.CompressGZIP(file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescData)
	})
	return file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDescData
}

var file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_autokitteh_runner_manager_v1_runner_manager_svc_proto_goTypes = []interface{}{
	(*ContainerConfig)(nil),      // 0: autokitteh.runner_manager.v1.ContainerConfig
	(*StartRunnerRequest)(nil),   // 1: autokitteh.runner_manager.v1.StartRunnerRequest
	(*StartRunnerResponse)(nil),  // 2: autokitteh.runner_manager.v1.StartRunnerResponse
	(*RunnerHealthRequest)(nil),  // 3: autokitteh.runner_manager.v1.RunnerHealthRequest
	(*RunnerHealthResponse)(nil), // 4: autokitteh.runner_manager.v1.RunnerHealthResponse
	(*StopRunnerRequest)(nil),    // 5: autokitteh.runner_manager.v1.StopRunnerRequest
	(*StopRunnerResponse)(nil),   // 6: autokitteh.runner_manager.v1.StopRunnerResponse
	(*HealthRequest)(nil),        // 7: autokitteh.runner_manager.v1.HealthRequest
	(*HealthResponse)(nil),       // 8: autokitteh.runner_manager.v1.HealthResponse
	nil,                          // 9: autokitteh.runner_manager.v1.StartRunnerRequest.VarsEntry
}
var file_autokitteh_runner_manager_v1_runner_manager_svc_proto_depIdxs = []int32{
	0, // 0: autokitteh.runner_manager.v1.StartRunnerRequest.container_config:type_name -> autokitteh.runner_manager.v1.ContainerConfig
	9, // 1: autokitteh.runner_manager.v1.StartRunnerRequest.vars:type_name -> autokitteh.runner_manager.v1.StartRunnerRequest.VarsEntry
	1, // 2: autokitteh.runner_manager.v1.RunnerManagerService.StartRunner:input_type -> autokitteh.runner_manager.v1.StartRunnerRequest
	3, // 3: autokitteh.runner_manager.v1.RunnerManagerService.RunnerHealth:input_type -> autokitteh.runner_manager.v1.RunnerHealthRequest
	5, // 4: autokitteh.runner_manager.v1.RunnerManagerService.StopRunner:input_type -> autokitteh.runner_manager.v1.StopRunnerRequest
	7, // 5: autokitteh.runner_manager.v1.RunnerManagerService.Health:input_type -> autokitteh.runner_manager.v1.HealthRequest
	2, // 6: autokitteh.runner_manager.v1.RunnerManagerService.StartRunner:output_type -> autokitteh.runner_manager.v1.StartRunnerResponse
	4, // 7: autokitteh.runner_manager.v1.RunnerManagerService.RunnerHealth:output_type -> autokitteh.runner_manager.v1.RunnerHealthResponse
	6, // 8: autokitteh.runner_manager.v1.RunnerManagerService.StopRunner:output_type -> autokitteh.runner_manager.v1.StopRunnerResponse
	8, // 9: autokitteh.runner_manager.v1.RunnerManagerService.Health:output_type -> autokitteh.runner_manager.v1.HealthResponse
	6, // [6:10] is the sub-list for method output_type
	2, // [2:6] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_autokitteh_runner_manager_v1_runner_manager_svc_proto_init() }
func file_autokitteh_runner_manager_v1_runner_manager_svc_proto_init() {
	if File_autokitteh_runner_manager_v1_runner_manager_svc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContainerConfig); i {
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
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartRunnerRequest); i {
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
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartRunnerResponse); i {
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
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunnerHealthRequest); i {
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
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunnerHealthResponse); i {
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
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopRunnerRequest); i {
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
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopRunnerResponse); i {
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
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthRequest); i {
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
		file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthResponse); i {
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
			RawDescriptor: file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_autokitteh_runner_manager_v1_runner_manager_svc_proto_goTypes,
		DependencyIndexes: file_autokitteh_runner_manager_v1_runner_manager_svc_proto_depIdxs,
		MessageInfos:      file_autokitteh_runner_manager_v1_runner_manager_svc_proto_msgTypes,
	}.Build()
	File_autokitteh_runner_manager_v1_runner_manager_svc_proto = out.File
	file_autokitteh_runner_manager_v1_runner_manager_svc_proto_rawDesc = nil
	file_autokitteh_runner_manager_v1_runner_manager_svc_proto_goTypes = nil
	file_autokitteh_runner_manager_v1_runner_manager_svc_proto_depIdxs = nil
}