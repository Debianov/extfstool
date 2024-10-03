package extfs

import (
	"bytes"
	"encoding/binary"
	"github.com/ImSingee/mmap"
	"log"
)

type MmapCustomReader struct {
	cursorPosition int64
	mmapInstance   *mmap.Mmap
}

func (m *MmapCustomReader) ReadN(offset int64) (result []byte) {
	var err error
	result = make([]byte, offset)
	_, err = m.mmapInstance.ReadAt(result, m.cursorPosition)
	if err != nil {
		log.Panicf("ReadN: %v", err)
	}
	m.cursorPosition += offset
	return
}

// Read32le reads and returns an unsigned int32 at current MmapCustomReader.cursorPosition. The offset are applied after reading.
func (m *MmapCustomReader) Read32le(offset int64) uint32 {
	defer func() { m.cursorPosition += offset }()
	return binary.LittleEndian.Uint32(m.read())
}

// Read16le reads and returns an unsigned int16 at current MmapCustomReader.cursorPosition. The offset are applied after reading.
func (m *MmapCustomReader) Read16le(offset int64) uint16 {
	defer func() { m.cursorPosition += offset }()
	return binary.LittleEndian.Uint16(m.read())
}

func (m *MmapCustomReader) Read8(offset int64) uint8 {
	defer func() { m.cursorPosition += offset }()
	var result uint8
	currentByte := m.read()
	binary.Read(bytes.NewReader(currentByte), binary.BigEndian, &result)
	return result
}

func (m *MmapCustomReader) read() (result []byte) {
	var err error
	result = make([]byte, 8)
	_, err = m.mmapInstance.ReadAt(result, m.cursorPosition)
	if err != nil {
		log.Panicf("read: %v", err)
	}
	return
}

func (m *MmapCustomReader) GetCursorValue() *int64 {
	return &m.cursorPosition
}

func (m *MmapCustomReader) SetCursorValue(n int64) {
	m.cursorPosition = n
}
