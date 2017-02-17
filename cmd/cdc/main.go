// Command cdc helps reading disk cache on command line.
//
//  Usage:
//	cdc command [flag] CACHEDIR
//
// 	The commands are:
//		list        list entries
//		header      print entry header
//		body        print entry body
//
//	The flags are:
//		-url string        entry url
//		-addr string       entry addr
//
//	CACHEDIR is the path to the chromium cache directory.
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
    cdc command [flag] CACHEDIR

The commands are:
    list        list entries
    header      print entry header
    body        print entry body

The flags are:
    -url string        entry url
    -addr string       entry addr

CACHEDIR is the path to the chromium cache directory.
`

func main() {
	log.SetFlags(0)

	var cmd, url, addr, cachedir string
	parseArgs(&cmd, &url, &addr, &cachedir)

	cache, err := cdc.OpenCache(cachedir)
	if err != nil {
		log.Fatal(err)
	}

	if cmd == "list" {
		for _, url := range cache.URLs() {
			fmt.Printf("%d\t%s\n", cache.GetAddr(url), url)
		}

	} else {
		entry := openEntry(cache, url, addr, cachedir)

		if cmd == "header" {
			printHeader(entry)

		} else if cmd == "body" {
			printBody(entry)

		} else {
			log.Fatalf("unknown command: %q", cmd)
		}
	}
}

func parseArgs(cmd, url, addr, cachedir *string) {
	if len(os.Args) == 1 {
		log.Fatal(usage)
	}

	// cmd
	*cmd = os.Args[1]

	// flags
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = func() { log.Println(usage) }

	flags.StringVar(url, "url", "", "entry url")
	flags.StringVar(addr, "addr", "", "entry addr")

	err := flags.Parse(os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}

	if *cmd != "list" && flags.NFlag() != 1 {
		log.Fatal(usage)
	}

	if flags.NArg() != 1 {
		log.Fatal(usage)
	}

	*cachedir = flags.Arg(0)
}

func openEntry(cache *cdc.DiskCache, url, addr, dir string) *cdc.Entry {
	var entry *cdc.Entry
	var err error

	if addr != "" {
		id, era := strconv.ParseUint(addr, 10, 32)
		if era != nil {
			log.Fatal(era)
		}
		entry, err = cdc.OpenEntry(cdc.CacheAddr(id), dir)

	} else if url != "" {
		entry, err = cache.OpenURL(url)
	}

	if err != nil {
		log.Fatal(err)
	}

	return entry
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

	_, err = io.Copy(os.Stdout, body)
	if err != nil {
		log.Println(err)
	}
}
