LINUX := linux
DARWIN := darwin
OSARCH := amd64
TRAVIS_TAG ?= $(shell git rev-list -1 HEAD)

.PHONY: check format test build

all: check format test build

check:
	@echo ">> checking code"
	go vet ./...

format:
	@echo ">> formatting code"
	go fmt ./...

test:
	@echo ">> running tests"
	go test -v -cover ./...

build:
	@echo ">> building binaries"
	GOOS=$(LINUX) GOARCH=$(OSARCH) go build -ldflags "-X main.version=$(TRAVIS_TAG)" -o release/snitch-$(LINUX)-$(OSARCH)
	GOOS=$(LINUX) GOARCH=$(OSARCH) go build -ldflags "-X main.version=$(TRAVIS_TAG)" -o release/snitch-$(DARWIN)-$(OSARCH)

