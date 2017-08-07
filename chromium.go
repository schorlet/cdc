// Package cdc provides support for reading Chromium disk cache v2.
// https://www.chromium.org/developers/design-documents/network-stack/disk-cache
package cdc

// Helpful resources:
// https://chromium.googlesource.com/chromium/src/net/+/master/disk_cache/blockfile/disk_format_base.h
// https://chromium.googlesource.com/chromium/src/net/+/master/disk_cache/blockfile/disk_format.h
// https://chromium.googlesource.com/chromium/src/net/+/master/disk_cache/blockfile/addr.h
// https://chromium.googlesource.com/chromium/src/net/+/master/http/http_response_info.cc
// https://chromium.googlesource.com/chromium/src/base/+/master/pickle.cc
// http://chip-dfir.techanarchy.net/?p=8

import (
	"encoding/binary"
	"fmt"
	"log"
)

// IndexHeader
const magicNumber uint32 = 0xc103cac3

// BlockFileHeader
const blockHeaderSize int = 8192
const maxBlocks int = (blockHeaderSize - 80) * 8

// EntryStore
const blockKeyLen int32 = 256 - 24*4

// CacheAddr
const initializedMask uint32 = 0x80000000
const fileTypeMask uint32 = 0x70000000
const fileTypeOffset uint32 = 28
const fileNameMask uint32 = 0x0fffffff
const fileSelectorMask uint32 = 0x00ff0000
const fileSelectorOffset uint32 = 16
const startBlockMask uint32 = 0x0000ffff
const numBlocksMask uint32 = 0x03000000
const numBlocksOffset uint32 = 24

// indexHeader for the master index file.
type indexHeader struct {
	Magic      uint32
	Version    uint32
	NumEntries int32     // Number of entries currently stored.
	NumBytes   int32     // Total size of the stored data.
	LastFile   int32     // Last external file created.
	ThisID     int32     // Id for all entries being changed (dirty flag).
	Stats      CacheAddr // Storage for usage data.
	TableLen   int32     // Actual size of the table (0 == kIndexTablesize).
	Crash      int32     // Signals a previous crash.
	Experiment int32     // Id of an ongoing test.
	CreateTime uint64    // Creation time for this set of files.
	Pad        [52]int32
	Lru        [28]int32 // Eviction control data.
}

// blockFileHeader is the header of a block-file.
// A block-file is the file used to store information in blocks (could be
// EntryStore blocks, RankingsNode blocks or user-data blocks).
type blockFileHeader struct {
	Magic         uint32
	Version       uint32
	ThisFile      int16    // Index of this file.
	NextFile      int16    // Next file when this one is full.
	EntrySize     int32    // Size of the blocks of this file.
	NumEntries    int32    // Number of stored entries.
	MaxEntries    int32    // Current maximum number of entries.
	Empty         [4]int32 // Counters of empty entries for each type.
	Hints         [4]int32 // Last used position for each entry type.
	Updating      int32    // Keep track of updates to the header.
	User          [5]int32
	AllocationMap [maxBlocks / 32]uint32 // 2028, to track used blocks on a block-file.
}

// entryStore is the main structure for an entry on the backing storage.
//
// Breakdown of the metadata:
//  0c 18 5b c2 00 00 00 00  2e 1d 00 90 06 00 00 00  |..[.............|
//  hash        next         ranking     reuse_count
//  00 00 00 00 00 00 00 00  5b f0 8e 69 de 85 2e 00  |........[..i....|
//  refetch     state        creation_time
//  33 00 00 00 00 00 00 00  3c 12 00 00 d1 51 00 00  |3.......<....Q..|
//  key_len     long_key     data_size
//  00 00 00 00 00 00 00 00  de 70 03 c1 87 25 04 80  |.........p...%..|
//  data_size                data_addr
//  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//  data_addr                flags       pad
//  00 00 00 00 00 00 00 00  00 00 00 00 b1 e0 d4 cf  |................|
//  pad                      pad         self_hash
//  68 74 74 70 73 3a 2f 2f  77 77 77 2e 72 65 74 68  |https://www.reth|
//  69 6e 6b 64 62 2e 63 6f  6d 2f 64 6f 63 73 2f 63  |inkdb.com/docs/c|
//  6f 6f 6b 62 6f 6f 6b 2f  6a 61 76 61 73 63 72 69  |ookbook/javascri|
//  70 74 2f 00 00 00 00 00  00 00 00 00 00 00 00 00  |pt/.............|
//  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
type entryStore struct {
	Hash         uint32    // Full hash of the key.
	Next         CacheAddr // Next entry with the same hash or bucket.
	RankingsNode CacheAddr // Rankings node for this entry.
	ReuseCount   int32     // How often is this entry used.
	RefetchCount int32     // How often is this fetched from the net.
	State        int32     // Current state.
	CreationTime uint64
	KeyLen       int32
	LongKey      CacheAddr    // Optional address of a long key.
	DataSize     [4]int32     // We can store up to 4 data streams for each
	DataAddr     [4]CacheAddr // entry.
	Flags        uint32       // Any combination of EntryFlags.
	Pad          [4]int32
	SelfHash     uint32            // The hash of EntryStore up to this point.
	Key          [blockKeyLen]byte // null terminated
}

// CacheAddr defines a storage address for an Entry.
type CacheAddr uint32

// Initialized returns the initialization state.
func (addr CacheAddr) Initialized() bool {
	return (uint32(addr) & initializedMask) != 0
}

// SeparateFile returns true if the cache record
// is located in a separated file.
func (addr CacheAddr) SeparateFile() bool {
	return (uint32(addr) & fileTypeMask) == 0
}

// FileType returns one of these values:
//  EXTERNAL = 0,
//  RANKINGS = 1,
//  BLOCK_256 = 2,
//  BLOCK_1K = 3,
//  BLOCK_4K = 4,
//  BLOCK_FILES = 5,
//  BLOCK_ENTRIES = 6,
//  BLOCK_EVICTED = 7
func (addr CacheAddr) FileType() uint32 {
	return (uint32(addr) & fileTypeMask) >> fileTypeOffset
}

// FileNumber returns the file number.
func (addr CacheAddr) FileNumber() uint32 {
	if addr.SeparateFile() {
		return uint32(addr) & fileNameMask
	}
	return (uint32(addr) & fileSelectorMask) >> fileSelectorOffset
}

// FileName returns the file name.
func (addr CacheAddr) FileName() (name string) {
	if !addr.Initialized() {
		// ""
	} else if addr.SeparateFile() {
		name = fmt.Sprintf("f_%06x", addr.FileNumber())
	} else {
		name = fmt.Sprintf("data_%d", addr.FileNumber())
	}
	return
}

// StartBlock returns the start block.
func (addr CacheAddr) StartBlock() uint32 {
	if addr.SeparateFile() {
		return 0
	}
	return uint32(addr) & startBlockMask
}

// BlockSize returns the block size.
func (addr CacheAddr) BlockSize() uint32 {
	switch addr.FileType() {
	case 1: // RANKINGS
		return 36
	case 2: // BLOCK_256
		return 256
	case 3: // BLOCK_1K
		return 1024
	case 4: // BLOCK_4K
		return 4096
	case 5: // BLOCK_FILES
		return 8
	case 6: // BLOCK_ENTRIES
		return 104
	case 7: // BLOCK_EVICTED
		return 48
	}
	return 0 // EXTERNAL
}

// NumBlocks returns the number of blocks.
func (addr CacheAddr) NumBlocks() uint32 {
	if addr.SeparateFile() {
		return 0
	}
	return ((uint32(addr) & numBlocksMask) >> numBlocksOffset) + 1
}

func init() {
	var ih indexHeader
	if n := binary.Size(ih); n != 368 {
		log.Fatalf("IndexHeader size error: %d, want: 368", n)
	}

	var bh blockFileHeader
	if n := binary.Size(bh); n != blockHeaderSize {
		log.Fatalf("BlockFileHeader size error: %d, want: %d", n, blockHeaderSize)
	}

	var entry entryStore
	if n := binary.Size(entry); n != 256 {
		log.Fatalf("EntryStore size error: %d, want: 256", n)
	}
}
