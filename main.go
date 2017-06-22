package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/qpliu/qrencode-go/qrencode"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net/http"
)

func addLabel(img *draw.Image, ttfont *truetype.Font, dpi, size float64, x, y int, label string) {

	col := color.RGBA{0, 0, 0, 255}
	hint := font.HintingNone
	d := &font.Drawer{
		Dst: *img,
		Src: image.NewUniform(col),
		Face: truetype.NewFace(ttfont, &truetype.Options{
			Size:    size,
			DPI:     dpi,
			Hinting: hint,
		}),
	}

	d.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	d.DrawString(label)
}

func generate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// Parse url params or use defaults
	qrdata := "No data provided :)"
	if len(r.Form["qrdata"]) > 0 {
		qrdata = r.Form["qrdata"][0]
	}

	toptext := ""
	if len(r.Form["toptext"]) > 0 {
		toptext = r.Form["toptext"][0]
	}

	bottext := ""
	if len(r.Form["bottext"]) > 0 {
		bottext = r.Form["bottext"][0]
	}

	ttfont := fonts["Go-Regular.ttf"]
	if len(r.Form["font"]) > 0 {
		wantedFont := r.Form["font"][0]
		mappedFont := fonts[wantedFont]
		if mappedFont != nil {
			fmt.Printf("Using %v!\n", wantedFont)
			ttfont = mappedFont
		} else {
			fmt.Printf("Rejecting want of %v!\n", wantedFont)
		}
	}

	// Generate a QR code
	grid, err := qrencode.Encode(qrdata, qrencode.ECLevelH)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Generated a QR code, width: %v, height: %v\n", grid.Width(), grid.Height())

	blocksize := 8
	size := 16.0
	dpi := 72.0
	img := grid.Image(blocksize).(draw.Image)
	addLabel(&img, ttfont, dpi, size, 32, 26, toptext)
	addLabel(&img, ttfont, dpi, size, 32, grid.Height()*(blocksize)+int(size*3), bottext)

	// Raw PNG data for img element initial load
	if len(r.Form["raw"]) > 0 {
		w.Header().Add("Content-Type", "image/png")
		png.Encode(w, img)
		return
	}

	// Base64 encode for AJAX
	var buf bytes.Buffer
	png.Encode(&buf, img)
	imgBase64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	io.WriteString(w, "data:image/png;base64,"+imgBase64Str)

}

func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, `
	<!DOCTYPE html>
	<html>
	<head>
	<title>QR codes and font rendering with Go</title>
	<style>
	body {
		background: #000;
	}
	#qrcode {
		display: block;
		margin: 1em auto 0;
	}
	.qrapp {
		background:#fff;
		padding: 1em;
		width: 24em;
		margin: 1em auto 0;
	}
	.qrapp div {
		margin: 8px 0;
		clear: both;
	}
	label {
		width: 9em;
		display: inline-block;
	}
	input[type=text],select{
		width: 15em;
		float: right;
		display: inline-block;
	}
	select {
		display: block;
		width: 15.3em
	}
	button {
		width: 100%;
		display: inline-block;
		padding: 0.5em;
	}
	</style>
	</head>
	<body>
		<div class="qrapp">
			<img id="qrcode" src="generate?toptext=Top%20text&bottext=Bottom%20text&qrdata=http://192.168.1.1:8000&raw" />
			<div>
				<label for="toptext">Top text</label>
				<input id="toptext" type="text" value="Top text"/>
			</div>
			<div>
				<label for="bottext">Bottom text</label>
				<input id="bottext" type="text" value="Bottom text"/>
			</div>
			<div>
				<label for="qrdata">QR data</label>
				<input id="qrdata" type="text" value="http://192.168.1.1:8000"/>
			</div>
			<div>
				<label for="font">Font</label>
				<select id="font">
					<option value="Go-Mono.ttf">Go-Mono.ttf</option>
					<option value="Go-Regular.ttf">Go-Regular.ttf</option>
				</select>
			</div>
			<div>
				<button id="generate">Generate</button>
			</div>
		</div>
		<script type="text/javascript">
			var button = document.getElementById("generate")
			var qrdata = document.getElementById("qrdata")
			var bottext = document.getElementById("bottext")
			var toptext = document.getElementById("toptext")
			var font = document.getElementById("font")
			var qrcode = document.getElementById("qrcode")
			var clickHandler = function(){
				var url = "generate" +
				"?qrdata=" + qrdata.value +
				"&bottext=" + bottext.value +
				"&toptext=" + toptext.value +
				"&font=" + font.value
				console.log(url)
				var xhr = new XMLHttpRequest()
				xhr.open("GET", url)
				xhr.onreadystatechange = ()=>{
					if (xhr.readyState == 4){
						qrcode.src = xhr.responseText
					}
				}
				xhr.send()
			}
			button.addEventListener("click", clickHandler)
		</script>
	</body>
	</html>
	`)
}

func parseFontAsset(path string) *truetype.Font {
	fontBytes, err := Asset(path)
	if err != nil {
		panic(err)
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}
	return font
}

var fonts = make(map[string]*truetype.Font)

func main() {
	// Font parsing
	fonts["Go-Mono.ttf"] = parseFontAsset("fonts/Go-Mono.ttf")
	fonts["Go-Regular.ttf"] = parseFontAsset("fonts/Go-Regular.ttf")

	// Page routes
	http.HandleFunc("/", index)
	http.HandleFunc("/generate", generate)

	// Serve forever
	port := ":9090"
	fmt.Printf("Trying to bind to http://127.0.0.1%v\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
