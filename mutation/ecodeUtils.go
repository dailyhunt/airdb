package mutation

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
)

const (
	SingleValueHeaderLength = 20
	TimestampSize           = 8
)

func calculateBuffSize(m *Mutation) int {
	return SingleValueHeaderLength + len(m.Key) + len(m.Family) + len(m.Col) + len(m.Value) + TimestampSize
}

// https://docs.google.com/document/d/1CsVAuxWh-jQW1AFzPF8m5Xj7aY_erSIVjjxlrOQpVQY/edit?ts=5b0bda0a#heading=h.tpkrpywuyj7w
func Encode(m *Mutation) []byte {
	buffSize := calculateBuffSize(m)
	buf := bytes.NewBuffer(make([]byte, 0, buffSize))

	kLen := len(m.Key)
	fLen := len(m.Family)
	cLen := len(m.Col)
	vLen := len(m.Value)

	// Header
	var header [SingleValueHeaderLength]byte
	binary.BigEndian.PutUint16(header[0:2], uint16(SINGLE))          // Single Value/Array
	binary.BigEndian.PutUint16(header[2:4], uint16(m.MutationType))  // Mutation Type
	binary.BigEndian.PutUint16(header[4:6], uint16(kLen))            // Key Len
	binary.BigEndian.PutUint16(header[6:8], uint16(fLen))            // Column Family Len
	binary.BigEndian.PutUint16(header[8:10], uint16(cLen))           // Column Len
	binary.BigEndian.PutUint64(header[10:18], uint64(TimestampSize)) // Bytes for TimeStamp
	binary.BigEndian.PutUint16(header[18:20], uint16(vLen))          // Value Len Type

	// Timestamp
	var enc [TimestampSize]byte
	binary.BigEndian.PutUint64(enc[:], m.Timestamp)

	// Encode to bytes
	buf.Write(header[:])
	buf.Write(m.Key)
	buf.Write(m.Family)
	buf.Write(m.Col)
	buf.Write(enc[:])
	buf.Write(m.Value)

	return buf.Bytes()
}

func Decode(buf []byte) (m *Mutation, err error) {
	if buf == nil {
		return nil, errors.New("Can not decode empty byte array")
	}

	//	dataType := binary.BigEndian.Uint16(buf[0:2])     // Single Value/Array
	mutationType := binary.BigEndian.Uint16(buf[2:4]) // Mutation Type
	kLen := int(binary.BigEndian.Uint16(buf[4:6]))    // Key Len
	fLen := int(binary.BigEndian.Uint16(buf[4:6]))    // Column Family Len
	cLen := int(binary.BigEndian.Uint16(buf[8:10]))   // Column Len
	tLen := int(binary.BigEndian.Uint64(buf[10:18]))  // Bytes for TimeStamp
	vLen := int(binary.BigEndian.Uint16(buf[18:20]))  // Value Len Type

	m = &Mutation{}

	// Key
	from := SingleValueHeaderLength

	m.Key = readBuf(buf, &from, kLen)
	m.Family = readBuf(buf, &from, fLen)
	m.Col = readBuf(buf, &from, cLen)
	m.Timestamp = binary.BigEndian.Uint64(readBuf(buf, &from, tLen))
	m.Value = readBuf(buf, &from, vLen)
	m.MutationType = Type(mutationType)

	return m, nil
}

func readBuf(buf []byte, from *int, toRead int) []byte {
	to := *from + toRead
	b := buf[*from:to]
	*from = to
	return b
}
