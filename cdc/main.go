package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/schorlet/cdc"
)

const usage = `cdc is a tool for for reading Chromium disk cache v2.

Usage:

    cdc command [arguments] CACHEDIR

The commands are:

    list        list entries
    info        print entry info
    header      print entry header
    body        print entry body

The arguments are:
    -url string        entry url
    -hash string       entry hash

CACHEDIR is the path to the chromium cache directory.
`

func printUsage() {
	log.Println(usage)
	os.Exit(2)
}

func main() {
	log.SetFlags(0)

	if len(os.Args) == 1 {
		printUsage()
	}

	// command
	command := os.Args[1]
	if command == "-h" || command == "--help" {
		printUsage()
	}

	// flags
	cmdline := flag.NewFlagSet("", flag.ExitOnError)
	cmdline.Usage = printUsage
	aURL := cmdline.String("url", "", "entry url")
	aHash := cmdline.String("hash", "", "entry hash")

	cmdline.Parse(os.Args[2:])
	if cmdline.NArg() != 1 {
		printUsage()
	}

	// init
	dir := cmdline.Arg(0)
	err := cdc.Init(dir)
	if err != nil {
		log.Fatal(err)
	}

	// exec
	if command == "list" {
		for _, url := range cdc.Urls() {
			fmt.Println(url)
		}

	} else if cmdline.NFlag() != 1 {
		printUsage()

	} else {
		entry, err := getEntry(*aURL, *aHash)
		if err != nil {
			log.Fatal(err)
		}

		if command == "info" {
			fmt.Println(entry)
		} else if command == "header" {
			printHeader(entry)
		} else if command == "body" {
			printBody(entry)
		} else {
			log.Fatalf("unknown command %s", command)
		}
	}
}

func printHeader(entry *cdc.EntryStore) {
	header, err := entry.OpenHeader()
	if err != nil {
		log.Fatal(err)
	}
	for key := range header {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}
}

func printBody(entry *cdc.EntryStore) {
	body, err := entry.OpenBody()
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	io.Copy(os.Stdout, body)
}

func getEntry(url, hash string) (*cdc.EntryStore, error) {
	if url != "" {
		h := cdc.Hash(url)
		return cdc.OpenHash(h)

	} else if hash != "" {
		h, err := strconv.ParseUint(hash, 16, 32)
		if err != nil {
			return nil, err
		}
		return cdc.OpenHash(uint32(h))
	}
	return nil, fmt.Errorf("empty args")
}
