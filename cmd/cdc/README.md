cdc client
==========

### List all entries

```sh
$ go run main.go list ../../testcache/
https://golang.org/lib/godoc/jquery.treeview.edit.js
https://golang.org/lib/godoc/playground.js
https://golang.org/pkg/os/
https://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js
https://golang.org/pkg/bytes/
https://golang.org/lib/godoc/godocs.js
https://golang.org/pkg/builtin/
https://ssl.google-analytics.com/ga.js
https://golang.org/pkg/bufio/
https://golang.org/pkg/io/
https://golang.org/pkg/
https://golang.org/doc/gopher/pkg.png
https://golang.org/pkg/strings/
https://golang.org/lib/godoc/jquery.treeview.js
https://golang.org/pkg/io/ioutil/
https://golang.org/favicon.ico
https://golang.org/lib/godoc/style.css
https://golang.org/pkg/strconv/
https://golang.org/lib/godoc/jquery.treeview.css
```

### Print entry info

```sh
$ go run main.go info -url https://golang.org/doc/gopher/pkg.png ../../testcache/
Hash:6cc46ffd Next:0 RankingsNode:90000003 ReuseCount:0 RefetchCount:0 State:0 CreationTime:13096853902296629 KeyLen:37 LongKey:0 DataSize:[6912 5409 0 0] DataAddr:[c103000a c103000c 0 0] Flags:0 SelfHash:e10f47ed Key:https://golang.org/doc/gopher/pkg.png
```

### Print entry header

```sh
$ go run main.go header -hash 6cc46ffd ../../testcache/
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
$ go run main.go body -hash 6cc46ffd ../../testcache/ | file -
/dev/stdin: PNG image data, 83 x 120, 8-bit grayscale, non-interlaced
```

```sh
$ go run main.go body -hash 6cc46ffd ../../testcache/ | hexdump -C -n 32
00000000  89 50 4e 47 0d 0a 1a 0a  00 00 00 0d 49 48 44 52  |.PNG........IHDR|
00000010  00 00 00 53 00 00 00 78  08 00 00 00 00 ab b2 91  |...S...x........|
00000020
```

