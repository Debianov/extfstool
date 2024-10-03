package extfs

import "log"

type SuperBlock struct {
	reader              MmapCustomReader
	s_inodes_count      uint32
	s_blocks_count      uint32
	s_r_blocks_count    uint32
	s_free_blocks_count uint32
	s_free_inodes_count uint32
	s_first_data_block  uint32 // 0 or 1
	s_log_block_size    uint32 // blocksize = 1024<<s_log_block_size
	s_log_frag_size     int32  // fragsize = 1024<<s_log_frag_size
	s_blocks_per_group  uint32
	s_frags_per_group   uint32
	s_inodes_per_group  uint32
	s_mtime             uint32
	s_wtime             uint32
	s_mnt_count         uint16
	s_max_mnt_count     uint16
	s_magic             uint16
	s_state             uint16
	s_errors            uint16
	s_minor_rev_level   uint16
	s_lastcheck         uint32
	s_checkinterval     uint32
	s_creator_os        uint32
	s_rev_level         uint32
	s_def_resuid        uint16
	s_def_resgid        uint16
	s_first_ino         uint32
	s_inode_size        uint16
	s_block_group_nr    uint16
	s_feature_compat    uint32
	s_feature_incompat  uint32
	s_feature_ro_compat uint32
	s_uuid              []uint8
	s_volume_name       []uint8
	s_last_mounted      []uint8
	s_algo_bitmap       uint32
}

func (e *SuperBlock) Parse(reader MmapCustomReader) {
	e.reader = reader
	e.s_inodes_count = reader.Read32le(4)
	e.s_blocks_count = reader.Read32le(4)
	e.s_r_blocks_count = reader.Read32le(4)
	e.s_free_blocks_count = reader.Read32le(4)
	e.s_free_inodes_count = reader.Read32le(4)
	e.s_first_data_block = reader.Read32le(4)
	e.s_log_block_size = reader.Read32le(4)
	e.s_log_frag_size = int32(reader.Read32le(4))
	e.s_blocks_per_group = reader.Read32le(4)
	e.s_frags_per_group = reader.Read32le(4)
	e.s_inodes_per_group = reader.Read32le(4)
	e.s_mtime = reader.Read32le(4)
	e.s_wtime = reader.Read32le(4)
	e.s_mnt_count = reader.Read16le(2)
	e.s_max_mnt_count = reader.Read16le(2)
	e.s_magic = reader.Read16le(2)
	e.s_state = reader.Read16le(2)
	e.s_errors = reader.Read16le(2)
	e.s_minor_rev_level = reader.Read16le(2)
	e.s_lastcheck = reader.Read32le(4)
	e.s_checkinterval = reader.Read32le(4)
	e.s_creator_os = reader.Read32le(4)
	e.s_rev_level = reader.Read32le(4)
	e.s_def_resuid = reader.Read16le(2)
	e.s_def_resgid = reader.Read16le(2)
	e.s_first_ino = reader.Read32le(4)
	e.s_inode_size = reader.Read16le(2)
	e.s_block_group_nr = reader.Read16le(2)
	e.s_feature_compat = reader.Read32le(4)
	e.s_feature_incompat = reader.Read32le(4)
	e.s_feature_ro_compat = reader.Read32le(4)
	e.s_uuid = reader.ReadN(16)
	e.s_volume_name = reader.ReadN(16)
	e.s_last_mounted = reader.ReadN(64)
	e.s_algo_bitmap = reader.Read32le(4)
}

func (e *SuperBlock) Blocksize() uint64 {
	return 1024 << e.s_log_block_size
}

func (e *SuperBlock) Ngroups() uint32 {
	return e.s_inodes_count / e.s_inodes_per_group
}

func (e *SuperBlock) BytesPerGroup() uint64 {
	return uint64(e.s_blocks_per_group) * e.Blocksize()
}

func (e *SuperBlock) GetBlock(n uint64) *MmapCustomReader {
	if uint32(n) >= e.s_blocks_count {
		log.Fatal("getBlock: blocknr too large")
	}
	e.reader.SetCursorValue(int64(e.Blocksize()) * int64(n))
	return &e.reader
}
