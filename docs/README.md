# MxlnAPI 项目文档

> AI Studio (Gemini) → OpenAI 格式反向代理工具

---

## 📚 文档目录

### 核心文档

| 文档 | 描述 | 最后更新 |
|------|------|----------|
| [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) | 项目实施计划与里程碑 | 2026-01-15 |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 系统架构设计与接口契约 | 2026-01-15 |
| [DEVELOPMENT.md](./DEVELOPMENT.md) | 开发工作流指南 | 2026-01-15 |

---

## 📊 开发进度

| 阶段 | 名称 | 状态 | 说明 |
|------|------|------|------|
| 阶段 0 | 准备 | ✅ 完成 | 计划、Skills 配置 |
| 阶段 1 | 架构设计 | ✅ 完成 | 类型定义、接口契约 |
| 阶段 2 | 核心逻辑 | ✅ 完成 | Key 池、转换器、客户端 |
| 阶段 3 | API 层 | ✅ 完成 | 路由、处理器、中间件 |
| 阶段 3.5 | 存储层 | ✅ 完成 | SQLite、会话管理 |
| 阶段 4 | 前端开发 | 🔜 进行中 | Vue3 + Naive UI |
| 阶段 5 | 打包发布 | ⏳ 待开始 | GoReleaser、CI/CD |

---

## 📂 目录结构

```
docs/
├── README.md                   # 📌 文档索引（本文件）
├── IMPLEMENTATION_PLAN.md      # 实施计划
├── ARCHITECTURE.md             # 系统架构
├── DEVELOPMENT.md              # 开发工作流
├── CodeReviewReport/           # 代码审核报告
│   ├── TYPE_REVIEW_REPORT.md
│   ├── KEYPOOL_REVIEW_REPORT.md
│   ├── CONVERTER_REVIEW_REPORT.md
│   ├── GEMINI_CLIENT_REVIEW_REPORT.md
│   └── API_LAYER_REVIEW_REPORT.md
├── prompts/                    # Agent 任务 Prompts
│   ├── TASK_CONVERTER.md
│   ├── REVIEW_CONVERTER.md
│   ├── TASK_GEMINI_CLIENT.md
│   ├── REVIEW_GEMINI_CLIENT.md
│   ├── TASK_API_LAYER.md
│   ├── REVIEW_API_LAYER.md
│   ├── TASK_DOCS_CLEANUP.md
│   └── TASK_FRONTEND.md        # 🔜 前端开发任务
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
- **[TASK_FORMAT.md](./prompts/TASK_FORMAT.md)** - Task 文档创建规范（**新建任务必读**）

### 已完成任务
- `TASK_CONVERTER.md` / `REVIEW_CONVERTER.md` - 格式转换器开发与审核
- `TASK_GEMINI_CLIENT.md` / `REVIEW_GEMINI_CLIENT.md` - Gemini 客户端开发与审核
- `TASK_API_LAYER.md` / `REVIEW_API_LAYER.md` - API 层开发与审核
- `TASK_STORAGE_LAYER.md` - 存储层开发任务
- `TASK_DOCS_CLEANUP.md` - 文档整理任务

### 待执行任务
- `TASK_FRONTEND.md` - 前端开发任务（Vue3 + Naive UI）
- `TASK_WINDOW_CUSTOMIZATION.md` - 窗口标题栏自定义（图标 + 按钮）

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

# 启动服务（开发模式）
go run ./cmd/server/main.go
```

### 相关链接
- [项目 README](../README.md)
- [Go 标准项目布局](https://github.com/golang-standards/project-layout)

---

*最后更新: 2026-01-15*
