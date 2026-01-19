# MxlnAPI 前端项目文档

> 本文档描述前端项目的当前状态、已完成的工作和待开发的功能。
> Agent 开发前必须阅读此文档以了解项目全貌。

---

## 📁 项目结构

```
web/
├── src/
│   ├── api/                    # API 层
│   │   ├── client.ts           # ✅ Axios 实例 (基础封装)
│   │   ├── chat.ts             # ✅ Chat API (SSE 流式调用)
│   │   ├── keys.ts             # ✅ Key API (CRUD + validate)
│   │   ├── config.ts           # ✅ Config API
│   │   └── types.ts            # ✅ API 类型定义 (含 Chat 类型)
│   ├── assets/                 # 静态资源
│   │   └── main.css            # 全局样式
│   ├── components/             # 通用组件
│   │   ├── HelloWorld.vue      # ⚠️ 脚手架示例，待删除
│   │   └── chat/               # ✅ Chat 组件
│   │       ├── ChatInput.vue   # ✅ 输入框组件
│   │       ├── MessageBubble.vue # ✅ 消息气泡 (Markdown + 代码高亮)
│   │       ├── MessageList.vue # ✅ 消息列表
│   │       └── ModelSelector.vue # ✅ 模型选择器
│   ├── layouts/                # 布局组件
│   │   └── MainLayout.vue      # ✅ 主布局 (侧边栏 + 内容区)
│   ├── router/
│   │   └── index.ts            # ✅ 路由配置 (5 个路由)
│   ├── stores/                 # Pinia 状态管理
│   │   ├── global.ts           # ✅ 全局状态 (isDark, sidebarCollapsed)
│   │   ├── keyStore.ts         # ✅ Key 管理状态
│   │   └── chatStore.ts        # ✅ Chat 状态管理
│   ├── views/                  # 页面视图
│   │   ├── ChatView.vue        # ✅ Chat 页面 (完整对话功能)
│   │   ├── DashboardView.vue   # ✅ API Endpoint + Quick Start + Stats
│   │   ├── KeyManagerView.vue  # ✅ Key CRUD + 验证向导
│   │   ├── SettingsView.vue    # 🔨 配置表单 (基础完成)
│   │   └── StatsView.vue       # ⏳ 骨架 (待开发)
│   ├── App.vue                 # ✅ 根组件 (NConfigProvider)
│   ├── main.ts                 # ✅ 入口
│   ├── style.css               # ✅ Tailwind CSS 引入
│   └── theme.ts                # ✅ Naive UI 主题配置
├── index.html
├── package.json
├── tailwind.config.js
├── tsconfig.json
└── vite.config.ts              # ✅ 别名 + Proxy 配置
```

---

## 🔌 路由表

| 路径 | 名称 | 组件 | 状态 |
|------|------|------|------|
| `/` | chat | ChatView.vue | ✅ 完成 |
| `/dashboard` | dashboard | DashboardView.vue | ✅ 完成 |
| `/keys` | keys | KeyManagerView.vue | ✅ 完成 |
| `/stats` | stats | StatsView.vue | ⏳ 骨架 |
| `/settings` | settings | SettingsView.vue | 🔨 基础完成 |

---

## 📦 API 类型 (`api/types.ts`)

已定义的类型：

| 类型 | 描述 |
|------|------|
| `KeyInfo` | API Key 信息 (id, key, name, status, stats, provider, default_model) |
| `KeyStats` | Key 使用统计 |
| `Session` | 会话信息 |
| `Message` | 聊天消息 |
| `ApiResponse<T>` | 通用 API 响应 |
| `ListResponse<T>` | 列表 API 响应 |
| `HealthStats` | 健康检查响应 |
| `ValidateKeyResult` | Key 验证结果 (在 keys.ts) |
| `ChatCompletionMessage` | OpenAI 格式 Chat 消息 |
| `ChatCompletionRequest` | Chat 请求参数 |
| `ChatCompletionChunk` | SSE 流式响应 chunk |

---

## 🗄️ 状态管理 (Pinia Stores)

### 现有 Store

| Store | 文件 | 状态 | 描述 |
|-------|------|------|------|
| `useGlobalStore` | stores/global.ts | ✅ | 主题 (isDark)、侧边栏收折 |
| `useKeyStore` | stores/keyStore.ts | ✅ | Key CRUD 操作 |
| `useChatStore` | stores/chatStore.ts | ✅ | 消息列表、生成状态、SSE 流式 |
| `useSessionStore` | stores/sessionStore.ts | ✅ | 会话列表、切换、消息持久化 |

### 待创建 Store

| Store | 描述 | 优先级 |
|-------|------|--------|
| `useStatsStore` | 统计数据缓存 | P1 |

---

## 🚧 开发进度

### ✅ 已完成
- 项目脚手架 (Vite + Vue3 + TS + Naive UI + Tailwind)
- 路由配置
- 主布局 (侧边栏导航)
- API 客户端基础封装
- 核心 API 类型定义
- 全局状态 (主题切换)
- 桌面应用 WebView 封装
- **Dashboard 页面** - API Endpoint + Quick Start + Key 统计
- [x] **Key 管理页面** - CRUD + 4 步验证向导 + 模型选择 + 按 key/name/tag 搜索 + 批量导入
- **Chat 页面**:
  - SSE 流式响应
  - Markdown 渲染 (标题、列表、表格、代码块、引用等)
  - 代码语法高亮 (highlight.js)
  - 模型选择器 (调用 Gemini API 实时获取可用模型)
  - 模型偏好 localStorage 持久化
  - 隐藏滚动条 (保持界面简洁)
- **会话持久化**:
  - 侧边栏会话列表 (SessionList)
  - 会话切换加载历史消息
  - 发送/接收消息自动保存到后端
  - 刷新页面保留对话上下文
- **新增后端 API**: `GET /api/models` - 获取当前 Key Pool 可用模型列表
- **App 图标设计** - MyGO 高松灯主题企鹅图标，集成到桌面应用
- **窗口图标修复 (2026-01-18)** - 使用 windres 正确嵌入图标资源

### 🔨 进行中
- **Settings**: 配置表单基础完成
- **UI 细节优化**: 暗色模式对比度、加载状态等

### ⏳ 待开发

| 功能 | 对应任务 | 描述 |
|------|----------|------|
| **Stats 页面** | TBD | 详细使用统计 |

### ❌ 已放弃

| 功能 | 原因 |
|------|------|
| 自定义标题栏按钮 | 技术难度高，需更换 WebView 库 |

---

## 📚 参考文档

| 文档 | 路径 | 描述 |
|------|------|------|
| API 文档 | `docs/API.md` | 所有后端 API 端点详情 |
| 开发工作流 | `docs/FRONTEND_WORKFLOW.md` | 代码规范、开发循环 |
| 架构设计 | `docs/ARCHITECTURE.md` | 系统整体架构 |

---

## 🛠️ 开发命令

```bash
cd web
npm install        # 安装依赖
npm run dev        # 启动开发服务器 (http://localhost:5173)
npm run build      # 构建生产版本 (输出到 dist/)
```

---

*最后更新: 2026-01-18 (API Keys UI Optimization Completed)*
