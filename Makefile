# 变量定义
APP_NAME := package-base
VERSION := 1.0.0
BUILD_DIR := build
CMD_DIR := cmd/server
BINARY_NAME := $(APP_NAME)

# Go相关变量
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "可用的命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发相关命令
.PHONY: dev
dev: ## 启动开发服务器
	@echo "🚀 启动开发服务器..."
	@go run $(CMD_DIR)/main.go

.PHONY: watch
watch: ## 监听文件变化并自动重启（需要安装air）
	@echo "👀 启动文件监听模式..."
	@air -c .air.toml

# 构建相关命令
.PHONY: build
build: ## 构建应用程序
	@echo "🔨 构建应用程序..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-linux
build-linux: ## 构建Linux版本
	@echo "🔨 构建Linux版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CMD_DIR)/main.go
	@echo "✅ Linux构建完成: $(BUILD_DIR)/$(BINARY_NAME)-linux"

.PHONY: build-windows
build-windows: ## 构建Windows版本
	@echo "🔨 构建Windows版本..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME).exe $(CMD_DIR)/main.go
	@echo "✅ Windows构建完成: $(BUILD_DIR)/$(BINARY_NAME).exe"

.PHONY: build-all
build-all: build build-linux build-windows ## 构建所有平台版本

# 测试相关命令
.PHONY: test
test: ## 运行测试
	@echo "🧪 运行测试..."
	@go test -v ./...

.PHONY: test-cover
test-cover: ## 运行测试并生成覆盖率报告
	@echo "🧪 运行测试覆盖率检查..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 覆盖率报告已生成: coverage.html"

.PHONY: benchmark
benchmark: ## 运行性能测试
	@echo "⚡ 运行性能测试..."
	@go test -bench=. -benchmem ./...

# 代码质量检查
.PHONY: lint
lint: ## 运行代码检查
	@echo "🔍 运行代码检查..."
	@golangci-lint run

.PHONY: fmt
fmt: ## 格式化代码
	@echo "💅 格式化代码..."
	@go fmt ./...

.PHONY: vet
vet: ## 运行go vet检查
	@echo "🔎 运行go vet检查..."
	@go vet ./...

.PHONY: tidy
tidy: ## 整理依赖
	@echo "📦 整理依赖..."
	@go mod tidy

# 数据库相关命令
.PHONY: migrate-up
migrate-up: ## 运行数据库迁移
	@echo "📊 运行数据库迁移..."
	@go run $(CMD_DIR)/main.go migrate up

.PHONY: migrate-down
migrate-down: ## 回滚数据库迁移
	@echo "📊 回滚数据库迁移..."
	@go run $(CMD_DIR)/main.go migrate down

.PHONY: migrate-status
migrate-status: ## 查看迁移状态
	@echo "📊 查看迁移状态..."
	@go run $(CMD_DIR)/main.go migrate status

# Docker相关命令
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "🐳 构建Docker镜像..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "🐳 运行Docker容器..."
	@docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

.PHONY: docker-compose-up
docker-compose-up: ## 使用docker-compose启动所有服务
	@echo "🐳 启动所有服务..."
	@docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## 使用docker-compose停止所有服务
	@echo "🐳 停止所有服务..."
	@docker-compose down

# 清理相关命令
.PHONY: clean
clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean

.PHONY: clean-all
clean-all: clean ## 清理所有生成文件
	@echo "🧹 清理所有文件..."
	@rm -rf vendor/
	@go clean -modcache

# 安装开发工具
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "🛠️ 安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest
	@echo "✅ 开发工具安装完成"

# 生成文档
.PHONY: docs
docs: ## 生成API文档
	@echo "📚 生成API文档..."
	@swag init -g $(CMD_DIR)/main.go

# 项目初始化
.PHONY: init
init: install-tools tidy ## 初始化项目
	@echo "🚀 项目初始化完成！"
	@echo "📖 查看可用命令: make help"
	@echo "🏃 启动开发服务器: make dev"

# 发布相关
.PHONY: release
release: clean test build-all ## 构建发布版本
	@echo "🎉 发布版本构建完成!"

# 健康检查
.PHONY: health
health: ## 检查应用程序健康状态
	@echo "🏥 检查应用程序健康状态..."
	@curl -f http://localhost:8080/health || echo "❌ 应用程序未运行"
