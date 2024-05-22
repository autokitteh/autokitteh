// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: autokitteh/common/v1/status.proto

package commonv1

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

type Status_Code int32

const (
	Status_CODE_UNSPECIFIED Status_Code = 0
	Status_CODE_OK          Status_Code = 1
	Status_CODE_WARNING     Status_Code = 2
	Status_CODE_ERROR       Status_Code = 3
)

// Enum value maps for Status_Code.
var (
	Status_Code_name = map[int32]string{
		0: "CODE_UNSPECIFIED",
		1: "CODE_OK",
		2: "CODE_WARNING",
		3: "CODE_ERROR",
	}
	Status_Code_value = map[string]int32{
		"CODE_UNSPECIFIED": 0,
		"CODE_OK":          1,
		"CODE_WARNING":     2,
		"CODE_ERROR":       3,
	}
)

func (x Status_Code) Enum() *Status_Code {
	p := new(Status_Code)
	*p = x
	return p
}

func (x Status_Code) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status_Code) Descriptor() protoreflect.EnumDescriptor {
	return file_autokitteh_common_v1_status_proto_enumTypes[0].Descriptor()
}

func (Status_Code) Type() protoreflect.EnumType {
	return &file_autokitteh_common_v1_status_proto_enumTypes[0]
}

func (x Status_Code) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status_Code.Descriptor instead.
func (Status_Code) EnumDescriptor() ([]byte, []int) {
	return file_autokitteh_common_v1_status_proto_rawDescGZIP(), []int{0, 0}
}

type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code    Status_Code `protobuf:"varint,1,opt,name=code,proto3,enum=autokitteh.common.v1.Status_Code" json:"code,omitempty"`
	Message string      `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_common_v1_status_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_common_v1_status_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_autokitteh_common_v1_status_proto_rawDescGZIP(), []int{0}
}

func (x *Status) GetCode() Status_Code {
	if x != nil {
		return x.Code
	}
	return Status_CODE_UNSPECIFIED
}

func (x *Status) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_autokitteh_common_v1_status_proto protoreflect.FileDescriptor

var file_autokitteh_common_v1_status_proto_rawDesc = []byte{
	0x0a, 0x21, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x14, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x22, 0xa6, 0x01, 0x0a, 0x06, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x35, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x21, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x2e, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x4b, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x14, 0x0a,
	0x10, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45,
	0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x4f, 0x4b, 0x10, 0x01,
	0x12, 0x10, 0x0a, 0x0c, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x57, 0x41, 0x52, 0x4e, 0x49, 0x4e, 0x47,
	0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52,
	0x10, 0x03, 0x42, 0xe2, 0x01, 0x0a, 0x18, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b,
	0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x42,
	0x0b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x47,
	0x67, 0x6f, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65,
	0x76, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69,
	0x74, 0x74, 0x65, 0x68, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x3b, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x41, 0x43, 0x58, 0xaa, 0x02, 0x14,
	0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x14, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x5c, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x20, 0x41, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x5c, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5c,
	0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x16, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x3a, 0x3a, 0x43, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_autokitteh_common_v1_status_proto_rawDescOnce sync.Once
	file_autokitteh_common_v1_status_proto_rawDescData = file_autokitteh_common_v1_status_proto_rawDesc
)

func file_autokitteh_common_v1_status_proto_rawDescGZIP() []byte {
	file_autokitteh_common_v1_status_proto_rawDescOnce.Do(func() {
		file_autokitteh_common_v1_status_proto_rawDescData = protoimpl.X.CompressGZIP(file_autokitteh_common_v1_status_proto_rawDescData)
	})
	return file_autokitteh_common_v1_status_proto_rawDescData
}

var file_autokitteh_common_v1_status_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_autokitteh_common_v1_status_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_autokitteh_common_v1_status_proto_goTypes = []interface{}{
	(Status_Code)(0), // 0: autokitteh.common.v1.Status.Code
	(*Status)(nil),   // 1: autokitteh.common.v1.Status
}
var file_autokitteh_common_v1_status_proto_depIdxs = []int32{
	0, // 0: autokitteh.common.v1.Status.code:type_name -> autokitteh.common.v1.Status.Code
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_autokitteh_common_v1_status_proto_init() }
func file_autokitteh_common_v1_status_proto_init() {
	if File_autokitteh_common_v1_status_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_autokitteh_common_v1_status_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
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
			RawDescriptor: file_autokitteh_common_v1_status_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_autokitteh_common_v1_status_proto_goTypes,
		DependencyIndexes: file_autokitteh_common_v1_status_proto_depIdxs,
		EnumInfos:         file_autokitteh_common_v1_status_proto_enumTypes,
		MessageInfos:      file_autokitteh_common_v1_status_proto_msgTypes,
	}.Build()
	File_autokitteh_common_v1_status_proto = out.File
	file_autokitteh_common_v1_status_proto_rawDesc = nil
	file_autokitteh_common_v1_status_proto_goTypes = nil
	file_autokitteh_common_v1_status_proto_depIdxs = nil
}
