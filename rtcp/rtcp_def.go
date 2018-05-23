package rtcp

import (
	"io"
)

const (
	RTCP_VERSION   = 2
	RTCP_TYPE_SR   = 200
	RTCP_TYPE_RR   = 201
	RTCP_TYPE_SDES = 202
	RTCP_TYPE_BYE  = 203
	RTCP_TYPE_APP  = 204
)

const (
	RTCP_SDES_END   = 0
	RTCP_SDES_CNAME = 1
	RTCP_SDES_NAME  = 2
	RTCP_SDES_EMAIL = 3
	RTCP_SDES_PHONE = 4
	RTCP_SDES_LOC   = 5
	RTCP_SDES_TOOL  = 6
	RTCP_SDES_NOTE  = 7
	RTCP_SDES_PRIV  = 8
)

//RTCP包格式
//
//  0               1               2               3
//  0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |V=2|P|  Count  |     PT        |           length              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                             SSRC                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

type RtcpPacket struct {
	Version     byte
	Padding     byte
	ReportCount byte
	PacketType  byte
	Length      uint
	SyncSource  uint
	Contents    []io.Reader
}

type RtcpSender struct {
	NTPSec       uint
	NTPFrac      uint
	RTPTimestamp uint
	PacketCount  uint
	ByteCount    uint
}

type RTCPReceptionReport struct {
	SyncSource         uint
	FractionLost       byte
	TotalLost          uint
	LastSeq            uint
	InterarrivalJitter uint
	LastSR             uint
	DelayLastSR        uint
}

type RTCPSDESItem struct {
	Type byte
	Data string
}

type RTCPSDES struct {
	SyncSource uint
	Items      []RTCPSDESItem
}

func ReadUint(buffer []byte, length int, value uint) {
	for i := 0; i < length; i++ {
		buffer[i] = byte((value >> uint(8*(length-i-1))) & 0xFF)
	}
}

func (packet *RtcpPacket) Read(buffer []byte) (int, error) {
	index := 0

	buffer[index] = 0
	buffer[index] |= ((packet.Version & 0x03) << 6)
	buffer[index] |= ((packet.Padding & 0x01) << 5)
	buffer[index] |= (packet.ReportCount & 0x1F)

	index += 1

	buffer[index] = packet.PacketType

	index += 1

	ReadUint(buffer[index:], 2, packet.Length)

	index += 2

	ReadUint(buffer[index:], 4, packet.SyncSource)

	index += 4

	for _, contents := range packet.Contents {
		length, err := contents.Read(buffer[index:])
		if err != nil {
			return index, err
		}

		index += length
	}

	packet.Length = uint(index)

	ReadUint(buffer[2:], 2, packet.Length)

	return index, nil
}

func (sender *RtcpSender) Read(buffer []byte) (int, error) {
	index := 0

	ReadUint(buffer[index:], 4, sender.NTPSec)

	index += 4

	ReadUint(buffer[index:], 4, sender.NTPFrac)

	index += 4

	ReadUint(buffer[index:], 4, sender.RTPTimestamp)

	index += 4

	ReadUint(buffer[index:], 4, sender.PacketCount)

	index += 4

	ReadUint(buffer[index:], 4, sender.ByteCount)

	index += 4

	return index, nil
}

func (report *RTCPReceptionReport) Read(buffer []byte) (int, error) {
	index := 0

	ReadUint(buffer[index:], 4, report.SyncSource)

	index += 4

	buffer[index] = report.FractionLost

	index += 1

	ReadUint(buffer[index:], 3, report.TotalLost)

	index += 3

	ReadUint(buffer[index:], 4, report.LastSeq)

	index += 4

	ReadUint(buffer[index:], 4, report.InterarrivalJitter)

	index += 4

	ReadUint(buffer[index:], 4, report.LastSR)

	index += 4

	ReadUint(buffer[index:], 4, report.DelayLastSR)

	index += 4

	return index, nil
}

func (sdes *RTCPSDES) Read(buffer []byte) (int, error) {
	index := 0

	ReadUint(buffer[index:], 4, sdes.SyncSource)

	index += 4

	for _, item := range sdes.Items {
		buffer[index] = item.Type

		index += 1

		dataLen := (len(item.Data) + 3) / 4

		buffer[index] = byte(dataLen)

		index += 1

		copy(buffer[index:], item.Data)

		index += dataLen
	}

	return index, nil
}
