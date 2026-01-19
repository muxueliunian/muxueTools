# 任务：Chat 对话功能完整实现

> **角色**: Senior Frontend Developer & UI/UX Designer  
> **技能**: `.agent/skills/ui-ux-pro-max/SKILL.md` (必读)  
> **参考文档**: `docs/FRONTEND_PROJECT.md`, `docs/FRONTEND_WORKFLOW.md`, `docs/API.md`

---

## 背景

MuxueTools 是一个 OpenAI 兼容的 API 反代服务。当前 `ChatView.vue` 仅有输入框占位符，需要实现完整的 Chat 对话功能，同时作为**验证本地反代 API 是否正常工作**的工具。

---

## 设计要求 🎨

### UI 风格：完全模仿 Claude.ai

> **必须**访问 [claude.ai](https://claude.ai) 网站查看真实聊天界面样式，确保 UI 100% 还原。

**关键设计元素**:
| 元素 | Claude.ai 风格 |
|------|---------------|
| 布局 | 居中对话区域，最大宽度约 768px |
| 消息区分 | 用户消息无背景/右侧，助手消息有淡色背景 |
| 输入框 | 底部固定，带圆角边框，多行自适应 |
| 发送按钮 | 圆形 Terracotta 色 (#D97757) |
| 代码块 | 深色背景 + 语法高亮 + 复制按钮 |
| 字体 | 衬线标题 (Merriweather)，无衬线正文 (Inter) |

### 色彩系统

使用现有 Tailwind 配置的 `claude-*` token：

| 元素 | Light | Dark |
|------|-------|------|
| 页面背景 | `bg-claude-bg` (#FAF8F5) | `dark:bg-claude-dark-bg` (#191919) |
| 助手消息背景 | `bg-[#F0EEEB]` | `dark:bg-[#2A2A2E]` |
| 用户消息 | 无背景 | 无背景 |
| 输入框 | `bg-white` + 边框 | `dark:bg-[#303030]` |
| 强调色 | `#D97757` (发送按钮) | 同 |

---

## 步骤

### 1. 阅读规范 📖

```
.agent/skills/ui-ux-pro-max/SKILL.md (UI/UX 设计规范)
docs/FRONTEND_WORKFLOW.md (开发工作流)
docs/FRONTEND_PROJECT.md (项目当前状态)
docs/API.md (API 接口文档 - 重点看 /v1/chat/completions 和会话管理)
```

### 2. 设计确认 (Design First) 🛑

**必须**：
1. 访问 https://claude.ai 查看真实 Chat 界面
2. 使用 `generate_image` 工具生成 Chat 界面 Mockup
3. 向用户展示设计并获取批准后再编码

### 3. 安装依赖

```bash
cd web
npm install marked highlight.js @types/marked
```

### 4. 创建 Chat Store

**文件**: `src/stores/chatStore.ts`

```typescript
interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  createdAt: Date
}

// 状态: messages, isGenerating, error, abortController
// Actions: sendMessage(), stopGeneration(), clearMessages()
```

### 5. 创建 SSE 流式 API

**文件**: `src/api/chat.ts`

```typescript
/**
 * 流式调用 /v1/chat/completions
 * 使用 fetch + ReadableStream 处理 SSE
 * 解析 data: {...} 格式，处理 [DONE] 信号
 */
export async function* streamChatCompletion(
  messages: { role: string; content: string }[],
  model: string,
  signal?: AbortSignal
): AsyncGenerator<string, void, unknown>
```

### 6. 创建消息组件

**文件**: `src/components/chat/MessageBubble.vue`

- 区分 `user` 和 `assistant` 样式
- Markdown 渲染（使用 `marked`）
- 代码块语法高亮（使用 `highlight.js`）
- 代码块复制按钮
- 支持 Light/Dark 主题

**文件**: `src/components/chat/MessageList.vue`

- 消息列表容器
- 自动滚动到底部
- 生成中显示打字机光标效果

### 7. 改造 ChatView

**文件**: `src/views/ChatView.vue`

**布局**:
```vue
<template>
  <div class="flex flex-col h-full">
    <!-- 消息区域 (可滚动) -->
    <div class="flex-1 overflow-y-auto">
      <MessageList :messages="store.messages" />
    </div>
    
    <!-- 输入区域 (固定底部) -->
    <div class="border-t p-4">
      <ChatInput 
        @send="handleSend" 
        @stop="handleStop"
        :loading="store.isGenerating" 
      />
    </div>
  </div>
</template>
```

**功能**:
- `Enter` 发送，`Shift+Enter` 换行
- 发送按钮在生成时变为"停止"按钮
- 取消正在进行的请求 (AbortController)
- 错误显示 Toast 提示

### 8. 输入组件

**文件**: `src/components/chat/ChatInput.vue`

模仿 Claude.ai 输入框样式：
- 多行自适应高度
- 圆角边框
- 发送按钮圆形居右
- 生成中显示停止按钮

---

## 产出文件

| 文件 | 类型 | 描述 |
|------|------|------|
| `src/stores/chatStore.ts` | NEW | Chat 状态管理 |
| `src/api/chat.ts` | NEW | SSE 流式调用封装 |
| `src/components/chat/MessageBubble.vue` | NEW | 消息气泡 (Markdown + 代码高亮) |
| `src/components/chat/MessageList.vue` | NEW | 消息列表 |
| `src/components/chat/ChatInput.vue` | NEW | 输入框组件 |
| `src/views/ChatView.vue` | MODIFY | 完整 Chat 页面 |
| `docs/FRONTEND_PROJECT.md` | MODIFY | 更新开发进度 |

---

## 验证

1. **TypeScript 构建**:
   ```bash
   cd web && npm run build
   ```

2. **功能测试**:
   - 启动后端: `go run cmd/server/main.go`
   - 启动前端: `npm run dev`
   - 打开 Chat 页面 (http://localhost:5173/)
   - 发送消息验证:
     - [ ] 消息发送成功
     - [ ] 流式响应逐字显示（打字机效果）
     - [ ] Markdown 正确渲染
     - [ ] 代码块有语法高亮 + 复制按钮
     - [ ] 停止按钮可中断生成
     - [ ] Light/Dark 模式切换正常

3. **浏览器截图**:
   使用 `browser_subagent` 工具截取 Chat 界面截图供用户确认。

---

## 约束

- 遵循 `docs/FRONTEND_WORKFLOW.md` 代码规范
- 使用 TypeScript 严格类型，**禁止 `any`**
- **禁止**使用 `window.open()`、`alert()`、`confirm()`
- 所有函数添加 **JSDoc 注释**
- 组件使用 Composition API (`<script setup lang="ts">`)
- 超过 200 行的组件必须抽离逻辑到 `composables/`
- UI 组件**禁止**直接调用 `axios`，必须通过 `stores/` 或 `api/` 层

---

## 交接

- **前置**: Dashboard 和 Key 管理已完成
- **完成后**: 用户可在 Chat 页面验证反代 API 是否正常工作

---

*提示: 如果对 Claude.ai 界面细节有疑问，直接使用 `browser_subagent` 工具访问 https://claude.ai 查看真实样式。*
