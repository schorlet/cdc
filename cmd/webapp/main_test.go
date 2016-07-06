package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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
		get(t, req{
			base:      base,
			view:      "https://golang.org/doc/gopher/pkg.png",
			ctype:     "image/png",
			cencoding: "",
			clength:   "5409",
			status:    http.StatusOK,
		})
		get(t, req{
			base:      base,
			view:      "https://golang.org/lib/godoc/godocs.js",
			ctype:     "application/x-javascript",
			cencoding: "gzip",
			clength:   "5186",
			status:    http.StatusOK,
		})
		get(t, req{
			base:      base,
			view:      "https://golang.org/pkg/",
			ctype:     "text/html; charset=utf-8",
			cencoding: "gzip",
			clength:   "8476",
			status:    http.StatusOK,
		})
		get(t, req{
			base:   base,
			view:   "https://golang.org/",
			status: http.StatusNotFound,
		})
	})
}

type req struct {
	base, view                string
	ctype, cencoding, clength string
	status                    int
}

func get(t *testing.T, r req) {
	q := url.Values{
		"view": []string{r.view},
	}
	u, _ := url.Parse(r.base)
	u.RawQuery = q.Encode()

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	res, err := client.Get(u.String())
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != r.status {
		t.Fatalf("got: %d, want: %d", res.StatusCode, r.status)
	}
	if res.StatusCode != http.StatusOK {
		return
	}

	cl := res.Header.Get("Content-Length")
	if cl != r.clength {
		t.Fatalf("got: %q, want: %q", cl, r.clength)
	}

	ct := res.Header.Get("Content-Type")
	if ct != r.ctype {
		t.Fatalf("got: %q, want: %q", ct, r.ctype)
	}

	ce := res.Header.Get("Content-Encoding")
	if ce != r.cencoding {
		t.Fatalf("got: %q, want: %q", ce, r.cencoding)
	}

	n, err := io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	nlength, _ := strconv.ParseInt(r.clength, 10, 64)
	if n != nlength {
		t.Fatalf("got: %d, want: %d", n, nlength)
	}
}
