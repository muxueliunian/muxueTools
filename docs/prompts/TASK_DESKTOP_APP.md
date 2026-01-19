# 任务：实现独立桌面窗口 (WebView Wrapper)

## 角色
Go System Engineer & Desktop Developer

## 必备技能
- **Senior Golang**: 必须阅读并应用 `.agent/skills/senior-golang/SKILL.md`，确保并发处理（Goroutine）和构建标记的正确性。
- **Architect**: 参考 `.agent/skills/architect/SKILL.md`，保持 `cmd/` 入口代码的简洁性，不要将业务逻辑泄漏到入口文件。

## 背景
MuxueTools 目前运行在浏览器中。用户希望脱离默认浏览器，运行在独立的桌面窗口中，体验更像原生应用。
我们将使用 `github.com/webview/webview_go` 来实现这一目标，它在 Windows 上使用 Edge WebView2，轻量且无需捆绑浏览器内核。

## 任务目标
创建一个新的入口 `cmd/desktop/main.go`，启动后端服务器并打开一个独立的 WebView 窗口加载前端页面。

## 详细步骤

### 1. 依赖管理
- 添加依赖：`go get github.com/webview/webview_go`
- **注意**：此库需要启用 CGO (`CGO_ENABLED=1`)。Windows 上需要安装 MinGW-w64 (gcc)。

### 2. 创建桌面入口 (`cmd/desktop/main.go`)
- **初始化服务器**：
  - 引用 `internal/api/server`。
  - 在单独的 Goroutine 中启动 HTTP 服务器。
  - 建议：服务器端口设为 `0` (随机端口) 或读取配置，并在启动后获取实际监听端口。
- **初始化 WebView**：
  - `w := webview.New(debug)`
  - `w.SetTitle("MuxueTools")`
  - `w.SetSize(1024, 768, webview.HintNone)`
  - `w.Navigate("http://localhost:<actual_port>")`
  - `w.Run()` (阻塞运行)
- **优雅退出**：
  - 当 WebView 窗口关闭时，确保 HTTP Server 也能优雅停止（Context Cancel）。

### 3. 构建脚本更新
- 更新 `scripts/build.ps1` 和 `scripts/build.sh`。
- 添加 `desktop` 构建目标。
- Windows 构建命令示例：
  ```powershell
  # Windows x64
  go build -ldflags "-H windowsgui" -o bin/MuxueTools-win-amd64.exe ./cmd/desktop
  
  # Windows x86 (需 32位 gcc)
  $env:GOARCH="386"
  go build -ldflags "-H windowsgui" -o bin/MuxueTools-win-386.exe ./cmd/desktop
  ```
  *(注：`-H windowsgui` 用于隐藏控制台窗口)*

### 4. 验证
- 确保存储层 (`sqlite`) 在桌面模式下能正确找到数据库文件路径（建议使用 `os.UserConfigDir` 或相对路径）。
- 验证窗口关闭后进程是否完全退出。

---

## 产出物
- `cmd/desktop/main.go`
- 更新后的构建脚本
- 可运行的 `MuxueTools-win-amd64.exe` (无控制台，独立窗口)

## 约束
- 保持现有的 `cmd/server/main.go` 不变，保留纯 Server 模式供 Linux 服务器使用。
- 确保 WebView 能够正确处理前端的 LocalStorage/Cookies (WebView 默认支持)。
