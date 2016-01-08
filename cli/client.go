package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"cdc"
)

var (
	aURL  = flag.String("url", "", "url")
	aHash = flag.String("hash", "", "hash")
	aAddr = flag.String("addr", "", "addr")
)

func printUsage() {
	log.SetFlags(0)
	log.Fatal(`Usage: client [OPTION] CACHEDIR
CACHEDIR is the path to chromium cache directory.
OPTION include:
    -url URL          print cache entry for URL
    -hash HASH        print cache entry for HASH
    -addr ADDR        print hex dump of block at ADDR
If no OPTION then client prints a listing of all entries URL.
`)
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		printUsage()
	}
	name := flag.Arg(0)

	err := cdc.Init(name)
	if err != nil {
		log.Fatal(err)
	}

	if flag.NFlag() == 0 {
		for url := range cdc.Urls() {
			fmt.Println(url)
		}

	} else if *aURL != "" {
		printURL(*aURL)

	} else if *aHash != "" {
		hash, err := strconv.ParseUint(*aHash, 16, 32)
		if err != nil {
			log.Fatal(err)
		}
		printHash(uint32(hash))

	} else if *aAddr != "" {
		addr, err := strconv.ParseUint(*aAddr, 16, 32)
		if err != nil {
			log.Fatal(err)
		}
		printAddr(cdc.CacheAddr(uint32(addr)))
	}
}

func printURL(key string) {
	printHash(cdc.Hash(key))
}

func printHash(hash uint32) {
	entry, err := cdc.OpenHash(hash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(entry)
	fmt.Println()

	printHeader(entry)
	printBody(entry)
}

func printAddr(addr cdc.CacheAddr) {
	b, err := cdc.ReadAddr(addr)
	if err != nil {
		log.Fatal(err)
	}
	b = bytes.TrimRight(b, "\x00")
	fmt.Println(hex.Dump(b))
}

func printHeader(entry *cdc.EntryStore) {
	header, err := entry.OpenHeader()
	if err != nil {
		log.Fatal(err)
	}
	for key := range header {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}
	fmt.Println()
}

func printBody(entry *cdc.EntryStore) {
	body, err := entry.OpenBody()
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	out := fmt.Sprintf("cdc_%08x", entry.DataAddr[1])
	w, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	fmt.Println("writing body to:", w.Name())
	io.Copy(w, body)
}
