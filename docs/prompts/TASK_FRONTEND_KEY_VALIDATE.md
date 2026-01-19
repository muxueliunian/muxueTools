# 任务: 增强版 API Key 管理 - 前端实现

> **角色**: Developer (Frontend Engineer)  
> **技能**: `docs/FRONTEND_WORKFLOW.md`, `.agent/skills/ui-ux-pro-max/SKILL.md` (设计参考)  
> **参考文档**: `docs/FRONTEND_PROJECT.md`, `docs/API.md`

---

## 背景

当前 "Add API Key" 弹窗仅支持输入密钥和名称。需要改造为多步骤向导：
1. 选择供应商 → 输入 Key → 验证
2. 从返回的模型列表中选择默认模型
3. 填写名称和标签
4. 创建 Key

---

## 前置条件

- **后端依赖**: `POST /api/keys/validate` 接口已实现 (参考 `TASK_BACKEND_KEY_VALIDATE.md`)

---

## 步骤

### 1. 阅读规范
```
docs/FRONTEND_WORKFLOW.md (开发循环)
.agent/skills/ui-ux-pro-max/SKILL.md (设计规范 - 可选参考)
```

### 2. 扩展 API 层

**文件**: `src/api/keys.ts`

```typescript
/** 验证 Key 并获取可用模型列表 */
export interface ValidateKeyResult {
    valid: boolean;
    latency_ms: number;
    models: string[];
    error?: string;
}

export const validateKey = async (data: { key: string; provider?: string }) =>
    (await apiClient.post<ApiResponse<ValidateKeyResult>>('/api/keys/validate', data)) as unknown as ApiResponse<ValidateKeyResult>
```

### 3. 扩展类型定义

**文件**: `src/api/types.ts`

```typescript
// 更新 KeyInfo 类型
export interface KeyInfo {
    // ...existing fields...
    provider: string;        // NEW
    default_model: string;   // NEW
}
```

### 4. 改造 Add Key Modal

**文件**: `src/views/KeyManagerView.vue`

重构弹窗为多步骤向导 (使用 `n-steps` 组件):

#### Step 1: Provider & Key Input
```vue
<n-step title="Provider & Key">
  <n-select v-model:value="newKeyForm.provider" :options="providerOptions" placeholder="Select Provider" />
  <n-input v-model:value="newKeyForm.key" type="password" placeholder="Enter API Key" />
  <n-button @click="handleValidate" :loading="validating">Validate & Fetch Models</n-button>
</n-step>
```

#### Step 2: Model Selection (条件显示)
```vue
<n-step title="Select Model" v-if="validateResult?.valid">
  <n-select v-model:value="newKeyForm.defaultModel" :options="modelOptions" placeholder="Choose Default Model" />
</n-step>
```

#### Step 3: Key Details
```vue
<n-step title="Key Details">
  <n-input v-model:value="newKeyForm.name" placeholder="Key Name (e.g., Production)" />
  <n-input v-model:value="newKeyForm.tags" placeholder="Tags (comma separated)" />
</n-step>
```

#### Step 4: Summary & Create
```vue
<n-step title="Confirm">
  <div>Provider: {{ newKeyForm.provider }}</div>
  <div>Default Model: {{ newKeyForm.defaultModel }}</div>
  <div>Name: {{ newKeyForm.name }}</div>
  <n-button @click="handleAddKey" :loading="loading">Create Key</n-button>
</n-step>
```

### 5. 实现验证逻辑

```typescript
const validating = ref(false)
const validateResult = ref<ValidateKeyResult | null>(null)
const modelOptions = computed(() => 
  validateResult.value?.models.map(m => ({ label: m, value: m })) || []
)

async function handleValidate() {
    validating.value = true
    try {
        const res = await validateKey({ key: newKeyForm.value.key, provider: newKeyForm.value.provider })
        if (res.success && res.data) {
            validateResult.value = res.data
            if (res.data.valid) {
                message.success(`Found ${res.data.models.length} models`)
                currentStep.value++ // 进入下一步
            } else {
                message.error(res.data.error || 'Invalid API Key')
            }
        }
    } catch (e) {
        message.error('Validation failed')
    } finally {
        validating.value = false
    }
}
```

### 6. 更新创建 Key 请求

```typescript
async function handleAddKey() {
    const success = await store.createKey({
        key: newKeyForm.value.key,
        name: newKeyForm.value.name,
        provider: newKeyForm.value.provider || 'google_aistudio',
        default_model: newKeyForm.value.defaultModel,
        tags: newKeyForm.value.tags ? newKeyForm.value.tags.split(',').map(t => t.trim()) : []
    })
    // ...
}
```

---

## 产出

| 文件 | 变更类型 |
|------|----------|
| `src/api/keys.ts` | 修改 (新增 validateKey) |
| `src/api/types.ts` | 修改 (扩展 KeyInfo) |
| `src/views/KeyManagerView.vue` | 大改 (多步骤向导弹窗) |

---

## UI/UX 规范

- **主题**: 遵循现有 Claude 风格 (暖色 Light / 深炭色 Dark)
- **组件**: 使用 Naive UI 的 `n-steps`, `n-select`, `n-input`
- **交互**:
  - 验证成功后自动进入下一步
  - 验证失败显示错误提示，不切换步骤
  - 每一步都可返回上一步修改

---

## 验证

1. 运行 `npm run dev`
2. 打开 API Keys 页面，点击 "Add Key"
3. 测试流程:
   - 输入有效 Key → 验证 → 选择模型 → 填写名称 → 创建
   - 输入无效 Key → 验证失败 → 显示错误

---

## 约束

- 遵循 `FRONTEND_WORKFLOW.md` 代码规范
- 使用 TypeScript 严格类型，禁止 `any`
- 函数添加 JSDoc 注释
- 确保 Light/Dark 模式均正常显示

---

## 交接

- **前置**: 需等待后端 `TASK_BACKEND_KEY_VALIDATE.md` 完成
- **完成后**: 用户可在 UI 中添加带模型选择的 API Key
