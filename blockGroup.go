package extfs

type BlockGroupDescriptor interface {
	parse(reader *MmapCustomReader)
	getLocalInodeTableStartBlock() uint64
	getSize() int
}

type DefaultBlockGroupDescriptor struct {
	size                 int
	bg_block_bitmap      uint32
	bg_inode_bitmap      uint32
	bg_inode_table       uint32
	bg_free_blocks_count uint16
	bg_free_inodes_count uint16
	bg_used_dirs_count   uint16
	bg_pad               uint16
}

func (b *DefaultBlockGroupDescriptor) parse(reader *MmapCustomReader) {
	b.bg_block_bitmap = reader.Read32le(4)
	b.bg_inode_bitmap = reader.Read32le(4)
	b.bg_inode_table = reader.Read32le(4)
	b.bg_free_blocks_count = reader.Read16le(2)
	b.bg_free_inodes_count = reader.Read16le(2)
	b.bg_used_dirs_count = reader.Read16le(2)
	b.bg_pad = reader.Read16le(2)
	b.size = 32
}

func (b *DefaultBlockGroupDescriptor) getLocalInodeTableStartBlock() uint64 {
	return uint64(b.bg_inode_table)
}

func (b *DefaultBlockGroupDescriptor) getSize() int {
	return b.size
}

func DefaultBlockGroupDescriptorFabric() BlockGroupDescriptor {
	return &DefaultBlockGroupDescriptor{size: 32}
}

type Ext4BlockGroupDescriptor struct {
	size                    int
	bg_block_bitmap_lo      uint32
	bg_inode_bitmap_lo      uint32
	bg_inode_table_lo       uint32
	bg_free_blocks_count_lo uint16
	bg_free_inodes_count_lo uint16
	bg_used_dirs_count_lo   uint16
	bg_flags                uint16
	bg_exclude_bitmap_lo    uint32
	bg_block_bitmap_csum_lo uint16
	bg_inode_bitmap_csum_lo uint16
	bg_itable_unused_lo     uint16
	bg_checksum             uint16
	bg_block_bitmap_hi      uint32
	bg_inode_bitmap_hi      uint32
	bg_inode_table_hi       uint32
	bg_free_blocks_count_hi uint16
	bg_free_inodes_count_hi uint16
	bg_used_dirs_count_hi   uint16
	bg_itable_unused_hi     uint16
	bg_exclude_bitmap_hi    uint32
	bg_block_bitmap_csum_hi uint16
	bg_inode_bitmap_csum_hi uint16
	bg_reserved             uint32
}

func (e *Ext4BlockGroupDescriptor) parse(reader *MmapCustomReader) {
	e.bg_block_bitmap_lo = reader.Read32le(4)
	e.bg_inode_bitmap_lo = reader.Read32le(4)
	e.bg_inode_table_lo = reader.Read32le(4)
	e.bg_free_blocks_count_lo = reader.Read16le(2)
	e.bg_free_inodes_count_lo = reader.Read16le(2)
	e.bg_used_dirs_count_lo = reader.Read16le(2)
	e.bg_flags = reader.Read16le(2)
	e.bg_exclude_bitmap_lo = reader.Read32le(4)
	e.bg_block_bitmap_csum_lo = reader.Read16le(2)
	e.bg_inode_bitmap_csum_lo = reader.Read16le(2)
	e.bg_itable_unused_lo = reader.Read16le(2)
	e.bg_checksum = reader.Read16le(2)
	e.bg_block_bitmap_hi = reader.Read32le(4)
	e.bg_inode_bitmap_hi = reader.Read32le(4)
	e.bg_inode_table_hi = reader.Read32le(4)
	e.bg_free_blocks_count_hi = reader.Read16le(2)
	e.bg_free_inodes_count_hi = reader.Read16le(2)
	e.bg_used_dirs_count_hi = reader.Read16le(2)
	e.bg_itable_unused_hi = reader.Read16le(2)
	e.bg_exclude_bitmap_hi = reader.Read32le(4)
	e.bg_block_bitmap_csum_hi = reader.Read16le(2)
	e.bg_inode_bitmap_csum_hi = reader.Read16le(2)
	e.bg_reserved = reader.Read32le(4)
	e.size = 64
}

func (e *Ext4BlockGroupDescriptor) getLocalInodeTableStartBlock() uint64 {
	return (uint64(e.bg_inode_table_hi) << 32) | uint64(e.bg_inode_table_lo)
}

func (e *Ext4BlockGroupDescriptor) getSize() int {
	return e.size
}

func Ext4BlockGroupDescriptorFabric() BlockGroupDescriptor {
	return &Ext4BlockGroupDescriptor{size: 64}
}

type BlockGroup struct {
	itableoffset uint64
	ninodes      uint64
	inodesize    uint64
	reader       MmapCustomReader
}

func (b *BlockGroup) parse(super SuperBlock, descBlock BlockGroupDescriptor) {
	b.reader = super.reader
	b.itableoffset = descBlock.getLocalInodeTableStartBlock() * super.Blocksize() // не дошла логика умножения,  но оставлю,
	// чтоб работало
	b.ninodes = uint64(super.s_inodes_per_group)
	b.inodesize = uint64(super.s_inode_size)
}

func (b *BlockGroup) getInode(inodeNum uint32) (inode DefaultInodeTable) {
	b.reader.SetCursorValue(int64(b.itableoffset + b.inodesize*uint64(inodeNum)))
	inode.parse(&b.reader)
	return
}
