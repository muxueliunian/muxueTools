# MuxueTools 项目文档

> AI Studio (Gemini) → OpenAI 格式反向代理工具

---

## 📚 文档目录

### 核心文档

| 文档 | 描述 | 最后更新 |
|------|------|----------|
| [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) | 项目实施计划与里程碑 | 2026-01-15 |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 系统架构设计与接口契约 | 2026-01-16 |
| [API.md](./API.md) | API 使用文档 | 2026-01-19 |
| [DEVELOPMENT.md](./DEVELOPMENT.md) | 开发工作流指南 | 2026-01-15 |
| [FRONTEND_PROJECT.md](./FRONTEND_PROJECT.md) | 前端项目说明 | 2026-01-16 |

---

## 📊 开发进度

| 阶段 | 名称 | 状态 | 说明 |
|------|------|------|------|
| 阶段 0 | 准备 | ✅ 完成 | 计划、Skills 配置 |
| 阶段 1 | 架构设计 | ✅ 完成 | 类型定义、接口契约 |
| 阶段 2 | 核心逻辑 | ✅ 完成 | Key 池、转换器、客户端 |
| 阶段 3 | API 层 | ✅ 完成 | 路由、处理器、中间件 |
| 阶段 3.5 | 存储层 | ✅ 完成 | SQLite、会话管理 |
| 阶段 4 | 前端开发 | ✅ 完成 | Vue3 + Naive UI |
| 阶段 5 | 桌面应用 | ✅ 完成 | WebView Desktop App |
| 阶段 6 | 打包发布 | ✅ 完成 | GitHub Actions CI/CD |

### 最新完成的功能

- **模型设置** - 全局模型参数配置 (System Prompt, Temperature, Top-P/K 等)
- **Inner Chat** - 内置聊天界面，支持多模态、流式响应、会话持久化
- **Desktop App** - Windows 桌面应用（WebView 封装）
- **统计页面** - API 使用统计和趋势图表
- **API Key 管理** - 完整的 Key 增删改查、验证、导入导出
- **项目改名** - mxlnapi → muxueTools (2026-01-20)

---

## 📂 目录结构

```
docs/
├── README.md                   # 📌 文档索引（本文件）
├── IMPLEMENTATION_PLAN.md      # 实施计划
├── ARCHITECTURE.md             # 系统架构
├── API.md                      # API 使用文档
├── DEVELOPMENT.md              # 开发工作流
├── FRONTEND_PROJECT.md         # 前端项目说明
├── FRONTEND_WORKFLOW.md        # 前端开发流程
├── TASK_FORMAT.md              # Task 文档格式规范
├── CodeReviewReport/           # 代码审核报告
│   ├── TYPE_REVIEW_REPORT.md
│   ├── KEYPOOL_REVIEW_REPORT.md
│   ├── CONVERTER_REVIEW_REPORT.md
│   ├── GEMINI_CLIENT_REVIEW_REPORT.md
│   └── API_LAYER_REVIEW_REPORT.md
├── prompts/                    # Agent 任务 Prompts
│   ├── TASK_*.md               # 开发任务文档
│   └── REVIEW_*.md             # 代码审查任务
├── bugs/                       # Bug 记录
│   └── BUG-001_WINDOW_ICON.md
├── reports/                    # 开发报告
│   └── DESKTOP_APP_REPORT.md
└── gemini/                     # Gemini API 参考文档
    ├── README.md
    ├── API_REFERENCE.md
    ├── STREAMING.md
    ├── VISION.md
    └── MODELS.md
```

---

## 📋 审核报告

所有代码审核报告位于 [CodeReviewReport/](./CodeReviewReport/) 目录：

| 报告 | 模块 | 结果 |
|------|------|------|
| [TYPE_REVIEW_REPORT.md](./CodeReviewReport/TYPE_REVIEW_REPORT.md) | 类型定义 | ✅ 通过 |
| [KEYPOOL_REVIEW_REPORT.md](./CodeReviewReport/KEYPOOL_REVIEW_REPORT.md) | Key 池管理 | ✅ 通过 |
| [CONVERTER_REVIEW_REPORT.md](./CodeReviewReport/CONVERTER_REVIEW_REPORT.md) | 格式转换器 | ✅ 通过 |
| [GEMINI_CLIENT_REVIEW_REPORT.md](./CodeReviewReport/GEMINI_CLIENT_REVIEW_REPORT.md) | Gemini 客户端 | ✅ 通过 |
| [API_LAYER_REVIEW_REPORT.md](./CodeReviewReport/API_LAYER_REVIEW_REPORT.md) | API 层 | ✅ 通过 |

---

## 🤖 Agent 任务 Prompts

Agent 任务 Prompts 位于 [prompts/](./prompts/) 目录，用于指导 AI Agent 完成特定开发任务。

### 📝 格式规范
- **[TASK_FORMAT.md](./TASK_FORMAT.md)** - Task 文档创建规范（**新建任务必读**）

### 最近完成的任务
- [TASK_PROXYKEY_RENAME.md](./prompts/TASK_PROXYKEY_RENAME.md) - Proxy Key 同步 + 项目改名 ✅
- [TASK_STATS_PAGE.md](./prompts/TASK_STATS_PAGE.md) - 统计页面开发 ✅
- [TASK_FRONTEND_CHAT.md](./prompts/TASK_FRONTEND_CHAT.md) - 聊天功能开发 ✅
- [TASK_DESKTOP_APP.md](./prompts/TASK_DESKTOP_APP.md) - 桌面应用开发 ✅
- [TASK_WINDOW_CUSTOMIZATION.md](./prompts/TASK_WINDOW_CUSTOMIZATION.md) - 窗口自定义 ✅

### 待开发任务
- [TASK_I18N.md](./prompts/TASK_I18N.md) - 国际化 (中/英/日) ✅
- [TASK_ANDROID_APK.md](./prompts/TASK_ANDROID_APK.md) - Android APK 适配 (Android 12+) ⏳

---

## 📖 参考文档

### Gemini API 文档
[gemini/](./gemini/) 目录包含 Google AI Studio (Gemini) API 的参考文档摘要：

| 文档 | 内容 |
|------|------|
| [README.md](./gemini/README.md) | 概述 |
| [API_REFERENCE.md](./gemini/API_REFERENCE.md) | API 参考 |
| [STREAMING.md](./gemini/STREAMING.md) | 流式响应 |
| [VISION.md](./gemini/VISION.md) | 图片输入 |
| [MODELS.md](./gemini/MODELS.md) | 可用模型 |

---

## 🚀 快速开始

### 开发环境
```powershell
# 运行所有测试
go test ./...

# 运行测试 + 竞态检测
go test -race ./...

# 构建桌面应用
.\scripts\build.ps1 desktop

# 启动开发服务器
go run ./cmd/server/main.go
```

### 前端开发
```powershell
cd web
npm install
npm run dev
```

### 相关链接
- [项目 README](../README.md)
- [Go 标准项目布局](https://github.com/golang-standards/project-layout)

---

*最后更新: 2026-01-20*
