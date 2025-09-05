# =============================================================================
# Minigame Admin Service Makefile
# 快速构建和部署工具
# =============================================================================

# 项目配置
PROJECT_NAME := minigame-admin-service
VERSION ?= 1.0.0
BUILD_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标签
LDFLAGS := -X main.version=$(VERSION) -X 'main.buildTime=$(BUILD_TIME)' -X main.gitCommit=$(GIT_COMMIT)

# 路径配置
BUILD_DIR := dist
RELEASE_DIR := release
BINARY_NAME := admin-service

# 默认目标
.DEFAULT_GOAL := help

# =============================================================================
# 帮助信息
# =============================================================================

.PHONY: help
help: ## 显示帮助信息
	@echo "Minigame Admin Service 构建工具"
	@echo ""
	@echo "使用方法:"
	@echo "  make <target>"
	@echo ""
	@echo "目标:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "示例:"
	@echo "  make build          # 构建当前平台"
	@echo "  make build-all      # 构建所有平台"
	@echo "  make release        # 创建发布包"
	@echo "  make install        # 安装到本地"

# =============================================================================
# 开发工具
# =============================================================================

.PHONY: deps
deps: ## 安装依赖
	@echo "🔄 安装Go依赖..."
	@go mod tidy
	@go mod download
	@echo "✅ 依赖安装完成"

.PHONY: test
test: ## 运行测试
	@echo "🔄 运行测试..."
	@go test -v ./...
	@echo "✅ 测试完成"

.PHONY: fmt
fmt: ## 格式化代码
	@echo "🔄 格式化代码..."
	@go fmt ./...
	@echo "✅ 代码格式化完成"

.PHONY: lint
lint: ## 代码检查
	@echo "🔄 代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过检查"; \
		go vet ./...; \
	fi
	@echo "✅ 代码检查完成"

.PHONY: clean
clean: ## 清理构建文件
	@echo "🔄 清理构建文件..."
	@rm -rf $(BUILD_DIR) $(RELEASE_DIR) bin/
	@echo "✅ 清理完成"

# =============================================================================
# 构建目标
# =============================================================================

.PHONY: build
build: deps ## 构建当前平台
	@echo "🔄 构建 $(PROJECT_NAME)..."
	@mkdir -p bin
	@go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .
	@echo "✅ 构建完成: bin/$(BINARY_NAME)"

.PHONY: build-linux
build-linux: deps ## 构建Linux版本
	@echo "🔄 构建Linux版本..."
	@mkdir -p $(BUILD_DIR)/linux-amd64
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/$(BINARY_NAME) .
	@cp -r conf views static $(BUILD_DIR)/linux-amd64/ 2>/dev/null || true
	@echo "✅ Linux版本构建完成"

.PHONY: build-windows
build-windows: deps ## 构建Windows版本
	@echo "🔄 构建Windows版本..."
	@mkdir -p $(BUILD_DIR)/windows-amd64
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/$(BINARY_NAME).exe .
	@cp -r conf views static $(BUILD_DIR)/windows-amd64/ 2>/dev/null || true
	@echo "✅ Windows版本构建完成"

.PHONY: build-darwin
build-darwin: deps ## 构建macOS版本
	@echo "🔄 构建macOS版本..."
	@mkdir -p $(BUILD_DIR)/darwin-amd64
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/darwin-amd64/$(BINARY_NAME) .
	@cp -r conf views static $(BUILD_DIR)/darwin-amd64/ 2>/dev/null || true
	@echo "✅ macOS版本构建完成"

.PHONY: build-all
build-all: build-linux build-windows build-darwin ## 构建所有平台
	@echo "🎉 所有平台构建完成"

# =============================================================================
# 前端构建
# =============================================================================

.PHONY: build-frontend
build-frontend: ## 构建前端项目
	@echo "🔄 构建前端项目..."
	@if [ -d "../game-admin" ]; then \
		cd ../game-admin && \
		if [ -f "package.json" ]; then \
			if command -v yarn >/dev/null 2>&1; then \
				yarn install && yarn build; \
			else \
				npm install && npm run build; \
			fi && \
			mkdir -p ../admin-service/$(BUILD_DIR)/frontend && \
			cp -r dist/* ../admin-service/$(BUILD_DIR)/frontend/; \
		else \
			echo "⚠️  未找到package.json，跳过前端构建"; \
		fi \
	else \
		echo "⚠️  前端目录不存在，跳过前端构建"; \
	fi
	@echo "✅ 前端构建完成"

# =============================================================================
# 发布和部署
# =============================================================================

.PHONY: release
release: clean build-all build-frontend ## 创建发布包
	@echo "🔄 创建发布包..."
	@mkdir -p $(RELEASE_DIR)
	@for platform in linux-amd64 windows-amd64 darwin-amd64; do \
		if [ -d "$(BUILD_DIR)/$$platform" ]; then \
			release_name="$(PROJECT_NAME)-$(VERSION)-$$platform"; \
			echo "📦 创建 $$release_name"; \
			mkdir -p "$(RELEASE_DIR)/$$release_name"; \
			cp -r "$(BUILD_DIR)/$$platform"/* "$(RELEASE_DIR)/$$release_name/"; \
			if [ -d "$(BUILD_DIR)/frontend" ]; then \
				mkdir -p "$(RELEASE_DIR)/$$release_name/static/admin"; \
				cp -r "$(BUILD_DIR)/frontend"/* "$(RELEASE_DIR)/$$release_name/static/admin/"; \
			fi; \
			cp install.sh install.bat "$(RELEASE_DIR)/$$release_name/" 2>/dev/null || true; \
			echo "# $(PROJECT_NAME) $(VERSION)" > "$(RELEASE_DIR)/$$release_name/README.md"; \
			echo "" >> "$(RELEASE_DIR)/$$release_name/README.md"; \
			echo "构建时间: $(BUILD_TIME)" >> "$(RELEASE_DIR)/$$release_name/README.md"; \
			echo "Git提交: $(GIT_COMMIT)" >> "$(RELEASE_DIR)/$$release_name/README.md"; \
			echo "平台: $$platform" >> "$(RELEASE_DIR)/$$release_name/README.md"; \
			cd $(RELEASE_DIR) && tar -czf "$$release_name.tar.gz" "$$release_name/" && cd ..; \
		fi \
	done
	@echo "🎉 发布包创建完成"

.PHONY: install
install: build ## 安装到本地
	@echo "🔄 安装到本地..."
	@sudo cp bin/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ 安装完成: /usr/local/bin/$(BINARY_NAME)"

# =============================================================================
# 运行和调试
# =============================================================================

.PHONY: run
run: build ## 运行服务
	@echo "🚀 启动服务..."
	@./bin/$(BINARY_NAME)

.PHONY: dev
dev: ## 开发模式运行
	@echo "🔧 开发模式启动..."
	@go run . -dev

.PHONY: debug
debug: ## 调试模式运行
	@echo "🐛 调试模式启动..."
	@dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient

# =============================================================================
# 部署相关
# =============================================================================

.PHONY: deploy-install
deploy-install: ## 运行安装脚本
	@echo "🔄 运行安装脚本..."
	@if [ -f "install.sh" ]; then \
		chmod +x install.sh && ./install.sh; \
	else \
		echo "❌ 安装脚本不存在"; \
	fi

.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "🔄 构建Docker镜像..."
	@if [ -f "Dockerfile" ]; then \
		docker build -t $(PROJECT_NAME):$(VERSION) .; \
		docker tag $(PROJECT_NAME):$(VERSION) $(PROJECT_NAME):latest; \
	else \
		echo "⚠️  Dockerfile不存在，跳过Docker构建"; \
	fi
	@echo "✅ Docker镜像构建完成"

# =============================================================================
# 信息显示
# =============================================================================

.PHONY: info
info: ## 显示项目信息
	@echo "📋 项目信息:"
	@echo "├── 项目名称: $(PROJECT_NAME)"
	@echo "├── 版本: $(VERSION)"
	@echo "├── 构建时间: $(BUILD_TIME)"
	@echo "├── Git提交: $(GIT_COMMIT)"
	@echo "├── Go版本: $(shell go version | awk '{print $$3}')"
	@echo "└── 构建标签: $(LDFLAGS)"
	@echo ""

.PHONY: status
status: ## 检查构建状态
	@echo "📊 构建状态:"
	@if [ -d "bin" ]; then \
		echo "├── 本地构建: ✅"; \
		ls -la bin/; \
	else \
		echo "├── 本地构建: ❌"; \
	fi
	@if [ -d "$(BUILD_DIR)" ]; then \
		echo "├── 多平台构建: ✅"; \
		ls -la $(BUILD_DIR)/; \
	else \
		echo "├── 多平台构建: ❌"; \
	fi
	@if [ -d "$(RELEASE_DIR)" ]; then \
		echo "└── 发布包: ✅"; \
		ls -la $(RELEASE_DIR)/; \
	else \
		echo "└── 发布包: ❌"; \
	fi
	@echo ""
