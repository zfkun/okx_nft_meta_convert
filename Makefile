NAME=okx-nft-metadata-convert
BUILD_VERSION=$(shell cat build.go | grep BuildVersion | grep -v "Apache" | awk -F '=' '{print $$2}' | awk -F '"' '{print $$2}')
MODULE_NAME=main

BIN_DIR=./bin

LDFLAGS=-X '$(MODULE_NAME).BuildVersion=$(BUILD_VERSION)' \
-X '$(MODULE_NAME).BuildUser=$(shell id -u -n)' \
-X '$(MODULE_NAME).BuildTime=$(shell date "+%F %T")' \
-X '$(MODULE_NAME).BuildGitCommit=$(shell git rev-parse HEAD)' \
-X '$(MODULE_NAME).BuildGoVersion=$(shell go version | cut -d ' ' -f 3 | cut -c3-)' \
-X '$(MODULE_NAME).BuildOsName=$(shell uname -s)' \
-X '$(MODULE_NAME).BuildArchName=$(shell uname -m)'

.PHONY: prepare
prepare:
	mkdir -p $(BIN_DIR)

.PHONY: build
build: prepare
	go build -ldflags "-s -w ${LDFLAGS}" -o $(BIN_DIR)/$(NAME) *.go

.PHONY: build-linux
build-linux: prepare
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w ${LDFLAGS}" -v -o $(BIN_DIR)/$(NAME)_v$(BUILD_VERSION)_linux_amd64 *.go

.PHONY: build-windows
build-windows: prepare
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w ${LDFLAGS}" -v -o $(BIN_DIR)/$(NAME)_v$(BUILD_VERSION)_windows_amd64.exe *.go

	echo "Embedding manifest..."
  mt -manifest okx-nft-metadata-convert.exe.manifest -outputresource:$(BIN_DIR)/$(NAME)_v$(BUILD_VERSION)_windows_amd64.exe;1

.PHONY: build-darwin
build-darwin: prepare
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w ${LDFLAGS}" -v -o $(BIN_DIR)/$(NAME)_v$(BUILD_VERSION)_darwin_amd64 *.go

.PHONY: build-darwin-arm64
build-darwin-arm64: prepare
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w ${LDFLAGS}" -v -o $(BIN_DIR)/$(NAME)_v$(BUILD_VERSION)_darwin_arm64 *.go

.PHONY: run
run:
	$(RUN_ENV) go run *.go

.PHONY: run-race
run-race:
	$(RUN_ENV) go run -race *.go

.PHONY: test
test:
	go test -v ./... -cover

# .PHONY: deploy
# deploy:
# 	scp ./bin/$(NAME)_linux_amd64 server:/root