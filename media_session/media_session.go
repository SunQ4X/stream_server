package media_session

import (
	"errors"
	"fmt"
	"logger"
	"reflect"
	"stream"
	"time"
	"unsafe"
)

type MediaSession struct {
	StreamName  string
	Stream      stream.Stream
	subSessions []MediaSubSession
}

func NewMediaSession(streamName string) *MediaSession {
	stream0 := stream.NewStream(streamName)
	if stream0 == nil || reflect.ValueOf(stream0).IsNil() {
		logger.Error("new stream failed:")
		return nil
	}

	sess := &MediaSession{
		StreamName:  streamName,
		Stream:      stream0,
		subSessions: make([]MediaSubSession, 0),
	}

	if sess.Stream.Type() == stream.STREAM_TYPE_INTERCOM_I8 {
		sess.subSessions = append(sess.subSessions, NewAudioSubSession(sess), NewAudioBackSubSession(sess, stream0.(*stream.IntercomStream).NewWriter()))
	} else if sess.Stream.Type() == stream.STREAM_TYPE_REALTIME_I8 ||
		sess.Stream.Type() == stream.STREAM_TYPE_DEVICE_REPLAY_I8 ||
		sess.Stream.Type() == stream.STREAM_TYPE_STORAGE_REPLAY_I8 {
		sess.subSessions = append(sess.subSessions, NewVideoSubSession(sess))
	} else {
		sess.subSessions = append(sess.subSessions, NewStandardVideoSubSession(sess), NewStandardAudioSubSession(sess))
	}

	//sess.subSessions[stream.FRAME_TYPE_AUDIO] = NewAudioSubSession(sess)
	//	sess.subSessions[stream.FRAME_TYPE_APP] = NewAppSubSession(sess)

	go sess.handleFrameData()

	return sess
}

type antsFrameHeader struct {
	startId        uint32
	frameType      uint32
	frameNo        uint32
	frameTime      uint32
	frameTickCount uint32
	frameLen       uint32
	videoframe     uint64
	timestamp      uint32
}

func (mediaSession *MediaSession) handleFrameData() {
	defer func() {
		if re := recover(); re != nil {
			logger.Error("panic:", re)
		}
	}()

	reader := mediaSession.Stream.Reader()
	readBuffer := make([]byte, 4*1024*1024)

	for {
		ch, length, err := reader.Read(readBuffer)
		if err != nil {
			logger.Error("read data error:", err)
			break
		}

		if length == 0 {
			time.Sleep(time.Millisecond * 2)
			continue
		}

		header := (*antsFrameHeader)(unsafe.Pointer(&readBuffer[0]))

		if mediaSession.Stream.Type() == stream.STREAM_TYPE_REALTIME_STANDARD ||
			mediaSession.Stream.Type() == stream.STREAM_TYPE_REALTIME_STANDARD ||
			mediaSession.Stream.Type() == stream.STREAM_TYPE_REALTIME_STANDARD {

			if header.frameType == 1 ||
				header.frameType == 9 ||
				header.frameType == 18 ||
				header.frameType == 19 {
				mediaSession.subSessions[0].InsertData(ch, uint(header.timestamp), readBuffer[:length])
			} else if header.frameType == 8 {
				mediaSession.subSessions[1].InsertData(ch, uint(header.timestamp), readBuffer[:length])
			}
		} else {
			mediaSession.subSessions[0].InsertData(ch, uint(header.timestamp), readBuffer[:length])
		}
	}

	for _, subSession := range mediaSession.subSessions {
		subSession.Close()
	}
}

func (mediaSession *MediaSession) Close() {
	mediaSession.Stream.Close()
}

func (mediaSession *MediaSession) SDPLines() (string, error) {
	var sdpLines string

	for _, subSession := range mediaSession.subSessions {
		sdp, err := subSession.SDPLines()
		if err != nil {
			return "", err
		}

		sdpLines += sdp
	}

	return sdpLines, nil
}

func (mediaSession *MediaSession) LookupSubSession(trackId string) MediaSubSession {
	for _, subSession := range mediaSession.subSessions {
		if subSession.TrackId() == trackId {
			return subSession
		}
	}

	return nil
}

func (mediaSession *MediaSession) StartPlay(sessID, trackId string) (string, error) {
	if trackId != "" {
		subSession := mediaSession.LookupSubSession(trackId)
		if subSession == nil {
			return "", errors.New("Sub Session Not Found")
		}

		seq, timestamp, err := subSession.Play(sessID)
		if err != nil {
			return "", errors.New("Sub Session Play Error")
		}

		return fmt.Sprintf("url=%s/%s;seq=%d;rtptime=%d", mediaSession.StreamName, subSession.TrackId(), seq, timestamp), nil
	}

	rtpInfo := ""

	for _, subSession := range mediaSession.subSessions {
		seq, timestamp, err := subSession.Play(sessID)
		if err != nil {
			return "", errors.New("Sub Session Play Error")
		}

		if rtpInfo != "" {
			rtpInfo += ","
		}

		rtpInfo += fmt.Sprintf("url=%s/%s;seq=%d;rtptime=%d", mediaSession.StreamName, subSession.TrackId(), seq, timestamp)
	}

	return rtpInfo, nil
}

func (mediaSession *MediaSession) HandleStreamData(tcpChannel int, data []byte) {
	if mediaSession.Stream == nil || mediaSession.Stream.Type() != stream.STREAM_TYPE_INTERCOM_I8 {
		return
	}

	mediaSession.subSessions[1].InsertData(tcpChannel, 0, data)
}
