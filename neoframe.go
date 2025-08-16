package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/golang/freetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"

	_ "embed"
	_ "image/png"
)

type Config struct {
	GetScreenInfo    bool
	ServerMode       bool
	WindowTitle      string
	WindowWidth      int
	WindowHeight     int
	WindowX          int
	WindowY          int
	WindowDecorated  bool
	WindowBgColor    string
	RunLuaScript     string
	MousePassthrough bool
	oldMouseX        int
	oldMouseY        int
	toolsVisible     bool
}

type button struct {
	id         string
	x, y, w, h int
	icon       image.Image
	onClick    func(bt *button)
	tag        string // Additional tag for the button, if needed
}

const (
	// 16 colors (old school)
	ColorBlack   = "000000FF"
	ColorRed     = "FF0000FF"
	ColorGreen   = "00FF00FF"
	ColorYellow  = "FFFF00FF"
	ColorBlue    = "0000FFFF"
	ColorMagenta = "FF00FFFF"
	ColorCyan    = "00FFFFFF"
	ColorWhite   = "FFFFFFFF"
	// 256 colors (modern)
	ColorGray         = "808080FF"
	ColorLightGray    = "C0C0C0FF"
	ColorDarkGray     = "404040FF"
	ColorLightRed     = "FF8080FF"
	ColorLightGreen   = "80FF80FF"
	ColorLightYellow  = "FFFF80FF"
	ColorLightBlue    = "8080FFFF"
	ColorLightMagenta = "FF80FFFF"
	ColorLightCyan    = "80FFFFFF"
	ColorLightWhite   = "FFFFFFFF"
	ColorTransparent  = "00000000" // Fully transparent
)

var (
	//go:embed assets/3270-Regular.ttf
	fontBytes []byte
)

/*
	CFG                    = &Config{}
	buttons                = []*button{}
	paintbrush        bool = false
	eraser            bool = false
	buttonBackground  *image.RGBA
	currentPaintColor string = ColorRed
	mouseX            int
	mouseY            int
	colorPalette      = []string{
		ColorBlack,
		ColorRed,
		ColorGreen,
		ColorYellow,
		ColorBlue,
		ColorMagenta,
		ColorCyan,
		ColorWhite,
		ColorGray,
		ColorLightGray,
		ColorDarkGray,
		ColorLightRed,
		ColorLightGreen,
		ColorLightYellow,
		ColorLightBlue,
		ColorLightMagenta,
		ColorLightCyan,
		ColorLightWhite,
		ColorTransparent,
	}
)
*/

type Leyer struct {
	img    *image.RGBA
	visibl bool
}

type NeoFrame struct {
	CFG               *Config
	buttonBackground  *image.RGBA
	buttons           []*button
	colorPalette      []string
	currentLayer      int
	currentPaintColor string
	eraser            bool
	fontBytes         []byte
	layer             []Leyer
	maxHeight         int
	maxWidth          int
	mouseX            int
	mouseY            int
	paintbrush        bool
}

func (nf *NeoFrame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return nf.maxWidth, nf.maxHeight
}

func (nf *NeoFrame) DrawTools() {
	if !nf.CFG.toolsVisible {
		nf.buttonBackground = nf.captureBackground(0, 0, nf.maxWidth, nf.maxHeight)

		for _, btn := range nf.buttons {
			nf.CopyImageToScreen(btn.icon, btn.x, btn.y)
		}
		nf.CFG.toolsVisible = true
	}
}

func (nf *NeoFrame) HideTools() {
	if nf.CFG.toolsVisible {
		nf.restoreBackground(nf.buttonBackground, 0, 0)
		nf.CFG.toolsVisible = false
	}
}

func (nf *NeoFrame) ButtonClick(x, y int) {
	for _, btn := range nf.buttons {
		if x >= btn.x && x <= btn.x+btn.w && y >= btn.y && y <= btn.y+btn.h {
			if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
				btn.onClick(btn)
			}
		}
	}
}

func (nf *NeoFrame) captureBackground(x, y, w, h int) *image.RGBA {
	bg := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(
		bg,
		bg.Bounds(),
		nf.layer[nf.currentLayer].img,
		image.Pt(x, y),
		draw.Src)
	return bg
}

func (nf *NeoFrame) restoreBackground(bg *image.RGBA, x, y int) {
	draw.Draw(
		nf.layer[nf.currentLayer].img,
		bg.Bounds().Add(image.Pt(x, y)),
		bg, image.Pt(0, 0),
		draw.Src)
}

func (nf *NeoFrame) moseOutsideBounds() bool {
	x, y := ebiten.CursorPosition()
	for _, btn := range nf.buttons {
		if x >= btn.x && x <= btn.x+btn.w && y >= btn.y && y <= btn.y+btn.h {
			return false
		}
	}
	return true
}

func (nf *NeoFrame) Update() error {
	x, y := ebiten.CursorPosition()
	nf.mouseX, nf.mouseY = x, y
	//log.Println("x:", x, "y:", y)

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		nf.SetMousePassthrough(true)
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if (!nf.paintbrush ||
			!nf.eraser) &&
			nf.moseOutsideBounds() {
			nf.HideTools()
			nf.SetMousePassthrough(false)
			return nil
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if nf.paintbrush {
			/*
				err := nf.DrawPixel(x, y, "FF0000FF")
				if err != nil {
					return err
				}
			*/

			err := nf.DrawLine(nf.CFG.oldMouseX, nf.CFG.oldMouseY, x, y, 3, nf.currentPaintColor)
			if err != nil {
				return err
			}
		}
		if nf.eraser {
			err := nf.DrawCircle(x, y, 10, 0, true, "00000000")
			if err != nil {
				return err
			}
		}

	}

	nf.ButtonClick(x, y)

	if x < 10 && y < 10 {
		nf.eraser = false
		nf.paintbrush = false

		nf.SetMousePassthrough(false)
		nf.DrawTools()
	}

	nf.CFG.oldMouseX = x
	nf.CFG.oldMouseY = y

	return nil
}

func (nf *NeoFrame) Draw(screen *ebiten.Image) {
	for i := 0; i < len(nf.layer); i++ {
		if nf.layer[i].visibl {
			screen.WritePixels(nf.layer[i].img.Pix)
		}
	}
}

func (nf *NeoFrame) DebugPrint(str string) {
	e := ebiten.NewImage(nf.maxWidth, nf.maxHeight)
	ebitenutil.DebugPrint(e, str)
	draw.Draw(nf.layer[nf.currentLayer].img, e.Bounds(), e, image.Pt(0, 0), draw.Src)
}

func RGBAImageToBytes(img *image.RGBA) []byte {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bytes := make([]byte, 0, w*h*4)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			bytes = append(bytes, byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8))
		}
	}
	return bytes
}

func RGBAstrToColor(str string) (r, g, b, a uint8, err error) {
	// RRGGBBAA or RRGGBB

	str = strings.TrimPrefix(str, "#")

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

	nf.layer[nf.currentLayer].img = image.NewRGBA(img.Bounds())
	draw.Draw(nf.layer[nf.currentLayer].img, img.Bounds(), img, image.Pt(0, 0), draw.Src)
}

func (nf *NeoFrame) GetScreenSize() (width, height int) {
	return nf.maxWidth, nf.maxHeight
}

func (nf *NeoFrame) SetBackgroudImageByData(data []byte) {
}

func (nf *NeoFrame) Clear() {
	nf.layer[nf.currentLayer].img = image.NewRGBA(image.Rect(0, 0, nf.maxWidth, nf.maxHeight))
}

func (nf *NeoFrame) ClearLayer(layer int) {
	if layer < 0 || layer >= len(nf.layer) {
		return
	}

	nf.layer[layer].img = image.NewRGBA(image.Rect(0, 0, nf.maxWidth, nf.maxHeight))
}

func (nf *NeoFrame) SetLayer(layer int) {
	if layer < 0 || layer >= len(nf.layer) {
		return
	}

	nf.currentLayer = layer
}

func (nf *NeoFrame) CreateLayer() {
	nf.layer = append(nf.layer, Leyer{
		img:    image.NewRGBA(image.Rect(0, 0, nf.maxWidth, nf.maxHeight)),
		visibl: true,
	})
}

func (nf *NeoFrame) DeleteLayer(layer int) {
	if layer < 0 || layer >= len(nf.layer) {
		return
	}

	nf.layer = append(nf.layer[:layer], nf.layer[layer+1:]...)
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

func (nf *NeoFrame) SetMousePassthrough(enabled bool) {
	nf.CFG.MousePassthrough = enabled
	ebiten.SetWindowMousePassthrough(enabled)
}

func (nf *NeoFrame) SetBackgroudImageAt(file string, x, y int) error {
	img, err := LoadImage(file)
	if err != nil {
		return err
	}

	draw.Draw(nf.layer[nf.currentLayer].img, img.Bounds().Add(image.Pt(x, y)), img, image.Pt(0, 0), draw.Src)
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
			nf.layer[nf.currentLayer].img.Set(i, j, c)
		}
	}

	return nil
}

func (nf *NeoFrame) DrawCircle(x, y, r, thickness int, filled bool, colorstr string) error {
	red, green, blue, alpha, err := RGBAstrToColor(colorstr)
	if err != nil {
		return err
	}

	c := color.RGBA{red, green, blue, alpha}

	if filled {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if dx*dx+dy*dy <= r*r {
					nf.layer[nf.currentLayer].img.Set(x+dx, y+dy, c)
				}
			}
		}
		return nil
	}
	if thickness <= 1 {
		thickness = 1
	}

	innerR := r - thickness/2
	outerR := r + (thickness-1)/2

	if innerR < 0 {
		innerR = 0
	}

	for dy := -outerR; dy <= outerR; dy++ {
		for dx := -outerR; dx <= outerR; dx++ {
			distSq := dx*dx + dy*dy
			if distSq <= outerR*outerR && distSq >= innerR*innerR {
				nf.layer[nf.currentLayer].img.Set(x+dx, y+dy, c)
			}
		}
	}

	return nil
}

func (nf *NeoFrame) DrawLine(x1, y1, x2, y2, thickness int, colorstr string) error {
	r, g, b, a, err := RGBAstrToColor(colorstr)
	if err != nil {
		return err
	}

	c := color.RGBA{r, g, b, a}

	if thickness <= 1 {
		thickness = 1
	}

	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 {
		if y1 > y2 {
			y1, y2 = y2, y1
		}

		for y := y1; y <= y2; y++ {
			for offset := -(thickness - 1) / 2; offset <= thickness/2; offset++ {
				nf.layer[nf.currentLayer].img.Set(x1+offset, y, c)
			}
		}
	} else if dy == 0 {
		if x1 > x2 {
			x1, x2 = x2, x1
		}

		for x := x1; x <= x2; x++ {
			for offset := -(thickness - 1) / 2; offset <= thickness/2; offset++ {
				nf.layer[nf.currentLayer].img.Set(x, y1+offset, c)
			}
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
				for offset := -(thickness - 1) / 2; offset <= thickness/2; offset++ {
					nf.layer[nf.currentLayer].img.Set(x, y+offset, c)
				}
			}
		} else {
			if y1 > y2 {
				x1, x2 = x2, x1
				y1, y2 = y2, y1
			}

			for y := y1; y <= y2; y++ {
				x := x1 + (y-y1)*(x2-x1)/(y2-y1)
				for offset := -(thickness - 1) / 2; offset <= thickness/2; offset++ {
					nf.layer[nf.currentLayer].img.Set(x+offset, y, c)
				}
			}

		}
	}
	return nil
}

func (nf *NeoFrame) DrawText(x, y int, size float64, textstr string, fgColor string) error {
	// TODO: reimplement using etxt https://github.com/tinne26/etxt

	r, g, b, a, err := RGBAstrToColor(fgColor)
	if err != nil {
		return err
	}

	fg := image.NewUniform(color.RGBA{r, g, b, a})

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return err
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(size)
	c.SetClip(nf.layer[nf.currentLayer].img.Bounds())
	c.SetDst(nf.layer[nf.currentLayer].img)
	c.SetSrc(fg)
	c.SetHinting(font.HintingFull)

	pt := freetype.Pt(x, y+int(c.PointToFixed(size)>>6))
	_, err = c.DrawString(textstr, pt)

	return err
}

func (nf *NeoFrame) DrawPixel(x, y int, colorstr string) error {
	r, g, b, a, err := RGBAstrToColor(colorstr)
	if err != nil {
		return err
	}

	nf.layer[nf.currentLayer].img.Set(x, y, color.RGBA{r, g, b, a})
	return nil
}

func (nf *NeoFrame) DrawGrid(ha, va int, colorstr string) error {

	r, g, b, a, err := RGBAstrToColor(colorstr)
	if err != nil {
		return err
	}

	c := color.RGBA{r, g, b, a}

	// draw horizontal lines
	for i := 0; i < nf.maxHeight; i += va {
		for j := 0; j < nf.maxWidth; j++ {
			nf.layer[nf.currentLayer].img.Set(j, i, c)
		}
	}

	// draw vertical lines
	for i := 0; i < nf.maxWidth; i += ha {
		for j := 0; j < nf.maxHeight; j++ {
			nf.layer[nf.currentLayer].img.Set(i, j, c)
		}
	}

	return nil
}

func (nf *NeoFrame) CopyImageToScreen(img image.Image, x, y int) {
	draw.Draw(
		nf.layer[nf.currentLayer].img,
		img.Bounds().Add(image.Pt(x, y)), img, image.Pt(0, 0), draw.Src)
}

func (nf *NeoFrame) SetWindowTitle(title string) {
	ebiten.SetWindowTitle(title)
}

func (nf *NeoFrame) SetWindowPosition(x, y int) {
	ebiten.SetWindowPosition(x, y)
}

func (nf *NeoFrame) ConfigureMonitorSize() {
	maxWidth, maxHeight := ebiten.Monitor().Size()
	if nf.CFG.WindowWidth == 0 {
		nf.CFG.WindowWidth = maxWidth
	}

	if nf.CFG.WindowHeight == 0 {
		nf.CFG.WindowHeight = maxHeight
	}

	nf.maxWidth = nf.CFG.WindowWidth
	nf.maxHeight = nf.CFG.WindowHeight
}

func (nf *NeoFrame) Run() {
	const (
		name = "NeoFrame"
	)

	nf.fontBytes = fontBytes
	nf.CFG = &Config{}
	nf.buttons = []*button{}
	nf.currentPaintColor = ColorRed
	nf.colorPalette = []string{
		ColorBlack,
		ColorRed,
		ColorGreen,
		ColorYellow,
		ColorBlue,
		ColorMagenta,
		ColorCyan,
		ColorWhite,
		ColorGray,
		ColorLightGray,
		ColorDarkGray,
		ColorLightRed,
		ColorLightGreen,
		ColorLightYellow,
		ColorLightBlue,
		ColorLightMagenta,
		ColorLightCyan,
		ColorLightWhite,
		ColorTransparent,
	}

	drawImg, err := LoadImage("assets/draw.png")
	if err != nil {
		log.Println("failed to load image:", err)
		return
	}

	closeImg, err := LoadImage("assets/close.png")
	if err != nil {
		log.Println("failed to load image:", err)
		return
	}

	clearImg, err := LoadImage("assets/clear.png")
	if err != nil {
		log.Println("failed to load image:", err)
		return
	}

	nf.buttons = []*button{
		{
			id:   "close",
			x:    0,
			y:    64,
			w:    32,
			h:    32,
			icon: closeImg,
			onClick: func(bt *button) {
				nf.HideTools()
				nf.SetMousePassthrough(true)
				nf.paintbrush = false
				nf.eraser = false
			},
		},
		{
			id:   "clear",
			x:    0,
			y:    64 + 32,
			w:    32,
			h:    32,
			icon: clearImg,
			onClick: func(bt *button) {
				nf.paintbrush = false
				nf.eraser = !nf.eraser
				nf.HideTools()
				nf.SetMousePassthrough(!nf.eraser)
				if nf.eraser {
					if ebiten.IsKeyPressed(ebiten.KeyControl) {
						nf.Clear()
						nf.eraser = false
						nf.SetMousePassthrough(true)
					}
				}
			},
		},
		{
			id:   "draw",
			x:    0,
			y:    64 + 32*2,
			w:    32,
			h:    32,
			icon: drawImg,
			onClick: func(bt *button) {
				nf.eraser = false
				nf.paintbrush = !nf.paintbrush
				nf.HideTools()
				nf.SetMousePassthrough(!nf.paintbrush)
			},
		},
	}

	for i, colorStr := range nf.colorPalette {
		//x := 32 + (i%4)*32
		//y := 64 + (i/4)*32

		x := 0
		y := 64 + 32*3 + (i * 32)

		r, g, b, a, err := RGBAstrToColor(colorStr)
		if err != nil {
			log.Printf("Error parsing color %s: %v", colorStr, err)
			continue
		}
		colorRect := image.NewRGBA(image.Rect(0, 0, 32, 32))
		draw.Draw(
			colorRect,
			colorRect.Bounds(),
			&image.Uniform{color.RGBA{r, g, b, a}},
			image.Pt(0, 0),
			draw.Src)

		btn := &button{
			id:   fmt.Sprintf("color_%d", i),
			x:    x,
			y:    y,
			w:    32,
			h:    32,
			tag:  colorStr,
			icon: colorRect,
			onClick: func(bt *button) {
				nf.currentPaintColor = bt.tag
				if !nf.paintbrush {
					nf.paintbrush = true
					nf.eraser = false
					nf.HideTools()
					nf.SetMousePassthrough(false)
				}
			},
		}
		nf.buttons = append(nf.buttons, btn)
	}

	nf.layer = make([]Leyer, 1)
	nf.layer[0].visibl = true
	nf.layer[0].img = image.NewRGBA(image.Rect(0, 0, nf.maxWidth, nf.maxHeight))

	if nf.CFG.WindowBgColor == "" {
		nf.CFG.WindowBgColor = "00000000" // Default to transparent
	}

	if nf.CFG.WindowBgColor != "00000000" {
		r, g, b, a, err := RGBAstrToColor(nf.CFG.WindowBgColor)
		if err != nil {
			log.Fatal(err)
		}

		c := color.RGBA{r, g, b, a}

		draw.Draw(nf.layer[nf.currentLayer].img,
			nf.layer[nf.currentLayer].img.Bounds(),
			&image.Uniform{c},
			image.Pt(0, 0),
			draw.Src)
	}

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowDecorated(nf.CFG.WindowDecorated)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(nf.CFG.MousePassthrough)
	ebiten.SetWindowPosition(nf.CFG.WindowX, nf.CFG.WindowY)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(nf.maxWidth, nf.maxHeight)
	ebiten.SetWindowTitle(name)

	err = ebiten.RunGameWithOptions(
		nf,
		&ebiten.RunGameOptions{
			InitUnfocused:     true,
			ScreenTransparent: nf.CFG.WindowBgColor == "00000000",
			SkipTaskbar:       true,
			X11ClassName:      name,
			X11InstanceName:   name,
		})
	if err != nil {
		log.Fatal(err)
	}
}
