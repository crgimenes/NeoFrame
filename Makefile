GIT_TAG := $(shell git describe --tags 2> /dev/null)

ifeq ($(GIT_TAG),)
    GIT_TAG = "dev"
endif

all: build

build:
	CGO_ENABLED=1 go build -a -ldflags "-X 'main.GitTag=$(GIT_TAG)' -extldflags '-s -w'" -o nf

clean:
	rm -f nf

