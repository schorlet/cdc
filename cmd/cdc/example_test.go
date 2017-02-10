package main_test

import (
	"bufio"
	"bytes"
	"fmt"
	"image/png"
	"io"
	"log"
	"os/exec"
	"sort"
)

func init() {
	cmd := exec.Command("go", "build")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func Example_list() {
	cmd := exec.Command("./cdc", "list", "../../testdata")

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	lines := read(output)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Output:
	// 2684420098	https://golang.org/pkg/
	// 2684420099	https://golang.org/lib/godoc/style.css
	// 2684420100	https://golang.org/lib/godoc/jquery.treeview.css
	// 2684420101	https://golang.org/doc/gopher/pkg.png
	// 2684420102	https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js
	// 2684420103	https://golang.org/lib/godoc/jquery.treeview.js
	// 2684420104	https://golang.org/lib/godoc/jquery.treeview.edit.js
	// 2684420105	https://golang.org/lib/godoc/playground.js
	// 2684420106	https://golang.org/lib/godoc/godocs.js
	// 2684420118	https://ssl.google-analytics.com/ga.js
	// 2684420119	https://golang.org/pkg/builtin/
	// 2684420120	https://golang.org/favicon.ico
	// 2684420123	https://golang.org/pkg/bufio/
	// 2684420134	https://golang.org/pkg/bytes/
	// 2684420137	https://golang.org/pkg/io/ioutil/
	// 2684420139	https://golang.org/pkg/os/
	// 2684420140	https://golang.org/pkg/io/
	// 2684420145	https://golang.org/pkg/strconv/
	// 2684420147	https://golang.org/pkg/strings/

}

func Example_header() {
	cmd := exec.Command("./cdc", "header", "-addr", "2684420101", "../../testdata")

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	lines := read(output)
	for _, line := range lines {
		fmt.Println(line)
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

func Example_body() {
	cmd := exec.Command("./cdc", "body", "-addr", "2684420101", "../../testdata")

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

func read(r io.Reader) []string {
	lines := make([]string, 0)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	sort.Strings(lines)
	return lines
}
