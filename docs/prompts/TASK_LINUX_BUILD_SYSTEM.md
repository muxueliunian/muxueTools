# 任务：实现 Linux 构建系统（Phase 2）

## 角色
Developer (senior-golang skill)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`

## 背景

MuxueTools 当前仅支持 Windows 平台构建，使用 PowerShell 脚本 `scripts/build.ps1`。为支持 Ubuntu 24 LTS，需要创建 Linux 构建系统。

**当前状态：**
- ✅ PowerShell 构建脚本 `scripts/build.ps1`（Windows Only）
- ❌ 无 Linux 构建脚本
- ❌ 无 `.goreleaser.yaml`（可选）

**依赖完成：**
- Phase 1: 数据路径重构 ✅ 已完成

## 目标

| 任务 | 目标 | 优先级 |
|------|------|--------|
| **2.1** | 新增 `scripts/build.sh` Linux Bash 构建脚本 | ⭐⭐⭐ 高 |
| **2.2** | 验证 Linux 构建流程 | ⭐⭐ 中 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范**
   - `.agent/skills/senior-golang/SKILL.md`

2. **项目文档**
   - `docs/ARCHITECTURE.md` - 系统架构
   - `docs/UBUNTU_ADAPTATION_PLAN.md` - 完整适配计划

3. **相关代码**
   - `scripts/build.ps1` - 现有 Windows 构建脚本（参考结构）
   - `cmd/server/main.go` - Server 入口
   - `cmd/desktop/main.go` - Desktop 入口

### 阶段 1：创建 Linux 构建脚本

**新增文件**: `scripts/build.sh`

```bash
#!/bin/bash
# MuxueTools Build Script (Linux/macOS)
# Usage: ./build.sh [target]
#   target: server | desktop | all | clean

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$PROJECT_ROOT/bin"

# Version info
VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "dev")}"
BUILD_TIME="$(date '+%Y-%m-%d %H:%M:%S')"
GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"

# LDFlags for version injection
LDFLAGS="-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME' -X 'main.GitCommit=$GIT_COMMIT'"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN} $1${NC}"
    echo -e "${CYAN}========================================${NC}\n"
}

ensure_bin_dir() {
    if [ ! -d "$BIN_DIR" ]; then
        mkdir -p "$BIN_DIR"
        echo -e "${GREEN}Created bin directory: $BIN_DIR${NC}"
    fi
}

build_frontend() {
    print_header "Building Frontend (Vite + Vue)"
    
    cd "$PROJECT_ROOT/web"
    
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}Installing dependencies...${NC}"
        npm ci
    fi
    
    echo -e "${CYAN}Building frontend...${NC}"
    npm run build
    
    echo -e "${GREEN}Frontend built successfully: web/dist${NC}"
    cd "$PROJECT_ROOT"
}

build_server() {
    print_header "Building Server (Pure Go)"
    
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
    
    OUTPUT_PATH="$BIN_DIR/muxuetools-server"
    
    cd "$PROJECT_ROOT"
    go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" ./cmd/server
    
    echo -e "${GREEN}Built: $OUTPUT_PATH${NC}"
}

build_desktop() {
    print_header "Building Desktop (CGO + WebView)"
    
    # Check for GTK dependencies
    if ! pkg-config --exists gtk+-3.0 webkit2gtk-4.1 2>/dev/null; then
        echo -e "${RED}Error: GTK/WebKit dependencies not found.${NC}"
        echo -e "${YELLOW}Please install: sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev${NC}"
        exit 1
    fi
    
    export CGO_ENABLED=1
    export GOOS=linux
    export GOARCH=amd64
    
    OUTPUT_PATH="$BIN_DIR/muxuetools-desktop"
    
    cd "$PROJECT_ROOT"
    go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" ./cmd/desktop
    
    echo -e "${GREEN}Built: $OUTPUT_PATH${NC}"
}

clean_build() {
    print_header "Cleaning build artifacts"
    
    if [ -d "$BIN_DIR" ]; then
        rm -rf "$BIN_DIR"
        echo -e "${YELLOW}Removed: $BIN_DIR${NC}"
    fi
    
    echo -e "${GREEN}Clean complete${NC}"
}

# Main
echo -e "${CYAN}MuxueTools Build Script (Linux)${NC}"
echo "Version: $VERSION | Commit: $GIT_COMMIT"

TARGET="${1:-all}"

ensure_bin_dir

case "$TARGET" in
    server)
        build_server
        ;;
    desktop)
        build_frontend
        build_desktop
        ;;
    all)
        build_frontend
        build_server
        build_desktop
        ;;
    clean)
        clean_build
        ;;
    *)
        echo "Usage: $0 [server|desktop|all|clean]"
        exit 1
        ;;
esac

echo -e "\n${GREEN}Build complete!${NC}"
```

### 阶段 2：验证构建

在 Ubuntu 环境中执行以下命令验证：

```bash
# 安装依赖（如果需要构建 desktop）
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev

# 运行构建脚本
chmod +x scripts/build.sh
./scripts/build.sh all

# 验证输出
ls -la bin/
./bin/muxuetools-server --version
```

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `scripts/build.sh` | **NEW** | Linux Bash 构建脚本 |

## 约束

### 技术约束
- Bash 4.0+
- Go 1.22+
- Node.js 18+ (用于前端构建)
- Desktop 构建需要 GTK3 + WebKit2GTK 依赖

### 质量约束
- 脚本需与 `build.ps1` 功能对等
- 支持 `server`, `desktop`, `all`, `clean` 四个目标
- 包含版本注入（-ldflags）

### 兼容性约束
- 支持 Ubuntu 24 LTS
- 支持 macOS（可选）

## 验收标准

- [x] `chmod +x scripts/build.sh` 无错误
- [x] `./scripts/build.sh server` 成功生成 `bin/muxuetools-server`
- [x] `./scripts/build.sh desktop` 成功生成 `bin/muxuetools-desktop`（需 GTK 依赖）
- [x] `./scripts/build.sh clean` 清理 bin 目录
- [x] `./bin/muxuetools-server --version` 显示正确版本

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/UBUNTU_ADAPTATION_PLAN.md` | 更新 Phase 2 状态为已完成 |
| `README.md` | 添加 Linux 构建说明 |

## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的开发流程。

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| GTK 依赖未安装 | Desktop 构建失败 | 脚本检测并提示安装命令 |
| Node.js 未安装 | 前端构建失败 | 脚本检测并提示 |
| 权限问题 | 脚本无法执行 | 使用 `chmod +x` |

---

*任务创建时间: 2026-01-25*
