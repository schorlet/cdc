package cdc_test

import (
	"fmt"
	"image/png"
	"log"

	"github.com/schorlet/cdc"
)

// Example gets an entry from the cache and prints informations to stdout.
func Example() {
	err := cdc.Init("testcache")
	if err != nil {
		log.Fatal(err)
	}

	entry, err := cdc.OpenURL("https://golang.org/doc/gopher/pkg.png")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(entry)

	header, err := entry.Header()
	if err != nil {
		log.Fatal(err)
	}
	for _, key := range []string{"Status", "Content-Length", "Content-Type"} {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}

	body, err := entry.Body()
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	config, err := png.DecodeConfig(body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("PNG image data, %d x %d\n", config.Width, config.Height)

	// Output:
	// Hash:6cc46ffd Next:0 RankingsNode:90000003 ReuseCount:0 RefetchCount:0 State:0 CreationTime:13096853902296629 KeyLen:37 LongKey:0 DataSize:[6912 5409 0 0] DataAddr:[c103000a c103000c 0 0] Flags:0 SelfHash:e10f47ed Key:https://golang.org/doc/gopher/pkg.png
	// Status: 200
	// Content-Length: 5409
	// Content-Type: image/png
	// PNG image data, 83 x 120
}
