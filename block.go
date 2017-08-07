package cdc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

// ErrNotFound is returned if the entry is not found.
var ErrNotFound = errors.New("entry not found")

// Entry represents a HTTP response as stored in the cache.
type Entry struct {
	*entryStore
	dir string
}

// OpenEntry returns the Entry at the specified address.
func OpenEntry(addr CacheAddr, dir string) (*Entry, error) {
	b, err := readAddr(addr, dir)
	if err != nil {
		return nil, fmt.Errorf("open entry: %d, %v", addr, err)
	}

	reader := bytes.NewReader(b)
	var block entryStore

	err = binary.Read(reader, binary.LittleEndian, &block)
	if err != nil {
		return nil, fmt.Errorf("read entry: %d, %v", addr, err)
	}

	entry := Entry{entryStore: &block, dir: dir}
	return &entry, nil
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
		// offset = sizeof(
		// 		infoSize     int32
		// 		flag         int32
		// 		requestTime  int64
		// 		responseTime int64
		// )
		offset     int64 = 24
		headerSize int32
	)

	size, addr := e.DataSize[0], e.DataAddr[0]
	b, err := readAddrSize(addr, e.dir, uint32(size))
	if err != nil {
		return nil, fmt.Errorf("read header: %v", err)
	}
	reader := bytes.NewReader(b)

	_, err = reader.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("seek header: %v", err)
	}

	err = binary.Read(reader, binary.LittleEndian, &headerSize)
	if err != nil {
		return nil, fmt.Errorf("read header size: %v", err)
	}

	p := make([]byte, headerSize)
	err = binary.Read(reader, binary.LittleEndian, p)
	if err != nil {
		return nil, fmt.Errorf("read header data: %v", err)
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
	if !addr.initialized() {
		return nil, fmt.Errorf("open body: invalid adress")
	}

	if addr.separateFile() {
		name := path.Join(e.dir, addr.fileName())
		file, err := os.Open(name)
		if err != nil {
			return nil, fmt.Errorf("open body: %v", err)
		}
		return file, nil
	}

	b, err := readAddrSize(addr, e.dir, uint32(size))
	if err != nil {
		return nil, fmt.Errorf("read body: %v", err)
	}
	reader := bytes.NewReader(b)
	return ioutil.NopCloser(reader), nil
}

func readAddr(addr CacheAddr, dir string) ([]byte, error) {
	if !addr.initialized() {
		return nil, fmt.Errorf("readAddr: invalid adress")
	}
	size := addr.blockSize() * addr.numBlocks()
	return readAddrSize(addr, dir, size)
}

func readAddrSize(addr CacheAddr, dir string, size uint32) ([]byte, error) {
	if !addr.initialized() {
		return nil, fmt.Errorf("readAddr: invalid adress")
	}

	name := path.Join(dir, addr.fileName())
	if addr.separateFile() {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("readAddr: %v", err)
		}
		return data, nil
	}

	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("readAddr: %v", err)
	}
	defer close(file)

	offset := addr.startBlock()*addr.blockSize() + uint32(blockHeaderSize)
	block := make([]byte, size)

	_, err = file.ReadAt(block, int64(offset))
	if err != nil {
		return nil, fmt.Errorf("readAddr: %v", err)
	}
	return block, nil
}

func close(f *os.File) {
	if err := f.Close(); err != nil {
		log.Printf("Error closing file %s: %v\n", f.Name(), err)
	}
}
