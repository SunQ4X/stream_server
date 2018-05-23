package rtp

import (
	"errors"
	"io"
)

const (
	RTP_VERSION = 2
)

// Packet as per https://tools.ietf.org/html/rfc1889#section-5.1
//
//  0                   1                   2                   3
//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |V=2|P|X|  CC   |M|     PT      |       sequence number         |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |                           timestamp                           |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
// |           synchronization source (SSRC) identifier            |
// +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
// |            contributing source (CSRC) identifiers             |
// |                             ....                              |
// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
type RtpPacket struct {
	Version        byte
	Padding        byte
	Ext            byte
	Marker         byte
	PayloadType    byte
	SequenceNumber uint
	Timestamp      uint
	SyncSource     uint

	CSRC []uint

	ExtHeader uint
	ExtData   []uint

	Payload io.ReadWriter
}

func NewRtpPacket() *RtpPacket {
	return &RtpPacket{
		Version: RTP_VERSION,
	}
}

func ReadRtp(buffer []byte) (*RtpPacket, error) {
	packet := &RtpPacket{}

	index := 0

	packet.Version = (buffer[index] & 0xC0) >> 6

	if buffer[index]&(1<<5) == 0 {
		packet.Padding = 0
	} else {
		packet.Padding = 1
	}

	if buffer[index]&(1<<4) == 0 {
		packet.Ext = 0
	} else {
		packet.Ext = 1
	}

	packet.CSRC = make([]uint, buffer[index]&0x0F)

	index += 1

	if buffer[index]&(1<<7) == 0 {
		packet.Marker = 0
	} else {
		packet.Marker = 1
	}

	packet.PayloadType = buffer[index] & 0x7F

	index += 1

	packet.SequenceNumber = ToUint(buffer[index : index+2])

	index += 2

	packet.Timestamp = ToUint(buffer[index : index+4])

	index += 4

	packet.SyncSource = ToUint(buffer[index : index+4])

	index += 4

	for i := range packet.CSRC {
		packet.CSRC[i] = ToUint(buffer[index : index+4])
		index += 4
	}

	if packet.Ext != 0 {
		packet.ExtHeader = ToUint(buffer[index : index+2])
		length := ToUint(buffer[index+2 : index+4])
		index += 4

		if length > 0 {
			packet.ExtData = make([]uint, length)
			for i := range packet.ExtData {
				packet.ExtData[i] = ToUint(buffer[index : index+4])
				index += 4
			}
		}
	}

	packet.Payload = &RawDataReadWriter{}
	_, err := packet.Payload.Write(buffer[index:])
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func (packet *RtpPacket) Read(buffer []byte) (int, error) {
	index := 0

	buffer[index] = 0
	buffer[index] |= ((packet.Version & 0x03) << 6)
	buffer[index] |= ((packet.Padding & 0x01) << 5)
	buffer[index] |= ((packet.Ext & 0x01) << 4)
	buffer[index] |= (byte(len(packet.CSRC)) & 0x0F)

	index += 1

	buffer[index] = 0
	buffer[index] |= ((packet.Marker & 0x01) << 7)
	buffer[index] |= (packet.PayloadType & 0x7F)

	index += 1

	ReadUint(buffer[index:], 2, packet.SequenceNumber)

	index += 2

	ReadUint(buffer[index:], 4, packet.Timestamp)

	index += 4

	ReadUint(buffer[index:], 4, packet.SyncSource)

	index += 4

	for _, CSRC := range packet.CSRC {
		ReadUint(buffer[index:], 4, CSRC)

		index += 4
	}

	if packet.Ext != 0 {
		ReadUint(buffer[index:], 2, packet.ExtHeader)

		index += 2

		ReadUint(buffer[index:], 2, uint(len(packet.ExtData)))

		index += 2

		for _, ext := range packet.ExtData {
			ReadUint(buffer[index:], 4, ext)

			index += 4
		}
	}

	length, err := packet.Payload.Read(buffer[index:])
	if err != nil {
		return length, err
	}

	index += length

	return index, nil
}

func ReadUint(buffer []byte, length int, value uint) {
	for i := 0; i < length; i++ {
		buffer[i] = byte((value >> uint(8*(length-i-1))) & 0xFF)
	}
}

func ToUint(buffer []byte) uint {
	var value uint

	length := len(buffer)

	for i, b := range buffer {
		value |= uint(b) << (8 * uint(length-i-1))
	}

	return value
}

type RawDataReadWriter struct {
	Data []byte
}

func (reader *RawDataReadWriter) Read(buffer []byte) (int, error) {
	if len(buffer) < len(reader.Data) {
		return 0, errors.New("Buffer too small")
	}

	copy(buffer, reader.Data)

	return len(reader.Data), nil
}

func (writer *RawDataReadWriter) Write(buffer []byte) (int, error) {
	writer.Data = make([]byte, len(buffer))

	return copy(writer.Data, buffer), nil
}
