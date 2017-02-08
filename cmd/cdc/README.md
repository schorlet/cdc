cdc client
==========

### List all entries

```sh
$ go run main.go list ../../testdata/
2684420103 https://golang.org/lib/godoc/jquery.treeview.js
2684420102 https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js
2684420140 https://golang.org/pkg/io/
2684420134 https://golang.org/pkg/bytes/
2684420098 https://golang.org/pkg/
2684420137 https://golang.org/pkg/io/ioutil/
2684420106 https://golang.org/lib/godoc/godocs.js
2684420119 https://golang.org/pkg/builtin/
2684420101 https://golang.org/doc/gopher/pkg.png
2684420118 https://ssl.google-analytics.com/ga.js
2684420147 https://golang.org/pkg/strings/
2684420120 https://golang.org/favicon.ico
2684420123 https://golang.org/pkg/bufio/
2684420104 https://golang.org/lib/godoc/jquery.treeview.edit.js
2684420099 https://golang.org/lib/godoc/style.css
2684420145 https://golang.org/pkg/strconv/
2684420100 https://golang.org/lib/godoc/jquery.treeview.css
2684420105 https://golang.org/lib/godoc/playground.js
2684420139 https://golang.org/pkg/os/
```

### Print entry header

```sh
$ go run main.go header -addr 2684420101 ../../testdata/
Alternate-Protocol: 443:quic,p=1
Content-Length: 5409
Last-Modified: Mon, 07 Dec 2015 14:14:26 GMT
Server: Google Frontend
Date: Sat, 09 Jan 2016 22:58:22 GMT
Status: 200
Accept-Ranges: bytes
Alt-Svc: quic=":443"; ma=604800; v="30,29,28,27,26,25"
Content-Type: image/png
```

### Print entry body

```sh
$ go run main.go body -addr 2684420101 ../../testdata/ | file -
/dev/stdin: PNG image data, 83 x 120, 8-bit grayscale, non-interlaced
```

```sh
$ go run main.go body -addr 2684420101 ../../testdata/ | hexdump -C -n 32
00000000  89 50 4e 47 0d 0a 1a 0a  00 00 00 0d 49 48 44 52  |.PNG........IHDR|
00000010  00 00 00 53 00 00 00 78  08 00 00 00 00 ab b2 91  |...S...x........|
00000020
```

