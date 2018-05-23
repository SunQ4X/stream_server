package stream

import (
	"errors"
	"logger"
	"protocol"
)

type IntercomStream struct {
	device  *Device
	handle  int
	channel int
}

func newIntercomStream(param protocol.OpenIntercomRequest) *IntercomStream {
	device := GetDevice(param.SerialNum)
	if device == nil {
		return nil
	}

	handle := device.OpenIntercom(param.Channel)
	if handle == 0 {
		logger.Debug("打开对讲失败")
		device.DelReference()
		return nil
	}

	intercomStream := &IntercomStream{
		device:  device,
		handle:  handle,
		channel: param.Channel,
	}

	return intercomStream
}

func (stream *IntercomStream) Close() {
	stream.device.CloseIntercom(stream.handle)

	stream.handle = 0

	stream.device.DelReference()
}

func (stream *IntercomStream) Reader() FrameReader {
	return &IntercomReader{
		stream:  stream,
		channel: stream.channel,
	}
}

func (stream *IntercomStream) NewWriter() FrameWriter {
	return &IntercomWriter{
		stream: stream,
	}
}

func (stream *IntercomStream) Type() int {
	return STREAM_TYPE_INTERCOM_I8
}

func (stream *IntercomStream) Period() (int, int) {
	return 0, 0
}

type IntercomReader struct {
	stream  *IntercomStream
	channel int
}

func (reader *IntercomReader) Read(buffer []byte) (int, int, error) {
	if reader.stream.handle == 0 {
		return reader.channel, 0, errors.New("Stream Closed")
	}

	length, err := reader.stream.device.GetIntercomData(reader.stream.handle, buffer)

	return reader.channel, length, err
}

type IntercomWriter struct {
	stream *IntercomStream
}

func (writer *IntercomWriter) Write(data []byte) error {
	if writer.stream.handle == 0 {
		return errors.New("Stream Closed")
	}

	return writer.stream.device.SendIntercomData(writer.stream.handle, data)
}
