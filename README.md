# go-qr-annotator
Generates and annotates QR codes in Go, exports PNGs, provides web interface

## Building
- Project uses [go-bindata](https://github.com/jteeuwen/go-bindata) to compile ttf fonts into an asset file, bindata.go
  - `go get go get -u github.com/jteeuwen/go-bindata/...`
  - `$GOPATH/bin/go-bindata fonts`
  - `go get`
  - Should now have bindata.go
- Build package
  - `go build`
## Running
- `./go-qr-annotator`
- Go to http://127.0.0.1:9090/ for the web interface
