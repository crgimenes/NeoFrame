package screen

import (
	"image"
	_ "image/png"
	"os"
)

var (
	imgBuf        []byte
	width, height int
)

func init() {
	width, height = GetScreenSize()
	imgBuf = make([]byte, width*height*4)
}

func LoadImage(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func SetBackgroudImageAt(file string, x, y int) error {
	img, err := LoadImage(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			r, g, b, a := img.At(i, j).RGBA()
			idx := (j+y)*width*4 + (i+x)*4
			imgBuf[idx] = byte(r >> 8)
			imgBuf[idx+1] = byte(g >> 8)
			imgBuf[idx+2] = byte(b >> 8)
			imgBuf[idx+3] = byte(a >> 8)
		}
	}

	SetBackgroudImageByData(imgBuf, width, height)

	return nil
}
