package gopack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// MaxTime maximum datetime
// var MaxTime = time.Date(2500, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

// ErrDecode means that a exception at the time of decoding
var ErrDecode = errors.New("decode error")

// MsgTypeSend message enum type
const MsgTypeSend = 0x1

// MsgTypeAck message enum type
const MsgTypeAck = 0x2

// MsgTypeReceived message enum type
const MsgTypeReceived = 0x3

// MsgTypeRelease message enum type
const MsgTypeRelease = 0x4

// MsgTypeCompleted message enum type
const MsgTypeCompleted = 0x5

// Qos0 quality of service level 0 (at most once)
const Qos0 = 0

// Qos1 quality of service level 1 (at least once)
const Qos1 = 1

// Qos2 quality of service level 2 (only once)
const Qos2 = 2

// Packet is a struct to hold a message
type Packet struct {
	MsgType         byte
	Qos             byte
	Dup             bool
	MsgID           uint16
	RemainingLength uint16
	TotalLength     uint16
	Payload         []byte
	Buffer          []byte

	// used to storage
	Confirm    bool
	RetryTimes int
	Timestamp  int64
}

// Clone copy packet
func (packet *Packet) Clone() (copyPacket *Packet) {
	copyPacket = new(Packet)
	copyPacket.MsgType = packet.MsgType
	copyPacket.Qos = packet.Qos
	copyPacket.Dup = packet.Dup
	copyPacket.MsgID = packet.MsgID
	copyPacket.RemainingLength = packet.RemainingLength
	copyPacket.TotalLength = packet.TotalLength
	copyPacket.Payload = packet.Payload
	copyPacket.Buffer = packet.Buffer
	copyPacket.Confirm = packet.Confirm
	copyPacket.RetryTimes = packet.RetryTimes
	copyPacket.Timestamp = packet.Timestamp
	return copyPacket
}

// Encode is used to convert bytes to packet struct
func Encode(msgType byte, qos byte, dup byte, msgID uint16, payload []byte) *Packet {
	var remainingLength uint16
	if payload != nil {
		remainingLength = uint16(len(payload))
	}
	var buffer bytes.Buffer
	fixedHeader := byte((msgType << 4) | (qos << 2) | (dup << 1))
	buffer.WriteByte(fixedHeader)
	buffer.Write(encodeUint16(msgID))
	buffer.Write(encodeUint16(remainingLength))
	if payload != nil {
		buffer.Write(payload)
	}
	return &Packet{
		MsgType:         msgType,
		Qos:             qos,
		Dup:             byteToBool(dup),
		MsgID:           msgID,
		RemainingLength: remainingLength,
		TotalLength:     5 + remainingLength,
		Payload:         payload,
		Buffer:          buffer.Bytes(),
		Timestamp:       0,
	}
}

// Decode is used to convert packet struct to bytes
func Decode(buf []byte) (packet *Packet, err error) {
	buffer := bytes.NewBuffer(buf)
	fixedHeader, err := buffer.ReadByte()
	if err == io.EOF {
		return nil, ErrDecode
	}
	packet = new(Packet)
	packet.MsgType = fixedHeader >> 4
	packet.Qos = (fixedHeader & 0xf) >> 2
	packet.Dup = byteToBool((fixedHeader & 0x3) >> 1)
	packet.MsgID, err = decodeUint16(buffer)
	if err == ErrDecode {
		return nil, ErrDecode
	}
	packet.RemainingLength, err = decodeUint16(buffer)
	if err == ErrDecode {
		return nil, ErrDecode
	}
	packet.Payload = make([]byte, packet.RemainingLength)
	n, err := buffer.Read(packet.Payload)
	if err == io.EOF || n != int(packet.RemainingLength) {
		return nil, ErrDecode
	}
	packet.Buffer = buf
	packet.Timestamp = 0
	return packet, nil
}

func boolToByte(b bool) byte {
	switch b {
	case true:
		return 1
	default:
		return 0
	}
}

func byteToBool(b byte) bool {
	switch b {
	case 1:
		return true
	default:
		return false
	}
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

func decodeUint16(b io.Reader) (i uint16, err error) {
	num := make([]byte, 2)
	n, err := b.Read(num)
	if err == io.EOF || n != 2 {
		return i, ErrDecode
	}
	return binary.BigEndian.Uint16(num), nil
}
