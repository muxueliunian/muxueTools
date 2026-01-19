# 桌面应用 (WebView Wrapper) 开发报告

**任务**: 实现独立桌面窗口  
**执行时间**: 2026-01-15 22:39 - 23:22  
**状态**: ✅ 完成

---

## 概述

成功使用 `github.com/webview/webview_go` 将 MxlnAPI 封装为独立桌面应用。应用在 Windows 上使用 Edge WebView2 运行，无需捆绑浏览器内核，体积轻量，体验接近原生应用。

---

## 实现的功能

### 核心功能

| 功能 | 实现方式 |
|------|----------|
| **WebView 桌面窗口** | 使用 `webview/webview_go` 库，调用系统 WebView2 |
| **无控制台窗口** | 使用 `-ldflags "-H windowsgui"` 编译 |
| **开发模式** | `-dev` 标志连接 Vite 开发服务器 |
| **生产模式** | 自动检测并提供 `web/dist` 静态文件 |
| **随机端口** | 避免端口冲突，自动分配可用端口 |
| **优雅退出** | 窗口关闭时正确停止 HTTP 服务器和释放资源 |

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-config` | 配置文件路径 | 自动检测 |
| `-dev` | 开发模式，连接 Vite | false |
| `-dev-url` | 自定义开发服务器 URL | `http://localhost:5173` |
| `-webroot` | 前端静态文件目录 | 自动检测 `web/dist` |
| `-debug` | 启用 WebView DevTools | false |
| `-version` | 显示版本信息 | - |

---

## 新增/修改的文件

### 新增文件

#### 1. `cmd/desktop/main.go`
桌面应用入口点，约 240 行代码。

主要逻辑：
- 解析命令行参数
- 初始化配置和日志
- 创建 HTTP 服务器（随机端口）
- 自动检测 WebRoot 路径
- 创建 WebView 窗口 (1024×768)
- 根据模式导航到开发服务器或本地服务器
- 窗口关闭后优雅停止服务

#### 2. `scripts/build.ps1`
PowerShell 构建脚本，约 110 行代码。

支持的构建目标：
- `server`: 纯 Go 服务器 (CGO_ENABLED=0)
- `desktop`: 桌面应用 x64 (CGO_ENABLED=1)
- `desktop-x86`: 桌面应用 x86
- `all`: 构建全部
- `clean`: 清理构建产物

功能特性：
- 自动创建 `bin/` 目录
- 注入版本信息 (Version, BuildTime, GitCommit)
- 桌面构建使用 `-H windowsgui` 隐藏控制台

### 修改的文件

#### 1. `internal/api/server.go`
- 添加 `webRoot` 字段到 Server 结构体
- 添加 `WithWebRoot(path string)` Option 函数
- 在 RouterConfig 中传递 WebRoot

#### 2. `internal/api/router.go`
- 添加 `WebRoot` 字段到 RouterConfig
- 实现静态文件服务逻辑：
  - 当 WebRoot 配置时，提供 `/assets` 静态资源
  - 实现 SPA HTML5 History 模式回退 (NoRoute 处理)
  - 区分 API 路由和前端路由

#### 3. `go.mod`
- 添加依赖：`github.com/webview/webview_go`

---

## 技术要点

### CGO 依赖

`webview/webview_go` 需要 CGO 支持：
- Windows: 需要安装 MinGW-w64 (gcc)
- 编译时设置 `CGO_ENABLED=1`

### WebView2 Runtime

用户机器需要 Edge WebView2 Runtime：
- Windows 11: 默认已安装
- Windows 10: 可能需要手动安装

### 路径检测逻辑

生产模式下自动检测 WebRoot：
1. `web/dist` (相对于工作目录)
2. `../web/dist`
3. `<exe目录>/web/dist`
4. `<exe目录>/../web/dist`

---

## 验证结果

| 测试项 | 结果 |
|--------|------|
| 开发模式 (`-dev`) | ✅ 正确显示前端界面 |
| 生产模式 (静态文件) | ✅ 正确显示前端界面 |
| 窗口标题 | ✅ "MxlnAPI" |
| 窗口尺寸 | ✅ 1024×768 |
| 无控制台窗口 | ✅ 隐藏 |
| 数据库持久化 | ✅ data/ 目录正确创建 |
| 进程退出 | ✅ 窗口关闭后完全退出 |

---

## 使用说明

### 开发流程

```powershell
# 1. 启动前端开发服务器
cd web
npm run dev

# 2. 另开终端，以开发模式启动桌面应用
cd ..
.\bin\mxlnapi.exe -dev -debug
```

### 生产构建

```powershell
# 1. 构建前端
cd web
npm run build

# 2. 构建桌面应用
cd ..
.\scripts\build.ps1 desktop

# 3. 运行
.\bin\mxlnapi.exe
```

### 分发

分发时需要包含：
- `mxlnapi.exe`
- `web/dist/` 目录（或将其嵌入 exe 中，需额外开发）

---

## 后续优化建议

1. **嵌入前端资源**: 使用 Go `embed` 包将前端文件嵌入二进制，实现单文件分发
2. **自动更新**: 实现应用内自动更新功能
3. **系统托盘**: 添加最小化到系统托盘功能
4. **窗口记忆**: 保存/恢复窗口位置和大小
5. **多窗口支持**: 支持打开多个窗口

---

## 依赖清单

| 依赖 | 版本 | 用途 |
|------|------|------|
| `webview/webview_go` | v0.0.0-20240831 | WebView 窗口 |
| MinGW-w64 (gcc) | 任意版本 | CGO 编译 |
| Edge WebView2 | 系统自带 | 渲染引擎 |

---

**报告完成时间**: 2026-01-15 23:22  
**执行者**: Worker Agent (Antigravity)
