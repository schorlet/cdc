package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"os/exec"
	"sort"
)

func ExampleList() {
	cmd := exec.Command("go", "run", "main.go", "list", "../../testcache")

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	lines := read(output)
	for _, line := range lines {
		fmt.Print(line)
	}

	// Output:
	// https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js
	// https://golang.org/doc/gopher/pkg.png
	// https://golang.org/favicon.ico
	// https://golang.org/lib/godoc/godocs.js
	// https://golang.org/lib/godoc/jquery.treeview.css
	// https://golang.org/lib/godoc/jquery.treeview.edit.js
	// https://golang.org/lib/godoc/jquery.treeview.js
	// https://golang.org/lib/godoc/playground.js
	// https://golang.org/lib/godoc/style.css
	// https://golang.org/pkg/
	// https://golang.org/pkg/bufio/
	// https://golang.org/pkg/builtin/
	// https://golang.org/pkg/bytes/
	// https://golang.org/pkg/io/
	// https://golang.org/pkg/io/ioutil/
	// https://golang.org/pkg/os/
	// https://golang.org/pkg/strconv/
	// https://golang.org/pkg/strings/
	// https://ssl.google-analytics.com/ga.js
}

func ExampleHeader() {
	cmd := exec.Command("go", "run", "main.go", "header", "-addr", "2684420101", "../../testcache")

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	lines := read(output)
	for _, line := range lines {
		fmt.Print(line)
	}

	// Output:
	// Accept-Ranges: bytes
	// Alt-Svc: quic=":443"; ma=604800; v="30,29,28,27,26,25"
	// Alternate-Protocol: 443:quic,p=1
	// Content-Length: 5409
	// Content-Type: image/png
	// Date: Sat, 09 Jan 2016 22:58:22 GMT
	// Last-Modified: Mon, 07 Dec 2015 14:14:26 GMT
	// Server: Google Frontend
	// Status: 200
}

func ExampleBody() {
	cmd := exec.Command("go", "run", "main.go", "body", "-addr", "2684420101", "../../testcache")

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	config, err := png.DecodeConfig(output)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("PNG image data, %d x %d\n", config.Width, config.Height)
	// Output:
	// PNG image data, 83 x 120
}

func read(buf *bytes.Buffer) []string {
	lines := make([]string, 0)
	line, err := buf.ReadString('\n')
	for err == nil {
		lines = append(lines, line)
		line, err = buf.ReadString('\n')
	}

	sort.Strings(lines)
	return lines
}
