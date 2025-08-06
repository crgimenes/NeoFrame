BINARY_NAME=nf

GIT_TAG := $(shell git describe --tags --always)
BUILD_FLAGS := -trimpath -ldflags "-X 'main.Version=$(GIT_TAG)' -s -w"

.PHONY: all build build-cross clean

all: build

build:
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

