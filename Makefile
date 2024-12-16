# only on linux
IS_LINUX := $(shell [ -f /etc/os-release ] || [ -f /etc/lsb-release ] && echo 1 || echo 0)

all: check_os
	@echo "begin build"

check_os:
ifeq ($(IS_LINUX),0)
	$(error "This Makefile is only for Linux")
endif

# Binary names
BINARY_NAME=AfdianToMarkdown

# Build flags
LDFLAGS=-ldflags "-s -w"
# num of parallel builds
PARALLEL_BUILD=6

# Release directories
RELEASE_DIR=./release
BUILD_DIR=./build

.PHONY: build windows linux macos clean test release

# Build all platforms
build: windows linux macos

# Build and compress for Windows
windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe main.go

# Build and compress for Linux
linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/linux/$(BINARY_NAME) main.go

# Build and compress for macOS
macos:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/macos/$(BINARY_NAME) main.go

# Move and rename the compressed executables
release: build
	# sudo apt-get install upx
	mkdir -p $(RELEASE_DIR)
	# Kaspersky antivirus will report virus, so do not use upx compress
	# upx --best --lzma $(BUILD_DIR)/windows/$(BINARY_NAME).exe
	upx --best --lzma $(BUILD_DIR)/linux/$(BINARY_NAME)
	mv $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(RELEASE_DIR)/$(BINARY_NAME)_windows_amd64.exe
	mv $(BUILD_DIR)/linux/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)_linux_amd64
	mv $(BUILD_DIR)/macos/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)_darwin_amd64

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) $(RELEASE_DIR)
	mkdir -p $(BUILD_DIR)/windows $(BUILD_DIR)/linux $(BUILD_DIR)/macos $(RELEASE_DIR)