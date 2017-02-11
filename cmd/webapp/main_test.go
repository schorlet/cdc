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

type req struct {
	url       string
	ctype     string
	cencoding string
	clength   string
	status    int
}

func withContext(fn func(url string)) {
	cache, err := cdc.OpenCache("../../testdata")
	if err != nil {
		log.Fatal(err)
	}

	handler := CacheHandler(cache)
	server := httptest.NewServer(handler)
	defer server.Close()

	fn(server.URL)
}

func makeURL(base, view string) string {
	q := url.Values{
		"view": []string{view},
	}
	u, _ := url.Parse(base)
	u.RawQuery = q.Encode()
	return u.String()
}

func TestView(t *testing.T) {
	withContext(func(base string) {
		tts := []req{
			{
				url:       makeURL(base, "https://golang.org/doc/gopher/pkg.png"),
				ctype:     "image/png",
				cencoding: "",
				clength:   "5409",
				status:    http.StatusOK,
			},
			{
				url:       makeURL(base, "https://golang.org/lib/godoc/godocs.js"),
				ctype:     "application/x-javascript",
				cencoding: "gzip",
				clength:   "5186",
				status:    http.StatusOK,
			}, {
				url:       makeURL(base, "https://golang.org/pkg/"),
				ctype:     "text/html; charset=utf-8",
				cencoding: "gzip",
				clength:   "8476",
				status:    http.StatusOK,
			}, {
				url:    makeURL(base, "https://golang.org/"),
				status: http.StatusNotFound,
			},
		}

		client := &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
		}

		for _, tt := range tts {
			res, err := client.Get(tt.url)
			if err != nil {
				t.Fatal(err)
			}

			verify(t, tt, res)
		}
	})
}

func verify(t *testing.T, r req, res *http.Response) {
	if res.StatusCode != r.status {
		t.Fatalf("bad statuscode: %d, want: %d", res.StatusCode, r.status)
	}
	if res.StatusCode != http.StatusOK {
		return
	}

	cl := res.Header.Get("Content-Length")
	if cl != r.clength {
		t.Fatalf("bad content-length: %q, want: %q", cl, r.clength)
	}

	ct := res.Header.Get("Content-Type")
	if ct != r.ctype {
		t.Fatalf("bad content-type: %q, want: %q", ct, r.ctype)
	}

	ce := res.Header.Get("Content-Encoding")
	if ce != r.cencoding {
		t.Fatalf("bad content-encoding: %q, want: %q", ce, r.cencoding)
	}

	n, err := io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()

	nlength, _ := strconv.ParseInt(r.clength, 10, 64)
	if n != nlength {
		t.Fatalf("bad stream size: %d, want: %d", n, nlength)
	}
}
