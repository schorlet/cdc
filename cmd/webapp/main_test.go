package main

import (
	"image/png"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/schorlet/cdc"
)

func withContext(fn func(url string)) {
	cache, err := cdc.OpenCache("../../testcache")
	if err != nil {
		log.Fatal(err)
	}

	handler := CacheHandler(cache)
	server := httptest.NewServer(handler)
	defer server.Close()

	fn(server.URL)
}

func TestView(t *testing.T) {
	withContext(func(base string) {
		q := url.Values{"view": []string{"https://golang.org/doc/gopher/pkg.png"}}
		u, _ := url.Parse(base)
		u.RawQuery = q.Encode()

		res, err := http.Get(u.String())
		if err != nil {
			t.Fatal(err)
		}

		cl := res.Header.Get("Content-Length")
		if cl != "5409" {
			t.Fatal("go: %s, want: %s", cl, "5409")
		}

		ct := res.Header.Get("Content-Type")
		if ct != "image/png" {
			t.Fatal("go: %s, want: %s", ct, "image/png")
		}

		config, err := png.DecodeConfig(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if config.Width != 83 || config.Height != 120 {
			t.Fatalf("got: %d x %d, want: 83 x 120", config.Width, config.Height)
		}
	})
}
