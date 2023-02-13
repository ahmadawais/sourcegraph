// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: gitserver.proto

package proto

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

type ExecRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Repo           string   `protobuf:"bytes,1,opt,name=repo,proto3" json:"repo,omitempty"`
	EnsureRevision string   `protobuf:"bytes,2,opt,name=ensure_revision,json=ensureRevision,proto3" json:"ensure_revision,omitempty"`
	Args           []string `protobuf:"bytes,3,rep,name=args,proto3" json:"args,omitempty"`
	Stdin          []byte   `protobuf:"bytes,4,opt,name=stdin,proto3" json:"stdin,omitempty"`
	NoTimeout      bool     `protobuf:"varint,5,opt,name=no_timeout,json=noTimeout,proto3" json:"no_timeout,omitempty"`
}

func (x *ExecRequest) Reset() {
	*x = ExecRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitserver_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecRequest) ProtoMessage() {}

func (x *ExecRequest) ProtoReflect() protoreflect.Message {
	mi := &file_gitserver_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecRequest.ProtoReflect.Descriptor instead.
func (*ExecRequest) Descriptor() ([]byte, []int) {
	return file_gitserver_proto_rawDescGZIP(), []int{0}
}

func (x *ExecRequest) GetRepo() string {
	if x != nil {
		return x.Repo
	}
	return ""
}

func (x *ExecRequest) GetEnsureRevision() string {
	if x != nil {
		return x.EnsureRevision
	}
	return ""
}

func (x *ExecRequest) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

func (x *ExecRequest) GetStdin() []byte {
	if x != nil {
		return x.Stdin
	}
	return nil
}

func (x *ExecRequest) GetNoTimeout() bool {
	if x != nil {
		return x.NoTimeout
	}
	return false
}

type ExecResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ExecResponse) Reset() {
	*x = ExecResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitserver_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecResponse) ProtoMessage() {}

func (x *ExecResponse) ProtoReflect() protoreflect.Message {
	mi := &file_gitserver_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecResponse.ProtoReflect.Descriptor instead.
func (*ExecResponse) Descriptor() ([]byte, []int) {
	return file_gitserver_proto_rawDescGZIP(), []int{1}
}

func (x *ExecResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type NotFoundPayload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Repo            string `protobuf:"bytes,1,opt,name=repo,proto3" json:"repo,omitempty"`
	CloneInProgress bool   `protobuf:"varint,2,opt,name=clone_in_progress,json=cloneInProgress,proto3" json:"clone_in_progress,omitempty"`
	CloneProgress   string `protobuf:"bytes,3,opt,name=clone_progress,json=cloneProgress,proto3" json:"clone_progress,omitempty"`
}

func (x *NotFoundPayload) Reset() {
	*x = NotFoundPayload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitserver_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotFoundPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotFoundPayload) ProtoMessage() {}

func (x *NotFoundPayload) ProtoReflect() protoreflect.Message {
	mi := &file_gitserver_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotFoundPayload.ProtoReflect.Descriptor instead.
func (*NotFoundPayload) Descriptor() ([]byte, []int) {
	return file_gitserver_proto_rawDescGZIP(), []int{2}
}

func (x *NotFoundPayload) GetRepo() string {
	if x != nil {
		return x.Repo
	}
	return ""
}

func (x *NotFoundPayload) GetCloneInProgress() bool {
	if x != nil {
		return x.CloneInProgress
	}
	return false
}

func (x *NotFoundPayload) GetCloneProgress() string {
	if x != nil {
		return x.CloneProgress
	}
	return ""
}

type ExecStatusPayload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StatusCode int32  `protobuf:"varint,1,opt,name=status_code,json=statusCode,proto3" json:"status_code,omitempty"`
	Stderr     string `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
}

func (x *ExecStatusPayload) Reset() {
	*x = ExecStatusPayload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_gitserver_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecStatusPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecStatusPayload) ProtoMessage() {}

func (x *ExecStatusPayload) ProtoReflect() protoreflect.Message {
	mi := &file_gitserver_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecStatusPayload.ProtoReflect.Descriptor instead.
func (*ExecStatusPayload) Descriptor() ([]byte, []int) {
	return file_gitserver_proto_rawDescGZIP(), []int{3}
}

func (x *ExecStatusPayload) GetStatusCode() int32 {
	if x != nil {
		return x.StatusCode
	}
	return 0
}

func (x *ExecStatusPayload) GetStderr() string {
	if x != nil {
		return x.Stderr
	}
	return ""
}

var File_gitserver_proto protoreflect.FileDescriptor

var file_gitserver_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x67, 0x69, 0x74, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x09, 0x67, 0x69, 0x74, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x22, 0x93, 0x01, 0x0a,
	0x0b, 0x45, 0x78, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x72, 0x65, 0x70, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72, 0x65, 0x70, 0x6f,
	0x12, 0x27, 0x0a, 0x0f, 0x65, 0x6e, 0x73, 0x75, 0x72, 0x65, 0x5f, 0x72, 0x65, 0x76, 0x69, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x65, 0x6e, 0x73, 0x75, 0x72,
	0x65, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x12, 0x14, 0x0a,
	0x05, 0x73, 0x74, 0x64, 0x69, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x73, 0x74,
	0x64, 0x69, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x6e, 0x6f, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75,
	0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x6e, 0x6f, 0x54, 0x69, 0x6d, 0x65, 0x6f,
	0x75, 0x74, 0x22, 0x22, 0x0a, 0x0c, 0x45, 0x78, 0x65, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x78, 0x0a, 0x0f, 0x4e, 0x6f, 0x74, 0x46, 0x6f, 0x75,
	0x6e, 0x64, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x65, 0x70,
	0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x12, 0x2a, 0x0a,
	0x11, 0x63, 0x6c, 0x6f, 0x6e, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65,
	0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x63, 0x6c, 0x6f, 0x6e, 0x65, 0x49,
	0x6e, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x6c, 0x6f,
	0x6e, 0x65, 0x5f, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x63, 0x6c, 0x6f, 0x6e, 0x65, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73,
	0x22, 0x4c, 0x0a, 0x11, 0x45, 0x78, 0x65, 0x63, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x50, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f,
	0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x64, 0x65, 0x72, 0x72,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x64, 0x65, 0x72, 0x72, 0x32, 0x4f,
	0x0a, 0x10, 0x47, 0x69, 0x74, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x3b, 0x0a, 0x04, 0x45, 0x78, 0x65, 0x63, 0x12, 0x16, 0x2e, 0x67, 0x69, 0x74,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x17, 0x2e, 0x67, 0x69, 0x74, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x45,
	0x78, 0x65, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42,
	0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67,
	0x69, 0x74, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_gitserver_proto_rawDescOnce sync.Once
	file_gitserver_proto_rawDescData = file_gitserver_proto_rawDesc
)

func file_gitserver_proto_rawDescGZIP() []byte {
	file_gitserver_proto_rawDescOnce.Do(func() {
		file_gitserver_proto_rawDescData = protoimpl.X.CompressGZIP(file_gitserver_proto_rawDescData)
	})
	return file_gitserver_proto_rawDescData
}

var file_gitserver_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_gitserver_proto_goTypes = []interface{}{
	(*ExecRequest)(nil),       // 0: gitserver.ExecRequest
	(*ExecResponse)(nil),      // 1: gitserver.ExecResponse
	(*NotFoundPayload)(nil),   // 2: gitserver.NotFoundPayload
	(*ExecStatusPayload)(nil), // 3: gitserver.ExecStatusPayload
}
var file_gitserver_proto_depIdxs = []int32{
	0, // 0: gitserver.GitserverService.Exec:input_type -> gitserver.ExecRequest
	1, // 1: gitserver.GitserverService.Exec:output_type -> gitserver.ExecResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_gitserver_proto_init() }
func file_gitserver_proto_init() {
	if File_gitserver_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_gitserver_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecRequest); i {
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
		file_gitserver_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecResponse); i {
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
		file_gitserver_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotFoundPayload); i {
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
		file_gitserver_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecStatusPayload); i {
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
			RawDescriptor: file_gitserver_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_gitserver_proto_goTypes,
		DependencyIndexes: file_gitserver_proto_depIdxs,
		MessageInfos:      file_gitserver_proto_msgTypes,
	}.Build()
	File_gitserver_proto = out.File
	file_gitserver_proto_rawDesc = nil
	file_gitserver_proto_goTypes = nil
	file_gitserver_proto_depIdxs = nil
}
