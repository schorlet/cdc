package cdc

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// DiskCache reads the blocks files and maps URL to CacheAddr.
type DiskCache struct {
	dir  string               // cache directory
	addr map[uint32]CacheAddr // [entry.hash]addr
	key  []string             // []entry.key
}

// OpenCache opens the disk cache at dir.
func OpenCache(dir string) (*DiskCache, error) {
	return openCache(dir)
}

// URLs returns all the URLs currently stored.
func (cache *DiskCache) URLs() []string {
	urls := make([]string, len(cache.key))
	copy(urls, cache.key)
	return urls
}

// GetAddr returns the addr for url.
// The returned CacheAddr might not be initialized, meaning that the url is unknown.
func (cache *DiskCache) GetAddr(url string) CacheAddr {
	h := hash(url)
	return cache.addr[h]
}

// OpenURL returns the Entry for url.
func (cache *DiskCache) OpenURL(url string) (*Entry, error) {
	h := hash(url)
	addr, ok := cache.addr[h]
	if !ok {
		return nil, ErrNotFound
	}
	return OpenEntry(addr, cache.dir)
}

func openCache(dir string) (*DiskCache, error) {
	err := checkCache(dir)
	if err != nil {
		return nil, fmt.Errorf("invalid cache: %v", err)
	}

	file, err := os.Open(path.Join(dir, "index"))
	if err != nil {
		return nil, fmt.Errorf("unable to open index: %v", err)
	}
	defer file.Close()

	return readIndex(file)
}

func readIndex(file *os.File) (*DiskCache, error) {
	index := new(indexHeader)
	err := binary.Read(file, binary.LittleEndian, index)
	if err != nil {
		return nil, fmt.Errorf("unable to read index: %v", err)
	}

	if index.Magic != magicNumber {
		return nil, fmt.Errorf("bad magic number:%x, want:%x",
			index.Magic, magicNumber)
	}

	cache := &DiskCache{
		dir:  filepath.Dir(file.Name()),
		addr: make(map[uint32]CacheAddr),
		key:  make([]string, 0, index.NumEntries),
	}

	for i := index.TableLen; i > 0; i-- {
		addr := new(CacheAddr)
		err = binary.Read(file, binary.LittleEndian, addr)
		if err != nil {
			break
		}
		if addr.Initialized() {
			cache.readAddr(*addr)
		}
	}
	return cache, nil
}

func (cache *DiskCache) readAddr(addr CacheAddr) {
	entry, err := OpenEntry(addr, cache.dir)
	if err != nil {
		return
	}
	if entry.State == 0 &&
		// KeyLen may be larger, not managed
		entry.KeyLen <= blockKeyLen {

		cache.addr[entry.Hash] = addr
		cache.key = append(cache.key, entry.URL())
	}
}

func checkCache(dir string) error {
	name := path.Clean(dir)
	info, err := os.Stat(name)
	if err != nil {
		return fmt.Errorf("unable to stat %q: %v", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %q", dir)
	}

	_, err = os.Stat(path.Join(name, "index"))
	if err != nil {
		return fmt.Errorf("unable to stat index: %v", err)
	}

	// ignore err as the only possible returned error is filepath.ErrBadPattern
	blocks, _ := filepath.Glob(path.Join(name, "data_[0-3]"))
	if len(blocks) != 4 {
		return fmt.Errorf("missing block files")
	}
	return nil
}
