// Code generated by protoc-gen-go. DO NOT EDIT.
// source: playback.proto

/*
Package PlaybackProtocol is a generated protocol buffer package.

It is generated from these files:
	playback.proto

It has these top-level messages:
	Channel
	OpenRequest
	OpenReply
	StreamRequest
	StreamReply
	CloseRequest
	CloseReply
	SeekRequest
	SeekReply
*/
package PlaybackProtocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Channel struct {
	Device string `protobuf:"bytes,1,opt,name=device" json:"device,omitempty"`
	Ch     int32  `protobuf:"varint,2,opt,name=ch" json:"ch,omitempty"`
}

func (m *Channel) Reset()                    { *m = Channel{} }
func (m *Channel) String() string            { return proto.CompactTextString(m) }
func (*Channel) ProtoMessage()               {}
func (*Channel) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Channel) GetDevice() string {
	if m != nil {
		return m.Device
	}
	return ""
}

func (m *Channel) GetCh() int32 {
	if m != nil {
		return m.Ch
	}
	return 0
}

type OpenRequest struct {
	Channels   []*Channel `protobuf:"bytes,1,rep,name=channels" json:"channels,omitempty"`
	BeginTime  uint64     `protobuf:"varint,2,opt,name=beginTime" json:"beginTime,omitempty"`
	EndTime    uint64     `protobuf:"varint,3,opt,name=endTime" json:"endTime,omitempty"`
	StreamType uint32     `protobuf:"varint,4,opt,name=streamType" json:"streamType,omitempty"`
	RecordType uint32     `protobuf:"varint,5,opt,name=recordType" json:"recordType,omitempty"`
	OrAnd      int32      `protobuf:"varint,6,opt,name=orAnd" json:"orAnd,omitempty"`
}

func (m *OpenRequest) Reset()                    { *m = OpenRequest{} }
func (m *OpenRequest) String() string            { return proto.CompactTextString(m) }
func (*OpenRequest) ProtoMessage()               {}
func (*OpenRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *OpenRequest) GetChannels() []*Channel {
	if m != nil {
		return m.Channels
	}
	return nil
}

func (m *OpenRequest) GetBeginTime() uint64 {
	if m != nil {
		return m.BeginTime
	}
	return 0
}

func (m *OpenRequest) GetEndTime() uint64 {
	if m != nil {
		return m.EndTime
	}
	return 0
}

func (m *OpenRequest) GetStreamType() uint32 {
	if m != nil {
		return m.StreamType
	}
	return 0
}

func (m *OpenRequest) GetRecordType() uint32 {
	if m != nil {
		return m.RecordType
	}
	return 0
}

func (m *OpenRequest) GetOrAnd() int32 {
	if m != nil {
		return m.OrAnd
	}
	return 0
}

type OpenReply struct {
	Error  string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
	Handle uint64 `protobuf:"varint,2,opt,name=handle" json:"handle,omitempty"`
}

func (m *OpenReply) Reset()                    { *m = OpenReply{} }
func (m *OpenReply) String() string            { return proto.CompactTextString(m) }
func (*OpenReply) ProtoMessage()               {}
func (*OpenReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *OpenReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *OpenReply) GetHandle() uint64 {
	if m != nil {
		return m.Handle
	}
	return 0
}

type StreamRequest struct {
	Handle uint64 `protobuf:"varint,1,opt,name=handle" json:"handle,omitempty"`
}

func (m *StreamRequest) Reset()                    { *m = StreamRequest{} }
func (m *StreamRequest) String() string            { return proto.CompactTextString(m) }
func (*StreamRequest) ProtoMessage()               {}
func (*StreamRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *StreamRequest) GetHandle() uint64 {
	if m != nil {
		return m.Handle
	}
	return 0
}

type StreamReply struct {
	Error    string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
	DataSize int32  `protobuf:"varint,3,opt,name=dataSize" json:"dataSize,omitempty"`
	Data     []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	Device   string `protobuf:"bytes,5,opt,name=device" json:"device,omitempty"`
	Ch       int32  `protobuf:"varint,6,opt,name=ch" json:"ch,omitempty"`
	Pid      uint64 `protobuf:"varint,7,opt,name=pid" json:"pid,omitempty"`
	Checksum string `protobuf:"bytes,8,opt,name=checksum" json:"checksum,omitempty"`
}

func (m *StreamReply) Reset()                    { *m = StreamReply{} }
func (m *StreamReply) String() string            { return proto.CompactTextString(m) }
func (*StreamReply) ProtoMessage()               {}
func (*StreamReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *StreamReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *StreamReply) GetDataSize() int32 {
	if m != nil {
		return m.DataSize
	}
	return 0
}

func (m *StreamReply) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *StreamReply) GetDevice() string {
	if m != nil {
		return m.Device
	}
	return ""
}

func (m *StreamReply) GetCh() int32 {
	if m != nil {
		return m.Ch
	}
	return 0
}

func (m *StreamReply) GetPid() uint64 {
	if m != nil {
		return m.Pid
	}
	return 0
}

func (m *StreamReply) GetChecksum() string {
	if m != nil {
		return m.Checksum
	}
	return ""
}

type CloseRequest struct {
	Handle uint64 `protobuf:"varint,1,opt,name=handle" json:"handle,omitempty"`
}

func (m *CloseRequest) Reset()                    { *m = CloseRequest{} }
func (m *CloseRequest) String() string            { return proto.CompactTextString(m) }
func (*CloseRequest) ProtoMessage()               {}
func (*CloseRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *CloseRequest) GetHandle() uint64 {
	if m != nil {
		return m.Handle
	}
	return 0
}

type CloseReply struct {
	Error string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
}

func (m *CloseReply) Reset()                    { *m = CloseReply{} }
func (m *CloseReply) String() string            { return proto.CompactTextString(m) }
func (*CloseReply) ProtoMessage()               {}
func (*CloseReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *CloseReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

type SeekRequest struct {
	Handle     uint64 `protobuf:"varint,1,opt,name=handle" json:"handle,omitempty"`
	SeekTime   uint64 `protobuf:"varint,2,opt,name=seekTime" json:"seekTime,omitempty"`
	StreamType uint32 `protobuf:"varint,3,opt,name=streamType" json:"streamType,omitempty"`
	RecordType uint32 `protobuf:"varint,4,opt,name=recordType" json:"recordType,omitempty"`
	OrAnd      int32  `protobuf:"varint,5,opt,name=orAnd" json:"orAnd,omitempty"`
}

func (m *SeekRequest) Reset()                    { *m = SeekRequest{} }
func (m *SeekRequest) String() string            { return proto.CompactTextString(m) }
func (*SeekRequest) ProtoMessage()               {}
func (*SeekRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *SeekRequest) GetHandle() uint64 {
	if m != nil {
		return m.Handle
	}
	return 0
}

func (m *SeekRequest) GetSeekTime() uint64 {
	if m != nil {
		return m.SeekTime
	}
	return 0
}

func (m *SeekRequest) GetStreamType() uint32 {
	if m != nil {
		return m.StreamType
	}
	return 0
}

func (m *SeekRequest) GetRecordType() uint32 {
	if m != nil {
		return m.RecordType
	}
	return 0
}

func (m *SeekRequest) GetOrAnd() int32 {
	if m != nil {
		return m.OrAnd
	}
	return 0
}

type SeekReply struct {
	Error string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
}

func (m *SeekReply) Reset()                    { *m = SeekReply{} }
func (m *SeekReply) String() string            { return proto.CompactTextString(m) }
func (*SeekReply) ProtoMessage()               {}
func (*SeekReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *SeekReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*Channel)(nil), "PlaybackProtocol.Channel")
	proto.RegisterType((*OpenRequest)(nil), "PlaybackProtocol.OpenRequest")
	proto.RegisterType((*OpenReply)(nil), "PlaybackProtocol.OpenReply")
	proto.RegisterType((*StreamRequest)(nil), "PlaybackProtocol.StreamRequest")
	proto.RegisterType((*StreamReply)(nil), "PlaybackProtocol.StreamReply")
	proto.RegisterType((*CloseRequest)(nil), "PlaybackProtocol.CloseRequest")
	proto.RegisterType((*CloseReply)(nil), "PlaybackProtocol.CloseReply")
	proto.RegisterType((*SeekRequest)(nil), "PlaybackProtocol.SeekRequest")
	proto.RegisterType((*SeekReply)(nil), "PlaybackProtocol.SeekReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Playback service

type PlaybackClient interface {
	Open(ctx context.Context, in *OpenRequest, opts ...grpc.CallOption) (*OpenReply, error)
	Stream(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (Playback_StreamClient, error)
	Close(ctx context.Context, in *CloseRequest, opts ...grpc.CallOption) (*CloseReply, error)
	Seek(ctx context.Context, in *SeekRequest, opts ...grpc.CallOption) (*SeekReply, error)
}

type playbackClient struct {
	cc *grpc.ClientConn
}

func NewPlaybackClient(cc *grpc.ClientConn) PlaybackClient {
	return &playbackClient{cc}
}

func (c *playbackClient) Open(ctx context.Context, in *OpenRequest, opts ...grpc.CallOption) (*OpenReply, error) {
	out := new(OpenReply)
	err := grpc.Invoke(ctx, "/PlaybackProtocol.Playback/open", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *playbackClient) Stream(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (Playback_StreamClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Playback_serviceDesc.Streams[0], c.cc, "/PlaybackProtocol.Playback/stream", opts...)
	if err != nil {
		return nil, err
	}
	x := &playbackStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Playback_StreamClient interface {
	Recv() (*StreamReply, error)
	grpc.ClientStream
}

type playbackStreamClient struct {
	grpc.ClientStream
}

func (x *playbackStreamClient) Recv() (*StreamReply, error) {
	m := new(StreamReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *playbackClient) Close(ctx context.Context, in *CloseRequest, opts ...grpc.CallOption) (*CloseReply, error) {
	out := new(CloseReply)
	err := grpc.Invoke(ctx, "/PlaybackProtocol.Playback/close", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *playbackClient) Seek(ctx context.Context, in *SeekRequest, opts ...grpc.CallOption) (*SeekReply, error) {
	out := new(SeekReply)
	err := grpc.Invoke(ctx, "/PlaybackProtocol.Playback/seek", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Playback service

type PlaybackServer interface {
	Open(context.Context, *OpenRequest) (*OpenReply, error)
	Stream(*StreamRequest, Playback_StreamServer) error
	Close(context.Context, *CloseRequest) (*CloseReply, error)
	Seek(context.Context, *SeekRequest) (*SeekReply, error)
}

func RegisterPlaybackServer(s *grpc.Server, srv PlaybackServer) {
	s.RegisterService(&_Playback_serviceDesc, srv)
}

func _Playback_Open_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OpenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PlaybackServer).Open(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/PlaybackProtocol.Playback/Open",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PlaybackServer).Open(ctx, req.(*OpenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Playback_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PlaybackServer).Stream(m, &playbackStreamServer{stream})
}

type Playback_StreamServer interface {
	Send(*StreamReply) error
	grpc.ServerStream
}

type playbackStreamServer struct {
	grpc.ServerStream
}

func (x *playbackStreamServer) Send(m *StreamReply) error {
	return x.ServerStream.SendMsg(m)
}

func _Playback_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PlaybackServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/PlaybackProtocol.Playback/Close",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PlaybackServer).Close(ctx, req.(*CloseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Playback_Seek_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SeekRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PlaybackServer).Seek(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/PlaybackProtocol.Playback/Seek",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PlaybackServer).Seek(ctx, req.(*SeekRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Playback_serviceDesc = grpc.ServiceDesc{
	ServiceName: "PlaybackProtocol.Playback",
	HandlerType: (*PlaybackServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "open",
			Handler:    _Playback_Open_Handler,
		},
		{
			MethodName: "close",
			Handler:    _Playback_Close_Handler,
		},
		{
			MethodName: "seek",
			Handler:    _Playback_Seek_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "stream",
			Handler:       _Playback_Stream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "playback.proto",
}

func init() { proto.RegisterFile("playback.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 470 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x54, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0xed, 0x26, 0xb6, 0x93, 0x4c, 0xda, 0xaa, 0x1a, 0x21, 0x64, 0x42, 0x5b, 0xc2, 0x1e, 0x20,
	0xa7, 0x08, 0x8a, 0x38, 0x70, 0x44, 0xad, 0xc4, 0x05, 0x89, 0xca, 0xed, 0x0f, 0x38, 0xeb, 0x11,
	0xb6, 0xe2, 0x78, 0x97, 0x75, 0x8a, 0x14, 0x7e, 0x84, 0x6f, 0xe0, 0xc0, 0xcf, 0xf0, 0x45, 0x68,
	0x77, 0x6d, 0xc7, 0x6d, 0x62, 0xe5, 0xb6, 0x6f, 0xe7, 0x79, 0x66, 0xde, 0xbc, 0x59, 0xc3, 0xa9,
	0xca, 0xe3, 0xcd, 0x22, 0x16, 0xcb, 0xb9, 0xd2, 0x72, 0x2d, 0xf1, 0xec, 0xb6, 0xc2, 0xb7, 0x06,
	0x0a, 0x99, 0xf3, 0xf7, 0x30, 0xb8, 0x4e, 0xe3, 0xa2, 0xa0, 0x1c, 0x9f, 0x43, 0x90, 0xd0, 0xcf,
	0x4c, 0x50, 0xc8, 0xa6, 0x6c, 0x36, 0x8a, 0x2a, 0x84, 0xa7, 0xd0, 0x13, 0x69, 0xd8, 0x9b, 0xb2,
	0x99, 0x1f, 0xf5, 0x44, 0xca, 0xff, 0x31, 0x18, 0x7f, 0x53, 0x54, 0x44, 0xf4, 0xe3, 0x81, 0xca,
	0x35, 0x7e, 0x84, 0xa1, 0x70, 0x29, 0xca, 0x90, 0x4d, 0xfb, 0xb3, 0xf1, 0xd5, 0x8b, 0xf9, 0xd3,
	0x3a, 0xf3, 0xaa, 0x48, 0xd4, 0x50, 0xf1, 0x1c, 0x46, 0x0b, 0xfa, 0x9e, 0x15, 0xf7, 0xd9, 0x8a,
	0x6c, 0x76, 0x2f, 0xda, 0x5e, 0x60, 0x08, 0x03, 0x2a, 0x12, 0x1b, 0xeb, 0xdb, 0x58, 0x0d, 0xf1,
	0x12, 0xa0, 0x5c, 0x6b, 0x8a, 0x57, 0xf7, 0x1b, 0x45, 0xa1, 0x37, 0x65, 0xb3, 0x93, 0xa8, 0x75,
	0x63, 0xe2, 0x9a, 0x84, 0xd4, 0x89, 0x8d, 0xfb, 0x2e, 0xbe, 0xbd, 0xc1, 0x67, 0xe0, 0x4b, 0xfd,
	0xb9, 0x48, 0xc2, 0xc0, 0x2a, 0x72, 0x80, 0x7f, 0x82, 0x91, 0xd3, 0xa4, 0xf2, 0x8d, 0xa1, 0x90,
	0xd6, 0x52, 0x57, 0x83, 0x70, 0xc0, 0xcc, 0x27, 0x8d, 0x8b, 0x24, 0xaf, 0xbb, 0xad, 0x10, 0x7f,
	0x0b, 0x27, 0x77, 0xb6, 0x7c, 0x3d, 0x90, 0x2d, 0x91, 0x3d, 0x22, 0xfe, 0x61, 0x30, 0xae, 0x99,
	0xdd, 0x65, 0x26, 0x30, 0x4c, 0xe2, 0x75, 0x7c, 0x97, 0xfd, 0x72, 0xd2, 0xfd, 0xa8, 0xc1, 0x88,
	0xe0, 0x99, 0xb3, 0x55, 0x7d, 0x1c, 0xd9, 0x73, 0xcb, 0x36, 0x7f, 0x8f, 0x6d, 0x41, 0x6d, 0x1b,
	0x9e, 0x41, 0x5f, 0x65, 0x49, 0x38, 0xb0, 0x2d, 0x99, 0xa3, 0xa9, 0x24, 0x52, 0x12, 0xcb, 0xf2,
	0x61, 0x15, 0x0e, 0xed, 0xb7, 0x0d, 0xe6, 0x6f, 0xe0, 0xf8, 0x3a, 0x97, 0x25, 0x1d, 0xd2, 0xc4,
	0x01, 0x2a, 0x5e, 0xa7, 0x22, 0xfe, 0xdb, 0xe8, 0x26, 0x5a, 0x1e, 0xc8, 0x65, 0xfa, 0x29, 0x89,
	0x96, 0xad, 0x85, 0x68, 0xf0, 0x13, 0xd7, 0xfb, 0x07, 0x5c, 0xf7, 0xba, 0x5d, 0xf7, 0xdb, 0xae,
	0xbf, 0x86, 0x91, 0x6b, 0xac, 0xb3, 0xf9, 0xab, 0xbf, 0x3d, 0x18, 0xd6, 0xdb, 0x8c, 0x37, 0xe0,
	0x49, 0x45, 0x05, 0x5e, 0xec, 0x2e, 0x78, 0xeb, 0x45, 0x4c, 0x5e, 0x76, 0x85, 0x55, 0xbe, 0xe1,
	0x47, 0xf8, 0x15, 0x02, 0xd7, 0x39, 0xbe, 0xda, 0x25, 0x3e, 0x5a, 0xa5, 0xc9, 0x45, 0x37, 0xc1,
	0xe6, 0x7a, 0xc7, 0xf0, 0x0b, 0xf8, 0xc2, 0x38, 0x80, 0x97, 0x7b, 0x5e, 0x5d, 0xcb, 0xc2, 0xc9,
	0x79, 0x67, 0xdc, 0xb5, 0x75, 0x03, 0x9e, 0x19, 0xf7, 0x3e, 0x71, 0x2d, 0xf7, 0xf6, 0x89, 0x6b,
	0x66, 0xc8, 0x8f, 0x16, 0x81, 0xfd, 0xd3, 0x7c, 0xf8, 0x1f, 0x00, 0x00, 0xff, 0xff, 0x33, 0xf2,
	0xa8, 0xd9, 0x7b, 0x04, 0x00, 0x00,
}