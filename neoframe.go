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
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	_ "embed"
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
	cmdWin            minigui.Window
	eraser            bool
	gui               minigui.Context
	layer             []Leyer
	maxHeight         int
	maxWidth          int
	panelRect         image.Rectangle
	paintbrush        bool
	toolWin           minigui.Window
}

func (nf *NeoFrame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return nf.maxWidth, nf.maxHeight
}

// buildTools lays out the toolbar and command windows for one frame and records
// their rectangles, so Update can keep them clickable in the click-through overlay
// (see SetMousePassthrough). Both windows are draggable by their title bar.
func (nf *NeoFrame) buildTools() {
	nf.gui.Begin(minigui.InputFromEbiten(), 0, 0)

	if nf.gui.BeginWindow(&nf.toolWin) {
		nf.gui.SetItemWidth(80)

		// Draw and Erase are toggles: clicking the active one releases it (like
		// Done). Entering Draw paints with the current color, which defaults to red
		// and remembers the last swatch picked.
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
		if nf.gui.Toggle("cmd", "Cmd", nf.cmdWin.Open) {
			nf.cmdWin.Open = !nf.cmdWin.Open
			if nf.cmdWin.Open {
				nf.gui.Focus("cmd") // ready to type as soon as the window opens
			}
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

		nf.panelRect = nf.gui.EndWindow()
	} else {
		nf.panelRect = image.Rectangle{}
	}

	if nf.gui.BeginWindow(&nf.cmdWin) {
		nf.gui.TextField("cmd", &nf.cmdText)
		if nf.gui.Button("run", "Run") || nf.gui.Submitted("cmd") {
			nf.runCommand(nf.cmdText)
			nf.cmdText = ""
		}
		if nf.cmdStatus != "" {
			nf.gui.Label(nf.cmdStatus)
		}
		nf.cmdRect = nf.gui.EndWindow()
	} else {
		nf.cmdRect = image.Rectangle{}
	}

	nf.gui.End()
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

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		nf.releaseTool()
	}

	// The toolbar panel is always visible: build it every frame and remember its
	// rectangle for the cursor-over-panel test below.
	nf.buildTools()

	tool := nf.paintbrush || nf.eraser
	pt := image.Pt(x, y)
	overPanel := pt.In(nf.panelRect) || pt.In(nf.cmdRect)
	dragging := nf.gui.Dragging()

	// Paint only with a tool active, the cursor off the windows and none being
	// dragged, so interacting with the UI never leaves a mark on the canvas.
	if tool && !overPanel && !dragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
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
	// the overlay grabs the mouse while a tool is active, the cursor is over a
	// window, or a window is being dragged; otherwise it is click-through.
	capture := tool || overPanel || dragging
	nf.SetMousePassthrough(!capture)

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

func (nf *NeoFrame) Clear() {
	nf.layer[nf.currentLayer].img = image.NewRGBA(image.Rect(0, 0, nf.maxWidth, nf.maxHeight))
}

func (nf *NeoFrame) SetMousePassthrough(enabled bool) {
	if enabled == nf.CFG.MousePassthrough {
		return // avoid hammering the window API; Update calls this every frame
	}
	nf.CFG.MousePassthrough = enabled
	ebiten.SetWindowMousePassthrough(enabled)
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
		ColorTransparent,
	}

	// Toolbar font: a system font by default, falling back to the embedded 3270
	// (always present; also the retro look — see useRetroFont).
	nf.gui.SetFace(nf.toolbarFace())

	// The toolbar is always present (no close box); the command window starts
	// hidden. Both are draggable by their title bar and remember their position.
	nf.toolWin = minigui.Window{Title: "NeoFrame", X: 8, Y: 8, Open: true, NoClose: true}
	nf.cmdWin = minigui.Window{Title: "Commands", X: 220, Y: 8, Open: false}

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
		f := embeddedFace()
		if f != nil {
			return f
		}
	}
	f, err := minigui.SystemFace(16)
	if err == nil {
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
