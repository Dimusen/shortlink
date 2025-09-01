# Makefile for Short Link Server

# 变量定义
BINARY_NAME=shorturl-server
BUILD_DIR=./bin
SRC_DIR=./cmd/shorturl-server

# 默认目标
all: build

# 构建项目
build: 
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)/main.go
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# 运行项目
run: 
	@go run $(SRC_DIR)/main.go

# 清理构建产物
clean: 
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed"

# 安装依赖
install: 
	@echo "Installing dependencies..."
	@go mod tidy
	@echo "Dependencies installed"

# 格式化代码
fmt: 
	@go fmt ./...

# 检查代码
vet: 
	@go vet ./...

# 构建并运行
build-run: build run

.PHONY: all build run clean install fmt vet build-run