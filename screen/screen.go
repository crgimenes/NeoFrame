package screen

import (
	"image"
	"image/draw"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	_ "image/png"
)

const (
	name = "NeoFrame"
)

var nf *neoframe

type neoframe struct {
	img                         *image.RGBA
	monitorWidth, monitorHeight int
}

func (nf *neoframe) Layout(outsideWidth, outsideHeight int) (int, int) {
	return nf.monitorWidth, nf.monitorHeight
}

func (nf *neoframe) Update() error {
	//x, y := ebiten.CursorPosition()
	//log.Println("x:", x, "y:", y)

	return nil
}

func (nf *neoframe) Draw(screen *ebiten.Image) {
	screen.WritePixels(nf.img.Pix)
}

func SetBackgroudImage(path string) {
	img, err := LoadImage(path)
	if err != nil {
		log.Println("failed to load image:", err)
		return
	}

	nf.img = image.NewRGBA(img.Bounds())
	draw.Draw(nf.img, img.Bounds(), img, image.Pt(0, 0), draw.Src)
	return
}

func GetScreenSize() (width, height int) {
	return 0, 0
}

func SetBackgroudImageByData(data []byte) {
}

func Clean() {
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

	draw.Draw(nf.img, img.Bounds().Add(image.Pt(x, y)), img, image.Pt(0, 0), draw.Src)
	return nil
}

func DrawBox(x, y, w, h int, color string) error {
	return nil
}

func DrawText(x, y, w, h int, text string, fgColor, bgColor string) error {
	return nil
}

func CopyImageToScreen(img image.Image, x, y int) {
	draw.Draw(nf.img, img.Bounds().Add(image.Pt(x, y)), img, image.Pt(0, 0), draw.Src)
}

func RunApp() {
	nf = &neoframe{}
	nf.monitorWidth, nf.monitorHeight = ebiten.Monitor().Size()
	nf.img = image.NewRGBA(image.Rect(0, 0, nf.monitorWidth, nf.monitorHeight))

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetTPS(50)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(true)
	ebiten.SetWindowPosition(0, 0)
	ebiten.SetWindowSize(nf.monitorWidth, nf.monitorHeight)
	ebiten.SetWindowTitle(name)

	err := ebiten.RunGameWithOptions(nf, &ebiten.RunGameOptions{
		InitUnfocused:     true,
		ScreenTransparent: true,
		SkipTaskbar:       true,
		X11ClassName:      name,
		X11InstanceName:   name,
	})
	if err != nil {
		log.Fatal(err)
	}
}
