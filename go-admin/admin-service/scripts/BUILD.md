# 构建说明文档

## 📋 概述

本项目提供了多种构建方式，支持跨平台编译和自动化部署。

## 🛠️ 构建工具

### 1. 快速构建（开发用）

最简单的构建方式，只构建当前平台：

```bash
# Linux/macOS
./quick-build.sh

# Windows
quick-build.bat
```

### 2. 完整构建脚本

支持多平台构建和发布包创建：

```bash
# Linux/macOS
./build.sh

# Windows
build.bat
```

### 3. Makefile（推荐）

提供丰富的构建选项：

```bash
make help          # 查看所有可用命令
make build         # 构建当前平台
make build-all     # 构建所有平台
make release       # 创建发布包
```

## 🎯 构建选项

### 环境变量

- `VERSION`: 设置版本号（默认：1.0.0）
- `BUILD_BACKEND`: 是否构建后端（true/false）
- `BUILD_FRONTEND`: 是否构建前端（true/false）
- `CREATE_RELEASE`: 是否创建发布包（true/false）

### 命令行参数

#### build.sh / build.bat
```bash
./build.sh --help                    # 显示帮助
./build.sh --backend-only            # 只构建后端
./build.sh --frontend-only           # 只构建前端
./build.sh --release                 # 创建发布包
./build.sh --clean                   # 清理构建文件
./build.sh --version 2.0.0           # 设置版本号
```

#### Makefile
```bash
make build          # 构建当前平台
make build-linux    # 构建Linux版本
make build-windows  # 构建Windows版本
make build-darwin   # 构建macOS版本
make build-all      # 构建所有平台
make release        # 创建发布包
make clean          # 清理构建文件
```

## 🏗️ 支持的平台

- **Linux**: amd64, arm64
- **Windows**: amd64
- **macOS**: amd64, arm64

## 📁 输出目录结构

```
dist/
├── backend/
│   ├── linux-amd64/
│   │   ├── admin-service
│   │   ├── conf/
│   │   ├── views/
│   │   ├── static/
│   │   └── start.sh
│   ├── windows-amd64/
│   │   ├── admin-service.exe
│   │   ├── conf/
│   │   ├── views/
│   │   ├── static/
│   │   └── start.bat
│   └── ...
└── frontend/
    ├── index.html
    ├── assets/
    └── ...

release/
├── minigame-admin-service-1.0.0-linux-amd64.tar.gz
├── minigame-admin-service-1.0.0-windows-amd64.zip
└── ...
```

## 🚀 快速开始

### 开发环境

1. **快速构建和运行**：
   ```bash
   ./quick-build.sh    # 或 quick-build.bat
   ./bin/admin-service
   ```

2. **使用Makefile**：
   ```bash
   make build
   make run
   ```

### 生产环境

1. **创建发布包**：
   ```bash
   ./build.sh --release
   # 或
   make release
   ```

2. **部署**：
   ```bash
   # 解压发布包
   tar -xzf minigame-admin-service-1.0.0-linux-amd64.tar.gz
   cd minigame-admin-service-1.0.0-linux-amd64/
   
   # 运行安装脚本
   ./install.sh
   
   # 或手动启动
   ./start.sh
   ```

## 🔧 依赖要求

### 必需
- **Go 1.21+**: 后端编译
- **Git**: 获取版本信息（可选）

### 可选
- **Node.js**: 前端构建
- **Make**: 使用Makefile
- **Docker**: 容器化部署

## 📝 构建示例

### 示例1：开发构建
```bash
# 快速构建当前平台
./quick-build.sh

# 运行服务
./bin/admin-service
```

### 示例2：生产构建
```bash
# 设置版本号并创建发布包
VERSION=2.1.0 ./build.sh --release

# 或使用Makefile
make release VERSION=2.1.0
```

### 示例3：只构建后端
```bash
# 只构建Go后端服务
./build.sh --backend-only

# 或使用环境变量
BUILD_FRONTEND=false ./build.sh
```

### 示例4：自定义构建
```bash
# 使用环境变量控制构建
export VERSION=1.2.0
export BUILD_BACKEND=true
export BUILD_FRONTEND=true
export CREATE_RELEASE=true

./build.sh
```

## 🐛 故障排除

### 常见问题

1. **Go环境未找到**
   ```
   ❌ 未找到Go环境，请先安装Go 1.21+
   ```
   解决：安装Go语言环境

2. **前端构建失败**
   ```
   ⚠️ 未找到Node.js环境，跳过前端构建
   ```
   解决：安装Node.js或设置`BUILD_FRONTEND=false`

3. **权限错误**
   ```bash
   chmod +x build.sh
   chmod +x quick-build.sh
   ```

4. **清理构建文件**
   ```bash
   ./build.sh --clean
   # 或
   make clean
   ```

### 调试模式

```bash
# 使用详细输出
set -x  # Linux/macOS
./build.sh

# 检查构建状态
make status
```

## 📚 更多信息

- 查看 `install.sh` / `install.bat` 了解部署脚本
- 查看 `Makefile` 了解所有可用命令
- 查看源码中的构建标签和版本信息

## 🎉 构建完成

构建成功后，你将得到：
- 可执行的二进制文件
- 完整的配置文件
- 启动脚本
- 发布包（如果选择）

现在可以部署和运行你的Minigame Admin Service了！
