# NeoFrame (nf)

A minimal on-screen overlay that creates a transparent, always‑on‑top window which lets normal mouse and keyboard input pass through. An always‑visible toolbar window lets you pick a color and draw on the screen or erase; **Esc** or **Done** returns control to the desktop. A command window (the **Cmd** button) offers basic commands (e.g., `clear`, `exit`). The UI is built with [minigui](https://github.com/crgimenes/minigui).

> Status: early prototype. Released for macOS (signed, notarized universal app), Windows, and Linux.

![NeoFrame](https://github.com/crgimenes/NeoFrame/blob/trunk/nf.gif)

## Features

- Transparent, click‑through overlay window (stays above all apps).
- Always‑visible, draggable toolbar window for tool and color selection.
- Freehand drawing with color selection and an eraser.
- In‑app command window (`clear`, `help`, `exit`).
- Skips taskbar / dock and uses a window class/name for X11 when available.
- Auto-detects monitor size on startup.

## Build

Requirements:

- Go **1.26+**.
- No C toolchain on macOS or Windows: Ebitengine v2.10+ talks to the OS through
  [purego](https://github.com/ebitengine/purego), so `CGO_ENABLED=0` builds
  natively (no Xcode or MinGW). Linux still requires CGO and the X11 dev
  libraries, and is not a release target yet.
- Module deps are managed via `go.mod`: Ebitengine v2.10 plus the sibling
  packages [`minigui`](https://github.com/crgimenes/minigui) (UI toolkit) and
  [`native`](https://github.com/crgimenes/native) (clipboard), and FreeType and
  x/image.

Quick build:

```sh
# Using the Makefile (produces ./nf)
make build

# Or directly:
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o nf .
```

## Release

`release.sh` (run on a macOS host, from a clean worktree on a Git tag) builds a
**signed, notarized macOS universal `.app`** (amd64 + arm64 via `lipo`, hardened
runtime, stapled) plus the Windows binaries (386/amd64/arm64, CGo‑free), then
creates the GitHub release. Linux binaries (amd64 + arm64, built with CGO on
native runners) are produced by the `release-linux` GitHub Actions workflow when
the release is published. The `ci` workflow tests on Linux and compile‑checks the
macOS/Windows cross‑builds on every push and pull request.

Requires `APPLE_DEVELOPER_ID` (a Developer ID Application certificate) and
`APPLE_NOTARY_PROFILE` (notarytool credentials stored in Keychain). The macOS app
is an accessory overlay (`LSUIElement`), so it has no Dock icon.

## Run
```sh
./nf
```

Behavior:

- The overlay starts transparent and on top of other windows, with a small
  always‑visible **toolbar window** in the top‑left; the rest of the screen stays
  click‑through, so you keep using the desktop.
- Pick a color (or **Draw** / **Erase**) to start drawing; **Done** or **Esc**
  releases the tools and hands control back to the desktop.
- The toolbar windows are draggable by their title bar.
- The **Cmd** button opens a command window with a text field (`clear`, `help`,
  `exit`).

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

