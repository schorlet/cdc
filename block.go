package cdc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

// ErrNotFound is returned when an entry is not found in cache.
var ErrNotFound = errors.New("cdc: entry not found")

// ErrBadAddr is returned if the addr is not initialized.
var ErrBadAddr = errors.New("cdc: addr is not initialized")

// Hash returns the url hash.
func Hash(url string) uint32 {
	return superFastHash([]byte(url))
}

// OpenURL returns the EntryStore for url.
func OpenURL(url string) (*EntryStore, error) {
	hash := Hash(url)
	return OpenHash(hash)
}

// OpenHash returns the EntryStore for hash.
func OpenHash(hash uint32) (*EntryStore, error) {
	addr, ok := cacheAddr[hash]
	if !ok {
		return nil, ErrNotFound
	}
	return OpenAddr(addr)
}

// OpenAddr returns the EntryStore for addr.
func OpenAddr(addr CacheAddr) (*EntryStore, error) {
	b, err := addr.ReadAll()
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(b)
	entry := new(EntryStore)

	err = binary.Read(reader, binary.LittleEndian, entry)
	return entry, err
}

// Header returns the HTTP header.
func (e EntryStore) Header() (http.Header, error) {
	var (
		infoSize     int32
		flag         int32
		requestTime  int64
		responseTime int64
		headerSize   int32
	)

	size, addr := e.DataSize[0], e.DataAddr[0]
	b, err := readAddrSize(addr, uint32(size))
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(b)

	binary.Read(reader, binary.LittleEndian, &infoSize)
	binary.Read(reader, binary.LittleEndian, &flag)
	binary.Read(reader, binary.LittleEndian, &requestTime)
	binary.Read(reader, binary.LittleEndian, &responseTime)
	binary.Read(reader, binary.LittleEndian, &headerSize)

	// unix epoch - win epoch (µsec)
	// (1970-01-01 - 1601-01-01)
	// const delta = int64(11644473600000000)

	// fmt.Printf("infoSize:%d\n", infoSize)
	// fmt.Printf("flag:%x\n", flag)
	// fmt.Printf("requestTime:%s\n", time.Unix(0, (requestTime-delta)*1000))
	// fmt.Printf("responseTime:%s\n", time.Unix(0, (responseTime-delta)*1000))
	// fmt.Printf("headerSize:%d\n", headerSize)

	p := make([]byte, headerSize)
	binary.Read(reader, binary.LittleEndian, p)
	// fmt.Println(hex.Dump(p))

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

// Body returns the response body.
func (e EntryStore) Body() (io.ReadCloser, error) {
	size, addr := e.DataSize[1], e.DataAddr[1]
	if !addr.Initialized() {
		return nil, ErrBadAddr
	}

	if addr.SeparateFile() {
		name := path.Join(cacheDir, addr.FileName())
		return os.Open(name)
	}

	b, err := readAddrSize(addr, uint32(size))
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(b)
	return ioutil.NopCloser(reader), nil
}

// ReadAll reads the data block at addr.
// The len of the returned byte array will be addr.BlockSize() * addr.NumBlocks().
func (addr CacheAddr) ReadAll() ([]byte, error) {
	if !addr.Initialized() {
		return nil, ErrBadAddr
	}

	size := addr.BlockSize() * addr.NumBlocks()
	return readAddrSize(addr, size)
}

func readAddrSize(addr CacheAddr, size uint32) ([]byte, error) {
	if !addr.Initialized() {
		return nil, ErrBadAddr
	}

	name := path.Join(cacheDir, addr.FileName())

	if addr.SeparateFile() {
		return ioutil.ReadFile(name)
	}

	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	offset := addr.StartBlock()*addr.BlockSize() + uint32(kBlockHeaderSize)
	block := make([]byte, size)

	_, err = file.ReadAt(block, int64(offset))
	return block, err
}
