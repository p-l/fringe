BIN_DIR := "bin"
BIN_FILE := fringe-server
BUILD_ARTIFACTS_DIR := "artifacts"
DB_DIR := "db"
VERSION := $(shell cat VERSION)-$(shell git rev-parse --short HEAD)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)


.PHONY: all
all: test build run

clean:
	go clean
	rm -rf $(BIN_DIR)/*
	rm -rf $(DB_DIR)/*
	rm -rf $(BUILD_ARTIFACTS_DIR)

dep:
	go mod tidy
	go mod download

lint:
	golangci-lint run --enable-all

lint-fix:
	golangci-lint run --enable-all --fix

test: lint
	go test

build: dep
	mkdir -p $(BIN_DIR)
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BIN_DIR)/$(BIN_FILE)-$(GOOS)-$(GOARCH)

.PHONY: run
run: build
	mkdir -p $(DB_DIR)
	$(BIN_DIR)/$(BIN_FILE)-$(GOOS)-$(GOARCH)

.PHONY: version
version:
	@echo "$(VERSION)"

ARCHITECTURES := 386 amd64 arm arm64
.PHONY: packages
packages:
	mkdir -p $(BUILD_ARTIFACTS_DIR)
	GOOS=linux GOARCH=386 ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=amd64 ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=arm ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=arm64 ./scripts/build-deb-package.sh
