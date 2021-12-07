BIN_DIR := "bin"
BIN_FILE := fringe-server
BUILD_ARTIFACTS_DIR := "artifacts"
DB_DIR := "db"
VERSION := $(shell cat VERSION)-$(shell git rev-parse --short HEAD)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOSEC := $(shell which gosec)

# Packages architecture list
ARCHITECTURES := 386 amd64 arm arm64


.PHONY: all
all: test build run

.PHONY: clean
clean:
	go clean
	rm -rf $(BIN_DIR)/*
	rm -rf $(DB_DIR)/*
	rm -rf $(BUILD_ARTIFACTS_DIR)

.PHONY: dep
dep:
	go get -t -v ./...
	go mod tidy
	go mod download

.PHONY: lint
lint:
	golangci-lint run --enable-all

.PHONY: lint-fix
lint-fix:
	golangci-lint run --enable-all --fix

.PHONY: security
security: gosec-check
	@[[ -x "$(GOSEC)" ]] && GO111MODULE=on && $(GOSEC) -conf .gosec.json ./...

.PHONY: gosec-check
gosec-check:
	@[[ -x "$(GOSEC)" ]] || echo "gosec is required: go install github.com/securego/gosec/v2/cmd/gosec@latest"

.PHONY: test
test: lint
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: build
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

.PHONY: packages
packages:
	mkdir -p $(BUILD_ARTIFACTS_DIR)
	GOOS=linux GOARCH=386 ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=amd64 ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=arm ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=arm64 ./scripts/build-deb-package.sh
