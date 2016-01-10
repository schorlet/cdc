package cdc_test

import (
	"testing"

	"github.com/schorlet/cdc"
)

func TestBasic(t *testing.T) {
	err := cdc.Init("testcache")
	if err != nil {
		t.Fatal(err)
	}

	for _, url := range cdc.Urls() {
		entry, err := cdc.OpenURL(url)
		if err != nil {
			t.Fatal(err)
		}

		addr := cdc.GetAddr(entry.Hash)
		if !addr.Initialized() {
			t.Errorf("got: addr not initialized")
			continue
		}

		header, err := entry.OpenHeader()
		if err != nil {
			t.Fatal(err)
		}
		if len(header) == 0 {
			t.Errorf("got: empty header")
			continue
		}

		body, err := entry.OpenBody()
		if err != nil {
			t.Fatal(err)
		}
		body.Close()
	}
}
