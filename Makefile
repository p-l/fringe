BIN_DIR := "bin"
BIN_FILE := fringe-server
BUILD_ARTIFACTS_DIR := "artifacts"
DB_DIR := "db"
CERT_DIR := "certs"
VERSION := $(shell cat VERSION)-$(shell git rev-parse --short HEAD)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOSEC := $(shell go env GOPATH)/bin/gosec

# Packages architecture list
ARCHITECTURES := 386 amd64 arm arm64

.PHONY: all
all: security lint test build

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
	@test -x "$(GOSEC)" && GO111MODULE=on && $(GOSEC) -conf .gosec.json ./...

.PHONY: gosec-check
gosec-check:
	@test -x "$(GOSEC)" || echo "gosec is required: go install github.com/securego/gosec/v2/cmd/gosec@latest"

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

.PHONY: run-air
run-air:
	mkdir -p $(DB_DIR)
	mkdir -p $(BIN_DIR)
	@echo "Running air (https://github.com/cosmtrek/air)"
	air

dev-cert:
	mkdir -p $(CERT_DIR)
	openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes -keyout $(CERT_DIR)/server.key -out $(CERT_DIR)/server.crt -subj "/CN=fringe.local" -addext "subjectAltName=DNS:fringe.local,DNS:www.fringe.local,IP:127.0.0.1"

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
