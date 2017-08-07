package cdc_test

import (
	"io"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/schorlet/cdc"
)

func TestCrawl(t *testing.T) {
	cache, err := cdc.OpenCache("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, url := range cache.URLs() {
		entry, err := cache.OpenURL(url)
		if err != nil {
			t.Fatal(err)
		}
		if entry.URL() != url {
			t.Fatalf("url: %s, want: %s", entry.URL(), url)
		}

		header, err := entry.Header()
		if err != nil {
			t.Fatalf("header: %v", err)
		}
		if len(header) == 0 {
			t.Fatal("header is empty")
		}

		clength := header.Get("Content-Length")
		nlength, err := strconv.ParseInt(clength, 10, 64)
		if err != nil {
			t.Fatal(err)
		}

		body, err := entry.Body()
		if err != nil {
			t.Fatalf("body: %v", err)
		}
		n, err := io.Copy(ioutil.Discard, body)
		if err != nil {
			t.Fatalf("discard body: %v", err)
		}
		_ = body.Close()

		if n != nlength {
			t.Fatalf("body stream-length: %d, want: %d", n, nlength)
		}
	}
}

func TestBadURL(t *testing.T) {
	cache, err := cdc.OpenCache("testdata")
	if err != nil {
		t.Fatal(err)
	}

	_, err = cache.OpenURL("http://foo.com")
	if err == nil || err != cdc.ErrNotFound {
		t.Fatalf("err: %v, want: %v", err, cdc.ErrNotFound)
	}
}
