package cdc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

// ErrNotFound is returned when an entry is not found in cache.
var ErrNotFound = errors.New("entry not found")

// ErrBadAddr is returned if the addr is not initialized.
var ErrBadAddr = errors.New("addr is not initialized")

// Entry represents a cache record as stored in the disk cache.
type Entry struct {
	*entryStore
	dir string
}

// OpenEntry returns the Entry at addr, in the disk cache at dir.
func OpenEntry(addr CacheAddr, dir string) (*Entry, error) {
	b, err := readAddr(addr, dir)
	if err != nil {
		return nil, fmt.Errorf("unable to open entry: %v", err)
	}

	reader := bytes.NewReader(b)
	block := new(entryStore)

	err = binary.Read(reader, binary.LittleEndian, block)
	if err != nil {
		return nil, fmt.Errorf("unable to read entry: %v", err)
	}

	entry := &Entry{entryStore: block, dir: dir}
	return entry, nil
}

// URL returns the entry URL.
func (e *Entry) URL() string {
	var key []byte
	if e.LongKey == 0 {
		if e.KeyLen <= blockKeyLen {
			key = e.Key[0:e.KeyLen]
		} else {
			// KeyLen may be larger, return trimmed
			key = e.Key[:]
		}
	}
	return string(key)
}

// Header returns the HTTP header.
func (e *Entry) Header() (http.Header, error) {
	var (
		// infoSize     int32
		// flag         int32
		// requestTime  int64
		// responseTime int64
		// offset = sizeof(infoSize+flag+requestTime+responseTime)
		offset     int64 = 24
		headerSize int32
	)

	size, addr := e.DataSize[0], e.DataAddr[0]
	b, err := readAddrSize(addr, e.dir, uint32(size))
	if err != nil {
		return nil, fmt.Errorf("unable to read header: %v", err)
	}
	reader := bytes.NewReader(b)

	_, err = reader.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("unable to seek header: %v", err)
	}

	err = binary.Read(reader, binary.LittleEndian, &headerSize)
	if err != nil {
		return nil, fmt.Errorf("unable to read header size: %v", err)
	}

	p := make([]byte, headerSize)
	err = binary.Read(reader, binary.LittleEndian, p)
	if err != nil {
		return nil, fmt.Errorf("unable to read header data: %v", err)
	}

	header := make(http.Header)
	lines := bytes.Split(p, []byte{0})

	for _, line := range lines {
		kv := bytes.SplitN(line, []byte{':'}, 2)
		if len(kv) == 2 {
			header.Add(
				string(bytes.TrimSpace(kv[0])),
				string(bytes.TrimSpace(kv[1])))
		}
	}
	return header, nil
}

// Body returns the HTTP body.
func (e *Entry) Body() (io.ReadCloser, error) {
	size, addr := e.DataSize[1], e.DataAddr[1]
	if !addr.Initialized() {
		return nil, ErrBadAddr
	}

	if addr.SeparateFile() {
		name := path.Join(e.dir, addr.FileName())
		return os.Open(name)
	}

	b, err := readAddrSize(addr, e.dir, uint32(size))
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %v", err)
	}
	reader := bytes.NewReader(b)
	return ioutil.NopCloser(reader), nil
}

func readAddr(addr CacheAddr, dir string) ([]byte, error) {
	if !addr.Initialized() {
		return nil, ErrBadAddr
	}

	size := addr.BlockSize() * addr.NumBlocks()
	return readAddrSize(addr, dir, size)
}

func readAddrSize(addr CacheAddr, dir string, size uint32) ([]byte, error) {
	if !addr.Initialized() {
		return nil, ErrBadAddr
	}

	name := path.Join(dir, addr.FileName())

	if addr.SeparateFile() {
		return ioutil.ReadFile(name)
	}

	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	offset := addr.StartBlock()*addr.BlockSize() + uint32(blockHeaderSize)
	block := make([]byte, size)

	_, err = file.ReadAt(block, int64(offset))
	return block, err
}

func hash(s string) uint32 {
	return superFastHash([]byte(s))
}
