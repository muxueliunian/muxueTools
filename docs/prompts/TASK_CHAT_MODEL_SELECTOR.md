# 任务: Chat 模型选择功能实现

> **角色**: Frontend Developer  
> **技能**: `docs/FRONTEND_WORKFLOW.md`  
> **参考文档**: `docs/FRONTEND_PROJECT.md`, `docs/API.md`

---

## 背景

Chat 功能已实现基础对话能力，现需添加**模型选择功能**，允许用户从后端返回的可用模型列表中选择当前对话使用的模型。

---

## 现有 API

后端已提供 `GET /v1/models` 接口（OpenAI 兼容格式），返回可用模型列表：

```json
{
  "object": "list",
  "data": [
    { "id": "gemini-2.0-flash", "object": "model", "created": 1677610602, "owned_by": "google" },
    { "id": "gemini-1.5-pro", "object": "model", "created": 1677610602, "owned_by": "google" }
  ]
}
```

---

## 步骤

### 1. 阅读规范 📖

```
docs/FRONTEND_WORKFLOW.md (开发工作流)
docs/FRONTEND_PROJECT.md (项目当前状态)
docs/API.md (API 接口文档 - /v1/models)
```

### 2. 扩展 API 层

**文件**: `src/api/chat.ts`

添加获取模型列表函数：

```typescript
/**
 * 模型信息（OpenAI 兼容格式）
 */
export interface ModelInfo {
    /** 模型 ID (e.g., 'gemini-2.0-flash') */
    id: string
    /** 对象类型，固定为 'model' */
    object: string
    /** 创建时间戳 */
    created: number
    /** 模型所有者 (e.g., 'google') */
    owned_by: string
}

/**
 * 获取可用模型列表
 * @returns 模型信息数组
 * @throws 网络错误或解析错误
 */
export async function fetchModels(): Promise<ModelInfo[]> {
    const response = await fetch('/v1/models')
    if (!response.ok) {
        throw new Error(`Failed to fetch models: ${response.status}`)
    }
    const data = await response.json()
    return data.data || []
}
```

### 3. 扩展 Store 层

**文件**: `src/stores/chatStore.ts`

新增状态和方法：

```typescript
// ========== State ==========
/** 可用模型列表 */
const availableModels = ref<string[]>([])
/** 模型加载状态 */
const isLoadingModels = ref(false)

// ========== Actions ==========
/**
 * 从后端加载可用模型列表
 */
async function loadModels(): Promise<void> {
    if (isLoadingModels.value) return
    isLoadingModels.value = true
    try {
        const models = await fetchModels()
        availableModels.value = models.map(m => m.id)
        // 如果当前模型不在列表中，选择第一个
        if (availableModels.value.length > 0 && 
            !availableModels.value.includes(currentModel.value)) {
            currentModel.value = availableModels.value[0]
        }
        saveModelPreference()
    } finally {
        isLoadingModels.value = false
    }
}

/**
 * 设置当前模型并持久化到 localStorage
 */
function setModel(modelId: string): void {
    currentModel.value = modelId
    localStorage.setItem('mxln_preferred_model', currentModel.value)
}

/**
 * 从 localStorage 恢复模型偏好
 */
function restoreModelPreference(): void {
    const saved = localStorage.getItem('mxln_preferred_model')
    if (saved) currentModel.value = saved
}
```

### 4. 创建模型选择器组件

**文件**: `src/components/chat/ModelSelector.vue`

```vue
<script setup lang="ts">
/**
 * 模型选择器组件
 * 职责: 显示可用模型下拉列表，允许用户切换当前对话模型
 * 依赖: ChatStore
 */
import { computed } from 'vue'
import { NSelect } from 'naive-ui'
import { useChatStore } from '../../stores/chatStore'

const chatStore = useChatStore()

const modelOptions = computed(() => 
    chatStore.availableModels.map(id => ({
        label: formatModelLabel(id),
        value: id
    }))
)

/**
 * 格式化模型名称 (gemini-2.0-flash -> Gemini 2.0 Flash)
 */
function formatModelLabel(modelId: string): string {
    return modelId
        .split('-')
        .map(part => part.charAt(0).toUpperCase() + part.slice(1))
        .join(' ')
}
</script>

<template>
    <n-select
        :value="chatStore.currentModel"
        :options="modelOptions"
        :loading="chatStore.isLoadingModels"
        size="small"
        style="width: 180px"
        placeholder="选择模型"
        @update:value="chatStore.setModel"
    />
</template>
```

### 5. 集成到 ChatView

**文件**: `src/views/ChatView.vue`

在输入区域上方添加模型选择器:

```vue
<template>
    <!-- 输入区域 -->
    <div class="border-t p-4">
        <div class="flex items-center gap-4 mb-3 max-w-3xl mx-auto">
            <ModelSelector />
            <span class="text-xs text-claude-secondaryText dark:text-gray-500">
                按 Enter 发送，Shift + Enter 换行
            </span>
        </div>
        <ChatInput @send="handleSend" :loading="chatStore.isGenerating" />
    </div>
</template>

<script setup>
onMounted(() => {
    chatStore.restoreModelPreference()
    chatStore.loadModels()
})
</script>
```

---

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `src/api/chat.ts` | MODIFY | 添加 `ModelInfo` 类型和 `fetchModels()` 函数 |
| `src/stores/chatStore.ts` | MODIFY | 添加模型列表状态、加载方法、持久化逻辑 |
| `src/components/chat/ModelSelector.vue` | NEW | 模型选择器组件 |
| `src/views/ChatView.vue` | MODIFY | 集成模型选择器 |

---

## 验证

1. **TypeScript 构建**:
   ```bash
   cd web && npm run build
   ```

2. **功能测试**:
   - [ ] 页面加载时自动获取模型列表
   - [ ] 下拉框显示所有可用模型（格式化名称）
   - [ ] 切换模型后发送消息使用新模型
   - [ ] 刷新页面后恢复上次选择的模型
   - [ ] Light/Dark 模式下样式正确

---

## 约束

- 遵循 `docs/FRONTEND_WORKFLOW.md` 代码规范
- 使用 TypeScript 严格类型，**禁止 `any`**
- 所有函数添加 **JSDoc 注释**
- Store 层处理 API 调用，组件层只负责 UI
