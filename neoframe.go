package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/crgimenes/minigui"
	"github.com/golang/freetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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

type Leyer struct {
	img    *image.RGBA
	visibl bool
}

type NeoFrame struct {
	CFG               *Config
	colorPalette      []string
	currentLayer      int
	currentPaintColor string
	cmdRect           image.Rectangle
	cmdStatus         string
	cmdText           string
	cmdVisible        bool
	eraser            bool
	fontBytes         []byte
	gui               minigui.Context
	layer             []Leyer
	maxHeight         int
	maxWidth          int
	mouseX            int
	mouseY            int
	panelRect         image.Rectangle
	paintbrush        bool
}

func (nf *NeoFrame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return nf.maxWidth, nf.maxHeight
}

// buildTools lays out the always-visible toolbar panel for one frame and records
// its screen rectangle, so Update can keep it clickable in the click-through
// overlay (see SetMousePassthrough).
func (nf *NeoFrame) buildTools() {
	nf.gui.Begin(minigui.InputFromEbiten(), 8, 8)
	nf.gui.BeginPanel("NeoFrame", 8, 8)
	nf.gui.SetItemWidth(80)

	// Draw and Erase are toggles: clicking the active one releases it (like Done).
	// Entering Draw paints with the current color, which defaults to red and
	// remembers the last swatch picked.
	if nf.gui.Toggle("draw", "Draw", nf.paintbrush) {
		if nf.paintbrush {
			nf.releaseTool()
		} else {
			nf.paintbrush, nf.eraser = true, false
		}
	}
	if nf.gui.Toggle("erase", "Erase", nf.eraser) {
		if nf.eraser {
			nf.releaseTool()
		} else {
			nf.eraser, nf.paintbrush = true, false
		}
	}
	if nf.gui.Button("clear", "Clear") {
		nf.Clear()
	}
	if nf.gui.Toggle("cmd", "Cmd", nf.cmdVisible) {
		nf.cmdVisible = !nf.cmdVisible
	}
	if nf.gui.Button("done", "Done") {
		nf.releaseTool()
	}

	const swatchCols = 4
	for i, hex := range nf.colorPalette {
		r, g, b, a, err := RGBAstrToColor(hex)
		if err != nil {
			continue
		}
		id := minigui.ID(fmt.Sprintf("color_%d", i))
		if nf.gui.Swatch(id, color.RGBA{r, g, b, a}, hex == nf.currentPaintColor) {
			nf.currentPaintColor = hex
			nf.paintbrush = true
			nf.eraser = false
		}
		if (i+1)%swatchCols != 0 && i != len(nf.colorPalette)-1 {
			nf.gui.SameLine()
		}
	}

	nf.panelRect = nf.gui.EndPanel()

	if nf.cmdVisible {
		nf.buildCmdPanel()
	} else {
		nf.cmdRect = image.Rectangle{}
	}

	nf.gui.End()
}

// buildCmdPanel lays out the command panel — a text field to type commands into —
// to the right of the toolbar, and records its rectangle. Pressing Enter or the
// Run button executes the command.
func (nf *NeoFrame) buildCmdPanel() {
	x := float64(nf.panelRect.Max.X + 12)
	nf.gui.BeginPanel("Commands", x, 8)
	nf.gui.TextField("cmd", &nf.cmdText)
	run := nf.gui.Button("run", "Run")
	if run || nf.gui.Submitted("cmd") {
		nf.runCommand(nf.cmdText)
		nf.cmdText = ""
	}
	if nf.cmdStatus != "" {
		nf.gui.Label(nf.cmdStatus)
	}
	nf.cmdRect = nf.gui.EndPanel()
}

// runCommand executes a typed command. The set is intentionally small while the
// app moves to a graphical workflow; more can be added later.
func (nf *NeoFrame) runCommand(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	switch line {
	case "clear", "cls":
		nf.Clear()
		nf.cmdStatus = "cleared"
	case "help":
		nf.cmdStatus = "commands: clear, help, exit"
	case "exit", "quit", "q":
		os.Exit(0)
	default:
		nf.cmdStatus = "unknown command: " + line
	}
}

// releaseTool deselects the active tool, returning to the idle state where the
// overlay is click-through and the desktop is usable again.
func (nf *NeoFrame) releaseTool() {
	nf.paintbrush = false
	nf.eraser = false
}

func (nf *NeoFrame) Update() error {
	x, y := ebiten.CursorPosition()
	nf.mouseX, nf.mouseY = x, y

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		nf.releaseTool()
	}

	// The toolbar panel is always visible: build it every frame and remember its
	// rectangle for the cursor-over-panel test below.
	nf.buildTools()

	tool := nf.paintbrush || nf.eraser
	overPanel := image.Pt(x, y).In(nf.panelRect) || image.Pt(x, y).In(nf.cmdRect)

	// Paint only with a tool active and the cursor off the panel, so clicking the
	// toolbar never leaves a mark on the canvas.
	if tool && !overPanel && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if nf.paintbrush {
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

	// Dynamic passthrough simulates a per-region hit area on a window-wide flag:
	// the overlay grabs the mouse while a tool is active or the cursor is over the
	// panel, and is click-through (desktop usable) otherwise.
	nf.SetMousePassthrough(!tool && !overPanel)

	nf.CFG.oldMouseX = x
	nf.CFG.oldMouseY = y

	return nil
}

func (nf *NeoFrame) Draw(screen *ebiten.Image) {
	for i := range nf.layer {
		if nf.layer[i].visibl {
			screen.WritePixels(nf.layer[i].img.Pix)
		}
	}
	nf.gui.Render(screen)
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
	if enabled == nf.CFG.MousePassthrough {
		return // avoid hammering the window API; Update calls this every frame
	}
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

	// Toolbar font: a system font by default, falling back to the embedded 3270
	// (always present; also the retro look — see useRetroFont).
	nf.gui.SetFace(nf.toolbarFace())

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

	nf.SetMousePassthrough(true)

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowDecorated(nf.CFG.WindowDecorated)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowPosition(nf.CFG.WindowX, nf.CFG.WindowY)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(nf.maxWidth, nf.maxHeight)
	ebiten.SetWindowTitle(name)

	err := ebiten.RunGameWithOptions(
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

// useRetroFont prefers the embedded 3270 font for the toolbar (a retro terminal
// look); the default is a system font, which also covers non-Latin scripts.
const useRetroFont = false

// toolbarFace picks the toolbar font: a system font by default, with the embedded
// 3270 as an opt-in retro face (useRetroFont) and as the guaranteed fallback.
func (nf *NeoFrame) toolbarFace() minigui.Face {
	if useRetroFont {
		if f := embeddedFace(); f != nil {
			return f
		}
	}
	if f, err := minigui.SystemFace(16); err == nil {
		return f
	}
	return embeddedFace() // the 3270 font is always embedded
}

// embeddedFace builds a text face from the embedded 3270 font, or returns nil
// (the toolkit then falls back to Ebitengine's debug font).
func embeddedFace() minigui.Face {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fontBytes))
	if err != nil {
		log.Println("toolbar font: parsing embedded 3270:", err)
		return nil
	}
	return &text.GoTextFace{Source: src, Size: 16}
}
