// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: autokitteh/deployments/v1/deployment.proto

package deploymentsv1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	v1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DeploymentState int32

const (
	DeploymentState_DEPLOYMENT_STATE_UNSPECIFIED DeploymentState = 0
	DeploymentState_DEPLOYMENT_STATE_ACTIVE      DeploymentState = 1
	DeploymentState_DEPLOYMENT_STATE_TESTING     DeploymentState = 2
	DeploymentState_DEPLOYMENT_STATE_DRAINING    DeploymentState = 3
	DeploymentState_DEPLOYMENT_STATE_INACTIVE    DeploymentState = 4
)

// Enum value maps for DeploymentState.
var (
	DeploymentState_name = map[int32]string{
		0: "DEPLOYMENT_STATE_UNSPECIFIED",
		1: "DEPLOYMENT_STATE_ACTIVE",
		2: "DEPLOYMENT_STATE_TESTING",
		3: "DEPLOYMENT_STATE_DRAINING",
		4: "DEPLOYMENT_STATE_INACTIVE",
	}
	DeploymentState_value = map[string]int32{
		"DEPLOYMENT_STATE_UNSPECIFIED": 0,
		"DEPLOYMENT_STATE_ACTIVE":      1,
		"DEPLOYMENT_STATE_TESTING":     2,
		"DEPLOYMENT_STATE_DRAINING":    3,
		"DEPLOYMENT_STATE_INACTIVE":    4,
	}
)

func (x DeploymentState) Enum() *DeploymentState {
	p := new(DeploymentState)
	*p = x
	return p
}

func (x DeploymentState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DeploymentState) Descriptor() protoreflect.EnumDescriptor {
	return file_autokitteh_deployments_v1_deployment_proto_enumTypes[0].Descriptor()
}

func (DeploymentState) Type() protoreflect.EnumType {
	return &file_autokitteh_deployments_v1_deployment_proto_enumTypes[0]
}

func (x DeploymentState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DeploymentState.Descriptor instead.
func (DeploymentState) EnumDescriptor() ([]byte, []int) {
	return file_autokitteh_deployments_v1_deployment_proto_rawDescGZIP(), []int{0}
}

type Deployment struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// immutable fields.
	ProjectId    string `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	DeploymentId string `protobuf:"bytes,2,opt,name=deployment_id,json=deploymentId,proto3" json:"deployment_id,omitempty"`
	BuildId      string `protobuf:"bytes,3,opt,name=build_id,json=buildId,proto3" json:"build_id,omitempty"`
	// mutable fields.
	State         DeploymentState            `protobuf:"varint,4,opt,name=state,proto3,enum=autokitteh.deployments.v1.DeploymentState" json:"state,omitempty"`
	CreatedAt     *timestamppb.Timestamp     `protobuf:"bytes,10,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt     *timestamppb.Timestamp     `protobuf:"bytes,11,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	SessionsStats []*Deployment_SessionStats `protobuf:"bytes,12,rep,name=sessions_stats,json=sessionsStats,proto3" json:"sessions_stats,omitempty"`
}

func (x *Deployment) Reset() {
	*x = Deployment{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_deployments_v1_deployment_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Deployment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Deployment) ProtoMessage() {}

func (x *Deployment) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_deployments_v1_deployment_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Deployment.ProtoReflect.Descriptor instead.
func (*Deployment) Descriptor() ([]byte, []int) {
	return file_autokitteh_deployments_v1_deployment_proto_rawDescGZIP(), []int{0}
}

func (x *Deployment) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *Deployment) GetDeploymentId() string {
	if x != nil {
		return x.DeploymentId
	}
	return ""
}

func (x *Deployment) GetBuildId() string {
	if x != nil {
		return x.BuildId
	}
	return ""
}

func (x *Deployment) GetState() DeploymentState {
	if x != nil {
		return x.State
	}
	return DeploymentState_DEPLOYMENT_STATE_UNSPECIFIED
}

func (x *Deployment) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Deployment) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

func (x *Deployment) GetSessionsStats() []*Deployment_SessionStats {
	if x != nil {
		return x.SessionsStats
	}
	return nil
}

type Deployment_SessionStats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State v1.SessionStateType `protobuf:"varint,1,opt,name=state,proto3,enum=autokitteh.sessions.v1.SessionStateType" json:"state,omitempty"`
	Count uint32              `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *Deployment_SessionStats) Reset() {
	*x = Deployment_SessionStats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_deployments_v1_deployment_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Deployment_SessionStats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Deployment_SessionStats) ProtoMessage() {}

func (x *Deployment_SessionStats) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_deployments_v1_deployment_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Deployment_SessionStats.ProtoReflect.Descriptor instead.
func (*Deployment_SessionStats) Descriptor() ([]byte, []int) {
	return file_autokitteh_deployments_v1_deployment_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Deployment_SessionStats) GetState() v1.SessionStateType {
	if x != nil {
		return x.State
	}
	return v1.SessionStateType(0)
}

func (x *Deployment_SessionStats) GetCount() uint32 {
	if x != nil {
		return x.Count
	}
	return 0
}

var File_autokitteh_deployments_v1_deployment_proto protoreflect.FileDescriptor

var file_autokitteh_deployments_v1_deployment_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x64, 0x65, 0x70,
	0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x61, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x24, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x2f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x2f,
	0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x62,
	0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x83, 0x04, 0x0a, 0x0a,
	0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x27, 0x0a, 0x0a, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08,
	0xfa, 0xf7, 0x18, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x64, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x08, 0x62, 0x75, 0x69, 0x6c,
	0x64, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0xf7, 0x18, 0x04,
	0x72, 0x02, 0x10, 0x01, 0x52, 0x07, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x64, 0x12, 0x4b, 0x0a,
	0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2a, 0x2e, 0x61,
	0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x09, 0xfa, 0xf7, 0x18, 0x05, 0x82, 0x01,
	0x02, 0x10, 0x01, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64,
	0x5f, 0x61, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x59, 0x0a, 0x0e, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x5f, 0x73, 0x74, 0x61,
	0x74, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b,
	0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74,
	0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x2e,
	0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x0d, 0x73, 0x65,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x53, 0x74, 0x61, 0x74, 0x73, 0x1a, 0x64, 0x0a, 0x0c, 0x53,
	0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x3e, 0x0a, 0x05, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x28, 0x2e, 0x61, 0x75, 0x74,
	0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x2a, 0xac, 0x01, 0x0a, 0x0f, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x20, 0x0a, 0x1c, 0x44, 0x45, 0x50, 0x4c, 0x4f, 0x59, 0x4d,
	0x45, 0x4e, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1b, 0x0a, 0x17, 0x44, 0x45, 0x50, 0x4c, 0x4f,
	0x59, 0x4d, 0x45, 0x4e, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x41, 0x43, 0x54, 0x49,
	0x56, 0x45, 0x10, 0x01, 0x12, 0x1c, 0x0a, 0x18, 0x44, 0x45, 0x50, 0x4c, 0x4f, 0x59, 0x4d, 0x45,
	0x4e, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x54, 0x45, 0x53, 0x54, 0x49, 0x4e, 0x47,
	0x10, 0x02, 0x12, 0x1d, 0x0a, 0x19, 0x44, 0x45, 0x50, 0x4c, 0x4f, 0x59, 0x4d, 0x45, 0x4e, 0x54,
	0x5f, 0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x44, 0x52, 0x41, 0x49, 0x4e, 0x49, 0x4e, 0x47, 0x10,
	0x03, 0x12, 0x1d, 0x0a, 0x19, 0x44, 0x45, 0x50, 0x4c, 0x4f, 0x59, 0x4d, 0x45, 0x4e, 0x54, 0x5f,
	0x53, 0x54, 0x41, 0x54, 0x45, 0x5f, 0x49, 0x4e, 0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x10, 0x04,
	0x42, 0x89, 0x02, 0x0a, 0x1d, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2e,
	0x76, 0x31, 0x42, 0x0f, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x51, 0x67, 0x6f, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69,
	0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f,
	0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x64, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x64, 0x65, 0x70, 0x6c, 0x6f,
	0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x41, 0x44, 0x58, 0xaa, 0x02,
	0x19, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x44, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x19, 0x41, 0x75, 0x74,
	0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x5c, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65,
	0x6e, 0x74, 0x73, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x25, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x5c, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x5c,
	0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x1b, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x3a, 0x3a, 0x44, 0x65, 0x70,
	0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_autokitteh_deployments_v1_deployment_proto_rawDescOnce sync.Once
	file_autokitteh_deployments_v1_deployment_proto_rawDescData = file_autokitteh_deployments_v1_deployment_proto_rawDesc
)

func file_autokitteh_deployments_v1_deployment_proto_rawDescGZIP() []byte {
	file_autokitteh_deployments_v1_deployment_proto_rawDescOnce.Do(func() {
		file_autokitteh_deployments_v1_deployment_proto_rawDescData = protoimpl.X.CompressGZIP(file_autokitteh_deployments_v1_deployment_proto_rawDescData)
	})
	return file_autokitteh_deployments_v1_deployment_proto_rawDescData
}

var file_autokitteh_deployments_v1_deployment_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_autokitteh_deployments_v1_deployment_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_autokitteh_deployments_v1_deployment_proto_goTypes = []interface{}{
	(DeploymentState)(0),            // 0: autokitteh.deployments.v1.DeploymentState
	(*Deployment)(nil),              // 1: autokitteh.deployments.v1.Deployment
	(*Deployment_SessionStats)(nil), // 2: autokitteh.deployments.v1.Deployment.SessionStats
	(*timestamppb.Timestamp)(nil),   // 3: google.protobuf.Timestamp
	(v1.SessionStateType)(0),        // 4: autokitteh.sessions.v1.SessionStateType
}
var file_autokitteh_deployments_v1_deployment_proto_depIdxs = []int32{
	0, // 0: autokitteh.deployments.v1.Deployment.state:type_name -> autokitteh.deployments.v1.DeploymentState
	3, // 1: autokitteh.deployments.v1.Deployment.created_at:type_name -> google.protobuf.Timestamp
	3, // 2: autokitteh.deployments.v1.Deployment.updated_at:type_name -> google.protobuf.Timestamp
	2, // 3: autokitteh.deployments.v1.Deployment.sessions_stats:type_name -> autokitteh.deployments.v1.Deployment.SessionStats
	4, // 4: autokitteh.deployments.v1.Deployment.SessionStats.state:type_name -> autokitteh.sessions.v1.SessionStateType
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_autokitteh_deployments_v1_deployment_proto_init() }
func file_autokitteh_deployments_v1_deployment_proto_init() {
	if File_autokitteh_deployments_v1_deployment_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_autokitteh_deployments_v1_deployment_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Deployment); i {
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
		file_autokitteh_deployments_v1_deployment_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Deployment_SessionStats); i {
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
			RawDescriptor: file_autokitteh_deployments_v1_deployment_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_autokitteh_deployments_v1_deployment_proto_goTypes,
		DependencyIndexes: file_autokitteh_deployments_v1_deployment_proto_depIdxs,
		EnumInfos:         file_autokitteh_deployments_v1_deployment_proto_enumTypes,
		MessageInfos:      file_autokitteh_deployments_v1_deployment_proto_msgTypes,
	}.Build()
	File_autokitteh_deployments_v1_deployment_proto = out.File
	file_autokitteh_deployments_v1_deployment_proto_rawDesc = nil
	file_autokitteh_deployments_v1_deployment_proto_goTypes = nil
	file_autokitteh_deployments_v1_deployment_proto_depIdxs = nil
}
