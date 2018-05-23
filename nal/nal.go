package nal

import (
	"errors"
	//	"fmt"
	//	"logger"
)

var (
	NAL_UNIT_START_CODE = [4]byte{0x00, 0x00, 0x00, 0x01}
)

type NalUnit struct {
	Forbidden   byte
	NalRefIdc   byte
	NalUnitType byte
	Payload     []byte
}

type NalFragmentA struct {
	Forbidden   byte
	NalRefIdc   byte
	PayloadType byte
	S           byte
	E           byte
	R           byte
	NalUnitType byte
	Payload     []byte
}

func FindNalUnitFromBuffer(data []byte) ([]byte, int, error) {
	findNalStartPos := func(data []byte, startPos int) int {
		position := startPos
		for position < len(data)-4 {
			if data[position] == NAL_UNIT_START_CODE[0] && data[position+1] == NAL_UNIT_START_CODE[1] && data[position+2] == NAL_UNIT_START_CODE[2] && data[position+3] == NAL_UNIT_START_CODE[3] {
				return position
			}

			position += 1
		}

		return -1
	}

	firstStartPos := findNalStartPos(data, 0)

	if firstStartPos == -1 {
		return nil, 0, errors.New("NalUnitStartCode Not found")
	}

	nextStartPos := findNalStartPos(data, firstStartPos+4)

	if nextStartPos == -1 {
		return data[firstStartPos:], len(data), nil
	} else {
		return data[firstStartPos:nextStartPos], nextStartPos, nil
	}
}

func ParseNalUnit(data []byte) *NalUnit {
	unit := &NalUnit{}

	unit.Forbidden = (data[4] & 0x80) >> 7
	unit.NalRefIdc = (data[4] & 0x60) >> 5
	unit.NalUnitType = data[4] & 0x1F

	unit.Payload = data[5:]

	return unit
}

func (unit *NalUnit) Read(buffer []byte) (int, error) {
	index := 0

	buffer[index] = 0
	buffer[index] |= ((unit.Forbidden & 0x01) << 7)
	buffer[index] |= ((unit.NalRefIdc & 0x03) << 5)
	buffer[index] |= (unit.NalUnitType & 0x1F)

	index += 1

	copy(buffer[index:], unit.Payload)

	index += len(unit.Payload)

	//	fmt.Println("NalUnit Forbidden", unit.Forbidden, "NalRefIdc", unit.NalRefIdc, "NalUnitType", unit.NalUnitType, "header:", buffer[0])

	return index, nil
}

func (unit *NalUnit) Write(buffer []byte) (int, error) {
	return 0, nil
}

func (unit *NalFragmentA) Read(buffer []byte) (int, error) {
	index := 0

	buffer[index] = 0
	buffer[index] |= ((unit.Forbidden & 0x01) << 7)
	buffer[index] |= ((unit.NalRefIdc & 0x03) << 5)
	buffer[index] |= (unit.PayloadType & 0x1F)

	index += 1

	buffer[index] = 0
	buffer[index] |= ((unit.S & 0x01) << 7)
	buffer[index] |= ((unit.E & 0x01) << 6)
	buffer[index] |= ((unit.R & 0x01) << 5)
	buffer[index] |= (unit.NalUnitType & 0x1F)

	index += 1

	copy(buffer[index:], unit.Payload)

	index += len(unit.Payload)

	//	fmt.Println("NalFragmentA Forbidden", unit.Forbidden, "NalRefIdc", unit.NalRefIdc, "PayloadType", unit.PayloadType,
	//		"S", unit.S, "E", unit.E, "R", unit.R, "NalUnitType", unit.NalUnitType, "header:", buffer[0:2])

	return index, nil
}

func (unit *NalFragmentA) Write(buffer []byte) (int, error) {
	return 0, nil
}
