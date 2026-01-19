# 任务: 会话持久化 - 前端实现

> **角色**: Senior Frontend Developer  
> **技能**: `.agent/skills/ui-ux-pro-max/SKILL.md` (UI 设计规范 - 必读), `.agent/skills/qa-automation/SKILL.md` (测试)  
> **必读文档**: `docs/FRONTEND_WORKFLOW.md`, `docs/API.md`, `docs/FRONTEND_PROJECT.md`

---

## 背景

MxlnAPI 的 Chat 功能已实现流式对话，但会话记录仅存于内存，刷新页面后丢失。后端会话 API 已完备，需要前端实现持久化功能。

---

## 需求决策 (已确认)

| 需求项 | 决策 |
|--------|------|
| 会话标题 | 使用第一条用户消息截取 (前 50 字符) |
| 删除交互 | 鼠标悬停显示删除按钮 |
| 模型绑定 | 全局模型，所有会话共用 |

---

## 🎨 UI 设计规范

### 设计风格: 完全模仿 Claude.ai

> **必须**参考现有 `MainLayout.vue` 和 Chat 组件的样式，保持一致性。
> 如有疑问，可访问 https://claude.ai 查看真实界面。

### 色彩系统 (使用现有 Tailwind token)

| 元素 | Light Mode | Dark Mode |
|------|------------|-----------|
| 侧边栏背景 | `bg-claude-sidebar` (#F5F4F1) | `dark:bg-claude-dark-sidebar` (#191919) |
| 悬停背景 | `bg-claude-hover` (#E6E4E1) | `dark:bg-claude-dark-hover` (#212124) |
| 选中背景 | `bg-[#E1DFDD]` | `dark:bg-[#303030]` |
| 边框 | `border-claude-border` | `dark:border-claude-dark-border` |
| 主文本 | `text-claude-text` (#1F1E1D) | `dark:text-claude-dark-text` (#E5E7EB) |
| 次要文本 | `text-claude-secondaryText` (#6F6F78) | `dark:text-gray-400` |
| 强调色 | `#D97757` (Terracotta) | 同 |

### SessionItem 组件样式要求

```
┌─────────────────────────────────────┐
│  会话标题...              🗑️ (悬停) │
│  2 hours ago                        │
└─────────────────────────────────────┘
```

**交互细节**:
- **默认状态**: 显示标题 (截取后加 `...`) + 相对时间
- **悬停状态**: 背景变色，右侧显示删除图标
- **选中状态**: 背景加深，左侧可选添加指示条
- **删除按钮**: 使用 `Trash2` 图标 (lucide-vue-next)，悬停变红

**字体样式**:
- 标题: `text-sm font-medium truncate`
- 时间: `text-xs text-claude-secondaryText`

### SessionList 容器

- 位于侧边栏 "New Chat" 按钮下方
- 可滚动，使用隐藏滚动条样式 (参考 `MessageList.vue`)
- 会话按 `updated_at` 降序排列 (最新在上)

### 删除确认

使用 Naive UI 的 `useDialog`:
```typescript
dialog.warning({
    title: '删除会话',
    content: '确定要删除这个会话吗？此操作不可撤销。',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: () => { /* 执行删除 */ }
})
```

### 图标使用

从 `lucide-vue-next` 导入:
- `Plus` - New Chat 按钮 (已存在)
- `Trash2` - 删除按钮
- `MessageSquare` - 会话图标 (可选)

---

## ⚠️ 约束 - 严格遵守

### 文件修改范围 (仅限以下文件)

**新建**:
- `src/api/sessions.ts`
- `src/stores/sessionStore.ts`
- `src/components/chat/SessionList.vue`
- `src/components/chat/SessionItem.vue`

**修改**:
- `src/api/types.ts` (添加 Session 类型)
- `src/stores/chatStore.ts` (集成 sessionStore)
- `src/layouts/MainLayout.vue` (侧边栏集成会话列表)
- `docs/FRONTEND_PROJECT.md` (更新开发进度)

**禁止修改任何其他文件**，包括但不限于:
- 后端 Go 代码
- 配置文件
- 其他 View 或组件

### 代码规范

- 遵循 `docs/FRONTEND_WORKFLOW.md` 全部规范
- TypeScript 严格类型，**禁止 `any`**
- 所有函数添加 **JSDoc 注释**
- 组件使用 Composition API (`<script setup lang="ts">`)
- 超过 200 行的组件必须抽离逻辑到 `composables/`
- UI 组件**禁止**直接调用 `axios`，必须通过 `stores/` 或 `api/` 层
- **禁止**使用 `window.open()`、`alert()`、`confirm()`

---

## 步骤

### 1. 阅读规范 📖

```
docs/FRONTEND_WORKFLOW.md (开发工作流、编码规范)
docs/API.md (会话管理 API - 搜索 "/api/sessions")
docs/FRONTEND_PROJECT.md (项目当前状态)
```

### 2. 扩展 API 类型

**文件**: `src/api/types.ts`

```typescript
// ==================== Session Types ====================

export interface Session {
    id: string
    title: string
    model: string
    created_at: string
    updated_at: string
}

export interface SessionMessage {
    id: string
    session_id: string
    role: 'user' | 'assistant' | 'system'
    content: string
    prompt_tokens?: number
    completion_tokens?: number
    created_at: string
}

export interface SessionListResponse {
    success: boolean
    sessions: Session[]
    total: number
}

export interface SessionDetailResponse {
    success: boolean
    session: Session
    messages: SessionMessage[]
}

export interface CreateSessionRequest {
    title?: string
    model?: string
}

export interface AddMessageRequest {
    role: 'user' | 'assistant' | 'system'
    content: string
    prompt_tokens?: number
    completion_tokens?: number
}
```

### 3. 实现 Session API

**文件**: `src/api/sessions.ts`

```typescript
import apiClient from './client'
import type {
    Session,
    SessionListResponse,
    SessionDetailResponse,
    CreateSessionRequest,
    AddMessageRequest,
    SessionMessage
} from './types'

/**
 * 获取会话列表
 * @param limit 每页数量 (默认 20)
 * @param offset 偏移量
 */
export async function getSessions(limit = 20, offset = 0): Promise<SessionListResponse>

/**
 * 创建新会话
 */
export async function createSession(data: CreateSessionRequest): Promise<{ success: boolean; data: Session }>

/**
 * 获取会话详情 (含消息)
 */
export async function getSession(id: string): Promise<SessionDetailResponse>

/**
 * 更新会话
 */
export async function updateSession(id: string, data: { title?: string; model?: string }): Promise<{ success: boolean; data: Session }>

/**
 * 删除会话
 */
export async function deleteSession(id: string): Promise<{ success: boolean; message: string }>

/**
 * 添加消息到会话
 */
export async function addMessage(sessionId: string, data: AddMessageRequest): Promise<{ success: boolean; data: SessionMessage }>
```

### 4. 实现 Session Store

**文件**: `src/stores/sessionStore.ts`

```typescript
/**
 * Session Store - 会话状态管理
 * 
 * 职责: 管理会话列表、当前会话、消息持久化
 */

export const useSessionStore = defineStore('session', () => {
    // State
    const sessions = ref<Session[]>([])
    const currentSessionId = ref<string | null>(null)
    const isLoading = ref(false)

    // Computed
    const currentSession = computed(...)

    // Actions
    async function loadSessions(): Promise<void>
    async function createNewSession(): Promise<Session>
    async function switchSession(id: string): Promise<void>
    async function deleteSession(id: string): Promise<void>
    async function saveMessage(role: 'user' | 'assistant', content: string): Promise<void>
    
    // 根据第一条消息更新会话标题
    async function updateSessionTitle(firstMessage: string): Promise<void>
})
```

### 5. 修改 Chat Store

**文件**: `src/stores/chatStore.ts`

修改 `sendMessage` 函数:
1. 发送用户消息后，调用 `sessionStore.saveMessage('user', content)`
2. 收到助手回复后，调用 `sessionStore.saveMessage('assistant', content)`
3. 首次发送消息时，自动更新会话标题

### 6. 实现会话列表组件

**文件**: `src/components/chat/SessionList.vue`

```vue
<template>
    <div class="session-list">
        <SessionItem 
            v-for="session in sessions" 
            :key="session.id"
            :session="session"
            :active="session.id === currentSessionId"
            @click="handleClick(session.id)"
            @delete="handleDelete(session.id)"
        />
    </div>
</template>
```

**文件**: `src/components/chat/SessionItem.vue`

- 显示标题 (截取后)、时间
- 当前会话高亮样式
- 悬停显示删除按钮
- 删除确认 (使用 Naive UI 的 `useDialog`)

### 7. 集成到侧边栏

**文件**: `src/layouts/MainLayout.vue`

在现有 "New Chat" 按钮下方添加 `<SessionList />` 组件:
- 替换或增强现有侧边栏会话入口
- 保持现有样式风格一致

---

## 产出文件

| 文件 | 类型 | 描述 |
|------|------|------|
| `src/api/types.ts` | MODIFY | 添加 Session 相关类型 |
| `src/api/sessions.ts` | NEW | Session API 封装 |
| `src/stores/sessionStore.ts` | NEW | 会话状态管理 |
| `src/stores/chatStore.ts` | MODIFY | 集成消息持久化 |
| `src/components/chat/SessionList.vue` | NEW | 会话列表组件 |
| `src/components/chat/SessionItem.vue` | NEW | 会话项组件 |
| `src/layouts/MainLayout.vue` | MODIFY | 侧边栏集成 |
| `docs/FRONTEND_PROJECT.md` | MODIFY | 更新开发进度 |

---

## 验证

### 1. TypeScript 构建
```bash
cd web && npm run build
```

### 2. 功能测试

启动后端和前端:
```bash
# Terminal 1
go run cmd/server/main.go

# Terminal 2
cd web && npm run dev
```

测试检查清单:
- [ ] 首次打开时自动创建新会话
- [ ] "New Chat" 按钮创建新会话
- [ ] 侧边栏显示会话列表
- [ ] 点击会话可切换
- [ ] 发送消息后自动保存
- [ ] 收到回复后自动保存
- [ ] 切换会话可加载历史消息
- [ ] 悬停会话显示删除按钮
- [ ] 删除会话功能正常
- [ ] **刷新页面后历史消息仍在**
- [ ] Light/Dark 模式显示正常

---

## 后端 API 参考

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/sessions` | GET | 获取会话列表 (支持分页 `?limit=20&offset=0`) |
| `/api/sessions` | POST | 创建会话 `{ title?, model? }` |
| `/api/sessions/:id` | GET | 获取会话详情 (含消息) |
| `/api/sessions/:id` | PUT | 更新会话 `{ title?, model? }` |
| `/api/sessions/:id` | DELETE | 删除会话 |
| `/api/sessions/:id/messages` | POST | 添加消息 `{ role, content, prompt_tokens?, completion_tokens? }` |

详见 `docs/API.md` 会话管理章节。

---

## 交接

- **前置**: Chat 流式对话功能已完成
- **完成后**: 用户对话历史可持久化，刷新页面不丢失
