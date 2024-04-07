package screen

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	name = "NeoFrame"
)

type neoframe struct {
	keys []ebiten.Key
}

var (
	monitorWidth, monitorHeight = ebiten.ScreenSizeInFullscreen()
)

func (nf *neoframe) Layout(outsideWidth, outsideHeight int) (int, int) {
	return monitorWidth, monitorHeight
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

func RunApp() {
	nf := &neoframe{}

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetTPS(50)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(true)
	ebiten.SetWindowPosition(0, 0)
	ebiten.SetWindowSize(monitorWidth, monitorHeight)
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
