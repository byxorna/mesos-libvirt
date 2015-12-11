BINARY_NAME=framework
BRANCH=`git rev-parse --abbrev-ref HEAD`
COMMIT=`git rev-parse --short HEAD`
VERSION=`git describe --always --tags --dirty=-hacky`
GOLDFLAGS="-X main.branch=$(BRANCH) -X main.commit=$(COMMIT) -X main.version=$(VERSION)"
PACKAGENAME=`go list .`
#GO15VENDOREXPERIMENT=1

all: test build

setup:
	@echo "-> install build deps"
	@go get -u "golang.org/x/tools/cmd/vet"
	@go get -u "github.com/tools/godep"

vet:
	@echo "-> go vet"
	@go vet $(PACKAGENAME)

fmt:
	@echo "-> go fmt"
	@go fmt ./...

install: test
	@echo "-> go install"
	@godep go install -ldflags=$(GOLDFLAGS)

build:
	@echo "-> go build framework"
	@godep go build -ldflags=$(GOLDFLAGS) -o $(BINARY_NAME)
	@echo "-> go build executor"
	@(cd executor && godep go build -ldflags=$(GOLDFLAGS))

test: fmt vet errcheck
	@echo "-> go test"
	@godep go test ./... -cover

linux: test
	@echo "-> building linux binary for testing"
	GOARCH=amd64 GOOS=linux godep go build -ldflags=$(GOLDFLAGS) -o $(BINARY_NAME)

.PHONY: setup errcheck vet fmt install build test
