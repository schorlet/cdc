package cdc

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
)

// Cache gives read access to the chromium disk cache.
//
// The cache is composed of one "index" file, four or more "data_[0-9]" files
// and many of "f_[0-9]+" separate files.
//
// Learn more:
// http://www.forensicswiki.org/wiki/Google_Chrome#Disk_Cache
// http://www.forensicswiki.org/wiki/Chrome_Disk_Cache_Format
type Cache struct {
	dir  string          // cache directory
	addr map[uint32]Addr // [entry.hash]addr
	urls []string        // []entry.key
}

// URLs returns all the URLs currently stored.
func (c *Cache) URLs() []string {
	urls := make([]string, len(c.urls))
	copy(urls, c.urls)
	return urls
}

// GetAddr returns the address of the URL.
// An error is returned if the URL is not found.
func (c *Cache) GetAddr(url string) (Addr, error) {
	hash := superFastHash([]byte(url))
	addr, ok := c.addr[hash]
	if !ok {
		return addr, ErrNotFound
	}
	return addr, nil
}

// OpenURL returns the Entry for the specified URL.
// An error is returned if the URL is not found.
func (c *Cache) OpenURL(url string) (*Entry, error) {
	addr, err := c.GetAddr(url)
	if err != nil {
		return nil, err
	}
	entry, err := OpenEntry(addr, c.dir)
	if err != nil {
		return nil, fmt.Errorf("open url %s: %v", url, err)
	}
	return entry, nil
}

// OpenCache opens the cache in dir.
// Opens the "index" file to read the addresses and then
// opens each Entry to read the URL and associate it to an address.
func OpenCache(dir string) (*Cache, error) {
	err := checkCache(dir)
	if err != nil {
		return nil, fmt.Errorf("invalid cache: %s, %v", dir, err)
	}

	file, err := os.Open(path.Join(dir, "index"))
	if err != nil {
		return nil, fmt.Errorf("open cache: %v", err)
	}
	defer close(file)

	var index indexHeader
	err = binary.Read(file, binary.LittleEndian, &index)
	if err != nil {
		return nil, fmt.Errorf("open cache: %v", err)
	}
	if index.Magic != magicNumber {
		return nil, fmt.Errorf("magic: %x, want: %x",
			index.Magic, magicNumber)
	}

	cache := Cache{
		dir:  filepath.Dir(file.Name()),
		addr: make(map[uint32]Addr),
		urls: make([]string, 0, index.NumEntries),
	}

	var addr Addr
	for i := index.TableLen; i > 0; i-- {
		err = binary.Read(file, binary.LittleEndian, &addr)
		if err != nil {
			return nil, fmt.Errorf("open cache: %v", err)
		}
		if addr.initialized() {
			cache.readAddr(addr)
		}
	}
	return &cache, nil
}

func (c *Cache) readAddr(addr Addr) {
	entry, err := OpenEntry(addr, c.dir)
	if err != nil {
		log.Printf("open cache: %v", err)
		return
	}
	if entry.State == 0 &&
		// KeyLen may be larger, not managed
		entry.KeyLen <= blockKeyLen {

		c.addr[entry.Hash] = addr
		c.urls = append(c.urls, entry.URL())
	}
}

func checkCache(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}

	_, err = os.Stat(path.Join(dir, "index"))
	if err != nil {
		return err
	}

	// ignore err as the only possible returned error is filepath.ErrBadPattern
	blocks, _ := filepath.Glob(path.Join(dir, "data_[0-3]"))
	if len(blocks) != 4 {
		return fmt.Errorf("missing block files")
	}
	return nil
}
