package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net/http"
	"fmt"
	"encoding/json"
)

type canvas struct {
	image *image.RGBA
}

func NewCanvas(rect image.Rectangle) *canvas {
	return &canvas{
		image: image.NewRGBA(rect),
	}
}

func (c *canvas) DrawRect(rect image.Rectangle, clr color.RGBA) {
	draw.Draw(c.image, rect, &image.Uniform{clr}, image.ZP, draw.Src)
}

func (c *canvas) EncodeTo(w io.Writer) error {
	return png.Encode(w, c.image)
}

func drawCell(w io.Writer, data [][]color.RGBA, imageWidth, lineWidth int) error {
	rectWidth := imageWidth / len(data)
	imageWidth = rectWidth * len(data) + 1

	cnvs := NewCanvas(image.Rect(0, 0, imageWidth, imageWidth))

	for rowIndex, columns := range data {
		for columnIndex, clr := range columns {
			cnvs.DrawRect(image.Rect(rowIndex * rectWidth, columnIndex * rectWidth,
				(rowIndex + 1) * rectWidth, (columnIndex + 1) * rectWidth), clr)
		}
	}

	for rowIndex := range data {
		cnvs.DrawRect(image.Rect(rowIndex * rectWidth, 0, rowIndex * rectWidth + lineWidth, imageWidth), color.RGBA{})

		cnvs.DrawRect(image.Rect(0, rowIndex * rectWidth, imageWidth, rowIndex * rectWidth + lineWidth), color.RGBA{})
	}

	return cnvs.EncodeTo(w)
}

func parseRequestData(req *http.Request) ([][]color.RGBA, error) {
	if err := req.ParseForm(); err != nil {
		return nil, err
	}

	var result [][]color.RGBA
	if err := json.Unmarshal([]byte(req.PostForm.Get("data")), &result); err != nil {
		return nil, err
	}

	return result, nil
}

const (
	imageFilename = "image.png"
	imageWidth = 300
	lineWidth = 1
)

func main() {
	//green := color.RGBA{0, 100, 0, 255}
	//myred := color.RGBA{200, 0, 0, 255}

	//data := [][]color.RGBA{
	//	{
	//		green, green, green, green, green, green,
	//	},
	//	{
	//		green, myred, green, myred, green, green,
	//	},
	//	{
	//		green, green, myred, green, green, green,
	//	},
	//	{
	//		green, myred, green, myred, green, green,
	//	},
	//	{
	//		green, green, myred, green, myred, green,
	//	},
	//	{
	//		green, green, green, green, green, green,
	//	},
	//}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
<html>
	<body>
		<form action="/image" method="POST">
			<input type="hidden" name="data" value='[[{"R": 200, "G": 0, "B": 0, "A": 255}, {"R": 200, "G": 0, "B": 0, "A": 255}],[{"R": 0, "G": 100, "B": 0, "A": 255}, {"R": 200, "G": 0, "B": 0, "A": 255}]]'/>
			<button type="submit">Download Image</button>
		</form>
	</body>
</html>`)
	})
	mux.HandleFunc("/image", func(w http.ResponseWriter, req *http.Request) {
		data, err := parseRequestData(req)
		if err != nil {
			log.Printf(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, imageFilename))
		if err := drawCell(w, data, imageWidth, lineWidth); err != nil {
			log.Printf(err.Error())
		}
	})

	log.Fatalln(http.ListenAndServe(":3000", mux))
}