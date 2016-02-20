package cdc

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var (
	cacheDir  string               // cache directory
	cacheAddr map[uint32]CacheAddr // [entry.hash]addr
	cacheKey  map[uint32]string    // [entry.hash]entry.key
)

// Init reads cache at dir.
func Init(dir string) error {
	name := path.Clean(dir)
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	_, err = os.Stat(path.Join(name, "index"))
	if err != nil {
		return err
	}

	blocks, err := filepath.Glob(path.Join(name, "data_[0-3]"))
	if err != nil {
		return err
	}
	if len(blocks) != 4 {
		return fmt.Errorf("not a cache directory: %s", dir)
	}

	return initMaps(name)
}

func initMaps(dir string) error {
	file, err := os.Open(path.Join(dir, "index"))
	if err != nil {
		return err
	}
	defer file.Close()

	index := new(IndexHeader)
	err = binary.Read(file, binary.LittleEndian, index)
	if err != nil {
		return err
	}

	cacheDir = dir
	cacheAddr = make(map[uint32]CacheAddr)
	cacheKey = make(map[uint32]string)

	i := index.TableLen
	for ; i > 0; i-- {
		addr := new(CacheAddr)
		err = binary.Read(file, binary.LittleEndian, addr)
		if err != nil {
			break
		}
		addEntry(*addr)
	}
	return err
}

func addEntry(addr CacheAddr) {
	entry, err := OpenAddr(addr)
	if err == nil &&
		entry.State == 0 &&
		// KeyLen may be larger, not managed
		entry.KeyLen <= blockKeyLen {

		cacheAddr[entry.Hash] = addr
		cacheKey[entry.Hash] = entry.URL()
	}
}

// URLs returns all the URLs currently stored.
func URLs() []string {
	urls := make([]string, 0, len(cacheKey))
	for _, value := range cacheKey {
		urls = append(urls, value)
	}
	return urls
}

// GetAddr does a lookup for CacheAddr from hash.
// The return CacheAddr may be not initialized,
// meaning that the hash is invalid.
func GetAddr(hash uint32) CacheAddr {
	return cacheAddr[hash]
}
