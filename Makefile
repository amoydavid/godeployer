# Deployer编译配置
BINARY_NAME=deployer
VERSION=0.1.0
BUILD_DIR=build
GO_FILES=./cmd/deployer/main.go
LDFLAGS=-ldflags "-X main.VERSION=$(VERSION)"

# 默认编译当前平台
.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(GO_FILES)

# 清理构建文件
.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)

# 构建所有平台
.PHONY: build-all
build-all: build-linux build-mac build-windows

# Linux构建
.PHONY: build-linux
build-linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(GO_FILES)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 $(GO_FILES)

# MacOS构建
.PHONY: build-mac
build-mac:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 $(GO_FILES)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 $(GO_FILES)

# Windows构建
.PHONY: build-windows
build-windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe $(GO_FILES)

# 测试
.PHONY: test
test:
	go test -v ./...

# 运行
.PHONY: run
run:
	go run $(GO_FILES)

# 帮助信息
.PHONY: help
help:
	@echo "可用命令:"
	@echo "  make build         - 为当前平台构建"
	@echo "  make build-all     - 为所有平台构建"
	@echo "  make build-linux   - 为Linux构建"
	@echo "  make build-mac     - 为MacOS构建"
	@echo "  make build-windows - 为Windows构建"
	@echo "  make clean         - 清理构建文件"
	@echo "  make test          - 运行测试"
	@echo "  make run           - 直接运行程序" 