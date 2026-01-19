# BUG-001: Windows 窗口图标显示为默认图标

## 问题概述

| 属性 | 值 |
|------|-----|
| **状态** | ✅ 已解决 |
| **严重程度** | Medium |
| **发现日期** | 2026-01-15 |
| **解决日期** | 2026-01-18 |
| **影响版本** | Desktop v1.0 |

## 问题描述

MxlnAPI 桌面应用在 Windows 系统上运行时，窗口左上角、任务栏和 Alt+Tab 切换器中显示的是 Windows 默认图标，而非自定义的企鹅图标。

## 根本原因

使用 `rsrc` 工具生成 `.syso` 资源文件时，生成的图标资源 ID 不符合 Windows Shell 的要求：

- **预期**: Windows Shell 查找资源 ID `32512` (`IDI_APPLICATION`) 作为主窗口图标
- **实际**: `rsrc` 工具使用任意 ID，导致 Windows Shell 无法识别

## 解决方案

### 1. 安装 MinGW (MSYS2)

```bash
# 在 MSYS2 UCRT64 终端中运行
pacman -Syu
pacman -S mingw-w64-x86_64-binutils
```

### 2. 创建资源脚本 `cmd/desktop/app.rc`

```rc
// Windows Resource Script for MxlnAPI Desktop
// Defines application icon with standard Windows resource IDs

// IDI_APPLICATION (32512) - Standard Windows application icon ID
// Used by Windows Shell for window icon and taskbar
32512 ICON "../../assets/icon.ico"

// ID 1 - Alternative icon ID (backup for some contexts)
1 ICON "../../assets/icon.ico"
```

**关键点**: 必须使用资源 ID `32512`，这是 Windows 定义的 `IDI_APPLICATION` 常量。

### 3. 使用 `windres` 编译资源

`build.ps1` 已修改为优先使用 `windres`：

```powershell
& $windresPath -i app.rc -o $SysoPath
```

## 受影响的文件

| 文件 | 操作 |
|------|------|
| `cmd/desktop/app.rc` | **新建** - Windows 资源脚本 |
| `scripts/build.ps1` | **修改** - 添加 windres 支持 |

## 验证步骤

1. 运行 `.\scripts\build.ps1 desktop`
2. 启动 `.\bin\mxlnapi.exe`
3. 检查以下位置图标是否正确：
   - [x] 窗口左上角
   - [x] Windows 任务栏
   - [x] Alt+Tab 切换器

## 技术参考

- Windows 资源 ID 定义: `IDI_APPLICATION = 32512`
- MinGW windres 文档: https://sourceware.org/binutils/docs/binutils/windres.html
- Go rsrc 工具限制: 不支持自定义资源 ID

## 教训总结

1. Windows 图标资源必须使用标准 ID (`32512`) 才能被 Shell 正确识别
2. `rsrc` 工具简单但功能有限，复杂场景需使用 `windres`
3. 构建脚本应保留回退方案以兼容不同开发环境
