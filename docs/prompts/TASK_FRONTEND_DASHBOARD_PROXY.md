# 任务: Proxy Dashboard 功能实现

> **角色**: Frontend Developer  
> **技能**: `.agent/skills/ui-ux-pro-max/SKILL.md`, `docs/FRONTEND_WORKFLOW.md`  
> **参考文档**: `docs/API.md`, `docs/FRONTEND_PROJECT.md`

---

## 背景

MxlnAPI 是一个 OpenAI 兼容的反代服务，核心功能是将请求转发到 Gemini API。当前 Dashboard 页面为占位符，需要实现实际内容，向用户展示反代 API 的使用方法，以便他们在外部工具（如 Cursor、ChatGPT 客户端）中使用。

---

## 设计要求

### UI 风格
- **必须**遵循 Claude.ai 官网聊天界面风格
- **必须**支持 Light/Dark 双主题，使用现有 `claude-*` Tailwind 颜色 token
- 使用 Naive UI 组件库

### 色彩系统 (参考 `tailwind.config.js`)
| Token | Light | Dark |
|-------|-------|------|
| 背景 | `bg-claude-bg` (#FAF8F5) | `dark:bg-claude-dark-bg` (#191919) |
| 卡片 | `bg-white` | `dark:bg-[#212124]` |
| 边框 | `border-claude-border` | `dark:border-[#2A2A2E]` |
| 主文字 | `text-claude-text` | `dark:text-white` |
| 次文字 | `text-claude-secondaryText` | `dark:text-gray-400` |
| 强调色 | `#D97757` (Terracotta) | 同 |

---

## 步骤

### 1. 阅读规范文档
```
.agent/skills/ui-ux-pro-max/SKILL.md (设计规范)
docs/FRONTEND_WORKFLOW.md (开发流程)
```

### 2. 实现 Dashboard 页面

**文件**: `src/views/DashboardView.vue`

#### 页面结构

```vue
<template>
  <div class="min-h-screen bg-claude-bg dark:bg-claude-dark-bg p-8 transition-colors">
    <!-- Header -->
    <h1>Dashboard</h1>
    <p>MxlnAPI Proxy Service</p>

    <!-- API Endpoint Card -->
    <Card title="🔌 API Endpoint">
      <div>Base URL: <code>{{ baseUrl }}</code> <CopyButton /></div>
      <div>Status: ● Running ({{ health.keys.active }} keys active)</div>
    </Card>

    <!-- Quick Start Card -->
    <Card title="📋 Quick Start">
      <CodeBlock :code="curlExample" />
      <CopyButton />
      <Tip>无需 API Key，本地反代已配置密钥池</Tip>
    </Card>

    <!-- Stats Row -->
    <div class="grid grid-cols-3 gap-4">
      <StatCard label="Total Keys" :value="health.keys.total" />
      <StatCard label="Active Keys" :value="health.keys.active" />
      <StatCard label="System Status" :value="health.status" />
    </div>
  </div>
</template>
```

### 3. 数据获取

```typescript
import { ref, computed, onMounted } from 'vue'
import { useGlobalStore } from '../stores/global'

interface HealthInfo {
  status: 'ok' | 'degraded';
  version: string;
  uptime: number;
  keys: { total: number; active: number; rate_limited: number; disabled: number };
}

const health = ref<HealthInfo | null>(null)
const globalStore = useGlobalStore()

// 动态生成 Base URL
const baseUrl = computed(() => `${window.location.origin}/v1`)

// curl 示例
const curlExample = computed(() => `curl -X POST ${baseUrl.value}/chat/completions \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`)

async function loadHealth() {
  try {
    const res = await fetch('/health')
    health.value = await res.json()
  } catch (e) {
    console.error('Failed to load health:', e)
  }
}

onMounted(loadHealth)
```

### 4. 复制功能

```typescript
import { useMessage } from 'naive-ui'

const message = useMessage()

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  message.success('Copied to clipboard!')
}
```

---

## 验证

1. 运行 `npm run dev`
2. 确保后端运行 (`go run cmd/server/main.go`)
3. 访问 Dashboard 页面
4. 验证:
   - [ ] Base URL 显示正确 (含端口)
   - [ ] 健康状态显示 "ok" 或 "degraded"
   - [ ] Key 统计数字正确
   - [ ] Copy 按钮功能正常
   - [ ] Light/Dark 模式切换正常
   - [ ] curl 示例可直接使用

---

## 约束

- 遵循 `docs/FRONTEND_WORKFLOW.md` 代码规范
- 使用 TypeScript 严格类型，禁止 `any`
- 添加 JSDoc 注释
- 使用动态 Tailwind 类实现主题切换 (`dark:` 前缀)
- 卡片使用 `n-card` 组件并应用自定义样式覆盖

---

## 产出

| 文件 | 变更类型 |
|------|----------|
| `src/views/DashboardView.vue` | 重写 (从占位符到完整页面) |
| `docs/DEVELOPMENT.md` | 更新 (Dashboard 状态标记为完成) |

---

## 完成后更新

完成开发后，需更新以下文档:
1. `docs/DEVELOPMENT.md` - 将 Dashboard 状态改为 ✅ 完成
2. 在 `walkthrough.md` 中记录功能实现
