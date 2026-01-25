# MuxueTools 跨平台适配与安装程序计划

> 版本: 2.0  
> 创建日期: 2026-01-25  
> 状态: 待实施

---

## 1. 概述

本文档描述 MuxueTools 跨平台适配方案，包括：
- **平台支持**: Windows 11 + Ubuntu 24 LTS
- **数据路径标准化**: 使用系统标准用户数据目录
- **安装程序**: Windows Inno Setup + Linux .deb 包
- **CI/CD 集成**: GitHub Actions 自动构建安装包

### 1.1 当前状态

| 项目 | 状态 |
|------|------|
| **支持平台** | Windows (桌面应用) |
| **数据存储** | 相对路径 `data/MuxueTools.db` |
| **安装方式** | 手动解压运行 |
| **构建脚本** | PowerShell Only |

### 1.2 目标状态

| 项目 | 目标 |
|------|------|
| **支持平台** | Windows 11 + Ubuntu 24 LTS |
| **数据存储** | 系统标准路径 (见下表) |
| **安装方式** | Setup 安装程序 |
| **构建产物** | `.exe` 安装包 (Win) + `.deb` 包 (Linux) |

---

## 2. 数据路径标准化

### 2.1 路径规范

| 平台 | 类型 | 路径 |
|------|------|------|
| **Windows** | 用户数据 | `%APPDATA%\MuxueTools\` |
| | 配置文件 | `%APPDATA%\MuxueTools\config.yaml` |
| | 数据库 | `%APPDATA%\MuxueTools\data\muxuetools.db` |
| | 日志 | `%APPDATA%\MuxueTools\logs\` |
| | 安装目录 | `%ProgramFiles%\MuxueTools\` (可自定义) |
| **Linux** | 用户数据 | `~/.local/share/muxuetools/` |
| | 配置文件 | `~/.config/muxuetools/config.yaml` |
| | 数据库 | `~/.local/share/muxuetools/muxuetools.db` |
| | 日志 | `~/.local/share/muxuetools/logs/` |
| | 安装目录 | `/opt/muxuetools/` 或 `/usr/local/bin/` |

> [!NOTE]
> 用户数据路径是固定的系统标准路径，程序自动扫描。  
> 安装目录可由用户在 Setup 安装时自定义。

### 2.2 代码实现

**新增文件**: `internal/config/paths.go`

```go
package config

import (
    "os"
    "path/filepath"
    "runtime"
)

const AppName = "MuxueTools"

// GetDataDir 返回用户数据目录（自动扫描，不可自定义）
func GetDataDir() string {
    switch runtime.GOOS {
    case "windows":
        // Windows: %APPDATA%\MuxueTools\data
        appData := os.Getenv("APPDATA")
        if appData == "" {
            home, _ := os.UserHomeDir()
            appData = filepath.Join(home, "AppData", "Roaming")
        }
        return filepath.Join(appData, AppName, "data")
    case "linux":
        // Linux: ~/.local/share/muxuetools
        if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
            return filepath.Join(xdgData, "muxuetools")
        }
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".local", "share", "muxuetools")
    default:
        return "data"
    }
}

// GetConfigDir 返回配置目录
func GetConfigDir() string {
    switch runtime.GOOS {
    case "windows":
        // Windows: %APPDATA%\MuxueTools
        appData := os.Getenv("APPDATA")
        if appData == "" {
            home, _ := os.UserHomeDir()
            appData = filepath.Join(home, "AppData", "Roaming")
        }
        return filepath.Join(appData, AppName)
    case "linux":
        // Linux: ~/.config/muxuetools
        if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
            return filepath.Join(xdgConfig, "muxuetools")
        }
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".config", "muxuetools")
    default:
        return "."
    }
}

// GetLogDir 返回日志目录
func GetLogDir() string {
    return filepath.Join(GetDataDir(), "logs")
}

// GetDatabasePath 返回数据库文件路径
func GetDatabasePath() string {
    return filepath.Join(GetDataDir(), "muxuetools.db")
}

// EnsureDirectories 确保所有必需目录存在
func EnsureDirectories() error {
    dirs := []string{
        GetConfigDir(),
        GetDataDir(),
        GetLogDir(),
    }
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return err
        }
    }
    return nil
}
```

**修改**: `internal/config/loader.go`

```diff
func (l *Loader) Load() (*types.Config, error) {
    l.setupDefaults()
    l.setupEnvBindings()

-   // Add default search paths
-   l.v.AddConfigPath(".")
-   l.v.AddConfigPath("./configs")
+   // Add default search paths (优先级从高到低)
+   l.v.AddConfigPath(GetConfigDir())  // 用户配置目录
+   l.v.AddConfigPath(".")             // 当前目录 (便携模式)
+   l.v.AddConfigPath("./configs")     // 开发模式

    // ... rest of the code
}
```

**修改**: `internal/types/config.go` (默认值)

```diff
func DefaultConfig() *Config {
    return &Config{
        // ...
        Database: DatabaseConfig{
-           Path: "data/MuxueTools.db",
+           Path: "", // 留空表示使用 GetDatabasePath()
        },
        // ...
    }
}
```

**新增文件**: `internal/config/migration.go`

```go
package config

import (
    "os"
    "path/filepath"

    "github.com/sirupsen/logrus"
)

// LegacyPaths 定义旧版本使用的相对路径
var legacyPaths = []string{
    "data/MuxueTools.db",
    "data/muxuetools.db",
}

// CheckLegacyData 检查是否存在旧版本数据，返回旧路径（如存在）
func CheckLegacyData() string {
    newDBPath := GetDatabasePath()
    
    // 如果新路径已存在数据，无需迁移
    if fileExists(newDBPath) {
        return ""
    }
    
    // 检查旧路径
    for _, oldPath := range legacyPaths {
        if fileExists(oldPath) {
            return oldPath
        }
    }
    
    return ""
}

// LogMigrationHint 记录迁移提示日志
func LogMigrationHint(logger *logrus.Logger, oldPath string) {
    newPath := GetDatabasePath()
    logger.Warn("========================================")
    logger.Warnf("发现旧版本数据: %s", oldPath)
    logger.Warnf("新数据路径: %s", newPath)
    logger.Warn("请手动迁移数据或继续使用旧路径")
    logger.Warn("========================================")
}

func fileExists(path string) bool {
    info, err := os.Stat(path)
    return err == nil && !info.IsDir()
}
```

---

## 3. 安装程序方案

### 3.1 Windows 安装程序 (Inno Setup)

**工具**: [Inno Setup](https://jrsoftware.org/isinfo.php) (免费, 轻量, 支持中文)

**特性**:
- ✅ 自定义安装路径
- ✅ 开始菜单快捷方式
- ✅ 桌面快捷方式 (可选)
- ✅ 卸载程序
- ✅ 中文界面

**新增文件**: `scripts/installer/windows/setup.iss`

```iss
; MuxueTools Windows Installer Script
; Inno Setup 6.x

#define MyAppName "MuxueTools"
#define MyAppVersion "0.2.0"
#define MyAppPublisher "muxueliunian"
#define MyAppURL "https://github.com/muxueliunian/MuxueTools"
#define MyAppExeName "MuxueTools.exe"

[Setup]
AppId={{GUID-TO-GENERATE}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
OutputDir=..\..\..\dist
OutputBaseFilename=MuxueTools-Setup-{#MyAppVersion}
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

; 权限设置 (不需要管理员权限)
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog

[Languages]
Name: "chinesesimplified"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
; 主程序
Source: "..\..\..\bin\MuxueTools.exe"; DestDir: "{app}"; Flags: ignoreversion

; 配置模板
Source: "..\..\..\configs\config.example.yaml"; DestDir: "{app}"; DestName: "config.example.yaml"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent
```

**构建命令**:

```powershell
# 需要先安装 Inno Setup 并添加到 PATH
iscc scripts/installer/windows/setup.iss
```

---

### 3.2 Linux 安装包 (.deb)

**工具**: `dpkg-deb` (Ubuntu 原生)

**新增目录结构**: `scripts/installer/linux/`

```
scripts/installer/linux/
├── build-deb.sh                 # 构建脚本
└── DEBIAN/
    ├── control                  # 包元信息
    ├── postinst                 # 安装后脚本
    └── prerm                    # 卸载前脚本
```

**文件**: `scripts/installer/linux/DEBIAN/control`

```
Package: muxuetools
Version: 0.2.0
Section: net
Priority: optional
Architecture: amd64
Depends: libgtk-3-0, libwebkit2gtk-4.1-0
Maintainer: muxueliunian <your-email@example.com>
Description: Gemini to OpenAI API Proxy
 MuxueTools is a reverse proxy that converts Google AI Studio (Gemini) API
 to OpenAI-compatible format, enabling seamless integration with existing
 OpenAI clients.
Homepage: https://github.com/muxueliunian/MuxueTools
```

**文件**: `scripts/installer/linux/DEBIAN/postinst`

```bash
#!/bin/bash
set -e

# 设置程序执行权限（数据目录由程序首次运行时自动创建）
chmod 755 /opt/muxuetools/muxuetools

echo "MuxueTools installed successfully!"
echo "Run: muxuetools"
echo "Data will be stored in: ~/.local/share/muxuetools/"
```

**文件**: `scripts/installer/linux/build-deb.sh`

```bash
#!/bin/bash
set -e

VERSION="${VERSION:-0.2.0}"
ARCH="amd64"
PKG_NAME="muxuetools_${VERSION}_${ARCH}"

# 创建包目录结构
BUILD_DIR="build/$PKG_NAME"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/opt/muxuetools"
mkdir -p "$BUILD_DIR/usr/local/bin"
mkdir -p "$BUILD_DIR/DEBIAN"

# 复制程序文件
cp ../../bin/muxuetools "$BUILD_DIR/opt/muxuetools/"
cp ../../configs/config.example.yaml "$BUILD_DIR/opt/muxuetools/"

# 创建符号链接脚本
echo '#!/bin/bash' > "$BUILD_DIR/usr/local/bin/muxuetools"
echo 'exec /opt/muxuetools/muxuetools "$@"' >> "$BUILD_DIR/usr/local/bin/muxuetools"
chmod +x "$BUILD_DIR/usr/local/bin/muxuetools"

# 复制 DEBIAN 控制文件
cp DEBIAN/* "$BUILD_DIR/DEBIAN/"
sed -i "s/Version: .*/Version: $VERSION/" "$BUILD_DIR/DEBIAN/control"
chmod 755 "$BUILD_DIR/DEBIAN/postinst" "$BUILD_DIR/DEBIAN/prerm" 2>/dev/null || true

# 构建 deb 包
dpkg-deb --build "$BUILD_DIR"
mv "build/$PKG_NAME.deb" "../../dist/"

echo "Created: dist/$PKG_NAME.deb"
```

**安装命令**:

```bash
sudo dpkg -i muxuetools_0.2.0_amd64.deb
```

---

## 4. CI/CD 集成

### 4.1 GitHub Actions 工作流

**文件**: `.github/workflows/release.yml` (需更新)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Build Frontend
        run: |
          cd web
          npm ci
          npm run build
      
      - name: Build Windows Executable
        run: |
          .\scripts\build.ps1 desktop
      
      - name: Install Inno Setup
        run: choco install innosetup -y
      
      - name: Build Windows Installer
        run: |
          iscc scripts/installer/windows/setup.iss
      
      - name: Upload Windows Installer
        uses: actions/upload-artifact@v4
        with:
          name: windows-installer
          path: dist/MuxueTools-Setup-*.exe

  build-linux:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Build Frontend
        run: |
          cd web
          npm ci
          npm run build
      
      - name: Install GTK/WebKit dependencies
        run: sudo apt-get update && sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
      
      - name: Build Linux Binary
        run: |
          chmod +x scripts/build.sh
          ./scripts/build.sh
      
      - name: Build .deb Package
        run: |
          cd scripts/installer/linux
          chmod +x build-deb.sh
          ./build-deb.sh
      
      - name: Upload Linux Package
        uses: actions/upload-artifact@v4
        with:
          name: linux-deb
          path: dist/*.deb

  release:
    needs: [build-windows, build-linux]
    runs-on: ubuntu-latest
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4
      
      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            windows-installer/*.exe
            linux-deb/*.deb
```

---

## 5. 实施阶段

### Phase 1: 数据路径重构 ✅ (已完成)

| 任务 | 文件 | 状态 |
|------|------|------|
| 1.1 | `internal/config/paths.go` | ✅ 已完成 |
| 1.2 | `internal/config/loader.go` | ✅ 已完成 |
| 1.3 | `internal/types/config.go` | ✅ 已完成 |
| 1.4 | `cmd/server/main.go` | ✅ 已完成 |
| 1.5 | `cmd/desktop/main.go` | ✅ 已完成 |
| 1.6 | `internal/config/migration.go` | ✅ 已完成 |

### Phase 2: Linux 构建系统 ✅ (已完成)

| 任务 | 文件 | 状态 |
|------|------|------|
| 2.1 | `scripts/build.sh` | ✅ 已完成 |
| 2.2 | `README.md` Linux 构建说明 | ✅ 已完成 |

### Phase 3: Windows 安装程序 ✅ (已完成)

| 任务 | 文件 | 状态 |
|------|------|------|
| 3.1 | `scripts/installer/windows/setup.iss` | ✅ 已完成 |
| 3.2 | `scripts/build.ps1` | ✅ 已完成 |

### Phase 4: Linux 安装包 ✅ (已完成)

| 任务 | 文件 | 状态 |
|------|------|------|
| 4.1 | `scripts/installer/linux/DEBIAN/control` | ✅ 已完成 |
| 4.2 | `scripts/installer/linux/DEBIAN/postinst` | ✅ 已完成 |
| 4.3 | `scripts/installer/linux/DEBIAN/prerm` | ✅ 已完成 |
| 4.4 | `scripts/installer/linux/build-deb.sh` | ✅ 已完成 |
| 4.5 | `scripts/installer/linux/muxuetools.desktop` | ✅ 已完成 |

### Phase 5: CI/CD 集成 ✅ (已完成)

| 任务 | 文件 | 状态 |
|------|------|------|
| 5.1 | `.github/workflows/release.yml` | ✅ 已完成 - Windows 安装程序构建 |
| 5.2 | `.github/workflows/release.yml` | ✅ 已完成 - Linux 构建 Job |
| 5.3 | `.github/workflows/release.yml` | ✅ 已完成 - .deb 包构建 |
| 5.4 | `.github/workflows/release.yml` | ✅ 已完成 - 独立 Release Job |
| 5.5 | `.github/workflows/release.yml` | ✅ 已完成 - 更新 latest.json

### Phase 6: 文档更新 ✅ (已完成)

| 任务 | 文件 | 状态 |
|------|------|------|
| 6.1 | 更新 `README.md` | ✅ 已完成 - 添加安装说明 |
| 6.2 | 更新 `docs/ARCHITECTURE.md` | ✅ 已完成 - 添加安装程序架构说明 |


---

## 6. 文件清单

### 新增文件

| 文件 | 描述 |
|------|------|
| `internal/config/paths.go` | 跨平台路径工具函数 |
| `internal/config/migration.go` | 旧数据迁移检测与提示 |
| `scripts/build.sh` | Linux 构建脚本 |
| `scripts/installer/windows/setup.iss` | Windows Inno Setup 脚本 |
| `scripts/installer/linux/DEBIAN/control` | Debian 包元信息 (含 GTK 依赖声明) |
| `scripts/installer/linux/DEBIAN/postinst` | 安装后脚本 |
| `scripts/installer/linux/DEBIAN/prerm` | 卸载前脚本 |
| `scripts/installer/linux/build-deb.sh` | .deb 构建脚本 |

### 修改文件

| 文件 | 修改内容 |
|------|----------|
| `internal/config/loader.go` | 使用新路径函数 |
| `internal/types/config.go` | 更新默认配置 |
| `cmd/server/main.go` | 调用 `EnsureDirectories()` |
| `cmd/desktop/main.go` | 调用 `EnsureDirectories()` |
| `.goreleaser.yaml` | 添加 Linux 构建目标 |
| `.github/workflows/release.yml` | 添加安装程序构建 |
| `README.md` | 添加安装说明 |

---

## 7. 验证计划

### 7.1 Windows 安装程序验证

```powershell
# 构建安装程序
iscc scripts/installer/windows/setup.iss

# 运行安装程序
.\dist\MuxueTools-Setup-0.2.0.exe

# 验证安装
# 1. 检查 C:\Program Files\MuxueTools\ 目录
# 2. 检查开始菜单快捷方式
# 3. 运行程序，验证数据写入 %APPDATA%\MuxueTools\
```

### 7.2 Linux .deb 包验证

```bash
# 构建 deb 包
cd scripts/installer/linux
./build-deb.sh

# 安装
sudo dpkg -i ../../dist/muxuetools_0.2.0_amd64.deb

# 验证
muxuetools --version
ls ~/.local/share/muxuetools/

# 卸载
sudo dpkg -r muxuetools
```

### 7.3 路径迁移验证

```bash
# 验证旧数据不受影响
# 验证新安装使用正确路径
# 验证便携模式仍可用（当前目录有 config.yaml）
```

---

## 8. 时间估算

| 阶段 | 预计工时 |
|------|----------|
| Phase 1: 数据路径重构 | 2 小时 |
| Phase 2: Linux 构建系统 | 1 小时 |
| Phase 3: Windows 安装程序 | 2 小时 |
| Phase 4: Linux 安装包 | 2 小时 |
| Phase 5: CI/CD 集成 | 2 小时 |
| Phase 6: 文档更新 | 1 小时 |
| 测试与调试 | 2 小时 |
| **总计** | **12 小时** |

---

## 9. 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 旧版本数据迁移 | 用户升级后旧数据丢失 | 首次运行检测旧路径并提示迁移 (Phase 1.6) |
| Inno Setup 版本兼容 | 构建失败 | 锁定 Inno Setup 6.x 版本 |
| GTK/WebKit 依赖 | Linux 桌面版需要 GTK3 | CI 安装 `libgtk-3-dev libwebkit2gtk-4.1-dev`；.deb 声明运行时依赖 |
| 权限问题 | 写入系统目录失败 | 使用 `PrivilegesRequired=lowest` 选项 |

> [!NOTE]
> **关于 SQLite**: 项目使用 `glebarez/sqlite` 纯 Go 驱动，**无需 CGO**，可直接交叉编译，无需安装 gcc。

---

## 10. 后续扩展 (TODO)

- [ ] macOS 支持 (`.dmg` 安装包)
- [ ] 自动更新机制（检测新版本后下载并提示安装）
- [ ] 数据迁移工具（从旧路径迁移到新路径）
- [ ] 便携模式开关（命令行参数 `--portable`）

---

## 11. 参考资料

- [Inno Setup 官方文档](https://jrsoftware.org/ishelp/)
- [Debian 打包指南](https://www.debian.org/doc/manuals/maint-guide/)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
- [GitHub Actions 官方文档](https://docs.github.com/en/actions)

---

*最后更新: 2026-01-26*

---

## 12. Ubuntu 测试计划

详见 [UBUNTU_TEST_CHECKLIST.md](./UBUNTU_TEST_CHECKLIST.md)

### 测试内容概要

| Phase | 测试项 | 说明 |
|-------|--------|------|
| Phase 2 | 构建系统 | `./scripts/build.sh all` 编译 server/desktop |
| Phase 4 | .deb 安装包 | 构建、安装、验证、卸载完整流程 |

### 前置条件

```bash
# 安装编译依赖
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev

# 安装 Go 1.22+ 和 Node.js 18+
```

### 验收标准

- [ ] `build.sh all` 无错误执行
- [ ] 生成 `bin/muxuetools-server` 和 `bin/muxuetools-desktop`
- [ ] `build-deb.sh` 成功生成 `dist/muxuetools_*.deb`
- [ ] `dpkg -i` 安装成功，程序可正常启动
- [ ] 应用程序菜单中可见 MuxueTools 图标
- [ ] `dpkg -r` 卸载成功，用户数据保留
