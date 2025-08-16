# NeoFrame (nf)

A minimal on-screen overlay that creates a transparent, always‑on‑top window which lets normal mouse and keyboard input pass through. When the cursor touches the **top‑left corner**, a small toolbar appears so you can pick a color and draw on the screen or select an eraser. A simple command prompt is also available with basic commands (e.g., `clear`, `exit`).

> Status: early prototype intended for macOS and Windows; Linux may work but is not part of the release targets yet.

## Features

- Transparent, click‑through overlay window (stays above all apps).
- Hot‑corner (**top‑left**) toolbar for quick tool selection.
- Freehand drawing with color selection and an eraser.
- Basic terminal prompt with `clear` and `exit`.
- Skips taskbar / dock and uses a window class/name for X11 when available.
- Auto-detects monitor size on startup.

## Build

Requirements:

- Go **1.25+** (CGO enabled).
- macOS: Xcode Command Line Tools; Windows: a recent MinGW toolchain is recommended.
- Module deps are managed via `go.mod` (Ebitengine v2, FreeType, x/image, readline).

Quick build:

```sh
# Using the Makefile (produces ./nf)
make build

# Or directly:
CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o nf .
```

Cross builds (macOS/Windows) are automated by `release.sh` (requires a `GITHUB_TOKEN`; optional Apple Developer ID signing).

## Run
```sh
./nf
```

Behavior:

- The overlay starts transparent and on top of other windows.
- Move the mouse to the **top‑left corner** to reveal the toolbar.
- Select a color to draw; choose the eraser to remove strokes.
- Use the prompt for `clear` (wipe drawings) or `exit` (quit).

## Notes & Limitations

- Click‑through and transparency depend on platform window APIs; exact behavior may vary between OS versions and, on Linux, between window managers.
- High‑DPI are expected to work; edge cases may still exist.
- Multi‑monitor are not supported yet; the overlay appears on the primary display or on the used to start the app.

## License

BSD 3‑Clause. See [`LICENSE`](LICENSE).

## References

- Ebitengine (Ebiten) v2: https://ebitengine.org
- Go CGO docs: https://pkg.go.dev/cmd/cgo
- FreeType for Go: https://pkg.go.dev/github.com/golang/freetype

