package stream

import (
	"errors"
	"protocol"
	"time"
)

type DeviceReplayStream struct {
	device     *Device
	handle     int
	begin      int
	end        int
	streamType int
}

func newDeviceReplayStream(param protocol.OpenDeviceReplayRequest) *DeviceReplayStream {
	device := GetDevice(param.ReplayRequest.SerialNum)
	if device == nil {
		return nil
	}

	handle := device.OpenReplay(param.ReplayRequest)
	if handle == 0 {
		device.DelReference()
		return nil
	}

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

	streamType := STREAM_TYPE_DEVICE_REPLAY_I8
	if param.RtpFormat == protocol.RtpFormatStandard {
		streamType = STREAM_TYPE_DEVICE_REPLAY_STANDARD
	}

	historyStream := &DeviceReplayStream{
		device:     device,
		handle:     handle,
		begin:      int(begin.Unix()),
		end:        int(end.Unix()),
		streamType: streamType,
	}

	return historyStream
}

func (stream *DeviceReplayStream) Close() {
	stream.device.CloseReplay(stream.handle)

	stream.handle = 0

	stream.device.DelReference()
}

func (stream *DeviceReplayStream) SeekTime(year, month, day, hour, minute, second, timeZone int) error {
	return stream.device.SeekReplayTime(stream.handle, year, month, day, hour, minute, second, timeZone)
}

func (stream *DeviceReplayStream) Reader() FrameReader {
	return &DeviceReplayReader{
		stream: stream,
	}
}

func (stream *DeviceReplayStream) Type() int {
	return stream.streamType
}

func (stream *DeviceReplayStream) Period() (int, int) {
	return stream.begin, stream.end
}

type DeviceReplayReader struct {
	stream *DeviceReplayStream
}

func (reader *DeviceReplayReader) Read(buffer []byte) (int, int, error) {
	if reader.stream.handle == 0 {
		return 0, 0, errors.New("Stream Closed")
	}

	return reader.stream.device.GetReplayData(reader.stream.handle, buffer)
}
