package stream

import (
	"errors"
	"fmt"
	"logger"
	"protocol"
	"sync"
	"time"
	"utility"
)

var (
	bufferCaches      = make(map[string]*syncBufferCache)
	bufferCachesMutex = &sync.Mutex{}
)

type antsFrameHeader struct {
	startId        uint32 //!帧同步头
	frameType      uint32 //!帧类型
	frameNo        uint32 //!帧号
	frameTime      uint32 //!UTC时间
	frameTickCount uint32 //!毫秒为单位的毫秒时间
}

const DATA_IN_BUFFER_COUNT = 25

type RealtimeBuffer struct {
	key         string
	device      *Device
	handle      int
	channel     int
	dataChans   [](chan *dataBuffer)
	operateChan chan chanOperate
	*utility.ReferenceCounter
	cache *syncBufferCache
}

func getRealtimeBuffer(param protocol.OpenRealtimeRequest) *RealtimeBuffer {
	bufferCachesMutex.Lock()
	cache, ok := bufferCaches[param.SerialNum]
	if !ok {
		cache = &syncBufferCache{
			bufferCache: make(map[string]*RealtimeBuffer),
		}
		bufferCaches[param.SerialNum] = cache
	}
	bufferCachesMutex.Unlock()

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	key := fmt.Sprintf("%d_%d", param.StreamType, param.Channel)

	if buffer, ok := cache.bufferCache[key]; ok {
		buffer.AddReference()

		return buffer
	}

	device := GetDevice(param.SerialNum)
	if device == nil {
		return nil
	}

	handle := device.OpenRealtime(param.Channel, param.StreamType)
	if handle == 0 {
		device.DelReference()
		return nil
	}

	buffer := &RealtimeBuffer{
		device:           device,
		handle:           handle,
		channel:          param.Channel,
		dataChans:        make([]chan *dataBuffer, 0),
		operateChan:      make(chan chanOperate, 10),
		ReferenceCounter: utility.NewReferenceCounter(key, time.Second*10),
		cache:            cache,
	}

	go buffer.handleData()

	cache.bufferCache[key] = buffer

	buffer.AddReference()

	return buffer
}

func (buffer *RealtimeBuffer) close() {
	buffer.cache.mutex.Lock()
	defer buffer.cache.mutex.Unlock()

	buffer.device.CloseRealtime(buffer.handle)

	buffer.handle = 0

	delete(buffer.cache.bufferCache, buffer.Key())

	close(buffer.operateChan)

	buffer.device.DelReference()
}

func (buffer *RealtimeBuffer) handleData() {
	defer func() {
		logger.Debug("realtime buffer handledata end.")
	}()

	bufferPool := newDataBufferPool(4*1024*1024, 100)

	for {
		func() {
			for {
				select {
				case operate := <-buffer.operateChan:
					if operate.operate == add_chan {
						buffer.dataChans = append(buffer.dataChans, operate.dataChan)
					} else {
						for index := 0; index < len(buffer.dataChans); index += 1 {
							if buffer.dataChans[index] == operate.dataChan {
								buffer.dataChans = append(buffer.dataChans[:index], buffer.dataChans[index+1:]...)

								close(operate.dataChan)
								buffer.DelReference()

								break
							}
						}
					}
				default:
					return
				}
			}
		}()

		if buffer.Invalid() {
			buffer.close()
			return
		}

		func() {
			defer time.Sleep(time.Microsecond)

			readBuffer := bufferPool.getBuffer()
			if readBuffer == nil {
				logger.Debug("获取不到数据缓存", buffer.channel)
				return
			}

			defer readBuffer.release()

			length, err := buffer.device.GetRealtimeData(buffer.handle, readBuffer.data)
			if err != nil || length == 0 {
				logger.Debug("获取数据失败", buffer.channel)
				return
			}

			readBuffer.length = length

			for _, dataChan := range buffer.dataChans {
				readBuffer.addReference()
				select {
				case dataChan <- readBuffer:
				default:
					readBuffer.release()
					logger.Debug("数据插入失败", buffer.channel)
				}
			}
		}()

	}
}

type RealtimeStream struct {
	buffer     *RealtimeBuffer
	dataChan   chan *dataBuffer
	streamType int
}

func newRealtimeStream(param protocol.OpenRealtimeRequest) *RealtimeStream {
	buffer := getRealtimeBuffer(param)
	if buffer == nil {
		return nil
	}

	dataChan := make(chan *dataBuffer, DATA_IN_BUFFER_COUNT)

	buffer.operateChan <- chanOperate{dataChan, add_chan}

	streamType := STREAM_TYPE_REALTIME_I8
	if param.RtpFormat == protocol.RtpFormatStandard {
		streamType = STREAM_TYPE_REALTIME_STANDARD
	}

	return &RealtimeStream{
		buffer:     buffer,
		dataChan:   dataChan,
		streamType: streamType,
	}
}

func (stream *RealtimeStream) Close() {
	stream.buffer.operateChan <- chanOperate{stream.dataChan, remove_chan}
}

func (stream *RealtimeStream) Reader() FrameReader {
	return &RealtimeReader{
		stream:  stream,
		channel: stream.buffer.channel,
	}
}

func (stream *RealtimeStream) Type() int {
	return stream.streamType
}

func (stream *RealtimeStream) Period() (int, int) {
	return 0, 0
}

type RealtimeReader struct {
	stream  *RealtimeStream
	channel int
}

func (reader *RealtimeReader) Read(buffer []byte) (int, int, error) {
	data, ok := <-reader.stream.dataChan
	if !ok {
		return reader.channel, 0, errors.New("Stream DataChan Closed")
	}

	length := copy(buffer, data.data[:data.length])

	data.release()

	return reader.channel, length, nil
}

const (
	add_chan    = 1
	remove_chan = 2
)

type chanOperate struct {
	dataChan chan *dataBuffer
	operate  int
}

type syncBufferCache struct {
	bufferCache map[string]*RealtimeBuffer
	mutex       sync.Mutex
}

type dataBufferPool struct {
	unusedBuffer       []*dataBuffer
	mutex              sync.Mutex
	bufferSize         int
	currentBufferCount int
	maxBufferCount     int
}

func newDataBufferPool(bufferSize, maxBufferCount int) *dataBufferPool {
	return &dataBufferPool{
		unusedBuffer:       make([]*dataBuffer, 0),
		bufferSize:         bufferSize,
		currentBufferCount: 0,
		maxBufferCount:     maxBufferCount,
	}
}

func (pool *dataBufferPool) getBuffer() *dataBuffer {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if len(pool.unusedBuffer) > 0 {
		buffer := pool.unusedBuffer[0]
		pool.unusedBuffer = pool.unusedBuffer[1:]
		buffer.referenceCount += 1
		return buffer
	}

	if pool.currentBufferCount >= pool.maxBufferCount {
		return nil
	}

	pool.currentBufferCount += 1

	return &dataBuffer{
		data:           make([]byte, pool.bufferSize),
		referenceCount: 1,
		pool:           pool,
	}
}

type dataBuffer struct {
	data           []byte
	length         int
	referenceCount int
	pool           *dataBufferPool
}

func (buffer *dataBuffer) addReference() {
	buffer.pool.mutex.Lock()
	defer buffer.pool.mutex.Unlock()

	buffer.referenceCount += 1
}

func (buffer *dataBuffer) release() {
	buffer.pool.mutex.Lock()
	defer buffer.pool.mutex.Unlock()

	buffer.referenceCount -= 1

	if buffer.referenceCount == 0 {
		buffer.pool.unusedBuffer = append(buffer.pool.unusedBuffer, buffer)
	}
}
