package screen

import (
	"errors"
	"image"
	_ "image/png"
	"os"
	"strings"
)

var (
	imgBuf        []byte
	width, height int

	ErrInvalidColor = errors.New("Invalid color format")
	ErrInvalidHex   = errors.New("Invalid hex format")
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

	SetBackgroudImageByData(imgBuf)

	return nil
}

func HexToByte(hex string) (byte, error) {
	if len(hex) != 2 {
		return 0, ErrInvalidHex
	}

	hex = strings.ToUpper(hex)

	var b byte
	for _, c := range hex {
		b <<= 4
		switch {
		case c >= '0' && c <= '9':
			b |= byte(c - '0')
		case c >= 'A' && c <= 'F':
			b |= byte(c - 'A' + 10)
		default:
			return 0, ErrInvalidHex
		}
	}

	return b, nil
}

func HexToRGBA(color string) (byte, byte, byte, byte, error) {
	// Hex color format: RRGGBBAA or RRGGBB
	if len(color) != 6 && len(color) != 8 {
		return 0, 0, 0, 0, ErrInvalidColor
	}

	if len(color) == 6 {
		color += "FF"
	}

	r, g, b, a := color[0:2], color[2:4], color[4:6], color[6:8]
	ra, err := HexToByte(r)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	ga, err := HexToByte(g)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	ba, err := HexToByte(b)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	aa, err := HexToByte(a)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return ra, ga, ba, aa, nil
}

func DrawBox(x, y, w, h int, color string) error {

	r, g, b, a, err := HexToRGBA(color)
	if err != nil {
		return err
	}

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			idx := (j+y)*width*4 + (i+x)*4
			imgBuf[idx] = r
			imgBuf[idx+1] = g
			imgBuf[idx+2] = b
			imgBuf[idx+3] = a
		}
	}

	SetBackgroudImageByData(imgBuf)

	return nil
}

func DrawText(x, y int, text, color string) error {

	return nil
}
