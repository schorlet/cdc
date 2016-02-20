package cdc_test

import (
	"testing"

	"github.com/schorlet/cdc"
)

func TestCrawl(t *testing.T) {
	cache, err := cdc.OpenCache("testcache")
	if err != nil {
		t.Fatal(err)
	}

	for _, url := range cache.URLs() {
		addr := cache.GetAddr(url)
		if !addr.Initialized() {
			t.Errorf("got: addr not initialized")
			continue
		}

		entry, err := cache.OpenURL(url)
		if err != nil {
			t.Fatal(err)
		}
		if url != entry.URL() {
			t.Fatal("go: %s, want: %s", entry.URL(), url)
		}

		header, err := entry.Header()
		if err != nil {
			t.Fatal(err)
		}
		if len(header) == 0 {
			t.Errorf("got: empty header")
			continue
		}

		body, err := entry.Body()
		if err != nil {
			t.Fatal(err)
		}
		body.Close()
	}
}

func TestEntry(t *testing.T) {
	addr := cdc.CacheAddr(2684420101)
	entry, err := cdc.OpenEntry(addr, "testcache")
	if err != nil {
		t.Fatal(err)
	}

	url := "https://golang.org/doc/gopher/pkg.png"
	if url != entry.URL() {
		t.Fatal("go: %s, want: %s", entry.URL(), url)
	}

	header, err := entry.Header()
	if err != nil {
		t.Fatal(err)
	}

	cl := header.Get("Content-Length")
	if cl != "5409" {
		t.Fatal("go: %s, want: %s", cl, "5409")
	}

	ct := header.Get("Content-Type")
	if ct != "image/png" {
		t.Fatal("go: %s, want: %s", ct, "image/png")
	}

	body, err := entry.Body()
	if err != nil {
		t.Fatal(err)
	}
	body.Close()
}
