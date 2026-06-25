BINARY_NAME=nf

# Ebitengine v2.10+ uses purego, so no C toolchain is needed on macOS/Windows.
export CGO_ENABLED=0

GIT_TAG := $(shell git describe --tags --always)
BUILD_FLAGS := -trimpath -ldflags "-X 'main.Version=$(GIT_TAG)' -s -w"

.PHONY: all build clean

all: build

build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-* neoframe.history

