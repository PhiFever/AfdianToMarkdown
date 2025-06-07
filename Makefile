# Cross-platform Makefile
# Detect OS
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    RM_CMD := powershell -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue"
    MKDIR_CMD := powershell -Command "New-Item -ItemType Directory -Force -Path"
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        DETECTED_OS := Linux
    endif
    ifeq ($(UNAME_S),Darwin)
        DETECTED_OS := macOS
    endif
    RM_CMD := rm -rf
    MKDIR_CMD := mkdir -p
endif

# Binary names
BINARY_NAME=AfdianToMarkdown

# Build flags
# Debug builds (build target) - no flags, keeps debug info
DEBUG_FLAGS=

# Release builds - strip debug info and symbols
RELEASE_FLAGS=-ldflags "-s -w"

# num of parallel builds
PARALLEL_BUILD=6

# Release directories
ROOT_DIR=.
RELEASE_DIR=./release
BUILD_DIR=./build

.PHONY: build build-windows build-linux build-macos windows linux macos clean test release

all: build

# Debug builds (with debug info)
build-windows:
	GOOS=windows GOARCH=amd64 go build $(DEBUG_FLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build $(DEBUG_FLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/linux/$(BINARY_NAME) main.go

build-macos:
	GOOS=darwin GOARCH=amd64 go build $(DEBUG_FLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/macos/$(BINARY_NAME) main.go

# Build for current platform only (with debug info)
build:
	@echo "Building for current platform: $(DETECTED_OS)..."
ifeq ($(DETECTED_OS),Windows)
	$(MAKE) build-windows
	@echo "Copying Windows binary to build root..."
	powershell -Command "Copy-Item $(BUILD_DIR)/windows/$(BINARY_NAME).exe -Destination $(ROOT_DIR)/ -Force"
else ifeq ($(DETECTED_OS),Linux)
	$(MAKE) build-linux
	@echo "Copying Linux binary to build root..."
	cp $(BUILD_DIR)/linux/$(BINARY_NAME) $(BUILD_DIR)/
else ifeq ($(DETECTED_OS),macOS)
	$(MAKE) build-macos
	@echo "Copying macOS binary to build root..."
	cp $(BUILD_DIR)/macos/$(BINARY_NAME) $(BUILD_DIR)/
else
	@echo "Unknown platform: $(DETECTED_OS)"
	@exit 1
endif

# Release builds (stripped of debug info)
windows:
	GOOS=windows GOARCH=amd64 go build $(RELEASE_FLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe main.go

linux:
	GOOS=linux GOARCH=amd64 go build $(RELEASE_FLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/linux/$(BINARY_NAME) main.go

macos:
	GOOS=darwin GOARCH=amd64 go build $(RELEASE_FLAGS) -p $(PARALLEL_BUILD) -o $(BUILD_DIR)/macos/$(BINARY_NAME) main.go

# Build all platforms for release
release: windows linux macos
	@echo "Creating release directory..."
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "New-Item -ItemType Directory -Force -Path $(RELEASE_DIR)"
else
	mkdir -p $(RELEASE_DIR)
endif
	@echo "Building release for all platforms..."
ifneq ($(DETECTED_OS),Windows)
	# Only compress on Linux/macOS where upx is more commonly available
	@echo "Compressing Linux binary with upx..."
	-upx --best --lzma $(BUILD_DIR)/linux/$(BINARY_NAME) 2>/dev/null || echo "upx not available, skipping compression"
endif
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Move-Item $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(RELEASE_DIR)/$(BINARY_NAME)_windows_amd64.exe -Force"
	powershell -Command "Move-Item $(BUILD_DIR)/linux/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)_linux_amd64 -Force"
	powershell -Command "Move-Item $(BUILD_DIR)/macos/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)_darwin_amd64 -Force"
else
	mv $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(RELEASE_DIR)/$(BINARY_NAME)_windows_amd64.exe
	mv $(BUILD_DIR)/linux/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)_linux_amd64
	mv $(BUILD_DIR)/macos/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)_darwin_amd64
endif

# Run tests
test:
	go test -v ./...

# Clean build artifacts - Cross platform version
clean:
	@echo "Cleaning build artifacts on $(DETECTED_OS)..."
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $(BUILD_DIR), $(RELEASE_DIR)"
	powershell -Command "New-Item -ItemType Directory -Force -Path $(BUILD_DIR)/windows, $(BUILD_DIR)/linux, $(BUILD_DIR)/macos, $(RELEASE_DIR)"
else
	rm -rf $(BUILD_DIR) $(RELEASE_DIR)
	mkdir -p $(BUILD_DIR)/windows $(BUILD_DIR)/linux $(BUILD_DIR)/macos $(RELEASE_DIR)
endif
	@echo "Clean completed!"

# Show detected OS
info:
	@echo "Detected OS: $(DETECTED_OS)"
	@echo "RM Command: $(RM_CMD)"
	@echo "MKDIR Command: $(MKDIR_CMD)"
	@echo "Debug Flags: $(DEBUG_FLAGS)"
	@echo "Release Flags: $(RELEASE_FLAGS)"