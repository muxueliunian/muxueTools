# MuxueTools 实施计划

> AI Studio (Gemini) → OpenAI 格式反向代理工具
> 
> 技术栈：Go 后端 + Vue3 前端 | 目标平台：Windows / Linux / (Android)

---

## 📋 项目概述

### 目标
创建一个本地运行的工具，将 Google AI Studio (Gemini) API 反向代理成 OpenAI 兼容格式，支持：
- 多 API Key 轮询/负载均衡
- 流式响应 (SSE)
- 图片输入 (Vision)
- Token 消耗统计（如 API 支持）
- GitHub 自动更新检测与分发
- Web UI 管理界面

### 使用场景
- 个人/小型社区使用（5-10人）
- 每人 5-20 个 API Key
- 本地部署，数据不共享
- 通过 GitHub Release 分发更新

### 跨平台支持

| 平台 | 支持状态 | 分发格式 |
|------|---------|---------|
| 平台 | 支持状态 | 分发格式 |
|------|---------|---------|
| Windows x64 | ✅ 主要支持 | `MuxueTools-windows-amd64.exe` (WebView) |
| Windows x86 | ✅ 新增支持 | `MuxueTools-windows-386.exe` (WebView) |
| Linux x64 | ✅ 完全支持 | `MuxueTools-linux-amd64` |
| Linux ARM64 | ✅ 完全支持 | `MuxueTools-linux-arm64` |
| macOS x64 | ✅ 完全支持 | `MuxueTools-darwin-amd64` (WebView) |
| macOS ARM64 | ✅ 完全支持 | `MuxueTools-darwin-arm64` (WebView) |
| Android 15+ | ✅ 核心支持 | `MuxueTools.apk` (Native WebView) |

---

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                          MuxueTools                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │   WebView   │────│   Go API    │────│   Gemini API Pool   │  │
│  │  (独立窗口) │    │   Server    │    │   (Key 轮询管理)    │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│         ▲                 │                      │              │
│         │ (Embed)         ▼                      ▼              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │   Vue3 UI   │    │ OpenAI 兼容 │    │   - gemini-pro      │  │
│  │   (资源)    │    │ /v1/chat/.. │    │   - gemini-flash    │  │
│  │             │    │ /v1/models  │    │   - gemini-vision   │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 独立窗口模式 (Standalone Window)
不依赖用户默认浏览器，使用轻量级 WebView 包装器（如 `webview_go` 或 `wails`）创建独立应用窗口。
- **Windows**: 使用 Edge WebView2 运行时。
- **macOS/Linux**: 使用系统 WebKit。
- **Android**: 使用原生 System WebView。

---

## 📁 项目结构

```
MuxueTools/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── api/
│   │   ├── router.go            # 路由配置
│   │   ├── openai_handler.go    # OpenAI 兼容端点处理
│   │   ├── admin_handler.go     # 管理 API
│   │   └── middleware.go        # 中间件（CORS、日志等）
│   ├── gemini/
│   │   ├── client.go            # Gemini API 客户端
│   │   ├── models.go            # Gemini 请求/响应模型
│   │   ├── converter.go         # OpenAI ↔ Gemini 格式转换
│   │   └── token_counter.go     # Token 消耗统计
│   ├── keypool/
│   │   ├── pool.go              # Key 池管理器
│   │   ├── strategy.go          # 轮询策略（轮询/随机/智能）
│   │   └── stats.go             # 使用统计
│   ├── config/
│   │   ├── config.go            # 配置结构
│   │   └── loader.go            # 配置加载/保存
│   ├── storage/
│   │   └── sqlite.go            # SQLite 本地存储（统计数据）
│   └── updater/
│       └── github.go            # GitHub 更新检测
├── web/                         # Vue3 前端
│   ├── src/
│   │   ├── views/
│   │   │   ├── Dashboard.vue    # 仪表盘
│   │   │   ├── KeyManager.vue   # Key 管理
│   │   │   ├── Stats.vue        # 统计页面
│   │   │   └── Settings.vue     # 设置页面
│   │   ├── components/
│   │   ├── stores/              # Pinia 状态管理
│   │   └── App.vue
│   ├── package.json
│   └── vite.config.ts
├── embed.go                     # 嵌入前端静态文件
├── configs/
│   └── config.example.yaml      # 配置文件示例
├── scripts/
│   ├── build.ps1                # Windows 构建脚本
│   └── build.sh                 # Linux/macOS 构建脚本
├── .goreleaser.yaml             # GoReleaser 配置（多平台打包）
├── go.mod
├── go.sum
└── README.md
```

---

## 🔌 API 端点设计

### OpenAI 兼容端点（代理功能）

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/chat/completions` | POST | 对话补全（核心功能） |
| `/v1/models` | GET | 获取可用模型列表 |
| `/health` | GET | 健康检查 |

### 管理端点（Web UI 使用）

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/keys` | GET/POST/DELETE | Key 管理 CRUD |
| `/api/keys/:id/test` | POST | 测试单个 Key 可用性 |
| `/api/keys/import` | POST | 批量导入 Key |
| `/api/keys/export` | GET | 导出 Key 列表 |
| `/api/stats` | GET | 获取使用统计（请求数、Token 消耗） |
| `/api/stats/keys` | GET | 获取各 Key 使用情况 |
| `/api/config` | GET/PUT | 配置管理 |
| `/api/update/check` | GET | 检查 GitHub 更新 |
| `/api/update/info` | GET | 获取最新版本信息 |

---

## 🔄 开发阶段

### 阶段一：核心后端 ✅ (完成于 2026-01-15)

#### 1.1 项目初始化
- [x] 初始化 Go module (`MuxueTools`)
- [x] 安装依赖（gin, viper, logrus 等）
- [x] 创建基础目录结构
- [x] 配置文件结构定义

#### 1.2 Key 池管理器
- [x] Key 存储结构（支持分组/标签）
- [x] 轮询策略实现
  - Round Robin（轮询）
  - Random（随机）
  - Least Used（最少使用）
  - Weighted（加权，根据成功率）
- [x] Key 状态管理（可用/冷却中/禁用）
- [x] Rate Limit 检测与自动冷却
- [x] 使用次数统计

#### 1.3 Gemini API 客户端
- [x] HTTP 客户端封装
- [x] 请求签名/认证
- [x] 普通请求处理
- [x] 流式响应处理（SSE）
- [x] 图片输入处理（base64/URL）
- [x] Token 消耗解析（从响应 header/body 提取）
- [x] 错误处理与重试

#### 1.4 格式转换器
- [x] OpenAI Request → Gemini Request
- [x] Gemini Response → OpenAI Response
- [x] 流式响应格式转换
- [x] 模型名称映射（gpt-4 → gemini-pro 等）
- [x] Token 统计字段转换

#### 1.5 OpenAI 兼容端点
- [x] `/v1/chat/completions` 实现
- [x] 流式响应支持
- [x] `/v1/models` 实现
- [x] 错误响应格式化

### 阶段二：管理功能 ✅ (完成于 2026-01-15)

#### 2.1 配置与存储
- [x] YAML 配置文件读写
- [x] SQLite 本地存储（统计数据持久化）
- [ ] 运行时配置热更新 - P2 待优化
- [x] 配置验证

#### 2.2 管理 API
- [x] Key CRUD 接口
- [x] Key 批量导入/导出
- [x] 使用统计接口（包含 Token 消耗）
- [x] 配置管理接口

#### 2.3 更新检测
- [x] GitHub Release API 调用
- [x] 版本比较逻辑
- [x] 更新提示（返回下载链接）

### 阶段三：API 层 ✅ (完成于 2026-01-15)

#### 3.1 路由与服务器
- [x] Gin 路由配置 (`api/router.go`)
- [x] 服务器初始化 (`api/server.go`)
- [x] 主程序入口 (`cmd/server/main.go`)

#### 3.2 处理器开发
- [x] OpenAI 兼容端点 (`api/openai_handler.go`)
  - `/v1/chat/completions` (阻塞 + SSE 流式)
  - `/v1/models`
- [x] 管理端点 (`api/admin_handler.go`)
  - Key CRUD、导入导出
  - 统计接口
  - 配置管理
  - 更新检测
- [x] 健康检查 (`api/health_handler.go`)

#### 3.3 中间件
- [x] CORS 中间件
- [x] 请求日志中间件
- [x] Panic 恢复中间件
- [x] 请求 ID 中间件

**测试结果**: 29/29 通过 ✅

### 阶段 3.5：存储层与会话管理 ✅ (完成于 2026-01-15)

#### 3.5.1 SQLite 存储层
- [x] 存储器初始化 (`storage/sqlite.go`)
- [x] Key 持久化 CRUD (`storage/keys.go`)
- [x] 数据库迁移

#### 3.5.2 会话管理
- [x] 会话类型定义 (`types/session.go`)
- [x] 会话存储 CRUD (`storage/sessions.go`)
- [x] 消息存储 CRUD
- [x] 会话 API (`api/session_handler.go`)
  - GET/POST/PUT/DELETE `/api/sessions`
  - POST `/api/sessions/:id/messages`

**测试结果**: 17/17 通过 ✅

### 阶段四：前端开发 ✅ (完成于 2026-01-18)

#### 4.1 项目搭建
- [x] Vite + Vue3 + TypeScript 初始化
- [x] Pinia 状态管理
- [x] Vue Router
- [x] UI 框架（Naive UI）

#### 4.2 页面开发
- [x] 仪表盘 (`Dashboard.vue`)
  - 概览统计、Token 消耗
- [x] Key 管理 (`KeyManager.vue`)
  - Key 列表（状态、使用次数、Token 消耗）
  - 添加/编辑/删除 Key
  - 批量导入/导出
  - 测试 Key 可用性
- [x] 使用统计 (`Stats.vue`)
  - 请求次数图表
  - Token 消耗图表
  - 各 Key 使用分布
- [x] 设置 (`Settings.vue`)
  - 轮询策略选择
  - 更新检测与下载链接

#### 4.3 美化与交互
- [x] 暗色主题（默认）
- [x] 响应式布局
- [x] 加载动画与 Toast 反馈

#### 4.4 前端嵌入
- [x] 前端构建输出到 Go embed
- [x] 静态文件服务配置

### 阶段五：打包发布 ✅ (完成于 2026-01-18)

#### 5.1 多平台打包
- [x] GoReleaser 配置 (`.goreleaser.yaml`)
- [x] Windows exe 图标设置
- [x] 版本信息嵌入
- [x] 多平台交叉编译
  - windows/amd64
  - linux/amd64, linux/arm64
  - darwin/amd64, darwin/arm64

#### 5.2 GitHub Release
- [x] GitHub Actions CI/CD 配置
- [x] 自动发布 Action
- [x] Changelog 自动生成

#### 5.3 文档
- [x] README 使用说明
- [x] 配置文件说明
- [x] API 使用文档

---

### 阶段六：Android APK (Native Support)

#### 6.1 Android 项目架构
使用 `gomobile` 将 Go 后端编译为 Android 库 (`.aar`)，在 Android Studio 中构建原生外壳。

#### 6.2 适配工作
- [ ] Go: 适配 Android 文件系统路径 (Config/DB)
- [ ] Go: 封装 `StartServer(bindAddr string)` 供 Java/Kotlin 调用
- [ ] Android: 实现前台服务 (Foreground Service) 保持服务运行
- [ ] Android: 集成 WebView 加载 `http://localhost:{port}`
- [ ] Android 15+: 适配 Edge-to-Edge 全屏显示

#### 6.3 打包与测试
- [ ] 编译 x86/arm64 架构库
- [ ] 生成 release 签名 APK
- [ ] 真机测试 (Android 15+)

---

## 📦 技术依赖

### Go 后端
```go
// go.mod
module muxueTools

go 1.22

require (
    github.com/gin-gonic/gin v1.9+        // HTTP 框架
    github.com/gin-contrib/cors v1.5+     // CORS 支持
    github.com/spf13/viper v1.18+         // 配置管理
    github.com/sirupsen/logrus v1.9+      // 日志
    gopkg.in/yaml.v3 v3.0+                // YAML 解析
    github.com/google/uuid v1.5+          // UUID 生成
    github.com/mattn/go-sqlite3 v1.14+    // SQLite 驱动
    gorm.io/gorm v1.25+                   // ORM
    gorm.io/driver/sqlite v1.5+           // GORM SQLite
)
```

### Vue3 前端
```json
{
  "dependencies": {
    "vue": "^3.4",
    "vue-router": "^4.2",
    "pinia": "^2.1",
    "naive-ui": "^2.38",
    "@vicons/ionicons5": "^0.12",
    "axios": "^1.6",
    "echarts": "^5.5",
    "vue-echarts": "^6.6"
  },
  "devDependencies": {
    "vite": "^5.0",
    "typescript": "^5.3",
    "@vitejs/plugin-vue": "^5.0"
  }
}
```

---

## 🎯 模型映射配置

```yaml
# 模型名称映射
model_mappings:
  # OpenAI 模型名 → Gemini 模型名
  "gpt-4": "gemini-1.5-pro-latest"
  "gpt-4-turbo": "gemini-1.5-pro-latest"
  "gpt-4-vision-preview": "gemini-1.5-pro-latest"
  "gpt-4o": "gemini-1.5-flash-latest"
  "gpt-4o-mini": "gemini-1.5-flash-8b-latest"
  "gpt-3.5-turbo": "gemini-1.5-flash-latest"
  
  # 直接使用 Gemini 模型名也支持
  "gemini-pro": "gemini-1.5-pro-latest"
  "gemini-flash": "gemini-1.5-flash-latest"
  "gemini-2.0-flash": "gemini-2.0-flash"
  "gemini-2.5-pro": "gemini-2.5-pro-preview"
```

---

## 📊 配置文件示例

```yaml
# config.yaml
server:
  port: 8080
  host: "0.0.0.0"

# API Keys 配置
keys:
  - key: "AIzaSy..."
    name: "Key 1"
    enabled: true
    tags: ["personal"]
  - key: "AIzaSy..."
    name: "Key 2"
    enabled: true
    tags: ["backup"]

# Key 池策略
pool:
  strategy: "round_robin"  # round_robin | random | least_used | weighted
  cooldown_seconds: 60     # 触发限流后冷却时间
  max_retries: 3           # 最大重试次数

# 日志配置
logging:
  level: "info"
  file: "logs/MuxueTools.log"

# 更新检测
update:
  enabled: true
  check_interval: "24h"
  github_repo: "muxueliunian/MuxueTools"
```

---

## 📈 Token 统计说明

Google AI Studio / Gemini API 在响应中会返回 `usageMetadata`，包含：

```json
{
  "usageMetadata": {
    "promptTokenCount": 10,
    "candidatesTokenCount": 50,
    "totalTokenCount": 60
  }
}
```

MuxueTools 会：
1. 解析这些字段
2. 转换为 OpenAI 格式的 `usage` 字段
3. 存储到本地 SQLite 用于统计展示

---

## 🚀 快速开始（用户视角）

### Windows
```powershell
# 1. 下载 MuxueTools-windows-amd64.exe
# 2. 双击运行（首次会生成 config.yaml）
# 3. 编辑 config.yaml 添加 API Key
# 4. 重新运行
# 5. 访问 http://localhost:8080 管理界面
# 6. 配置客户端使用 http://localhost:8080/v1 作为 API 端点
```

### Linux
```bash
# 1. 下载对应架构的二进制文件
wget https://github.com/muxueliunian/MuxueTools/releases/latest/download/MuxueTools-linux-amd64
chmod +x MuxueTools-linux-amd64

# 2. 运行
./MuxueTools-linux-amd64

# 3. 访问 http://localhost:8080
```

### Android (WebView APK)
1. 下载并安装 `MuxueTools.apk`
2. 打开 App，等待服务器启动（约1-2秒）
3. 界面自动显示（WebView）
4. 其他 App（如 Chatbox）可连接 `http://localhost:8080/v1`


---

## ⏰ 时间估算

| 阶段 | 预估时间 | 实际进度 |
|------|---------|----------|
| 阶段一：核心后端 | 3-4 天 | ✅ 完成 |
| 阶段二：管理功能 | 2 天 | ✅ 完成 |
| 阶段三：API 层 | 1-2 天 | ✅ 完成 |
| 阶段四：前端开发 | 2-3 天 | ✅ 完成 |
| 阶段五：打包发布 | 1-2 天 | ✅ 完成 |
| 阶段六：Android APK | 2-3 天 | ⏳ 可选 |
| **总计** | **10-14 天** | **已完成** |

---

## ✅ 确认事项

- [x] 项目名称：**MuxueTools**
- [x] GitHub 仓库：`muxueliunian/MuxueTools`（待创建）
- [x] 无需用户认证，本地部署
- [x] Token 统计：使用 Gemini API 返回的 usageMetadata
- [x] 图标：用户自行绘制
- [x] 跨平台：Windows / Linux / macOS / Android (WebView APK)

---

## 📚 参考项目

1. [songquanpeng/one-api](https://github.com/songquanpeng/one-api) - 多模型聚合方案
2. [zhu327/gemini-openai-proxy](https://github.com/zhu327/gemini-openai-proxy) - Gemini 代理
3. [PublicAffairs/openai-gemini](https://github.com/PublicAffairs/openai-gemini) - Serverless 方案

---

*最后更新: 2026-01-20*
