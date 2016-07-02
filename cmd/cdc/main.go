// Package cdc helps reading disk cache on command line.
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

const usage = `cdc is a tool for reading Chromium disk cache v2.

Usage:

    cdc command [arguments] CACHEDIR

The commands are:

    list        list entries
    header      print entry header
    body        print entry body

The arguments are:
    -url string        entry url
    -addr string       entry addr

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
		fmt.Println(usage)
		return
	}

	// flags
	cmdline := flag.NewFlagSet("", flag.ExitOnError)
	cmdline.Usage = printUsage
	aURL := cmdline.String("url", "", "entry url")
	aAddr := cmdline.String("addr", "", "entry addr")

	cmdline.Parse(os.Args[2:])
	if cmdline.NArg() != 1 {
		printUsage()
	}

	// init
	dir := cmdline.Arg(0)
	cache, err := cdc.OpenCache(dir)
	if err != nil {
		log.Fatal(err)
	}

	// exec
	if command == "list" {
		for _, url := range cache.URLs() {
			fmt.Printf("%d %s\n", cache.GetAddr(url), url)
		}

	} else if cmdline.NFlag() != 1 {
		printUsage()

	} else {
		var entry *cdc.Entry

		if *aURL != "" {
			entry, err = cache.OpenURL(*aURL)

		} else if *aAddr != "" {
			addr, era := strconv.ParseUint(*aAddr, 10, 32)
			if era != nil {
				log.Fatal(era)
			}
			entry, err = cdc.OpenEntry(cdc.CacheAddr(addr), dir)
		}

		if err != nil {
			log.Fatal(err)
		}

		if command == "header" {
			printHeader(entry)
		} else if command == "body" {
			printBody(entry)
		} else {
			log.Fatalf("unknown command %s", command)
		}
	}
}

func printHeader(entry *cdc.Entry) {
	header, err := entry.Header()
	if err != nil {
		log.Fatal(err)
	}
	for key := range header {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}
}

func printBody(entry *cdc.Entry) {
	body, err := entry.Body()
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	io.Copy(os.Stdout, body)
}
