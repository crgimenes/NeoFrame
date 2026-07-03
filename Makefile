BINARY_NAME=nf

# Ebitengine v2.10+ uses purego, so no C toolchain is needed on macOS/Windows.
export CGO_ENABLED=0

BUILD_FLAGS := -trimpath -ldflags "-s -w"

.PHONY: all build clean

all: build

build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

