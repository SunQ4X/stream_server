package stream

import (
	"protocol"
	"sync"
)

const (
	FRAME_TYPE_VIDEO = 0
	FRAME_TYPE_AUDIO = 1
	FRAME_TYPE_APP   = 2
)

const (
	STREAM_TYPE_REALTIME_I8             = 0
	STREAM_TYPE_DEVICE_REPLAY_I8        = 1
	STREAM_TYPE_STORAGE_REPLAY_I8       = 2
	STREAM_TYPE_INTERCOM_I8             = 3
	STREAM_TYPE_REALTIME_STANDARD       = 4
	STREAM_TYPE_DEVICE_REPLAY_STANDARD  = 5
	STREAM_TYPE_STORAGE_REPLAY_STANDARD = 6
)

var (
	streamCache      = make(map[string]interface{})
	streamCacheMutex sync.RWMutex
)

type Stream interface {
	Type() int
	Period() (int, int)
	Reader() FrameReader
	Close()
}

type FrameReader interface {
	Read([]byte) (int, int, error)
}

type FrameWriter interface {
	Write([]byte) error
}

func AddStream(streamName string, param interface{}) {
	streamCacheMutex.Lock()
	streamCache[streamName] = param
	streamCacheMutex.Unlock()
}

func NewStream(streamName string) Stream {
	//查询streamName,得到对应的开流参数
	streamCacheMutex.RLock()
	param, ok := streamCache[streamName]
	if !ok {
		streamCacheMutex.RUnlock()
		return nil
	}
	streamCacheMutex.RUnlock()

	switch param.(type) {
	case protocol.OpenRealtimeRequest:
		return newRealtimeStream(param.(protocol.OpenRealtimeRequest))
	case protocol.OpenDeviceReplayRequest:
		return newDeviceReplayStream(param.(protocol.OpenDeviceReplayRequest))
	case protocol.OpenStorageReplayRequest:
		return newStorageReplayStream(param.(protocol.OpenStorageReplayRequest))
	case protocol.OpenIntercomRequest:
		return newIntercomStream(param.(protocol.OpenIntercomRequest))
	default:
		return nil
	}
}
