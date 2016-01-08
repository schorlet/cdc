package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"cdc"
)

// indexHandler handles all requests.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	host := r.FormValue("host")
	view := r.FormValue("view")

	if len(host) != 0 {
		handleHost(w, r, host)

	} else if len(view) != 0 {
		handleView(w, r, view)

	} else {
		view = assetView(r)

		if len(view) != 0 {
			handleView(w, r, view)

		} else if r.URL.Path == "/" {
			handleHost(w, r, host)

		} else {
			http.Error(w, "cdc: unknown resource", http.StatusBadRequest)
		}
	}
}

// handleHost prints all hosts or all urls from host.
func handleHost(w http.ResponseWriter, r *http.Request, host string) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data struct {
		Hosts map[string]bool
		Urls  []string
	}

	if len(host) == 0 {
		data.Hosts = cacheHost
	} else {
		data.Urls = cacheURL[host]
	}

	t.Execute(w, data)
}

// handleView prints the body of the view.
func handleView(w http.ResponseWriter, r *http.Request, view string) {
	entry, err := cdc.OpenURL(view)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	header, err := entry.OpenHeader()
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
			handleView(w, r, location)
		}
		return
	}

	body, err := entry.OpenBody()
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

	io.Copy(w, body)
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
func assetView(r *http.Request) (v string) {
	referer := r.Referer()

	if referer == "" {
		return
	}

	refererURL, err := url.Parse(referer)
	if err != nil {
		return
	}

	if refererURL.Host != r.Host {
		return
	}

	view := refererURL.Query().Get("view")
	if view == "" {
		return
	}

	baseView, err := url.Parse(view)
	if err != nil {
		return
	}

	nextView, err := baseView.Parse(r.URL.Path)
	if err != nil {
		return
	}

	v = nextView.String()
	return
}

var cacheHost map[string]bool    // [hostname]bool
var cacheURL map[string][]string // [hostname]urls

func initCache(name string) {
	err := cdc.Init(name)
	if err != nil {
		log.Fatal(err)
	}

	cacheHost = make(map[string]bool)
	cacheURL = make(map[string][]string)

	for ustr := range cdc.Urls() {
		u, err := url.Parse(ustr)
		if err != nil {
			continue
		}
		if len(u.Host) != 0 {
			if !cacheHost[u.Host] {
				cacheHost[u.Host] = true
			}
			cacheURL[u.Host] = append(cacheURL[u.Host], ustr)
		}
	}
	log.Printf("hosts count: %d\n", len(cacheHost))
}

func main() {
	if len(os.Args) != 2 {
		log.SetFlags(0)
		log.Fatal(`Usage: server CACHEDIR
CACHEDIR is the path to chromium cache directory.`)
	}

	initCache(os.Args[1])

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/favicon.png", http.NotFound)

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
