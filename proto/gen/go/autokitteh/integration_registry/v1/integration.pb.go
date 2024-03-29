// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: autokitteh/integration_registry/v1/integration.proto

package integration_registryv1

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

type Visibility int32

const (
	Visibility_VISIBILITY_UNSPECIFIED Visibility = 0
	// Visible only to the user who registered the integration (default).
	Visibility_VISIBILITY_PRIVATE Visibility = 1
	// Visible to all the users in the owner organization, but only to them.
	Visibility_VISIBILITY_INTERNAL Visibility = 2
	// Visible to all the users in the server.
	// TBD: May require coordination/approval with autokitteh personnel.
	Visibility_VISIBILITY_PUBLIC Visibility = 3
)

// Enum value maps for Visibility.
var (
	Visibility_name = map[int32]string{
		0: "VISIBILITY_UNSPECIFIED",
		1: "VISIBILITY_PRIVATE",
		2: "VISIBILITY_INTERNAL",
		3: "VISIBILITY_PUBLIC",
	}
	Visibility_value = map[string]int32{
		"VISIBILITY_UNSPECIFIED": 0,
		"VISIBILITY_PRIVATE":     1,
		"VISIBILITY_INTERNAL":    2,
		"VISIBILITY_PUBLIC":      3,
	}
)

func (x Visibility) Enum() *Visibility {
	p := new(Visibility)
	*p = x
	return p
}

func (x Visibility) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Visibility) Descriptor() protoreflect.EnumDescriptor {
	return file_autokitteh_integration_registry_v1_integration_proto_enumTypes[0].Descriptor()
}

func (Visibility) Type() protoreflect.EnumType {
	return &file_autokitteh_integration_registry_v1_integration_proto_enumTypes[0]
}

func (x Visibility) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Visibility.Descriptor instead.
func (Visibility) EnumDescriptor() ([]byte, []int) {
	return file_autokitteh_integration_registry_v1_integration_proto_rawDescGZIP(), []int{0}
}

type Integration struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IntegrationId string `protobuf:"bytes,1,opt,name=integration_id,json=integrationId,proto3" json:"integration_id,omitempty"`
	// Visible (non-internal) identifier of the integration.
	// Unique in the context of the owner's (org/user) name.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Owning organization or user ID.
	OwnerId    string     `protobuf:"bytes,3,opt,name=owner_id,json=ownerId,proto3" json:"owner_id,omitempty"`
	Visibility Visibility `protobuf:"varint,4,opt,name=visibility,proto3,enum=autokitteh.integration_registry.v1.Visibility" json:"visibility,omitempty"`
	// Integration gRPC service to which autokitteh can send requests.
	ApiUrl string `protobuf:"bytes,5,opt,name=api_url,json=apiUrl,proto3" json:"api_url,omitempty"`
	// User-visible details.
	DisplayName string `protobuf:"bytes,6,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	Description string `protobuf:"bytes,7,opt,name=description,proto3" json:"description,omitempty"`
	LogoUrl     string `protobuf:"bytes,8,opt,name=logo_url,json=logoUrl,proto3" json:"logo_url,omitempty"`
	HomepageUrl string `protobuf:"bytes,9,opt,name=homepage_url,json=homepageUrl,proto3" json:"homepage_url,omitempty"`
	ConnectUrl  string `protobuf:"bytes,10,opt,name=connect_url,json=connectUrl,proto3" json:"connect_url,omitempty"`
	ApiKey      string `protobuf:"bytes,11,opt,name=api_key,json=apiKey,proto3" json:"api_key,omitempty"`
	SigningKey  string `protobuf:"bytes,12,opt,name=signing_key,json=signingKey,proto3" json:"signing_key,omitempty"`
}

func (x *Integration) Reset() {
	*x = Integration{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_integration_registry_v1_integration_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Integration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Integration) ProtoMessage() {}

func (x *Integration) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_integration_registry_v1_integration_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Integration.ProtoReflect.Descriptor instead.
func (*Integration) Descriptor() ([]byte, []int) {
	return file_autokitteh_integration_registry_v1_integration_proto_rawDescGZIP(), []int{0}
}

func (x *Integration) GetIntegrationId() string {
	if x != nil {
		return x.IntegrationId
	}
	return ""
}

func (x *Integration) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Integration) GetOwnerId() string {
	if x != nil {
		return x.OwnerId
	}
	return ""
}

func (x *Integration) GetVisibility() Visibility {
	if x != nil {
		return x.Visibility
	}
	return Visibility_VISIBILITY_UNSPECIFIED
}

func (x *Integration) GetApiUrl() string {
	if x != nil {
		return x.ApiUrl
	}
	return ""
}

func (x *Integration) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *Integration) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Integration) GetLogoUrl() string {
	if x != nil {
		return x.LogoUrl
	}
	return ""
}

func (x *Integration) GetHomepageUrl() string {
	if x != nil {
		return x.HomepageUrl
	}
	return ""
}

func (x *Integration) GetConnectUrl() string {
	if x != nil {
		return x.ConnectUrl
	}
	return ""
}

func (x *Integration) GetApiKey() string {
	if x != nil {
		return x.ApiKey
	}
	return ""
}

func (x *Integration) GetSigningKey() string {
	if x != nil {
		return x.SigningKey
	}
	return ""
}

var File_autokitteh_integration_registry_v1_integration_proto protoreflect.FileDescriptor

var file_autokitteh_integration_registry_v1_integration_proto_rawDesc = []byte{
	0x0a, 0x34, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x22, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74,
	0x65, 0x68, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb5, 0x03, 0x0a, 0x0b, 0x49, 0x6e, 0x74, 0x65,
	0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x69, 0x6e, 0x74, 0x65, 0x67,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x49, 0x64, 0x12, 0x4e, 0x0a,
	0x0a, 0x76, 0x69, 0x73, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x2e, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x69,
	0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x69, 0x73, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74,
	0x79, 0x52, 0x0a, 0x76, 0x69, 0x73, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x12, 0x22, 0x0a,
	0x07, 0x61, 0x70, 0x69, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x42, 0x09,
	0xfa, 0xf7, 0x18, 0x05, 0x72, 0x03, 0x90, 0x01, 0x01, 0x52, 0x06, 0x61, 0x70, 0x69, 0x55, 0x72,
	0x6c, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x6c, 0x6f, 0x67, 0x6f, 0x5f, 0x75,
	0x72, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6c, 0x6f, 0x67, 0x6f, 0x55, 0x72,
	0x6c, 0x12, 0x21, 0x0a, 0x0c, 0x68, 0x6f, 0x6d, 0x65, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x75, 0x72,
	0x6c, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x68, 0x6f, 0x6d, 0x65, 0x70, 0x61, 0x67,
	0x65, 0x55, 0x72, 0x6c, 0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f,
	0x75, 0x72, 0x6c, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x55, 0x72, 0x6c, 0x12, 0x17, 0x0a, 0x07, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65, 0x79,
	0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x70, 0x69, 0x4b, 0x65, 0x79, 0x12, 0x1f,
	0x0a, 0x0b, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x0c, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x4b, 0x65, 0x79, 0x2a,
	0x70, 0x0a, 0x0a, 0x56, 0x69, 0x73, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x12, 0x1a, 0x0a,
	0x16, 0x56, 0x49, 0x53, 0x49, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x16, 0x0a, 0x12, 0x56, 0x49, 0x53,
	0x49, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x50, 0x52, 0x49, 0x56, 0x41, 0x54, 0x45, 0x10,
	0x01, 0x12, 0x17, 0x0a, 0x13, 0x56, 0x49, 0x53, 0x49, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f,
	0x49, 0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c, 0x10, 0x02, 0x12, 0x15, 0x0a, 0x11, 0x56, 0x49,
	0x53, 0x49, 0x42, 0x49, 0x4c, 0x49, 0x54, 0x59, 0x5f, 0x50, 0x55, 0x42, 0x4c, 0x49, 0x43, 0x10,
	0x03, 0x42, 0xc5, 0x02, 0x0a, 0x26, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69,
	0x74, 0x74, 0x65, 0x68, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x42, 0x10, 0x49, 0x6e,
	0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x63, 0x67, 0x6f, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x64, 0x65, 0x76, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x3b, 0x69,
	0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x41, 0x49, 0x58, 0xaa, 0x02, 0x21, 0x41, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x56, 0x31, 0xca,
	0x02, 0x21, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x5c, 0x49, 0x6e, 0x74,
	0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x5c, 0x56, 0x31, 0xe2, 0x02, 0x2d, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68,
	0x5c, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0xea, 0x02, 0x23, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68,
	0x3a, 0x3a, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_autokitteh_integration_registry_v1_integration_proto_rawDescOnce sync.Once
	file_autokitteh_integration_registry_v1_integration_proto_rawDescData = file_autokitteh_integration_registry_v1_integration_proto_rawDesc
)

func file_autokitteh_integration_registry_v1_integration_proto_rawDescGZIP() []byte {
	file_autokitteh_integration_registry_v1_integration_proto_rawDescOnce.Do(func() {
		file_autokitteh_integration_registry_v1_integration_proto_rawDescData = protoimpl.X.CompressGZIP(file_autokitteh_integration_registry_v1_integration_proto_rawDescData)
	})
	return file_autokitteh_integration_registry_v1_integration_proto_rawDescData
}

var file_autokitteh_integration_registry_v1_integration_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_autokitteh_integration_registry_v1_integration_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_autokitteh_integration_registry_v1_integration_proto_goTypes = []interface{}{
	(Visibility)(0),     // 0: autokitteh.integration_registry.v1.Visibility
	(*Integration)(nil), // 1: autokitteh.integration_registry.v1.Integration
}
var file_autokitteh_integration_registry_v1_integration_proto_depIdxs = []int32{
	0, // 0: autokitteh.integration_registry.v1.Integration.visibility:type_name -> autokitteh.integration_registry.v1.Visibility
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_autokitteh_integration_registry_v1_integration_proto_init() }
func file_autokitteh_integration_registry_v1_integration_proto_init() {
	if File_autokitteh_integration_registry_v1_integration_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_autokitteh_integration_registry_v1_integration_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Integration); i {
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
			RawDescriptor: file_autokitteh_integration_registry_v1_integration_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_autokitteh_integration_registry_v1_integration_proto_goTypes,
		DependencyIndexes: file_autokitteh_integration_registry_v1_integration_proto_depIdxs,
		EnumInfos:         file_autokitteh_integration_registry_v1_integration_proto_enumTypes,
		MessageInfos:      file_autokitteh_integration_registry_v1_integration_proto_msgTypes,
	}.Build()
	File_autokitteh_integration_registry_v1_integration_proto = out.File
	file_autokitteh_integration_registry_v1_integration_proto_rawDesc = nil
	file_autokitteh_integration_registry_v1_integration_proto_goTypes = nil
	file_autokitteh_integration_registry_v1_integration_proto_depIdxs = nil
}
