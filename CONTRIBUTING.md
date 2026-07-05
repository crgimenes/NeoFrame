# Contributing to NeoFrame

Thanks for your interest. NeoFrame is a small tool with one job: a transparent,
click-through overlay you can draw on, with a toolbar and a command window. The
main risk to a project like this is becoming a whiteboard suite; contributions
should make the one job better, not add a second job.

## Scope: an overlay, not a canvas app

The test for a feature: does it help someone annotate the screen during a demo,
a class, or a recording? Drawing tools, colors, visibility, getting in and out
of the way fast; that is the surface. Layers, documents, export pipelines, and
shape libraries are other apps.

YAGNI is the house rule. The default answer to a new feature is no until a real
use shows up; the shortest change that solves the problem wins; deletion beats
addition.

## The fragile part: the overlay itself

Transparent + always-on-top + click-through is the hard, platform-specific core,
and it behaves differently on macOS, Windows, and X11. Two consequences:

- Any change to window creation, transparency, input pass-through, or z-order
  must say which platforms it was actually run on. CI cannot see an overlay;
  a human has to.
- Platform quirks get a comment explaining why the code is the way it is;
  the next person cannot rediscover a compositor bug from a bare API call.

Where a platform genuinely cannot do something, degrade honestly (feature off,
clear note) instead of shipping something that half works.

## Dependencies and builds

Ebitengine and `minigui`; that is the list. UI widgets and their fixes belong
upstream in [minigui](https://github.com/crgimenes/minigui), and OS bindings in
[native](https://github.com/crgimenes/native). NeoFrame consumes them; it does
not fork them locally.

Builds are cgo-free on macOS and Windows (Ebitengine goes through purego), and
releases are single binaries; a contribution must not introduce cgo or a C
toolchain requirement, and must keep `CGO_ENABLED=0 go build` green there.

## Code style

`gofmt`, US English everywhere. No inline `if` init; assign, then `if`. No
`else` after a terminal branch; return early. Comments explain why, not what.

```sh
go fix ./...
gofmt -l .        # must print nothing
go vet ./...
golangci-lint run ./...
go test -timeout 30s -count 1 ./...
make build        # then actually draw on the screen with it
```

## Proposing a change

The default branch is `trunk`. Bug fixes and small improvements can go straight
to a PR (state the platforms you tested). For new tools or behavior changes,
open an issue first; it is a prototype with a clear direction, and a short
conversation beats a big rewrite.
