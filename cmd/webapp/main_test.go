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

type request struct {
	url      string
	mime     string
	encoding string
	length   string
	status   int
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
		requests := []request{
			{
				url:    makeURL(base, "https://golang.org/doc/gopher/pkg.png"),
				mime:   "image/png",
				length: "5409",
				status: http.StatusOK,
			},
			{
				url:      makeURL(base, "https://golang.org/lib/godoc/godocs.js"),
				mime:     "application/x-javascript",
				encoding: "gzip",
				length:   "5186",
				status:   http.StatusOK,
			}, {
				url:      makeURL(base, "https://golang.org/pkg/"),
				mime:     "text/html; charset=utf-8",
				encoding: "gzip",
				length:   "8476",
				status:   http.StatusOK,
			}, {
				url:    makeURL(base, "https://golang.org/"),
				mime:   "text/plain; charset=utf-8",
				length: "16",
				status: http.StatusNotFound,
			},
		}

		client := http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
		}

		for _, req := range requests {
			res, err := client.Get(req.url)
			if err != nil {
				t.Fatal(err)
			}
			verify(t, req, res)
		}
	})
}

func verify(t *testing.T, req request, res *http.Response) {
	if res.StatusCode != req.status {
		t.Fatalf("bad statuscode: %d, want: %d", res.StatusCode, req.status)
	}

	length := res.Header.Get("Content-Length")
	if length != req.length {
		t.Fatalf("bad content-length: %q, want: %q", length, req.length)
	}

	mime := res.Header.Get("Content-Type")
	if mime != req.mime {
		t.Fatalf("bad content-type: %q, want: %q", mime, req.mime)
	}

	encoding := res.Header.Get("Content-Encoding")
	if encoding != req.encoding {
		t.Fatalf("bad content-encoding: %q, want: %q", encoding, req.encoding)
	}

	n, err := io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()

	nlength, _ := strconv.ParseInt(req.length, 10, 64)
	if n != nlength {
		t.Fatalf("bad stream size: %d, want: %d", n, nlength)
	}
}
