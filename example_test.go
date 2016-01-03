package cdc_test

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"cdc"
)

func Example() {
	err := cdc.Init("cache")
	if err != nil {
		log.Fatal(err)
	}

	entry, err := cdc.OpenURL("https://golang.org/pkg/")
	if err != nil {
		log.Fatal(err)
	}

	header, err := entry.OpenHeader()
	if err != nil {
		log.Fatal(err)
	}
	for key := range header {
		fmt.Printf("%s: %s\n", key, header.Get(key))
	}

	body, err := entry.OpenBody()
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	dumper := hex.Dumper(os.Stdout)
	defer dumper.Close()

	io.Copy(dumper, body)
}
