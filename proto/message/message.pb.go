// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: proto/message/message.proto

package pb

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

type ElementMessageType int32

const (
	ElementMessageType_UNKNOWN        ElementMessageType = 0
	ElementMessageType_ERROR          ElementMessageType = 1
	ElementMessageType_CHAT_MESSAGE   ElementMessageType = 2
	ElementMessageType_PLAY           ElementMessageType = 3
	ElementMessageType_PAUSE          ElementMessageType = 4
	ElementMessageType_CHECK_SEEK     ElementMessageType = 5
	ElementMessageType_TOO_FAST       ElementMessageType = 6
	ElementMessageType_TOO_SLOW       ElementMessageType = 7
	ElementMessageType_CHANGE_RATE    ElementMessageType = 8
	ElementMessageType_CHANGE_SEEK    ElementMessageType = 9
	ElementMessageType_CHANGE_CURRENT ElementMessageType = 10
	ElementMessageType_CHANGE_MOVIES  ElementMessageType = 11
	ElementMessageType_CHANGE_PEOPLE  ElementMessageType = 12
)

// Enum value maps for ElementMessageType.
var (
	ElementMessageType_name = map[int32]string{
		0:  "UNKNOWN",
		1:  "ERROR",
		2:  "CHAT_MESSAGE",
		3:  "PLAY",
		4:  "PAUSE",
		5:  "CHECK_SEEK",
		6:  "TOO_FAST",
		7:  "TOO_SLOW",
		8:  "CHANGE_RATE",
		9:  "CHANGE_SEEK",
		10: "CHANGE_CURRENT",
		11: "CHANGE_MOVIES",
		12: "CHANGE_PEOPLE",
	}
	ElementMessageType_value = map[string]int32{
		"UNKNOWN":        0,
		"ERROR":          1,
		"CHAT_MESSAGE":   2,
		"PLAY":           3,
		"PAUSE":          4,
		"CHECK_SEEK":     5,
		"TOO_FAST":       6,
		"TOO_SLOW":       7,
		"CHANGE_RATE":    8,
		"CHANGE_SEEK":    9,
		"CHANGE_CURRENT": 10,
		"CHANGE_MOVIES":  11,
		"CHANGE_PEOPLE":  12,
	}
)

func (x ElementMessageType) Enum() *ElementMessageType {
	p := new(ElementMessageType)
	*p = x
	return p
}

func (x ElementMessageType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ElementMessageType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_message_message_proto_enumTypes[0].Descriptor()
}

func (ElementMessageType) Type() protoreflect.EnumType {
	return &file_proto_message_message_proto_enumTypes[0]
}

func (x ElementMessageType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ElementMessageType.Descriptor instead.
func (ElementMessageType) EnumDescriptor() ([]byte, []int) {
	return file_proto_message_message_proto_rawDescGZIP(), []int{0}
}

type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Seek    float64 `protobuf:"fixed64,1,opt,name=seek,proto3" json:"seek,omitempty"`
	Rate    float64 `protobuf:"fixed64,2,opt,name=rate,proto3" json:"rate,omitempty"`
	Playing bool    `protobuf:"varint,3,opt,name=playing,proto3" json:"playing,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_message_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_proto_message_message_proto_msgTypes[0]
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
	return file_proto_message_message_proto_rawDescGZIP(), []int{0}
}

func (x *Status) GetSeek() float64 {
	if x != nil {
		return x.Seek
	}
	return 0
}

func (x *Status) GetRate() float64 {
	if x != nil {
		return x.Rate
	}
	return 0
}

func (x *Status) GetPlaying() bool {
	if x != nil {
		return x.Playing
	}
	return false
}

type ElementMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type      ElementMessageType `protobuf:"varint,1,opt,name=type,proto3,enum=proto.ElementMessageType" json:"type,omitempty"`
	Sender    string             `protobuf:"bytes,2,opt,name=sender,proto3" json:"sender,omitempty"`
	Message   string             `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	Rate      float64            `protobuf:"fixed64,4,opt,name=rate,proto3" json:"rate,omitempty"`
	Seek      float64            `protobuf:"fixed64,5,opt,name=seek,proto3" json:"seek,omitempty"`
	PeopleNum int64              `protobuf:"varint,6,opt,name=peopleNum,proto3" json:"peopleNum,omitempty"`
	Time      int64              `protobuf:"varint,7,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *ElementMessage) Reset() {
	*x = ElementMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_message_message_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ElementMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ElementMessage) ProtoMessage() {}

func (x *ElementMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_message_message_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ElementMessage.ProtoReflect.Descriptor instead.
func (*ElementMessage) Descriptor() ([]byte, []int) {
	return file_proto_message_message_proto_rawDescGZIP(), []int{1}
}

func (x *ElementMessage) GetType() ElementMessageType {
	if x != nil {
		return x.Type
	}
	return ElementMessageType_UNKNOWN
}

func (x *ElementMessage) GetSender() string {
	if x != nil {
		return x.Sender
	}
	return ""
}

func (x *ElementMessage) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *ElementMessage) GetRate() float64 {
	if x != nil {
		return x.Rate
	}
	return 0
}

func (x *ElementMessage) GetSeek() float64 {
	if x != nil {
		return x.Seek
	}
	return 0
}

func (x *ElementMessage) GetPeopleNum() int64 {
	if x != nil {
		return x.PeopleNum
	}
	return 0
}

func (x *ElementMessage) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

var File_proto_message_message_proto protoreflect.FileDescriptor

var file_proto_message_message_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4a, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12,
	0x0a, 0x04, 0x73, 0x65, 0x65, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x73, 0x65,
	0x65, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01,
	0x52, 0x04, 0x72, 0x61, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x69, 0x6e,
	0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x69, 0x6e, 0x67,
	0x22, 0xcb, 0x01, 0x0a, 0x0e, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x2d, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e,
	0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x04, 0x72, 0x61, 0x74, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x65, 0x65, 0x6b,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x73, 0x65, 0x65, 0x6b, 0x12, 0x1c, 0x0a, 0x09,
	0x70, 0x65, 0x6f, 0x70, 0x6c, 0x65, 0x4e, 0x75, 0x6d, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x09, 0x70, 0x65, 0x6f, 0x70, 0x6c, 0x65, 0x4e, 0x75, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x2a, 0xdb,
	0x01, 0x0a, 0x12, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e,
	0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x01, 0x12, 0x10, 0x0a,
	0x0c, 0x43, 0x48, 0x41, 0x54, 0x5f, 0x4d, 0x45, 0x53, 0x53, 0x41, 0x47, 0x45, 0x10, 0x02, 0x12,
	0x08, 0x0a, 0x04, 0x50, 0x4c, 0x41, 0x59, 0x10, 0x03, 0x12, 0x09, 0x0a, 0x05, 0x50, 0x41, 0x55,
	0x53, 0x45, 0x10, 0x04, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x48, 0x45, 0x43, 0x4b, 0x5f, 0x53, 0x45,
	0x45, 0x4b, 0x10, 0x05, 0x12, 0x0c, 0x0a, 0x08, 0x54, 0x4f, 0x4f, 0x5f, 0x46, 0x41, 0x53, 0x54,
	0x10, 0x06, 0x12, 0x0c, 0x0a, 0x08, 0x54, 0x4f, 0x4f, 0x5f, 0x53, 0x4c, 0x4f, 0x57, 0x10, 0x07,
	0x12, 0x0f, 0x0a, 0x0b, 0x43, 0x48, 0x41, 0x4e, 0x47, 0x45, 0x5f, 0x52, 0x41, 0x54, 0x45, 0x10,
	0x08, 0x12, 0x0f, 0x0a, 0x0b, 0x43, 0x48, 0x41, 0x4e, 0x47, 0x45, 0x5f, 0x53, 0x45, 0x45, 0x4b,
	0x10, 0x09, 0x12, 0x12, 0x0a, 0x0e, 0x43, 0x48, 0x41, 0x4e, 0x47, 0x45, 0x5f, 0x43, 0x55, 0x52,
	0x52, 0x45, 0x4e, 0x54, 0x10, 0x0a, 0x12, 0x11, 0x0a, 0x0d, 0x43, 0x48, 0x41, 0x4e, 0x47, 0x45,
	0x5f, 0x4d, 0x4f, 0x56, 0x49, 0x45, 0x53, 0x10, 0x0b, 0x12, 0x11, 0x0a, 0x0d, 0x43, 0x48, 0x41,
	0x4e, 0x47, 0x45, 0x5f, 0x50, 0x45, 0x4f, 0x50, 0x4c, 0x45, 0x10, 0x0c, 0x42, 0x06, 0x5a, 0x04,
	0x2e, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_message_message_proto_rawDescOnce sync.Once
	file_proto_message_message_proto_rawDescData = file_proto_message_message_proto_rawDesc
)

func file_proto_message_message_proto_rawDescGZIP() []byte {
	file_proto_message_message_proto_rawDescOnce.Do(func() {
		file_proto_message_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_message_message_proto_rawDescData)
	})
	return file_proto_message_message_proto_rawDescData
}

var file_proto_message_message_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_message_message_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_message_message_proto_goTypes = []interface{}{
	(ElementMessageType)(0), // 0: proto.ElementMessageType
	(*Status)(nil),          // 1: proto.Status
	(*ElementMessage)(nil),  // 2: proto.ElementMessage
}
var file_proto_message_message_proto_depIdxs = []int32{
	0, // 0: proto.ElementMessage.type:type_name -> proto.ElementMessageType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_message_message_proto_init() }
func file_proto_message_message_proto_init() {
	if File_proto_message_message_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_message_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_proto_message_message_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ElementMessage); i {
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
			RawDescriptor: file_proto_message_message_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_message_message_proto_goTypes,
		DependencyIndexes: file_proto_message_message_proto_depIdxs,
		EnumInfos:         file_proto_message_message_proto_enumTypes,
		MessageInfos:      file_proto_message_message_proto_msgTypes,
	}.Build()
	File_proto_message_message_proto = out.File
	file_proto_message_message_proto_rawDesc = nil
	file_proto_message_message_proto_goTypes = nil
	file_proto_message_message_proto_depIdxs = nil
}
