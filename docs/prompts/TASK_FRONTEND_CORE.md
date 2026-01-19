# 任务：前端核心功能开发 (Key管理与设置)

## 角色
Senior Frontend Developer & UI Designer

## 必备技能
- **UI/UX Pro Max**: 必须阅读并应用 `.agent/skills/ui-ux-pro-max/SKILL.md`。

## 背景
前端基础架构已完成。请先阅读 **`docs/FRONTEND_PROJECT.md`** 了解当前项目结构和已有代码。
现在的任务是实现 Key 管理和系统设置页面。

## 任务目标
实现 `KeyManagerView.vue` 和 `SettingsView.vue` 页面，对接后端 API。

## 参考文档
- **项目全貌**: `docs/FRONTEND_PROJECT.md` ⬅️ **必读**
- **API 文档**: `docs/API.md` (Key 管理 /api/keys 、配置 /api/config)
- **工作流**: `docs/FRONTEND_WORKFLOW.md`

## 详细步骤

### 1. 界面设计与确认 (Design First) 🗣️
- **应用 Skill**: 使用 `ui-ux-pro-max` 的原则设计 Key 列表和设置表单。
- **生成预览**: 解析用户需求，必要时使用 `generate_image` 生成界面 Mockup。
- **沟通**: 向用户展示你的设计思路（由于是管理界面，确认表格交互、移动端卡片视图方案）。
- **获得批准**: 用户同意后方可编码。

### 2. API 层实现
在 `src/api/` 中创建以下文件：

#### `src/api/keys.ts`
```typescript
import { apiClient } from './client'
import type { KeyInfo, ApiResponse, ListResponse } from './types'

export const getKeys = () => apiClient.get<ListResponse<KeyInfo>>('/api/keys')
export const addKey = (data: { key: string; name?: string; tags?: string[] }) => apiClient.post<ApiResponse<KeyInfo>>('/api/keys', data)
export const deleteKey = (id: string) => apiClient.delete<ApiResponse<void>>(`/api/keys/${id}`)
export const testKey = (id: string) => apiClient.post<ApiResponse<{ valid: boolean; latency_ms: number }>>(`/api/keys/${id}/test`)
export const importKeys = (data: { keys: string; tag?: string }) => apiClient.post<ApiResponse<{ imported: number; skipped: number }>>('/api/keys/import', data)
export const exportKeys = () => apiClient.get('/api/keys/export', { responseType: 'blob' })
```

#### `src/api/config.ts`
```typescript
import { apiClient } from './client'
import type { ApiResponse } from './types'
// 补充 ConfigInfo 类型到 types.ts
export const getConfig = () => apiClient.get<ApiResponse<ConfigInfo>>('/api/config')
export const updateConfig = (data: Partial<ConfigInfo>) => apiClient.put<ApiResponse<ConfigInfo>>('/api/config', data)
export const checkUpdate = () => apiClient.get<ApiResponse<UpdateInfo>>('/api/update/check')
```

### 3. Key 管理页面 (`views/KeyManagerView.vue`)
- **Key 列表**:
  - 使用 `NDataTable` 展示 Key 信息（脱敏 Key、名称、状态、标签、使用统计）。
  - 支持分页（如果 API 支持）或前端分页。
- **操作栏**:
  - [添加 Key] 按钮 -> 弹出模态框 (Form: Key, Name, Tags)。
  - [批量导入] 按钮 -> 弹出模态框 (Textarea)。
  - [导出] 按钮 -> 调用导出 API 下载文件。
- **行操作**:
  - [测试] -> 调用 `/test` 接口，展示延迟和可用状态（Toast 提示）。
  - [删除] -> 二次确认对话框。
  - [复制] -> 复制完整 Key（如果前端持有）或脱敏 Key。

### 3. 设置页面 (`views/Settings.vue`)
- **配置表单**:
  - 端口/Host (只读展示)。
  - **Key 池策略**: 下拉选择 (RoundRobin, Random, LeastUsed, Weighted)。
  - **日志级别**: 下拉选择。
  - **更新检测**: 开启/关闭开关。
- [保存修改] 按钮 -> 调用 PUT `/api/config`。
- **更新检查卡片**:
  - 展示当前版本。
  - [检查更新] 按钮 -> 展示最新版本和下载链接。

### 4. 状态管理
- 在 `useKeyStore` (Pinia) 中管理 Key 列表数据，避免页面反复刷新。
- 实现 Key 列表的自动/手动刷新机制。

---

## 产出物
- 完善的 Key 管理和设置页面。
- 通过 API 真实交互的功能。

## 约束
- 遵循 Naive UI 设计规范。
- **移动端适配**: 表格在小屏下应改为卡片列表视图或可横向滚动。
- 错误处理：API 失败时必须弹出清晰的 `message.error()`。
