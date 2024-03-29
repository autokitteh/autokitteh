// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: autokitteh/module/v1/module.proto

package modulev1

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

type Module struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Functions map[string]*Function `protobuf:"bytes,1,rep,name=functions,proto3" json:"functions,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Variables map[string]*Variable `protobuf:"bytes,2,rep,name=variables,proto3" json:"variables,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Module) Reset() {
	*x = Module{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_module_v1_module_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Module) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Module) ProtoMessage() {}

func (x *Module) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_module_v1_module_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Module.ProtoReflect.Descriptor instead.
func (*Module) Descriptor() ([]byte, []int) {
	return file_autokitteh_module_v1_module_proto_rawDescGZIP(), []int{0}
}

func (x *Module) GetFunctions() map[string]*Function {
	if x != nil {
		return x.Functions
	}
	return nil
}

func (x *Module) GetVariables() map[string]*Variable {
	if x != nil {
		return x.Variables
	}
	return nil
}

type Variable struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Description string `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *Variable) Reset() {
	*x = Variable{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_module_v1_module_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Variable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Variable) ProtoMessage() {}

func (x *Variable) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_module_v1_module_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Variable.ProtoReflect.Descriptor instead.
func (*Variable) Descriptor() ([]byte, []int) {
	return file_autokitteh_module_v1_module_proto_rawDescGZIP(), []int{1}
}

func (x *Variable) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

type Function struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Description       string           `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
	DocumentationUrl  string           `protobuf:"bytes,2,opt,name=documentation_url,json=documentationUrl,proto3" json:"documentation_url,omitempty"`
	Input             []*FunctionField `protobuf:"bytes,3,rep,name=input,proto3" json:"input,omitempty"`
	Output            []*FunctionField `protobuf:"bytes,4,rep,name=output,proto3" json:"output,omitempty"`
	Examples          []*Example       `protobuf:"bytes,5,rep,name=examples,proto3" json:"examples,omitempty"`
	DeprecatedMessage string           `protobuf:"bytes,6,opt,name=deprecated_message,json=deprecatedMessage,proto3" json:"deprecated_message,omitempty"`
}

func (x *Function) Reset() {
	*x = Function{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_module_v1_module_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Function) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Function) ProtoMessage() {}

func (x *Function) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_module_v1_module_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Function.ProtoReflect.Descriptor instead.
func (*Function) Descriptor() ([]byte, []int) {
	return file_autokitteh_module_v1_module_proto_rawDescGZIP(), []int{2}
}

func (x *Function) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Function) GetDocumentationUrl() string {
	if x != nil {
		return x.DocumentationUrl
	}
	return ""
}

func (x *Function) GetInput() []*FunctionField {
	if x != nil {
		return x.Input
	}
	return nil
}

func (x *Function) GetOutput() []*FunctionField {
	if x != nil {
		return x.Output
	}
	return nil
}

func (x *Function) GetExamples() []*Example {
	if x != nil {
		return x.Examples
	}
	return nil
}

func (x *Function) GetDeprecatedMessage() string {
	if x != nil {
		return x.DeprecatedMessage
	}
	return ""
}

type FunctionField struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description  string     `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Type         string     `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"` // Flexible informative annotation, not parsed.
	Optional     bool       `protobuf:"varint,4,opt,name=optional,proto3" json:"optional,omitempty"`
	DefaultValue string     `protobuf:"bytes,5,opt,name=default_value,json=defaultValue,proto3" json:"default_value,omitempty"`
	Kwarg        bool       `protobuf:"varint,6,opt,name=kwarg,proto3" json:"kwarg,omitempty"`
	Examples     []*Example `protobuf:"bytes,7,rep,name=examples,proto3" json:"examples,omitempty"`
}

func (x *FunctionField) Reset() {
	*x = FunctionField{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_module_v1_module_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionField) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionField) ProtoMessage() {}

func (x *FunctionField) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_module_v1_module_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionField.ProtoReflect.Descriptor instead.
func (*FunctionField) Descriptor() ([]byte, []int) {
	return file_autokitteh_module_v1_module_proto_rawDescGZIP(), []int{3}
}

func (x *FunctionField) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FunctionField) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *FunctionField) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *FunctionField) GetOptional() bool {
	if x != nil {
		return x.Optional
	}
	return false
}

func (x *FunctionField) GetDefaultValue() string {
	if x != nil {
		return x.DefaultValue
	}
	return ""
}

func (x *FunctionField) GetKwarg() bool {
	if x != nil {
		return x.Kwarg
	}
	return false
}

func (x *FunctionField) GetExamples() []*Example {
	if x != nil {
		return x.Examples
	}
	return nil
}

type Example struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code        string `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
	Explanation string `protobuf:"bytes,2,opt,name=explanation,proto3" json:"explanation,omitempty"` // Optional.
}

func (x *Example) Reset() {
	*x = Example{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autokitteh_module_v1_module_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Example) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Example) ProtoMessage() {}

func (x *Example) ProtoReflect() protoreflect.Message {
	mi := &file_autokitteh_module_v1_module_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Example.ProtoReflect.Descriptor instead.
func (*Example) Descriptor() ([]byte, []int) {
	return file_autokitteh_module_v1_module_proto_rawDescGZIP(), []int{4}
}

func (x *Example) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *Example) GetExplanation() string {
	if x != nil {
		return x.Explanation
	}
	return ""
}

var File_autokitteh_module_v1_module_proto protoreflect.FileDescriptor

var file_autokitteh_module_v1_module_proto_rawDesc = []byte{
	0x0a, 0x21, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x6d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x14, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x82, 0x03, 0x0a, 0x06, 0x4d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x12, 0x5d, 0x0a, 0x09, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x75,
	0x6c, 0x65, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x42, 0x12, 0xfa, 0xf7, 0x18, 0x0e, 0x9a, 0x01, 0x0b, 0x22, 0x04, 0x72, 0x02, 0x10, 0x01,
	0x2a, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x09, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x5d, 0x0a, 0x09, 0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68,
	0x2e, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x2e, 0x56, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x42, 0x12, 0xfa, 0xf7, 0x18, 0x0e, 0x9a, 0x01, 0x0b, 0x22, 0x04, 0x72, 0x02, 0x10, 0x01, 0x2a,
	0x03, 0xc8, 0x01, 0x01, 0x52, 0x09, 0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x1a,
	0x5c, 0x0a, 0x0e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x34, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e,
	0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x5c, 0x0a,
	0x0e, 0x56, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x34, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1e, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x6d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x2c, 0x0a, 0x08, 0x56,
	0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xbb, 0x02, 0x0a, 0x08, 0x46, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2b, 0x0a, 0x11, 0x64, 0x6f, 0x63, 0x75,
	0x6d, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x10, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x55, 0x72, 0x6c, 0x12, 0x39, 0x0a, 0x05, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65,
	0x68, 0x2e, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x05, 0x69, 0x6e, 0x70, 0x75, 0x74,
	0x12, 0x3b, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x23, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x6d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x39, 0x0a,
	0x08, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x6d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x08,
	0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x12, 0x2d, 0x0a, 0x12, 0x64, 0x65, 0x70, 0x72,
	0x65, 0x63, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x64, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x65, 0x64,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0xf5, 0x01, 0x0a, 0x0d, 0x46, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x1c, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0xf7, 0x18, 0x04, 0x72, 0x02, 0x10,
	0x01, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x08, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x65, 0x66,
	0x61, 0x75, 0x6c, 0x74, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x14,
	0x0a, 0x05, 0x6b, 0x77, 0x61, 0x72, 0x67, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x6b,
	0x77, 0x61, 0x72, 0x67, 0x12, 0x39, 0x0a, 0x08, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73,
	0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x2e, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x08, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x22,
	0x3f, 0x0a, 0x07, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f,
	0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x20,
	0x0a, 0x0b, 0x65, 0x78, 0x70, 0x6c, 0x61, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x65, 0x78, 0x70, 0x6c, 0x61, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x42, 0xe2, 0x01, 0x0a, 0x18, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74,
	0x74, 0x65, 0x68, 0x2e, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x42, 0x0b, 0x4d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x47, 0x67, 0x6f,
	0x2e, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x64, 0x65, 0x76, 0x2f,
	0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74,
	0x65, 0x68, 0x2f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x6d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x41, 0x4d, 0x58, 0xaa, 0x02, 0x14, 0x41, 0x75,
	0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e,
	0x56, 0x31, 0xca, 0x02, 0x14, 0x41, 0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x5c,
	0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x20, 0x41, 0x75, 0x74, 0x6f,
	0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x5c, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x5c, 0x56, 0x31,
	0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x16, 0x41,
	0x75, 0x74, 0x6f, 0x6b, 0x69, 0x74, 0x74, 0x65, 0x68, 0x3a, 0x3a, 0x4d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_autokitteh_module_v1_module_proto_rawDescOnce sync.Once
	file_autokitteh_module_v1_module_proto_rawDescData = file_autokitteh_module_v1_module_proto_rawDesc
)

func file_autokitteh_module_v1_module_proto_rawDescGZIP() []byte {
	file_autokitteh_module_v1_module_proto_rawDescOnce.Do(func() {
		file_autokitteh_module_v1_module_proto_rawDescData = protoimpl.X.CompressGZIP(file_autokitteh_module_v1_module_proto_rawDescData)
	})
	return file_autokitteh_module_v1_module_proto_rawDescData
}

var file_autokitteh_module_v1_module_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_autokitteh_module_v1_module_proto_goTypes = []interface{}{
	(*Module)(nil),        // 0: autokitteh.module.v1.Module
	(*Variable)(nil),      // 1: autokitteh.module.v1.Variable
	(*Function)(nil),      // 2: autokitteh.module.v1.Function
	(*FunctionField)(nil), // 3: autokitteh.module.v1.FunctionField
	(*Example)(nil),       // 4: autokitteh.module.v1.Example
	nil,                   // 5: autokitteh.module.v1.Module.FunctionsEntry
	nil,                   // 6: autokitteh.module.v1.Module.VariablesEntry
}
var file_autokitteh_module_v1_module_proto_depIdxs = []int32{
	5, // 0: autokitteh.module.v1.Module.functions:type_name -> autokitteh.module.v1.Module.FunctionsEntry
	6, // 1: autokitteh.module.v1.Module.variables:type_name -> autokitteh.module.v1.Module.VariablesEntry
	3, // 2: autokitteh.module.v1.Function.input:type_name -> autokitteh.module.v1.FunctionField
	3, // 3: autokitteh.module.v1.Function.output:type_name -> autokitteh.module.v1.FunctionField
	4, // 4: autokitteh.module.v1.Function.examples:type_name -> autokitteh.module.v1.Example
	4, // 5: autokitteh.module.v1.FunctionField.examples:type_name -> autokitteh.module.v1.Example
	2, // 6: autokitteh.module.v1.Module.FunctionsEntry.value:type_name -> autokitteh.module.v1.Function
	1, // 7: autokitteh.module.v1.Module.VariablesEntry.value:type_name -> autokitteh.module.v1.Variable
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_autokitteh_module_v1_module_proto_init() }
func file_autokitteh_module_v1_module_proto_init() {
	if File_autokitteh_module_v1_module_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_autokitteh_module_v1_module_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Module); i {
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
		file_autokitteh_module_v1_module_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Variable); i {
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
		file_autokitteh_module_v1_module_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Function); i {
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
		file_autokitteh_module_v1_module_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionField); i {
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
		file_autokitteh_module_v1_module_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Example); i {
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
			RawDescriptor: file_autokitteh_module_v1_module_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_autokitteh_module_v1_module_proto_goTypes,
		DependencyIndexes: file_autokitteh_module_v1_module_proto_depIdxs,
		MessageInfos:      file_autokitteh_module_v1_module_proto_msgTypes,
	}.Build()
	File_autokitteh_module_v1_module_proto = out.File
	file_autokitteh_module_v1_module_proto_rawDesc = nil
	file_autokitteh_module_v1_module_proto_goTypes = nil
	file_autokitteh_module_v1_module_proto_depIdxs = nil
}
