/*
Библиотека для чтения и распаковки образов ext2/ext3/ext4 файловых систем
Перепись С++ либы https://github.com/nlitsme/extfstools
*/

package extfs

import (
	"github.com/ImSingee/mmap"
	"log"
	"os"
	"time"
)

const (
	EXT4_FT_UNKNOWN = iota
	EXT4_FT_REG_FILE
	EXT4_FT_DIR
	EXT4_FT_CHRDEV
	EXT4_FT_BLKDEV
	EXT4_FT_FIFO
	EXT4_FT_SOCK
	EXT4_FT_SYMLINK
)

const ROOTDIRINODE = 2
const EXT4SIFDIR = 0x4000
const EXT4SIFLNK = 0xa000
const EXT4EXTENTSFL = 0x00080000 /* Inode using extents */
const EXT4_FEATURE_INCOMPAT_EXTENTS = 0x40
const EXT4_FEATURE_INCOMPAT_64BIT = 0x80

type ExtFileSystem struct {
	super            SuperBlock
	bgdescs          []BlockGroupDescriptor
	bgroups          []BlockGroup
	superBlockOffset int64
}

func (e *ExtFileSystem) parse(reader MmapCustomReader) {
	reader.SetCursorValue(e.superBlockOffset)
	e.super.Parse(reader)
	if e.super.s_magic != 0xef53 {
		log.Panicf("extfs parse: not an ext2 filesystem.")
	}
	bgdescpos := e.getBlockGroupDescPosition()
	reader.SetCursorValue(int64(bgdescpos))

	if (e.super.s_feature_incompat&EXT4_FEATURE_INCOMPAT_EXTENTS != 0x0) &&
		(e.super.s_feature_incompat&EXT4_FEATURE_INCOMPAT_64BIT != 0x0) {
		e.parseGroupDescs(&reader, Ext4BlockGroupDescriptorFabric)
	} else {
		e.parseGroupDescs(&reader, DefaultBlockGroupDescriptorFabric)
	}
	e.parseBlockGroups()
}

func (e *ExtFileSystem) getBlockGroupDescPosition() (bgdescpos uint64) {
	if e.super.Blocksize() == 1024 {
		bgdescpos = 2048
	} else {
		bgdescpos = e.super.Blocksize()
	}
	return
}

func (e *ExtFileSystem) parseGroupDescs(reader *MmapCustomReader, blockGroupDescVersionFabric func() BlockGroupDescriptor) {
	ngroups := int(e.super.Ngroups())
	initialCursor := reader.cursorPosition

	for i := 0; i < ngroups; i++ {
		blockGroupDescInstance := blockGroupDescVersionFabric()
		reader.SetCursorValue(initialCursor + int64(i*blockGroupDescInstance.getSize()))
		blockGroupDescInstance.parse(reader)
		e.bgdescs = append(e.bgdescs, blockGroupDescInstance)
	}
}

func (e *ExtFileSystem) parseBlockGroups() {
	ngroups := int(e.super.Ngroups())
	for i := 0; i < ngroups; i++ {
		blockGroupInstance := BlockGroup{}
		blockGroupInstance.parse(e.super, e.bgdescs[i])
		e.bgroups = append(e.bgroups, blockGroupInstance)
	}
}

func (e *ExtFileSystem) getInode(inodeNumber uint32) DefaultInodeTable {
	inodeNumber--
	if inodeNumber >= e.super.s_inodes_count {
		inodeNumber = 0
	}
	blockGroupNumber := inodeNumber / e.super.s_inodes_per_group
	localInodeNumber := inodeNumber % e.super.s_inodes_per_group // relative to the current BlockGroup
	return e.bgroups[blockGroupNumber].getInode(localInodeNumber)
}

type FsUnpacker struct {
	fs       ExtFileSystem
	savePath string
}

func (f *FsUnpacker) perform() {
	inodeNumber := ROOTDIRINODE
	f.recurseDirs(uint32(inodeNumber), "", func(entry DirectoryEntry, currentPath string) {
		var pathForMkdir string
		if currentPath == "" {
			pathForMkdir = f.savePath + "/" + entry.name
		} else {
			pathForMkdir = f.savePath + "/" + currentPath + "/" + entry.name
		}
		if entry.filetype == EXT4_FT_DIR {
			if err := os.Mkdir(pathForMkdir, 0777); err != nil {
				log.Panicf("perform: mkdir wasn't completed: %v", err)
			}
		} else if entry.filetype == EXT4_FT_REG_FILE || entry.filetype == EXT4_FT_SYMLINK {
			f.exportInode(entry.inode, pathForMkdir)
		}
	})
}

func (f *FsUnpacker) recurseDirs(inodeNumber uint32, path string, callback func(DirectoryEntry, string)) {
	inodeTable := f.fs.getInode(inodeNumber)
	if (inodeTable.i_mode & 0xf000) != EXT4SIFDIR {
		return
	}
	inodeTable.enumBlocks(f.fs.super, func(reader *MmapCustomReader) bool {
		initialCursorPosition := reader.cursorPosition
		currentReader := *reader
		for currentReader.cursorPosition < initialCursorPosition+int64(f.fs.super.Blocksize()) {
			var e DirectoryEntry
			readerForDirectoryEntryParser := currentReader
			e.parse(&readerForDirectoryEntryParser)
			n := e.rec_len
			currentReader.cursorPosition += int64(n)
			if n == 0 {
				break
			}
			if e.filetype == EXT4_FT_UNKNOWN {
				continue
			}
			if e.name == "." || e.name == ".." {
				continue
			}
			callback(e, path)
			if e.filetype == EXT4_FT_DIR {
				var pathForRecurse string
				if path != "" {
					pathForRecurse = path + "/" + e.name
				} else {
					pathForRecurse = e.name
				}
				f.recurseDirs(e.inode, pathForRecurse, callback)
			}
		}
		return true
	})
}

func (f *FsUnpacker) exportInode(inodeNumber uint32, currentPath string) {
	var err error
	inodeTable := f.fs.getInode(inodeNumber)
	if inodeTable.emptyFlag {
		return
	}
	file, err := os.OpenFile(currentPath, os.O_CREATE|os.O_WRONLY, os.FileMode(inodeTable.i_mode))
	defer file.Close()
	if err != nil {
		log.Panicf("exportInode: Failed to create file: %v", err)
	}
	inodeTable.enumBlocks(f.fs.super, func(reader *MmapCustomReader) bool {
		_, err := file.Write(reader.ReadN(int64(f.fs.super.Blocksize())))
		if err != nil {
			log.Panicf("enumBlocks: Failed to write file: %v", err)
		}
		return true
	})
	_, err = file.Seek(int64(inodeTable.datasize()), 0)
	if err != nil {
		log.Panicf("exportInode: Failed to seek file: %v", err)
	}
	err = file.Truncate(int64(inodeTable.datasize()))
	if err != nil {
		log.Panicf("exportInode: Failed to truncate file: %v", err)
	}
	err = file.Chmod(os.FileMode(inodeTable.i_mode))
	if err != nil {
		log.Panicf("exportInode: Failed to chmod file: %v", err)
	}
	f.setTimeVal(inodeTable, currentPath)
}

func (f *FsUnpacker) setTimeVal(inode DefaultInodeTable, path string) {
	atime := time.Unix(int64(inode.i_atime), 0)
	mtime := time.Unix(int64(inode.i_mtime), 0)
	err := os.Chtimes(path, atime, mtime)
	if err != nil {
		log.Panicf("setTimeVal: Failed to chtimes: %v", err)
	}
}

func Unpack(targetPath string, pathForExtracting string) {
	file, err := mmap.New(mmap.NewReadOnly(targetPath))
	if err != nil {
		log.Panicf("extfs Unpack: %v", err)
	}
	fs := ExtFileSystem{superBlockOffset: 0x400}
	reader := MmapCustomReader{mmapInstance: file}
	fs.parse(reader)
	unpacker := FsUnpacker{fs: fs, savePath: pathForExtracting}
	unpacker.perform()
}
