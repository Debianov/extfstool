package extfs

import "log"

type ExtentHeader struct {
	magic      uint16
	entries    uint16
	max        uint16
	depth      uint16
	generation uint32
}

func (e *ExtentHeader) parse(reader *MmapCustomReader) {
	e.magic = reader.Read16le(2)
	e.entries = reader.Read16le(2)
	e.max = reader.Read16le(2)
	e.depth = reader.Read16le(2)
	e.generation = reader.Read32le(4)
}

type ExtentNode interface {
	extent()
	enumBlocks(SuperBlock, func(*MmapCustomReader) bool) bool
	parse(*MmapCustomReader)
}

type ExtentLeaf struct {
	block    uint32
	len      uint16
	start_hi uint16
	start_lo uint32
}

func (e *ExtentLeaf) extent() { // sign-method

}

func (e *ExtentLeaf) parse(reader *MmapCustomReader) {
	e.block = reader.Read32le(4)
	e.len = reader.Read16le(2)
	e.start_hi = reader.Read16le(2)
	e.start_lo = reader.Read32le(4)
}

func (e *ExtentLeaf) startblock() uint64 {
	return uint64(e.start_hi)<<32 | uint64(e.start_lo)
}

func (e *ExtentLeaf) enumBlocks(super SuperBlock, cb func(reader *MmapCustomReader) bool) bool {
	blk := e.startblock()
	for i := 0; int16(i) < int16(e.len); i++ {
		if !cb(super.GetBlock(blk)) {
			return false
		}
		blk++
	}
	return true
}

type ExtentInternal struct {
	block   uint32
	leaf_lo uint32
	leaf_hi uint16
	unused  uint16
}

func (e *ExtentInternal) extent() { // sign-method

}

func (e *ExtentInternal) parse(reader *MmapCustomReader) {
	e.block = reader.Read32le(4)
	e.leaf_lo = reader.Read32le(4)
	e.leaf_hi = reader.Read16le(2)
	e.unused = reader.Read16le(2)
}

func (e *ExtentInternal) enumBlocks(super SuperBlock, cb func(reader *MmapCustomReader) bool) bool {
	return true
}

type Extent struct {
	extHeader ExtentHeader
	extents   []ExtentNode
}

func (e *Extent) parse(reader *MmapCustomReader) {
	e.extHeader.parse(reader)
	if e.extHeader.magic != 0xf30a {
		log.Panicf("parse: invalid extent hdr magic: %d\n", e.extHeader.magic)
	}
	var i uint16
	for ; i < e.extHeader.entries; i++ {
		if e.extHeader.depth == 0 {
			extentInstance := &ExtentLeaf{}
			extentInstance.parse(reader)
			e.extents = append(e.extents, extentInstance)
		} else {
			extentInstance := &ExtentInternal{}
			extentInstance.parse(reader)
			e.extents = append(e.extents, extentInstance)
		}
		reader.cursorPosition += 12
	}
}

func (e *Extent) enumBlocks(super SuperBlock, cb func(reader *MmapCustomReader) bool) bool {
	for i := 0; i < int(e.extHeader.entries); i++ {
		if !e.extents[i].enumBlocks(super, cb) {
			return false
		}
	}
	return true
}
