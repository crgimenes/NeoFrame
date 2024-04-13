package neoframe

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"nf/config"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	_ "image/png"
)

const (
	name = "NeoFrame"
)

type NeoFrame struct {
	img                 *image.RGBA
	maxWidth, maxHeight int
}

func New() *NeoFrame {
	return &NeoFrame{}
}

func (nf *NeoFrame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return nf.maxWidth, nf.maxHeight
}

func (nf *NeoFrame) Update() error {
	//x, y := ebiten.CursorPosition()
	//log.Println("x:", x, "y:", y)

	return nil
}

func (nf *NeoFrame) Draw(screen *ebiten.Image) {
	screen.WritePixels(nf.img.Pix)
}

func RGBAstrToColor(str string) (r, g, b, a uint8, err error) {
	// RRGGBBAA or RRGGBB

	if len(str) != 8 && len(str) != 6 {
		return 0, 0, 0, 0, fmt.Errorf("invalid color string: %s", str)
	}

	rt, err := strconv.ParseUint(str[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	gt, err := strconv.ParseUint(str[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	bt, err := strconv.ParseUint(str[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	a = 0xff
	if len(str) == 8 {
		at, err := strconv.ParseUint(str[6:8], 16, 8)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		a = uint8(at)
	}

	r = uint8(rt)
	g = uint8(gt)
	b = uint8(bt)

	return r, g, b, a, nil
}

func (nf *NeoFrame) SetBackgroudImage(path string) {
	img, err := LoadImage(path)
	if err != nil {
		log.Println("failed to load image:", err)
		return
	}

	nf.img = image.NewRGBA(img.Bounds())
	draw.Draw(nf.img, img.Bounds(), img, image.Pt(0, 0), draw.Src)
	return
}

func (nf *NeoFrame) GetScreenSize() (width, height int) {
	return nf.maxWidth, nf.maxHeight
}

func (nf *NeoFrame) SetBackgroudImageByData(data []byte) {
}

func (nf *NeoFrame) Clean() {
	nf.img = image.NewRGBA(image.Rect(0, 0, nf.maxWidth, nf.maxHeight))
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

func (nf *NeoFrame) SetBackgroudImageAt(file string, x, y int) error {
	img, err := LoadImage(file)
	if err != nil {
		return err
	}

	draw.Draw(nf.img, img.Bounds().Add(image.Pt(x, y)), img, image.Pt(0, 0), draw.Src)
	return nil
}

func (nf *NeoFrame) DrawBox(x, y, w, h int, colorstr string) error {
	r, g, b, a, err := RGBAstrToColor(colorstr)
	if err != nil {
		return err
	}

	c := color.RGBA{r, g, b, a}

	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			nf.img.Set(i, j, c)
		}
	}

	return nil
}

func (nf *NeoFrame) DrawCircle(x, y, r int, color string) error {
	return nil
}

func (nf *NeoFrame) DrawLine(x1, y1, x2, y2 int, colorstr string) error {
	r, g, b, a, err := RGBAstrToColor(colorstr)
	if err != nil {
		return err
	}

	c := color.RGBA{r, g, b, a}

	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 {
		if y1 > y2 {
			y1, y2 = y2, y1
		}

		for y := y1; y <= y2; y++ {
			nf.img.Set(x1, y, c)
		}
	} else if dy == 0 {
		if x1 > x2 {
			x1, x2 = x2, x1
		}

		for x := x1; x <= x2; x++ {
			nf.img.Set(x, y1, c)
		}
	} else {
		if dx < 0 {
			dx = -dx
		}

		if dy < 0 {
			dy = -dy
		}

		if dx > dy {
			if x1 > x2 {
				x1, x2 = x2, x1
				y1, y2 = y2, y1
			}

			for x := x1; x <= x2; x++ {
				y := y1 + (x-x1)*(y2-y1)/(x2-x1)
				nf.img.Set(x, y, c)
			}
		} else {
			if y1 > y2 {
				x1, x2 = x2, x1
				y1, y2 = y2, y1
			}

			for y := y1; y <= y2; y++ {
				x := x1 + (y-y1)*(x2-x1)/(y2-y1)
				nf.img.Set(x, y, c)
			}

		}
	}
	return nil
}

func (nf *NeoFrame) DrawText(x, y, w, h int, text string, fgColor, bgColor string) error {
	return nil
}

func (nf *NeoFrame) DrawPixel(x, y int, colorstr string) error {
	r, g, b, a, err := RGBAstrToColor(colorstr)
	if err != nil {
		return err
	}

	nf.img.Set(x, y, color.RGBA{r, g, b, a})
	return nil
}

func (nf *NeoFrame) DrawGrid(ha, va int, colorstr string) error {

	r, g, b, a, err := RGBAstrToColor(colorstr)
	if err != nil {
		return err
	}

	c := color.RGBA{r, g, b, a}

	log.Println("Drawing grid with horizontal:", ha, "vertical:", va, "color:", c)

	// draw horizontal lines
	for i := 0; i < nf.maxHeight; i += va {
		for j := 0; j < nf.maxWidth; j++ {
			nf.img.Set(j, i, c)
		}
	}

	// draw vertical lines
	for i := 0; i < nf.maxWidth; i += ha {
		for j := 0; j < nf.maxHeight; j++ {
			nf.img.Set(i, j, c)
		}
	}

	draw.Draw(nf.img, nf.img.Bounds(), nf.img, image.Pt(0, 0), draw.Src)

	return nil
}

func (nf *NeoFrame) CopyImageToScreen(img image.Image, x, y int) {
	draw.Draw(nf.img, img.Bounds().Add(image.Pt(x, y)), img, image.Pt(0, 0), draw.Src)
}

func (nf *NeoFrame) Run() {
	maxWidth, maxHeight := ebiten.Monitor().Size()
	if config.CFG.WindowWidth == 0 {
		config.CFG.WindowWidth = maxWidth
		nf.maxWidth = maxWidth
	}

	if config.CFG.WindowHeight == 0 {
		config.CFG.WindowHeight = maxHeight
		nf.maxHeight = maxHeight
	}

	nf.img = image.NewRGBA(image.Rect(0, 0, nf.maxWidth, nf.maxHeight))

	if config.CFG.WindowBgColor != "" {
		r, g, b, a, err := RGBAstrToColor(config.CFG.WindowBgColor)
		if err != nil {
			log.Fatal(err)
		}

		c := color.RGBA{r, g, b, a}

		draw.Draw(nf.img, nf.img.Bounds(), &image.Uniform{c}, image.Pt(0, 0), draw.Src)
	}

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowDecorated(config.CFG.WindowBorder)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(config.CFG.MousePassthrough)
	ebiten.SetWindowPosition(config.CFG.WindowX, config.CFG.WindowY)
	ebiten.SetWindowSize(nf.maxWidth, nf.maxHeight)
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
