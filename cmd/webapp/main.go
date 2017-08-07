// Package webapp helps browsing disk cache.
package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/schorlet/cdc"
)

type cacheHandler struct {
	*cdc.DiskCache
	host map[string]bool     // [hostname]bool
	url  map[string][]string // [hostname]urls
}

// CacheHandler returns a handler that serves HTTP requests
// with the contents of the specified cache.
func CacheHandler(cache *cdc.DiskCache) http.Handler {
	handler := cacheHandler{
		DiskCache: cache,
		host:      make(map[string]bool),
		url:       make(map[string][]string),
	}

	for _, ustr := range cache.URLs() {
		u, err := url.Parse(ustr)
		if err != nil {
			continue
		}
		if len(u.Host) != 0 {
			if !handler.host[u.Host] {
				handler.host[u.Host] = true
			}
			handler.url[u.Host] = append(handler.url[u.Host], ustr)
		}
	}
	return &handler
}

// ServeHTTP responds to an HTTP request.
func (h *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	host := r.FormValue("host")
	view := r.FormValue("view")

	if len(host) != 0 {
		h.handleHost(w, r, host)

	} else if len(view) != 0 {
		h.handleView(w, r, view)

	} else {
		view = assetView(r)

		if len(view) != 0 {
			h.handleView(w, r, view)

		} else if r.URL.Path == "/" {
			h.handleHost(w, r, host)

		} else {
			http.Error(w, "cdc: unknown resource", http.StatusBadRequest)
		}
	}
}

// handleHost prints all hosts or all URLs from host.
func (h *cacheHandler) handleHost(w http.ResponseWriter, r *http.Request, host string) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data struct {
		Hosts map[string]bool
		URLs  []string
	}

	if len(host) == 0 {
		data.Hosts = h.host
	} else {
		data.URLs = h.url[host]
	}

	w.Header().Set("Cache-Control", "no-cache, no-store")
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleView prints the body of the view.
func (h *cacheHandler) handleView(w http.ResponseWriter, r *http.Request, view string) {
	entry, err := h.OpenURL(view)
	if err == cdc.ErrNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	header, err := entry.Header()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	location := header.Get("Location")
	if len(location) != 0 {
		location, err = redirectView(location, view)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			h.handleView(w, r, location)
		}
		return
	}

	body, err := entry.Body()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer body.Close()

	lst := []string{"Content-Type", "Content-Length", "Content-Encoding"}
	for _, item := range lst {
		value := header.Get(item)
		if len(value) != 0 {
			w.Header().Set(item, value)
		}
	}

	w.Header().Set("Cache-Control", "no-cache, no-store")
	_, err = io.Copy(w, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// redirectView handles view redirection to location.
func redirectView(location, view string) (string, error) {
	locationURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	if !locationURL.IsAbs() {
		viewURL, err := url.Parse(view)
		if err != nil {
			return "", err
		}
		locationURL, err = viewURL.Parse(location)
		if err != nil {
			return "", err
		}
	}
	return locationURL.String(), nil
}

// assetView handles the requested assets.
//  request      /doc/gopher/pkg.png
//  referer      http://localhost:8000/?view=https://golang.org/pkg/
//  returns      https://golang.org/doc/gopher/pkg.png
func assetView(r *http.Request) string {
	referer := r.Referer()

	if referer == "" {
		return ""
	}
	refererURL, err := url.Parse(referer)
	if err != nil {
		return ""
	}
	if refererURL.Host != r.Host {
		return ""
	}

	view := refererURL.Query().Get("view")
	if view == "" {
		return ""
	}
	baseView, err := url.Parse(view)
	if err != nil {
		return ""
	}
	nextView, err := baseView.Parse(r.URL.Path)
	if err != nil {
		return ""
	}
	return nextView.String()
}

const usage = `this is a webapp for reading Chromium disk cache v2.

Usage:

    go run main.go CACHEDIR

CACHEDIR is the path to the chromium cache directory.
`

func main() {
	if len(os.Args) != 2 {
		log.SetFlags(0)
		log.Fatal(usage)
	}

	cache, err := cdc.OpenCache(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	handler := CacheHandler(cache)
	http.Handle("/", handler)

	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/favicon.png", http.NotFound)
	http.HandleFunc("/opensearch.xml", http.NotFound)

	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
