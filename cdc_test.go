package cdc_test

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/schorlet/cdc"
)

func TestCrawl(t *testing.T) {
	cache, err := cdc.OpenCache("testcache")
	if err != nil {
		t.Fatal(err)
	}

	_, err = cache.OpenURL("http://foo.com")
	if err != cdc.ErrNotFound {
		t.Fatalf("got: %v, want: %v", err, cdc.ErrNotFound)
	}

	for _, url := range cache.URLs() {
		addr := cache.GetAddr(url)
		if !addr.Initialized() {
			t.Fatal("got: addr not initialized")
			continue
		}

		entry, err := cache.OpenURL(url)
		if err != nil {
			t.Fatal(err)
		}
		if url != entry.URL() {
			t.Fatalf("got: %s, want: %s", entry.URL(), url)
		}

		header, err := entry.Header()
		if err != nil {
			t.Fatal(err)
		}
		if len(header) == 0 {
			t.Fatal("got: empty header")
			continue
		}

		body, err := entry.Body()
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(ioutil.Discard, body)
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
		t.Fatalf("got: %s, want: %s", entry.URL(), url)
	}

	header, err := entry.Header()
	if err != nil {
		t.Fatal(err)
	}

	cl := header.Get("Content-Length")
	if cl != "5409" {
		t.Fatalf("got: %s, want: 5409", cl)
	}

	ct := header.Get("Content-Type")
	if ct != "image/png" {
		t.Fatalf("got: %s, want: image/png", ct)
	}

	body, err := entry.Body()
	if err != nil {
		t.Fatal(err)
	}
	n, err := io.Copy(ioutil.Discard, body)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5409 {
		t.Fatalf("got: %s, want: 5409", n)
	}
	body.Close()
}
