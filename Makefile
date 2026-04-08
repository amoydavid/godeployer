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

.PHONY: install
install:
	make build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

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

# 代码格式化检查
.PHONY: fmt
fmt:
	@echo "检查代码格式..."
	@test -z "$$(gofmt -s -l . | tee /dev/stderr)" || (echo "请运行 'make fmt-fix' 修复格式问题" && exit 1)

# 代码格式化修复
.PHONY: fmt-fix
fmt-fix:
	@echo "格式化代码..."
	gofmt -s -w .

# 静态检查
.PHONY: vet
vet:
	@echo "运行 go vet..."
	go vet ./...

# Lint 检查 (需要安装 golangci-lint)
.PHONY: lint
lint:
	@echo "运行 golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "警告: golangci-lint 未安装。跳过 lint 检查。"; \
		echo "安装: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
	fi

# 综合检查 (fmt + vet + lint)
.PHONY: check
check: fmt vet lint
	@echo "所有检查通过！"

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
	@echo "  make fmt           - 检查代码格式"
	@echo "  make fmt-fix       - 自动修复代码格式"
	@echo "  make vet           - 运行 go vet 静态检查"
	@echo "  make lint          - 运行 golangci-lint 检查"
	@echo "  make check         - 运行所有检查 (fmt + vet + lint)"
	@echo "  make run           - 直接运行程序" 