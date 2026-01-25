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

# 检查前端资源是否存在
WEB_DIST_PATH="$PROJECT_ROOT/web/dist"
if [ ! -d "$WEB_DIST_PATH" ]; then
    echo "Error: Frontend resources not found at $WEB_DIST_PATH"
    echo "Please run './scripts/build.sh desktop' first"
    exit 1
fi

# 创建包目录结构
BUILD_DIR="$SCRIPT_DIR/build/$PKG_NAME"
rm -rf "$BUILD_DIR"

# 核心目录
mkdir -p "$BUILD_DIR/opt/muxuetools"
mkdir -p "$BUILD_DIR/opt/muxuetools/web"
mkdir -p "$BUILD_DIR/usr/local/bin"
mkdir -p "$BUILD_DIR/usr/share/applications"
mkdir -p "$BUILD_DIR/usr/share/icons/hicolor/256x256/apps"
mkdir -p "$BUILD_DIR/DEBIAN"

# 复制程序文件
cp "$BINARY_PATH" "$BUILD_DIR/opt/muxuetools/muxuetools"
cp -r "$WEB_DIST_PATH" "$BUILD_DIR/opt/muxuetools/web/dist"
cp "$PROJECT_ROOT/configs/config.example.yaml" "$BUILD_DIR/opt/muxuetools/"

# 创建启动脚本 (wrapper)
cat > "$BUILD_DIR/usr/local/bin/muxuetools" << 'EOF'
#!/bin/bash
exec /opt/muxuetools/muxuetools "$@"
EOF
chmod +x "$BUILD_DIR/usr/local/bin/muxuetools"

# 复制图标 (优先使用 web/public/icon.png)
ICON_PATH="$PROJECT_ROOT/web/public/icon.png"
if [ -f "$ICON_PATH" ]; then
    cp "$ICON_PATH" "$BUILD_DIR/usr/share/icons/hicolor/256x256/apps/muxuetools.png"
elif [ -f "$PROJECT_ROOT/assets/icon.png" ]; then
    cp "$PROJECT_ROOT/assets/icon.png" "$BUILD_DIR/usr/share/icons/hicolor/256x256/apps/muxuetools.png"
fi

# 创建 .desktop 文件
cat > "$BUILD_DIR/usr/share/applications/muxuetools.desktop" << EOF
[Desktop Entry]
Name=MuxueTools
Comment=Gemini to OpenAI API Proxy
Comment[zh_CN]=Gemini 转 OpenAI API 代理
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
