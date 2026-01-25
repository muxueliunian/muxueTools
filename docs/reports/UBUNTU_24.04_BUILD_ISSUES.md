# Ubuntu 24.04 构建问题记录

> 测试日期：2026-01-26  
> 测试环境：Ubuntu 24.04.3 LTS (Noble Numbat)  
> MuxueTools 版本：0.3.1

## 测试概览

本次测试按照 `UBUNTU_TESTING_GUIDE.md` 中的流程，在 Ubuntu 24.04 双系统上进行了完整的构建和打包测试。

## 遇到的问题与解决方案

### 问题 1: webkit2gtk-4.0 依赖缺失 ⚠️ 重要

#### 问题描述

运行 `./scripts/build.sh all` 构建 Desktop 版本时，编译失败并报错：

```bash
# github.com/webview/webview_go
# [pkg-config --cflags  -- gtk+-3.0 webkit2gtk-4.0]
Package webkit2gtk-4.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `webkit2gtk-4.0.pc'
to the PKG_CONFIG_PATH environment variable
Package 'webkit2gtk-4.0', required by 'virtual:world', not found
```

#### 根本原因

1. **Ubuntu 24.04 移除了 webkit2gtk-4.0**  
   Ubuntu 24.04 官方软件仓库只提供 `libwebkit2gtk-4.1-dev`，不再提供 `libwebkit2gtk-4.0-dev`。

2. **webview_go 库硬编码依赖**  
   `github.com/webview/webview_go` 库在其源代码中硬编码了 CGO 指令：
   ```go
   #cgo linux openbsd freebsd netbsd pkg-config: gtk+-3.0 webkit2gtk-4.0
   ```

3. **API/ABI 差异**  
   - `webkit2gtk-4.0` 使用 **libsoup 2.4** 网络库
   - `webkit2gtk-4.1` 使用 **libsoup 3.0** 网络库
   - 两者二进制不兼容，但 API 几乎相同

#### 解决方案

修改 `scripts/build.sh` 中的 `build_desktop()` 函数，创建一个 **pkg-config 包装器**，将 `webkit2gtk-4.0` 的请求重定向到 `webkit2gtk-4.1`：

```bash
if [ "$webkit_version" = "webkit2gtk-4.1" ]; then
    echo -e "${YELLOW}Using webkit2gtk-4.1 (Ubuntu 24.04+)${NC}"
    echo -e "${CYAN}Creating pkg-config wrapper to redirect webkit2gtk-4.0 -> webkit2gtk-4.1${NC}"
    
    # 创建临时目录存放包装器
    local wrapper_dir=$(mktemp -d)
    trap "rm -rf $wrapper_dir" EXIT
    
    # 创建 pkg-config 包装器脚本
    cat > "$wrapper_dir/pkg-config" << 'WRAPPER_EOF'
#!/bin/bash
# Wrapper script to redirect webkit2gtk-4.0 requests to webkit2gtk-4.1

# 替换参数中的 webkit2gtk-4.0 为 webkit2gtk-4.1
args=()
for arg in "$@"; do
    if [ "$arg" = "webkit2gtk-4.0" ]; then
        args+=("webkit2gtk-4.1")
    else
        args+=("$arg")
    fi
done

# 调用真正的 pkg-config
exec /usr/bin/pkg-config "${args[@]}"
WRAPPER_EOF
    
    chmod +x "$wrapper_dir/pkg-config"
    
    # 将包装器目录放到 PATH 最前面
    export PATH="$wrapper_dir:$PATH"
    
    echo -e "${GREEN}pkg-config wrapper created at: $wrapper_dir/pkg-config${NC}"
fi
```

#### 测试结果

✅ 构建成功  
✅ Desktop 程序正常运行  
✅ WebView 窗口正常显示

---

### 问题 2: .deb 包安装后应用名称显示为中文

#### 问题描述

使用 `./scripts/installer/linux/build-deb.sh` 生成 .deb 包并安装后，在应用菜单中看到的名称为 **"沐雪工具"** 而不是 **"MuxueTools"**。

#### 根本原因

`scripts/installer/linux/build-deb.sh` 脚本在生成 `.desktop` 文件时，包含了中文本地化名称：

```ini
[Desktop Entry]
Name=MuxueTools
Name[zh_CN]=沐雪工具    # ← 这行导致在中文环境下显示中文名
Comment=Gemini to OpenAI API Proxy
Comment[zh_CN]=Gemini 转 OpenAI API 代理
```

#### 解决方案

修改 `scripts/installer/linux/build-deb.sh` 第 65-78 行，移除 `Name[zh_CN]` 字段：

```bash
cat > "$BUILD_DIR/usr/share/applications/muxuetools.desktop" << EOF
[Desktop Entry]
Name=MuxueTools
# Name[zh_CN]=沐雪工具  ← 已删除
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
```

同时为保持一致性，也修改了 `scripts/installer/linux/muxuetools.desktop` 模板文件（虽然脚本当前并未使用它）。

#### 测试结果

✅ 重新构建 .deb 包后，应用名称正确显示为 "MuxueTools"

---

## 构建产物验证

### 二进制文件

```bash
$ ls -la bin/
-rwxrwxr-x 1 muxueliunian muxueliunian 38032928 1月 26 06:31 muxuetools-desktop
-rwxrwxr-x 1 muxueliunian muxueliunian 38411744 1月 26 06:19 muxuetools-server
```

### .deb 包

```bash
$ ls -la dist/
-rw-r--r-- 1 muxueliunian muxueliunian 19138326 1月 26 06:39 muxuetools_0.3.1_amd64.deb
```

### 版本信息

```bash
$ ./bin/muxuetools-server --version
MuxueTools 0.3.1
  Build Time: 2026-01-26 06:18:41
  Git Commit: 79b448c
```

### 数据库路径确认

根据源代码 `internal/config/paths.go`：

- **标准模式**：`~/.local/share/muxuetools/muxuetools.db`
- **便携模式**：`./data/muxuetools.db`（当前目录下有 config.yaml 时）
- **配置指定**：使用 config.yaml 中指定的路径

---

## 其他观察与讨论

### Linux vs Windows 安装分发机制差异

在测试过程中，讨论了 Linux 与 Windows 在软件分发、安装和更新方面的差异：

| 方面 | Windows (Inno Setup) | Linux (.deb) |
|------|---------------------|--------------|
| **安装** | 双击 `.exe` → 安装向导 → 自动完成 | 双击 `.deb` → 软件中心 → 输入密码安装 |
| **更新** | 程序内检测 → 下载新版 Setup → 自动升级 | 需要 PPA/仓库 或 手动下载新 .deb 覆盖 |
| **卸载** | 控制面板 → 卸载程序 | `sudo dpkg -r package` 或软件中心 |
| **依赖处理** | 打包进安装包 | 由 apt 自动处理（如果配置了仓库） |

### 改进建议

由于 MuxueTools 目前没有搭建 APT 仓库（PPA），建议在 **程序内部实现自动更新检测**：

1. 启动时检查 GitHub Releases API
2. 发现新版本提示用户
3. 下载新 `.deb` 到 `/tmp`
4. 调用 `pkexec dpkg -i /tmp/muxuetools_x.x.x_amd64.deb` 自动安装

这样可以提供类似 Windows 的无缝更新体验。

---

## 测试检查清单

- [x] Go 1.22+ 已安装
- [x] Node.js 20+ 已安装
- [x] GTK3 + WebKit2GTK 4.1 开发依赖已安装
- [x] `./scripts/build.sh all` 成功
- [x] `bin/muxuetools-server` 存在并可运行
- [x] `bin/muxuetools-desktop` 存在并可运行
- [x] Server 版本 `--version` 显示正确版本
- [x] Desktop GUI 窗口正常显示
- [x] `.deb` 包成功生成
- [x] `.deb` 包可正常安装和卸载
- [x] 应用名称正确显示（无中文）

---

## 文件修改记录

### 修改的文件

1. **scripts/build.sh**  
   - 新增：webkit2gtk-4.1 pkg-config 包装器逻辑（L153-187）
   - 目的：解决 Ubuntu 24.04 上的构建问题

2. **scripts/installer/linux/build-deb.sh**  
   - 移除：`Name[zh_CN]=沐雪工具`（L68）
   - 目的：去除中文应用名称

3. **scripts/installer/linux/muxuetools.desktop**  
   - 移除：`Name[zh_CN]=沐雪工具`（L3）
   - 目的：保持模板一致性

### 构建命令总结

```bash
# 1. 构建所有目标
./scripts/build.sh all

# 2. 构建 .deb 包
cd scripts/installer/linux
./build-deb.sh 0.3.1

# 3. 安装
sudo dpkg -i ../../../dist/muxuetools_0.3.1_amd64.deb

# 4. 卸载
sudo dpkg -r muxuetools
```

---

## 下一步行动

- [ ] 考虑添加程序内自动更新功能
- [ ] 考虑发布到 Flathub 或 Snap Store
- [ ] 考虑构建 AppImage 版本（便携式）
- [ ] 建立 GitHub Actions 自动化构建流程
- [ ] 考虑为 Ubuntu 22.04/20.04 兼容性测试（仍使用 webkit2gtk-4.0）

---

*文档创建时间：2026-01-26*  
*测试人员：muxueliunian*
