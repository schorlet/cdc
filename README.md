# cdc [![GoDoc](https://godoc.org/github.com/schorlet/cdc?status.png)](https://godoc.org/github.com/schorlet/cdc)

The cdc package provides support for reading Chromium disk cache v2.

The disk cache stores resources fetched from the web so that they can be accessed quickly at a latter time if needed.

Learn more:
* https://www.chromium.org/developers/design-documents/network-stack/disk-cache
* http://www.forensicswiki.org/wiki/Google_Chrome#Disk_Cache
* http://www.forensicswiki.org/wiki/Chrome_Disk_Cache_Format

See the [example_test.go](example_test.go) for an example of how to read an image from cache in testdata.

This project also includes a tool to read the cache from command line, read this [README](cmd/cdc).
