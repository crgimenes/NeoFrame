package screen

import (
	"image"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	name = "NeoFrame"
)

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
}

func SetBackgroudImage(path string) {
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

func RunApp() {
	nf := &neoframe{}
	nf.monitorWidth, nf.monitorHeight = ebiten.ScreenSizeInFullscreen()

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
