# Binary names
BINARY_NAME=AfdianToMarkdown

# Build flags
LDFLAGS=-ldflags "-s -w"

# Release directories
RELEASE_DIR=./release
BUILD_DIR=./bin

.PHONY: build windows linux macos clean test release

# Build all platforms
build: windows linux macos

# Build and compress for Windows
windows:
	mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -p 4 -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe main.go
	#卡巴斯基会报毒，所以暂时不使用upx压缩
	#upx --best --lzma $(BUILD_DIR)/windows/$(BINARY_NAME).exe

# Build and compress for Linux
linux:
	mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -p 4 -o $(BUILD_DIR)/linux/$(BINARY_NAME) main.go
	upx --best --lzma $(BUILD_DIR)/linux/$(BINARY_NAME)

# Build and compress for macOS
macos:
	mkdir -p $(BUILD_DIR)/macos
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -p 4 -o $(BUILD_DIR)/macos/$(BINARY_NAME) main.go

# Move and rename the compressed executables
release: build
	mkdir -p $(RELEASE_DIR)
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