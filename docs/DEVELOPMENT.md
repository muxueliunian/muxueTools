# MxlnAPI 开发工作流指南

> 本文档规范了 MxlnAPI 项目的开发流程、Agent 协作方式和代码提交规范。

---

## 📋 开发方法论

本项目采用 **Contract-First + TDD 混合模式**：

| 层级 | 方法 | 说明 |
|------|------|------|
| **接口层** | Contract-First | 先定义类型和接口，再实现 |
| **逻辑层** | TDD | 先写测试，再写实现 |
| **集成层** | 端到端测试 | 验证完整流程 |

---

## 🔄 开发阶段与流程

```
阶段 0        阶段 1         阶段 2         阶段 3         阶段 4
 准备    →    架构设计   →   核心实现   →   API 实现   →   前端开发
  │            │              │              │              │
  ▼            ▼              ▼              ▼              ▼
PLAN.md    ARCH.md        types/         api/           web/
Skills     Types          keypool/       handlers       Vue3
           Errors         gemini/        middleware
```

### 阶段 0：准备 ✅
- [x] 创建 `IMPLEMENTATION_PLAN.md`
- [x] 配置 Skills（architect, senior-golang, qa-automation）
- [x] 创建 `DEVELOPMENT.md`

### 阶段 1：架构设计 ✅
- [x] 创建 `ARCHITECTURE.md`
- [x] 定义核心类型 (`internal/types/`)
- [x] 定义错误码 (`internal/types/errors.go`)

### 阶段 2：核心逻辑实现 (TDD)
按以下顺序开发，每个模块遵循 TDD 流程：

| 优先级 | 模块 | 测试 | 实现 | 审核 |
|--------|------|------|------|------|
| P0 | 配置加载 | 17/17 ✅ | ✅ | - |
| P0 | Key 池 | 28/28 ✅ | ✅ | ✅ 通过 |
| P0 | 格式转换 | 88.6% ✅ | ✅ | ✅ 通过 |
| P0 | Gemini 客户端 | ✅ | ✅ | ✅ 通过 |
| P1 | 存储层 | 17/17 ✅ | ✅ | - |
| P1 | Token 统计 | - | - | - |

### 阶段 3：API 层实现 ✅ 审核通过 (2026-01-15)
| 模块 | 文件 | 状态 | 测试 |
|------|------|------|------|
| 路由配置 | `api/router.go` | ✅ | ✅ |
| OpenAI 端点 | `api/openai_handler.go` | ✅ | ✅ |
| 管理端点 | `api/admin_handler.go` | ✅ | ✅ |
| 中间件 | `api/middleware.go` | ✅ | - |
| 服务器 | `api/server.go` | ✅ | - |
| 主程序 | `cmd/server/main.go` | ✅ | - |

**测试结果**: 29/29 通过 ✅

### 阶段 3.5：存储层与会话管理 ✅ (2026-01-15)
| 模块 | 文件 | 状态 | 测试 |
|------|------|------|------|
| SQLite 存储 | `storage/sqlite.go` | ✅ | ✅ |
| Key 持久化 | `storage/keys.go` | ✅ | ✅ |
| 会话存储 | `storage/sessions.go` | ✅ | ✅ |
| 会话 API | `api/session_handler.go` | ✅ | ✅ |
| 会话类型 | `types/session.go` | ✅ | - |

**测试结果**: 17/17 通过 ✅

### 阶段 4：前端开发 ✅ (2026-01-16)
| 任务 | 优先级 | 状态 |
|------|--------|------|
| Vite + Vue3 + Naive UI 初始化 | P0 | ✅ 完成 |
| Dashboard 仪表盘 | P0 | ✅ 框架完成 |
| Key Manager 页面 | P0 | ✅ 完成 (含增强版向导) |
| Stats 统计页面 | P1 | ⏳ 待开发 |
| Settings 设置页面 | P1 | ✅ 完成 |
| 前端嵌入 Go 二进制 | P0 | ✅ 完成 |
| Chat 聊天功能 | P1 | ✅ 完成 (SSE 流式 + Markdown) |
| 会话持久化 | P1 | ✅ 完成 (2026-01-16) |
| App 图标设计 | P1 | ✅ 完成 (2026-01-16) |

### 阶段 4.5：窗口自定义 ✅ (2026-01-18)
| 任务 | 优先级 | 状态 |
|------|--------|------|
| Phase 1: 窗口图标修复 | P0 | ✅ 完成 (使用 windres) |
| Phase 2: 自定义标题栏按钮 | P2 | ❌ 已放弃 (技术难度高) |

> **Phase 2 放弃原因**: 需要更换 WebView 库（如 Wails）才能实现无边框窗口和自定义标题栏按钮，改动过大。保持使用 Windows 原生标题栏。

### 阶段 5：打包发布 ⏳
| 任务 | 优先级 | 状态 |
|------|--------|------|
| GoReleaser 配置 | P0 | 待开始 |
| 多平台构建测试 | P0 | 待开始 |
| GitHub Actions CI/CD | P1 | 待开始 |
| README 使用文档 | P0 | 待开始 |
| CHANGELOG 生成 | P1 | 待开始 |

---

## 📦 单模块 TDD 开发流程

每个模块开发时，严格遵循以下步骤：

```
┌─────────────────────────────────────────────────────────────┐
│  1. 阅读 ARCHITECTURE.md 中该模块的设计                      │
│                         │                                   │
│                         ▼                                   │
│  2. 创建类型定义 (如果需要)                                  │
│     internal/types/xxx.go                                   │
│                         │                                   │
│                         ▼                                   │
│  3. 编写测试用例 (测试先行)                                  │
│     internal/xxx/xxx_test.go                                │
│     - Happy Path                                            │
│     - Error Cases                                           │
│     - Edge Cases                                            │
│                         │                                   │
│                         ▼                                   │
│  4. 运行测试 (确认失败)                                      │
│     go test ./internal/xxx/...                              │
│                         │                                   │
│                         ▼                                   │
│  5. 实现代码                                                │
│     internal/xxx/xxx.go                                     │
│                         │                                   │
│                         ▼                                   │
│  6. 运行测试 (确认通过)                                      │
│     go test ./internal/xxx/...                              │
│                         │                                   │
│                         ▼                                   │
│  7. 运行竞态检测                                            │
│     go test -race ./internal/xxx/...                        │
│                         │                                   │
│                         ▼                                   │
│  8. 提交代码                                                │
└─────────────────────────────────────────────────────────────┘
```

---

## 🤖 Agent 协作规范

### Agent 角色分工

| Agent | Skill | 职责 |
|-------|-------|------|
| **Architect** | architect | 架构设计、接口定义、技术决策 |
| **Developer** | senior-golang | 代码实现、重构 |
| **QA** | qa-automation | 测试编写、质量验证 |

### Agent 交接规范

每个 Agent 完成任务后，需要在对应文档中更新状态：

```markdown
## 任务状态
- [x] 完成的任务
- [ ] 待完成的任务

## 交接给下一个 Agent
- 下一步：xxx
- 注意事项：xxx
- 相关文件：xxx
```

### Agent Prompt 模板

```markdown
## 任务：[任务名称]

### 背景
[简要说明上下文]

### 步骤
1. 阅读 `.agent/skills/[skill-name]/SKILL.md`
2. 阅读相关文档：`docs/ARCHITECTURE.md`、`docs/DEVELOPMENT.md`
3. [具体任务步骤]

### 产出
- 文件：[输出文件路径]
- 格式：[格式要求]

### 约束
- [技术约束]
- [质量要求]
```

---

## 📝 代码规范

### 文件命名
```
internal/
├── api/
│   ├── router.go           # 路由配置
│   ├── openai_handler.go   # OpenAI 端点处理器
│   ├── admin_handler.go    # 管理端点处理器
│   └── middleware.go       # 中间件
├── gemini/
│   ├── client.go           # Gemini API 客户端
│   ├── client_test.go      # 客户端测试
│   ├── converter.go        # 格式转换器
│   ├── converter_test.go   # 转换器测试
│   └── models.go           # Gemini 类型定义
└── keypool/
    ├── pool.go             # Key 池实现
    ├── pool_test.go        # Key 池测试
    └── strategy.go         # 轮询策略
```

### Commit 规范
```
<type>(<scope>): <description>

类型(type):
- feat:     新功能
- fix:      Bug 修复
- refactor: 重构
- test:     测试相关
- docs:     文档更新
- chore:    构建/工具变更

示例:
feat(keypool): implement round-robin strategy
test(converter): add edge case tests for image input
fix(api): handle SSE connection timeout
```

---

## 🔧 构建依赖

### Windows 桌面版图标

为确保 Windows 桌面应用正确显示图标，需要安装 MinGW 工具链：

```bash
# 1. 安装 MSYS2: https://www.msys2.org/
# 2. 在 MSYS2 UCRT64 终端中运行:
pacman -Syu
pacman -S mingw-w64-x86_64-binutils
```

构建脚本 (`scripts/build.ps1`) 会自动检测并使用 `windres` 编译图标资源。如未安装 MinGW，将回退到 `rsrc` 工具。

> **注意**: 使用 `rsrc` 可能导致窗口图标显示不正确。详见 [BUG-001](./bugs/BUG-001_WINDOW_ICON.md)。

---

## 🚀 快速命令

```powershell
# 运行所有测试
go test ./...

# 运行测试 + 竞态检测
go test -race ./...

# 运行特定模块测试
go test ./internal/keypool/...

# 运行 Benchmark
go test -bench=. -benchmem ./...

# 构建所有平台
.\scripts\build.ps1

# 运行 Linter
golangci-lint run ./...
```

---

## 📚 相关文档

| 文档 | 描述 |
|------|------|
| [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) | 项目概览与里程碑 |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 系统架构与接口设计 |
| [DEVELOPMENT.md](./DEVELOPMENT.md) | 后端开发工作流 |
| [FRONTEND_WORKFLOW.md](./FRONTEND_WORKFLOW.md) | 前端开发工作流 (Vue3) |
| [API.md](./API.md) | API 使用文档 (待创建) |

---

*最后更新: 2026-01-18*
