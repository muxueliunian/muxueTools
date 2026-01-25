# 任务：实现 Linux 安装包（Phase 4）

## 角色
Developer (senior-golang skill)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`

## 背景

MuxueTools 已完成以下跨平台适配工作：
- **Phase 1**: 数据路径重构 ✅ 
- **Phase 2**: Linux 构建系统 ✅ (`scripts/build.sh`)
- **Phase 3**: Windows 安装程序 ✅ (`scripts/installer/windows/setup.iss`)

当前 Linux 版本仅能通过构建脚本生成二进制文件，需要用户手动安装。为提升 Ubuntu 用户体验，需要创建 `.deb` 安装包。

**目标平台**: Ubuntu 24 LTS (amd64)

**依赖项说明**:
- Desktop 版本需要 GTK3 + WebKit2GTK 运行时
- Server 版本是纯 Go 编译，无外部依赖

## 目标

| 任务 | 目标 | 优先级 |
|------|------|--------|
| **4.1** | 新增 `scripts/installer/linux/DEBIAN/control` | ⭐⭐⭐ 高 |
| **4.2** | 新增 `scripts/installer/linux/DEBIAN/postinst` | ⭐⭐ 中 |
| **4.3** | 新增 `scripts/installer/linux/DEBIAN/prerm` | ⭐⭐ 中 |
| **4.4** | 新增 `scripts/installer/linux/build-deb.sh` | ⭐⭐⭐ 高 |
| **4.5** | 新增 `.desktop` 文件 (桌面菜单项) | ⭐ 低 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **项目文档**
   - `docs/UBUNTU_ADAPTATION_PLAN.md` - 完整适配计划（Section 3.2, 4）
   - `docs/ARCHITECTURE.md` - 项目目录结构
   - `scripts/build.sh` - 现有 Linux 构建脚本

2. **外部参考**
   - [Debian 打包指南](https://www.debian.org/doc/manuals/maint-guide/)
   - [XDG Desktop Entry 规范](https://specifications.freedesktop.org/desktop-entry-spec/latest/)

### 阶段 1：创建 DEBIAN 控制文件

**新增目录**: `scripts/installer/linux/DEBIAN/`

#### 4.1 control 文件

```
Package: muxuetools
Version: 0.3.1
Section: net
Priority: optional
Architecture: amd64
Depends: libgtk-3-0, libwebkit2gtk-4.1-0
Maintainer: muxueliunian <muxueliunian@example.com>
Description: Gemini to OpenAI API Proxy
 MuxueTools is a reverse proxy that converts Google AI Studio (Gemini) API
 to OpenAI-compatible format, enabling seamless integration with existing
 OpenAI clients.
 .
 Features:
  - API key pool management
  - Request statistics and monitoring
  - Desktop GUI with embedded WebView
Homepage: https://github.com/muxueliunian/muxueTools
```

> [!NOTE]
> `Depends` 声明的是运行时依赖，安装时 dpkg 会检查并提示用户安装缺少的包。

#### 4.2 postinst 脚本 (安装后执行)

```bash
#!/bin/bash
set -e

# 设置执行权限
chmod 755 /opt/muxuetools/muxuetools

# 更新桌面菜单缓存
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database /usr/share/applications 2>/dev/null || true
fi

echo ""
echo "============================================"
echo " MuxueTools 安装成功!"
echo "============================================"
echo ""
echo " 启动命令: muxuetools"
echo " 数据目录: ~/.local/share/muxuetools/"
echo " 配置目录: ~/.config/muxuetools/"
echo ""
echo " 桌面用户也可从应用程序菜单启动"
echo "============================================"
```

#### 4.3 prerm 脚本 (卸载前执行)

```bash
#!/bin/bash
set -e

# 提示用户数据不会被删除
echo ""
echo "============================================"
echo " 正在卸载 MuxueTools..."
echo "============================================"
echo ""
echo " 注意: 用户数据将保留在以下位置:"
echo "   ~/.local/share/muxuetools/"
echo "   ~/.config/muxuetools/"
echo ""
echo " 如需完全删除，请手动删除上述目录"
echo "============================================"
```

### 阶段 2：创建构建脚本

#### 4.4 build-deb.sh

**文件**: `scripts/installer/linux/build-deb.sh`

```bash
#!/bin/bash
# MuxueTools .deb Package Builder
# Usage: ./build-deb.sh [version]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"

VERSION="${1:-0.3.1}"
ARCH="amd64"
PKG_NAME="muxuetools_${VERSION}_${ARCH}"

echo "Building MuxueTools .deb package v${VERSION}..."

# 检查二进制文件是否存在
BINARY_PATH="$PROJECT_ROOT/bin/muxuetools-desktop"
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH"
    echo "Please run './scripts/build.sh desktop' first"
    exit 1
fi

# 创建包目录结构
BUILD_DIR="$SCRIPT_DIR/build/$PKG_NAME"
rm -rf "$BUILD_DIR"

# 核心目录
mkdir -p "$BUILD_DIR/opt/muxuetools"
mkdir -p "$BUILD_DIR/usr/local/bin"
mkdir -p "$BUILD_DIR/usr/share/applications"
mkdir -p "$BUILD_DIR/usr/share/icons/hicolor/256x256/apps"
mkdir -p "$BUILD_DIR/DEBIAN"

# 复制程序文件
cp "$BINARY_PATH" "$BUILD_DIR/opt/muxuetools/muxuetools"
cp "$PROJECT_ROOT/web/dist" "$BUILD_DIR/opt/muxuetools/web/dist" -r
cp "$PROJECT_ROOT/configs/config.example.yaml" "$BUILD_DIR/opt/muxuetools/"

# 创建启动脚本 (wrapper)
cat > "$BUILD_DIR/usr/local/bin/muxuetools" << 'EOF'
#!/bin/bash
exec /opt/muxuetools/muxuetools "$@"
EOF
chmod +x "$BUILD_DIR/usr/local/bin/muxuetools"

# 复制图标
if [ -f "$PROJECT_ROOT/assets/icon.png" ]; then
    cp "$PROJECT_ROOT/assets/icon.png" "$BUILD_DIR/usr/share/icons/hicolor/256x256/apps/muxuetools.png"
fi

# 创建 .desktop 文件
cat > "$BUILD_DIR/usr/share/applications/muxuetools.desktop" << EOF
[Desktop Entry]
Name=MuxueTools
Comment=Gemini to OpenAI API Proxy
Exec=/opt/muxuetools/muxuetools
Icon=muxuetools
Terminal=false
Type=Application
Categories=Network;Development;
Keywords=gemini;openai;api;proxy;
StartupWMClass=muxuetools
EOF

# 复制 DEBIAN 控制文件
cp "$SCRIPT_DIR/DEBIAN/"* "$BUILD_DIR/DEBIAN/"

# 更新版本号
sed -i "s/^Version:.*/Version: $VERSION/" "$BUILD_DIR/DEBIAN/control"

# 设置权限
chmod 755 "$BUILD_DIR/DEBIAN/postinst" 2>/dev/null || true
chmod 755 "$BUILD_DIR/DEBIAN/prerm" 2>/dev/null || true

# 构建 deb 包
echo "Creating .deb package..."
dpkg-deb --build --root-owner-group "$BUILD_DIR"

# 移动到 dist 目录
mkdir -p "$PROJECT_ROOT/dist"
mv "$SCRIPT_DIR/build/$PKG_NAME.deb" "$PROJECT_ROOT/dist/"

# 清理构建目录
rm -rf "$SCRIPT_DIR/build"

echo ""
echo "============================================"
echo " Package created: dist/$PKG_NAME.deb"
echo "============================================"
echo ""
echo " Install: sudo dpkg -i dist/$PKG_NAME.deb"
echo " Fix deps: sudo apt-get install -f"
echo " Remove:  sudo dpkg -r muxuetools"
echo "============================================"
```

### 阶段 3：创建桌面文件

#### 4.5 muxuetools.desktop

已包含在 `build-deb.sh` 中动态生成，也可单独放置：

**文件**: `scripts/installer/linux/muxuetools.desktop`

```ini
[Desktop Entry]
Name=MuxueTools
Name[zh_CN]=沐雪工具
Comment=Gemini to OpenAI API Proxy
Comment[zh_CN]=Gemini 转 OpenAI API 代理
Exec=/opt/muxuetools/muxuetools
Icon=muxuetools
Terminal=false
Type=Application
Categories=Network;Development;
Keywords=gemini;openai;api;proxy;
StartupWMClass=muxuetools
```

### 阶段 4：验证安装包

```bash
# 1. 构建 Desktop 二进制 (需要在 Ubuntu 上执行)
./scripts/build.sh desktop

# 2. 构建 .deb 包
cd scripts/installer/linux
chmod +x build-deb.sh
./build-deb.sh 0.3.1

# 3. 安装
sudo dpkg -i ../../dist/muxuetools_0.3.1_amd64.deb

# 4. 安装缺失依赖 (如有)
sudo apt-get install -f

# 5. 验证
muxuetools --version
ls ~/.local/share/muxuetools/

# 6. 启动程序
muxuetools

# 7. 卸载
sudo dpkg -r muxuetools
```

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `scripts/installer/linux/DEBIAN/control` | **NEW** | 包元信息 (依赖声明) |
| `scripts/installer/linux/DEBIAN/postinst` | **NEW** | 安装后脚本 |
| `scripts/installer/linux/DEBIAN/prerm` | **NEW** | 卸载前脚本 |
| `scripts/installer/linux/build-deb.sh` | **NEW** | .deb 构建脚本 |
| `scripts/installer/linux/muxuetools.desktop` | **NEW** | 桌面菜单项 (可选) |

## 约束

### 技术约束
- 目标平台: Ubuntu 24 LTS (amd64)
- 依赖包: `libgtk-3-0`, `libwebkit2gtk-4.1-0`
- 安装路径: `/opt/muxuetools/`
- 用户数据: 遵循 XDG 规范 (`~/.local/share/`, `~/.config/`)

### 质量约束
- 遵循 `.agent/skills/senior-golang/SKILL.md` 代码规范
- Shell 脚本使用 `set -e` 确保错误时退出
- 所有脚本使用 `shellcheck` 检查

### 兼容性约束
- 支持 Ubuntu 22.04+ (WebKit 4.0/4.1 自动适配)
- 用户数据与 Windows 版本独立

## 验收标准

- [ ] 所有 DEBIAN 控制文件语法正确 (无 trailing whitespace)
- [ ] `./build-deb.sh` 成功生成 `dist/muxuetools_*.deb`
- [ ] `sudo dpkg -i *.deb` 安装成功
- [ ] `muxuetools --version` 显示正确版本
- [ ] 程序可正常启动并显示 GUI
- [ ] 应用程序菜单中可见 MuxueTools 图标
- [ ] `sudo dpkg -r muxuetools` 卸载成功
- [ ] 卸载后用户数据保留

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/UBUNTU_ADAPTATION_PLAN.md` | 更新 Phase 4 状态为已完成 |
| `README.md` | 添加 Linux 安装说明 |

## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的开发流程。

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| WebKit 版本不匹配 | 程序无法启动 | `Depends` 声明 libwebkit2gtk-4.1-0 |
| 权限问题 | 安装失败 | 使用 `--root-owner-group` 选项 |
| 图标缺失 | 菜单显示空白图标 | 提供 PNG 格式图标 (256x256) |
| 前端资源缺失 | 界面空白 | 检查 web/dist 是否存在 |

---

*任务创建时间: 2026-01-26*
