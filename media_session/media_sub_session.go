package media_session

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"logger"
	"math/rand"
	"nal"
	"rtp"
	"stream"
	"time"
)

const (
	RTP_MAX_PAYLOAD_SIZE = 1280
	PAYLOAD_TYPE_FU_A    = 28
)

type MediaSubSession interface {
	SDPLines() (string, error)
	TrackId() string
	SyncSource() uint
	SetDestination(dest *Destination)
	Play(sessionID string) (uint, uint, error)
	Close()
	InsertData(int, uint, []byte)
}

type frameData struct {
	channel int
	data    []byte
}

type baseSubSession struct {
	mediaSession        *MediaSession
	syncSource          uint
	destination         *Destination
	currentSeq          uint
	currentRtpTimestamp uint
}

func newBaseSubSesion(mediaSession *MediaSession) *baseSubSession {
	rand.Seed(time.Now().UTC().UnixNano())
	syncSource := rand.Uint32()

	return &baseSubSession{
		mediaSession: mediaSession,
		syncSource:   uint(syncSource),
	}
}

func (subSession *baseSubSession) SyncSource() uint {
	return subSession.syncSource
}

func (subSession *baseSubSession) SetDestination(dest *Destination) {
	subSession.destination = dest
}

type VideoSubSession struct {
	*baseSubSession
}

func NewVideoSubSession(mediaSession *MediaSession) *VideoSubSession {
	subSession := &VideoSubSession{
		baseSubSession: newBaseSubSesion(mediaSession),
	}

	return subSession
}

func (subSession *VideoSubSession) SDPLines() (string, error) {
	sdpLines := fmt.Sprintf(`m=video 0 RTP/AVP 110
c=IN IP4 0.0.0.0
a=rtpmap:110 AntsComb/90000
a=control:%s
a=recvonly
`, subSession.TrackId())

	return sdpLines, nil
}

func (subSession *VideoSubSession) TrackId() string {
	return "?ctype=video"
}

func (subSession *VideoSubSession) Close() {
	if subSession.destination != nil {
		subSession.destination.Close()
		subSession.destination = nil
	}
}

func (subSession *VideoSubSession) InsertData(ch int, timestamp uint, data []byte) {
	sendCombRtp := func(marker, start, streamType, ch uint, payload io.ReadWriter) {
		subSession.currentSeq += 1

		if subSession.destination == nil {
			return
		}

		pack := rtp.NewRtpPacket()

		pack.PayloadType = 110
		pack.SequenceNumber = subSession.currentSeq
		pack.Timestamp = subSession.currentRtpTimestamp
		pack.SyncSource = subSession.syncSource
		pack.Payload = payload
		pack.Marker = byte(marker)
		pack.Ext = 1
		pack.ExtHeader = 0xC4DB
		pack.ExtData = make([]uint, 4)
		pack.ExtData[0] |= ((marker & 0x01) << 31)
		pack.ExtData[0] |= ((start & 0x01) << 30)

		pack.ExtData[0] |= ((streamType & 0xFF) << 16)
		pack.ExtData[0] |= ((ch & 0x00FF) << 8)
		pack.ExtData[0] |= ((ch & 0xFF00) >> 8)

		subSession.destination.SendRtp(pack)
	}

	if subSession.destination == nil {
		return
	}

	subSession.currentRtpTimestamp = timestamp

	if len(data) <= RTP_MAX_PAYLOAD_SIZE {
		sendCombRtp(1, 0, 1, uint(ch), &rtp.RawDataReadWriter{Data: data})
		return
	}

	offset := 0

	for offset < len(data) {
		var payloadLength int

		if len(data)-offset > RTP_MAX_PAYLOAD_SIZE {
			payloadLength = RTP_MAX_PAYLOAD_SIZE
		} else {
			payloadLength = len(data) - offset
		}

		if offset == 0 {
			sendCombRtp(0, 1, 1, uint(ch), &rtp.RawDataReadWriter{data[offset : offset+payloadLength]})
		} else if offset+payloadLength == len(data) {
			sendCombRtp(1, 0, 1, uint(ch), &rtp.RawDataReadWriter{data[offset : offset+payloadLength]})
		} else {
			sendCombRtp(0, 0, 1, uint(ch), &rtp.RawDataReadWriter{data[offset : offset+payloadLength]})
		}

		offset += payloadLength
	}
}

func (subSession *VideoSubSession) Play(sessionID string) (uint, uint, error) {
	if subSession.destination == nil {
		return 0, 0, errors.New("Destination Not Found")
	}

	return subSession.currentSeq, subSession.currentRtpTimestamp, nil
}

type AudioSubSession struct {
	*baseSubSession
	readBuffer []byte
}

func NewAudioSubSession(mediaSession *MediaSession) *AudioSubSession {
	subSession := &AudioSubSession{
		baseSubSession: newBaseSubSesion(mediaSession),
		readBuffer:     make([]byte, 4*1024*1024),
	}

	return subSession
}

func (subSession *AudioSubSession) SDPLines() (string, error) {
	sdpLines := fmt.Sprintf(`m=audio 0 RTP/AVP 0
a=rtpmap:0 pcmu/8000/1
a=control:%s
a=recvonly
`, subSession.TrackId())

	return sdpLines, nil
}

func (subSession *AudioSubSession) TrackId() string {
	return "?ctype=audio"
}

func (subSession *AudioSubSession) InsertData(ch int, timestamp uint, data []byte) {
	if subSession.destination == nil {
		return
	}

	sendRtp := func(payload io.ReadWriter) {
		subSession.currentSeq += 1

		pack := rtp.NewRtpPacket()

		pack.PayloadType = 0
		pack.SequenceNumber = subSession.currentSeq
		pack.Timestamp = subSession.currentRtpTimestamp
		pack.SyncSource = subSession.syncSource
		pack.Payload = payload

		subSession.destination.SendRtp(pack)
	}

	subSession.currentRtpTimestamp = timestamp

	sendRtp(&rtp.RawDataReadWriter{Data: data})
}

func (subSession *AudioSubSession) Play(sessionID string) (uint, uint, error) {
	if subSession.destination == nil {
		return 0, 0, errors.New("Destination Not Found")
	}

	return subSession.currentSeq, subSession.currentRtpTimestamp, nil
}

func (subSession *AudioSubSession) Close() {
	if subSession.destination != nil {
		subSession.destination.Close()
		subSession.destination = nil
	}
}

type AudioBackSubSession struct {
	*baseSubSession
	writer stream.FrameWriter
}

func NewAudioBackSubSession(mediaSession *MediaSession, writer stream.FrameWriter) *AudioBackSubSession {
	return &AudioBackSubSession{
		baseSubSession: newBaseSubSesion(mediaSession),
		writer:         writer,
	}
}

func (subSession *AudioBackSubSession) SDPLines() (string, error) {
	sdpLines := fmt.Sprintf(`m=audio 0 RTP/AVP 0
a=rtpmap:0 pcmu/8000/1
a=control:%s
a=sendonly
`, subSession.TrackId())

	return sdpLines, nil
}

func (subSession *AudioBackSubSession) TrackId() string {
	return "?ctype=audioback"
}

func (subSession *AudioBackSubSession) InsertData(tcpChannel int, timestamp uint, data []byte) {
	if subSession.destination == nil || tcpChannel != subSession.destination.RtpChannelId || subSession.writer == nil {
		return
	}

	pack, err := rtp.ReadRtp(data)
	if err != nil {
		logger.Debug("解析rtp包失败:", err)
		return
	}

	buffer := make([]byte, 2048)
	length, err := pack.Payload.Read(buffer)
	if err != nil {
		logger.Debug("读取rtp数据失败:", err)
		return
	}

	subSession.writer.Write(buffer[:length])
}

func (subSession *AudioBackSubSession) Play(sessionID string) (uint, uint, error) {
	return 0, 0, nil
}

func (subSession *AudioBackSubSession) Close() {
	if subSession.destination != nil {
		subSession.destination.Close()
		subSession.destination = nil
	}
}

type AppSubSession struct {
	*baseSubSession
}

func NewAppSubSession(mediaSession *MediaSession) *AppSubSession {
	return &AppSubSession{
		baseSubSession: newBaseSubSesion(mediaSession),
	}
}

func (subSession *AppSubSession) SDPLines() (string, error) {
	sdpLines := fmt.Sprintf(`m=application 0 RTP/AVP 106
a=rtpmap:106 vnd.onvif.metadata/90000
a=control:%s
a=sendonly
`, subSession.TrackId())

	return sdpLines, nil
}

func (subSession *AppSubSession) TrackId() string {
	return "?ctype=app106"
}

func (subSession *AppSubSession) InsertData(ch int, data []byte) {

}

func (subSession *AppSubSession) Play(sessionID string) (uint, uint, error) {

	return 0, 0, nil
}

func (subSession *AppSubSession) Close() {
	if subSession.destination != nil {
		subSession.destination.Close()
		subSession.destination = nil
	}
}

type StandardVideoSubSession struct {
	*baseSubSession
	profileLevelId uint
	sps            string
	pps            string
	sdpLines       string
}

func NewStandardVideoSubSession(mediaSession *MediaSession) *StandardVideoSubSession {
	subSession := &StandardVideoSubSession{
		baseSubSession: newBaseSubSesion(mediaSession),
	}

	return subSession
}

func (subSession *StandardVideoSubSession) SDPLines() (string, error) {
	if subSession.sdpLines != "" {
		return subSession.sdpLines, nil
	}

	timer := time.NewTimer(time.Second * 3)

	for {
		select {
		case <-timer.C:
			return "", errors.New("Get SDPLines Timeout")
		default:
			if subSession.sps != "" && subSession.pps != "" {
				subSession.sdpLines = fmt.Sprintf(`m=video 0 RTP/AVP 96
c=IN IP4 0.0.0.0
a=rtpmap:96 h264/90000
a=fmtp:96 packetization-mode=1;profile-level-id=%08X;sprop-parameter-sets=%s,%s
a=control:%s
`,
					subSession.profileLevelId, subSession.sps, subSession.pps, subSession.TrackId())
				return subSession.sdpLines, nil
			}
		}
	}
	//b=AS:500
}

func (subSession *StandardVideoSubSession) TrackId() string {
	return "?ctype=video"
}

func (subSession *StandardVideoSubSession) Close() {
	if subSession.destination != nil {
		subSession.destination.Close()
		subSession.destination = nil
	}
}

func (subSession *StandardVideoSubSession) InsertData(ch int, timestamp uint, data []byte) {
	sendRtp := func(payload io.ReadWriter) {
		subSession.currentSeq += 1

		if subSession.destination == nil {
			return
		}

		pack := rtp.NewRtpPacket()

		pack.PayloadType = 96
		pack.SequenceNumber = subSession.currentSeq
		pack.Timestamp = subSession.currentRtpTimestamp
		pack.SyncSource = subSession.syncSource
		pack.Payload = payload

		subSession.destination.SendRtp(pack)
	}

	subSession.currentRtpTimestamp = timestamp

	nalOffset := 0

	for nalOffset < len(data) {
		nalData, readLen, err := nal.FindNalUnitFromBuffer(data[nalOffset:])
		if err != nil {
			logger.Error("FindNalUnitFromBuffer error:", err)
			break
		}

		nalOffset += readLen

		nalUnit := nal.ParseNalUnit(nalData)

		if subSession.sdpLines == "" {
			if nalUnit.NalUnitType == 7 {
				subSession.sps = base64.StdEncoding.EncodeToString(nalUnit.Payload)
				subSession.profileLevelId = (uint(nalUnit.Payload[1]) << 16) | (uint(nalUnit.Payload[2]) << 8) | uint(nalUnit.Payload[3])
				fmt.Println("====nal type 7, sps:", subSession.sps, "profile level id:", subSession.profileLevelId)
			} else if nalUnit.NalUnitType == 8 {
				subSession.pps = base64.StdEncoding.EncodeToString(nalUnit.Payload)
				fmt.Println("====nal type 8, pps:", subSession.pps)
			}
		}

		if len(nalUnit.Payload) <= RTP_MAX_PAYLOAD_SIZE {
			sendRtp(nalUnit)
		} else {
			offset := 0

			for offset < len(nalUnit.Payload) {
				fragment := &nal.NalFragmentA{
					Forbidden:   nalUnit.Forbidden,
					NalRefIdc:   nalUnit.NalRefIdc,
					PayloadType: PAYLOAD_TYPE_FU_A,
					NalUnitType: nalUnit.NalUnitType,
				}

				payloadLength := 0

				bytesRemain := len(nalUnit.Payload) - offset

				if bytesRemain > RTP_MAX_PAYLOAD_SIZE {
					payloadLength = RTP_MAX_PAYLOAD_SIZE
				} else {
					payloadLength = bytesRemain
				}

				fragment.Payload = nalUnit.Payload[offset : offset+payloadLength]

				if offset == 0 {
					//第一个FU-A
					fragment.S = 1
					fragment.E = 0
					fragment.R = 0
				} else if offset+payloadLength == len(nalUnit.Payload) {
					//最后一个FU-A
					fragment.S = 0
					fragment.E = 1
					fragment.R = 0
				} else {
					//中间的FU-A
					fragment.S = 0
					fragment.E = 0
					fragment.R = 0
				}

				sendRtp(fragment)

				offset += payloadLength
			}
		}
	}
}

func (subSession *StandardVideoSubSession) Play(sessionID string) (uint, uint, error) {
	if subSession.destination == nil {
		return 0, 0, errors.New("Destination Not Found")
	}

	return subSession.currentSeq, subSession.currentRtpTimestamp, nil
}

type StandardAudioSubSession struct {
	*baseSubSession
	readBuffer []byte
}

func NewStandardAudioSubSession(mediaSession *MediaSession) *StandardAudioSubSession {
	subSession := &StandardAudioSubSession{
		baseSubSession: newBaseSubSesion(mediaSession),
		readBuffer:     make([]byte, 4*1024*1024),
	}

	return subSession
}

func (subSession *StandardAudioSubSession) SDPLines() (string, error) {
	sdpLines := fmt.Sprintf(`m=audio 0 RTP/AVP 0
a=rtpmap:0 pcmu/8000/1
a=control:%s
a=recvonly
`, subSession.TrackId())

	return sdpLines, nil
}

func (subSession *StandardAudioSubSession) TrackId() string {
	return "?ctype=audio"
}

func (subSession *StandardAudioSubSession) InsertData(ch int, timestamp uint, data []byte) {
	if subSession.destination == nil {
		return
	}

	sendRtp := func(payload io.ReadWriter) {
		subSession.currentSeq += 1

		pack := rtp.NewRtpPacket()

		pack.PayloadType = 0
		pack.SequenceNumber = subSession.currentSeq
		pack.Timestamp = subSession.currentRtpTimestamp
		pack.SyncSource = subSession.syncSource
		pack.Payload = payload

		subSession.destination.SendRtp(pack)
	}

	subSession.currentRtpTimestamp = timestamp

	sendRtp(&rtp.RawDataReadWriter{Data: data})
}

func (subSession *StandardAudioSubSession) Play(sessionID string) (uint, uint, error) {
	if subSession.destination == nil {
		return 0, 0, errors.New("Destination Not Found")
	}

	return subSession.currentSeq, subSession.currentRtpTimestamp, nil
}

func (subSession *StandardAudioSubSession) Close() {
	if subSession.destination != nil {
		subSession.destination.Close()
		subSession.destination = nil
	}
}
