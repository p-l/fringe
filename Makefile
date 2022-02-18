CLIENT_DIR = client

BUILD_PACKAGES_DIR = packages
BUILD_REPO_DIR = repo
BUILD_CLIENT_DIST_DIR = "client/build"
BUILD_BIN_DIR = "bin"
BUILD_BIN_FILE := fringe-server

RUN_DB_DIR = "db"

FRINGE_VERSION = $(shell cat VERSION)-$(shell git rev-parse --short HEAD)
CLIENT_VERSION = $(shell cat $(CLIENT_DIR)/package.json | grep version | head -1 | awk -F: '{ print $$2 }' | sed 's/[",]//g' | sed 's/^[ \t]*//;s/[ \t]*$$//')

# Go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOSEC := $(shell go env GOPATH)/bin/gosec

# Packages architecture list
ARCHITECTURES := 386 amd64 arm arm64

.PHONY: help
help: ## 💬 This help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: ## 🔎 Lint & format, will not fix but sets exit code on error
	golangci-lint run --modules-download-mode=mod ./...
	cd $(CLIENT_DIR); npx eslint "src/**/*.ts" "src/**/*.tsx" "typings/*.ts"
	@# No app specific CSS at the moment
	@# cd $(CLIENT_DIR); npx stylelint "**/*.css"

.PHONY: lint-fix
lint-fix: ## 📜 Lint & format, will try to fix errors and modify code
	golangci-lint run --modules-download-mode=mod --fix ./...
	cd $(CLIENT_DIR); npx eslint --fix "src/**/*.ts" "src/**/*.tsx" "typings/*.ts"
	@# No app specific CSS at the moment
	@# cd $(CLIENT_DIR); npx stylelint --fix "src/**/*.css" "public/**/*.css"

.PHONY: dep
dep: ## 📥 Download and install dependencies
	mkdir -p $(BUILD_BIN_DIR)
	mkdir -p $(BUILD_CLIENT_DIST_DIR)
	touch $(BUILD_CLIENT_DIST_DIR)/.tmp
	go mod download
	cd $(CLIENT_DIR); npm install --silent

.PHONY: watch
watch: dep ## 👀 Run Fringe go server and independent react service with hot reload file watcher, needs https://github.com/cosmtrek/air
	mkdir -p $(RUN_DB_DIR)
	cd $(CLIENT_DIR); npx concurrently "cd ..; air -c .air.toml" "npm run start" # "npm run test:watch"

.PHONY: build
build: dep ## 🔨 Build and bundle the server with the client built-in
	cd $(CLIENT_DIR); npm run build
	go build -ldflags "-X main.Version=$(FRINGE_VERSION)" -o $(BUILD_BIN_DIR)/$(BUILD_BIN_FILE)-$(GOOS)-$(GOARCH)


.PHONY: packages
packages: ## 📦 Build debian packages for easy deployment
	@rm -rf $(BUILD_PACKAGES_DIR) 
	mkdir -p $(BUILD_PACKAGES_DIR)
	GOOS=linux GOARCH=386 ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=amd64 ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=arm ./scripts/build-deb-package.sh
	GOOS=linux GOARCH=arm64 ./scripts/build-deb-package.sh

.PHONY: repo
repo: packages ## 📦 Create an APT repo structure for fringe packages
	@# Inspired by https://earthly.dev/blog/creating-and-hosting-your-own-deb-packages-and-apt-repo/
	@rm -rf $(BUILD_REPO_DIR)
	mkdir -p $(BUILD_REPO_DIR)
	mkdir -p $(BUILD_REPO_DIR)/pool/main/
	cp $(BUILD_PACKAGES_DIR)/*.deb $(BUILD_REPO_DIR)/pool/main/
	@./scripts/generate-repo-packages.sh $(BUILD_REPO_DIR) 386
	@./scripts/generate-repo-packages.sh $(BUILD_REPO_DIR) amd64 
	@./scripts/generate-repo-packages.sh $(BUILD_REPO_DIR) arm
	@./scripts/generate-repo-packages.sh $(BUILD_REPO_DIR) arm64
	@./scripts/generate-repo-release.sh $(BUILD_REPO_DIR) > $(BUILD_REPO_DIR)/dists/stable/Release
	@./scripts/sign-repo.sh $(BUILD_REPO_DIR)


.PHONY: test
test: ## 🎯 Unit test for server and client
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	cd $(CLIENT_DIR); npm run test

.PHONY: clean
clean: ## 🧹 Clean up project
	go clean
	rm -f coverage.txt
	rm -rf $(BUILD_BIN_DIR)
	rm -rf $(BUILD_PACKAGES_DIR)
	rm -rf $(RUN_DB_DIR)
	rm -rf $(RUN_CERTS_DIR)
	rm -rf $(CLIENT_DIR)/node_modules
	rm -rf $(CLIENT_DIR)/coverage
	rm -rf $(BUILD_CLIENT_DIST_DIR)




.PHONY: security
security: ## 🚓 Run go security checks (go install github.com/securego/gosec/v2/cmd/gosec@latest)
	@test -x "$(GOSEC)" && GO111MODULE=on && $(GOSEC) -conf .gosec.json ./...

.PHONY: run
run: build $(CLIENT_DIR)/node_modules ## 🏃 Run Fringe locally

	$(BIN_DIR)/$(BIN_FILE)-$(GOOS)-$(GOARCH)

.PHONY: version
version: ## #️⃣ Show current version number
	@echo "Fringe Go: $(FRINGE_VERSION)"
	@echo "Fringe Client: $(CLIENT_VERSION)"
