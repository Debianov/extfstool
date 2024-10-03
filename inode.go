package extfs

type DefaultInodeTable struct {
	i_mode        uint16
	i_uid         uint16
	i_size        uint32
	i_atime       uint32
	i_ctime       uint32
	i_mtime       uint32
	i_dtime       uint32
	i_gid         uint16
	i_links_count uint16
	i_blocks      uint32
	i_flags       uint32
	i_osd1        uint32
	i_block       [15]uint32
	symlink       string
	extent        Extent
	i_generation  uint32
	i_file_acl    uint32
	i_dir_acl     uint32
	i_faddr       uint32
	i_osd2        []uint8
	emptyFlag     bool
}

func (i *DefaultInodeTable) parse(reader *MmapCustomReader) {
	i.setEmptyFlag(*reader)
	i.i_mode = reader.Read16le(2)
	i.i_uid = reader.Read16le(2)
	i.i_size = reader.Read32le(4)
	i.i_atime = reader.Read32le(4)
	i.i_ctime = reader.Read32le(4)
	i.i_mtime = reader.Read32le(4)
	i.i_dtime = reader.Read32le(4)
	i.i_gid = reader.Read16le(2)
	i.i_links_count = reader.Read16le(2)
	i.i_blocks = reader.Read32le(4)
	i.i_flags = reader.Read32le(4)
	i.i_osd1 = reader.Read32le(4)
	if i.isSymlink() {
		i.symlink = string(reader.ReadN(60))
	} else if (i.i_flags & EXT4EXTENTSFL) != 0 {
		i.extent.parse(reader)
	} else {
		for ind := 0; ind < 15; ind++ {
			i.i_block[ind] = reader.Read32le(4)
		}
	}
	i.i_generation = reader.Read32le(4)
	i.i_file_acl = reader.Read32le(4)
	i.i_dir_acl = reader.Read32le(4)
	i.i_faddr = reader.Read32le(4)
	i.i_osd2 = reader.ReadN(12)
}

func (i *DefaultInodeTable) setEmptyFlag(reader MmapCustomReader) {
	buf := reader.ReadN(128)
	i.emptyFlag = true
	for _, elem := range buf {
		if elem != 0 {
			i.emptyFlag = false
			break
		}
	}
}

func (i *DefaultInodeTable) isSymlink() bool {
	return (i.i_mode&0xf000) == EXT4SIFLNK && i.i_size < 60
}

func (i *DefaultInodeTable) enumBlocks(super SuperBlock, callback func(reader *MmapCustomReader) bool) bool {
	if i.isSymlink() {

	} else if i.i_flags&EXT4EXTENTSFL != 0 {
		return i.enumExtents(super, callback)
	}
	var currentBytes uint64
	for ind := 0; ind < 12 && currentBytes < uint64(i.i_size); ind++ {
		if i.i_block[ind] != 0 {
			if !callback(super.GetBlock(uint64(i.i_block[ind]))) {
				return false
			}
			currentBytes += super.Blocksize()
		}
	}
	if (i.i_block[12]) != 0 {
		if !i.enum12Block(super, currentBytes, super.GetBlock(uint64(i.i_block[12])), callback) {
			return false
		}
	}
	if (i.i_block[13]) != 0 {
		if !i.enum13Block(super, currentBytes, super.GetBlock(uint64(i.i_block[13])), callback) {
			return false
		}
	}
	if (i.i_block[14]) != 0 {
		if !i.enum14Block(super, currentBytes, super.GetBlock(uint64(i.i_block[14])), callback) {
			return false
		}
	}
	return true
}

func (i *DefaultInodeTable) enum12Block(super SuperBlock, bytes uint64, reader *MmapCustomReader, callback func(reader *MmapCustomReader) bool) bool {
	currentReader := *reader
	readerInEnd := *reader
	readerInEnd.SetCursorValue(int64(super.Blocksize()))
	for currentReader.cursorPosition < readerInEnd.cursorPosition && bytes < uint64(i.i_size) {
		currentReader.cursorPosition++
		blockNumber := currentReader.cursorPosition
		if blockNumber != 0 {
			if !callback(super.GetBlock(uint64(blockNumber))) {
				return false
			}
		}
		bytes += super.Blocksize()
	}
	return true
}

func (i *DefaultInodeTable) enum13Block(super SuperBlock, bytes uint64, reader *MmapCustomReader, callback func(reader *MmapCustomReader) bool) bool {
	currentReader := *reader
	readerInEnd := *reader
	readerInEnd.SetCursorValue(int64(super.Blocksize()))
	for currentReader.cursorPosition < readerInEnd.cursorPosition && bytes < uint64(i.i_size) {
		currentReader.cursorPosition++
		blockNumber := currentReader.cursorPosition
		if !i.enum12Block(super, bytes, super.GetBlock(uint64(blockNumber)), callback) {
			return false
		}
	}
	return true
}

func (i *DefaultInodeTable) enum14Block(super SuperBlock, bytes uint64, reader *MmapCustomReader, callback func(reader *MmapCustomReader) bool) bool {
	currentReader := *reader
	readerInEnd := *reader
	readerInEnd.SetCursorValue(int64(super.Blocksize()))
	for currentReader.cursorPosition < readerInEnd.cursorPosition && bytes < uint64(i.i_size) {
		currentReader.cursorPosition++
		blockNumber := currentReader.cursorPosition
		if !i.enum13Block(super, bytes, super.GetBlock(uint64(blockNumber)), callback) {
			return false
		}
	}
	return true
}

func (i *DefaultInodeTable) enumExtents(super SuperBlock, callback func(reader *MmapCustomReader) bool) bool {
	return i.extent.enumBlocks(super, callback)
}

func (i *DefaultInodeTable) datasize() uint64 {
	return uint64(i.i_size)
}

type DirectoryEntry struct {
	inode    uint32
	filetype uint8
	name     string
	rec_len  uint16
	name_len uint8
}

func (d *DirectoryEntry) parse(reader *MmapCustomReader) {
	d.inode = reader.Read32le(4)
	d.rec_len = reader.Read16le(2)
	d.name_len = reader.Read8(1)
	d.filetype = reader.Read8(1)
	d.name = string(reader.ReadN(int64(d.name_len)))
}
