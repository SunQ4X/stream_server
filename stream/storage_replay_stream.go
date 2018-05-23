package stream

import (
	pp "PlaybackProtocol"
	//	"crypto/md5"
	//	"encoding/hex"
	"errors"
	"io"
	"logger"
	//"os"
	"protocol"
	"time"
	"unsafe"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type StorageReplayStream struct {
	conn        *grpc.ClientConn
	client      pp.PlaybackClient
	stream      pp.Playback_StreamClient
	handle      uint64
	streamType  uint32
	recordType  uint32
	orAnd       int32
	streamType0 int
	//fd *os.File
}

func newStorageReplayStream(param protocol.OpenStorageReplayRequest) *StorageReplayStream {
	//outputFile, _ := os.Create("./record.i8") //创建文件//
	//playbackHandle.fd = outputFile

	println(".........newStorageReplayStream...........")
	begin := time.Date(int(param.ReplayRequest.BeginTime.Year),
		time.Month(int(param.ReplayRequest.BeginTime.Month)),
		int(param.ReplayRequest.BeginTime.Day),
		int(param.ReplayRequest.BeginTime.Hour),
		int(param.ReplayRequest.BeginTime.Minute),
		int(param.ReplayRequest.BeginTime.Second),
		0,
		time.FixedZone("", int(param.ReplayRequest.BeginTime.TimeZone)))

	end := time.Date(int(param.ReplayRequest.EndTime.Year),
		time.Month(int(param.ReplayRequest.EndTime.Month)),
		int(param.ReplayRequest.EndTime.Day),
		int(param.ReplayRequest.EndTime.Hour),
		int(param.ReplayRequest.EndTime.Minute),
		int(param.ReplayRequest.EndTime.Second),
		0,
		time.FixedZone("", int(param.ReplayRequest.EndTime.TimeZone)))

	// 连接
	conn, err := grpc.Dial(param.RpcAddress, grpc.WithInsecure())
	if err != nil {
		println(".........rpc.Dial failed...........")
		logger.Error(err)
		return nil
	}

	// 初始化客户端
	client := pp.NewPlaybackClient(conn)

	var channels []*pp.Channel
	for _, r := range param.ReplayRequest.ChIds {
		channels = append(channels, &pp.Channel{
			Device: r.SerialNum,
			Ch:     uint32(r.Channel),
		})
	}
	openRequest := pp.OpenRequest{
		Channels:   channels,
		BeginTime:  uint64(begin.Unix()),
		EndTime:    uint64(end.Unix()),
		StreamType: uint32(param.ReplayRequest.StreamType),
		RecordType: uint32(param.ReplayRequest.RecordType),
		OrAnd:      uint32(param.ReplayRequest.RecordType_condition),
	}

	openReply, err := client.Open(context.Background(), &openRequest)
	if err != nil {
		logger.Error(err)
		return nil
	}
	if openReply.Error != "" {
		logger.Error(openReply.Error)
		return nil
	}

	streamRequest := pp.StreamRequest{
		Handle: openReply.Handle,
	}
	stream, err := client.Stream(context.Background(), &streamRequest)
	if err != nil {
		println(".........rpc.Stream failed...........")
		logger.Error(err)
		return nil
	}

	streamType0 := STREAM_TYPE_STORAGE_REPLAY_I8
	if param.RtpFormat == protocol.RtpFormatStandard {
		streamType0 = STREAM_TYPE_STORAGE_REPLAY_STANDARD
	}

	return &StorageReplayStream{
		client:      client,
		conn:        conn,
		stream:      stream,
		handle:      openReply.Handle,
		streamType0: streamType0,
	}
}

func (stream *StorageReplayStream) Close() {
	stream.conn.Close()
	//stream.fd.Close()
}

func (stream *StorageReplayStream) Reader() FrameReader {
	return &StorageReplayReader{
		stream: stream,
	}
}

func (stream *StorageReplayStream) Type() int {
	return stream.streamType0
}

func (stream *StorageReplayStream) Period() (int, int) {
	return 0, 0
}

func (stream *StorageReplayStream) SeekTime(year, month, day, hour, minute, second, timeZone int) error {
	seekTime := time.Date(int(year),
		time.Month(int(month)),
		int(day),
		int(hour),
		int(minute),
		int(second),
		0,
		time.FixedZone("", int(timeZone)))
	request := pp.SeekRequest{
		Handle:     stream.handle,
		SeekTime:   uint64(seekTime.Unix()),
		StreamType: stream.streamType,
		RecordType: stream.recordType,
		OrAnd:      uint32(stream.orAnd),
	}
	reply, err := stream.client.Seek(context.Background(), &request)
	if err != nil {
		logger.Error(err)
		return errors.New("err")
	}
	if reply.Error != "" {
		logger.Error(reply.Error)
		return errors.New(reply.Error)
	}
	return nil
}

type StorageReplayReader struct {
	stream *StorageReplayStream
}

type AntsFrameHeader struct {
	startId        uint32 //!帧同步头
	frameType      uint32 //!帧类型
	frameNo        uint32 //!帧号
	frameTime      uint32 //!UTC时间
	frameTickCount uint32 //!毫秒为单位的毫秒时间
}

func (reader *StorageReplayReader) Read(buffer []byte) (int, int, error) {
	reply, err := reader.stream.stream.Recv()
	if err == io.EOF {
		println("EOF break")
		return 0, 0, nil
	}
	if err != nil {
		print("failed to recv: ")
		logger.Error(err)
		return 0, 0, errors.New("failed to recv")
	}
	if reply.Error != "" {
		println("reply.Error", reply.Error)
		return 0, 0, nil
	}
	copy(buffer, reply.Data)
	//把i8帧头的时间戳改为存储服务器记录的时间
	var frameHeader *AntsFrameHeader = (*AntsFrameHeader)(unsafe.Pointer(&buffer[0]))
	//println("1:", reply.FrameTime)
	//println("2:", frameHeader.frameTime)
	frameHeader.frameTime = uint32(reply.FrameTime)
	//println("3:", frameHeader.frameTime)
	//		md5HashInBytes := md5.Sum(buffer[:reply.DataSize])
	//		checksum := hex.EncodeToString(md5HashInBytes[:])
	//		println("Pid:", reply.Pid, "md5:", reply.Checksum, reply.Checksum == checksum)

	//reader.stream.fd.Write(reply.Data[:reply.DataSize])//
	//reader.stream.fd.Sync()//
	return int(reply.Ch), int(reply.DataSize), nil
}
