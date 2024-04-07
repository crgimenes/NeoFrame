package screen

import (
	"image"
	"os"
)

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
	return nil
}

func DrawBox(x, y, w, h int, color string) error {
	return nil
}

func DrawText(x, y, w, h int, text string, fgColor, bgColor string) error {
	return nil
}

func CopyImageToScreen(img image.Image, x, y int) {
}
