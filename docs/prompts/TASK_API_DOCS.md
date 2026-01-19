# 任务：编写 API 使用文档

## 角色
Technical Writer / Architect

## 背景
MxlnAPI 项目后端已完成（阶段 1-3.5），但目前缺少独立的 API 使用文档。虽然 `ARCHITECTURE.md` 包含 API 契约定义，但需要创建面向用户和前端开发者的完整 API 文档。

## 任务目标

创建 `docs/API.md`，包含所有后端 API 端点的完整文档，便于：
1. 前端开发者对接 API
2. 用户了解如何使用代理服务
3. 第三方集成参考

---

## 文档结构要求

### 1. 概述
- API 基础 URL
- 认证方式（无认证 / 本地使用）
- 响应格式规范
- 错误处理规范

### 2. OpenAI 兼容端点
需要参考 `internal/api/openai_handler.go` 实际实现：

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/chat/completions` | POST | 对话补全（支持流式） |
| `/v1/models` | GET | 获取可用模型列表 |
| `/health` | GET | 健康检查 |

### 3. Key 管理 API
需要参考 `internal/api/admin_handler.go` 实际实现：

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/keys` | GET | 获取 Key 列表 |
| `/api/keys` | POST | 添加 Key |
| `/api/keys/:id` | DELETE | 删除 Key |
| `/api/keys/:id/test` | POST | 测试 Key |
| `/api/keys/import` | POST | 批量导入 |
| `/api/keys/export` | GET | 导出 Key |

### 4. 会话管理 API（阶段 3.5 新增）
需要参考 `internal/api/session_handler.go` 实际实现：

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/sessions` | GET | 获取会话列表（支持分页） |
| `/api/sessions` | POST | 创建新会话 |
| `/api/sessions/:id` | GET | 获取会话详情（含消息） |
| `/api/sessions/:id` | PUT | 更新会话 |
| `/api/sessions/:id` | DELETE | 删除会话 |
| `/api/sessions/:id/messages` | POST | 添加消息 |

### 5. 统计 API
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/stats` | GET | 总体统计 |
| `/api/stats/keys` | GET | 各 Key 统计 |

### 6. 配置 API
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/config` | GET | 获取配置 |
| `/api/config` | PUT | 更新配置 |

### 7. 更新检测 API
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/update/check` | GET | 检查更新 |

---

## 步骤

1. **阅读现有代码**
   - `internal/api/router.go` - 所有路由定义
   - `internal/api/openai_handler.go` - OpenAI 兼容端点
   - `internal/api/admin_handler.go` - 管理 API
   - `internal/api/session_handler.go` - 会话 API
   - `internal/api/response.go` - 响应格式
   - `internal/types/` - 所有类型定义

2. **参考 ARCHITECTURE.md**
   - 已有的 API 契约定义可作为基础
   - 需核对实际实现是否一致

3. **编写 API 文档**
   - 使用 Markdown 格式
   - 每个端点包含：请求方法、URL、请求体、响应体、示例
   - 使用代码块展示 JSON 示例

4. **添加实用示例**
   - cURL 命令示例
   - 常见使用场景

---

## 文档格式要求

每个端点应包含：

```markdown
### `METHOD /path`

**描述**: 简要说明

**请求参数**:
| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| param | string | 是 | 参数说明 |

**请求体**:
```json
{
  "field": "value"
}
```

**响应体**:
```json
{
  "success": true,
  "data": {}
}
```

**示例**:
```bash
curl -X METHOD http://localhost:8080/path \
  -H "Content-Type: application/json" \
  -d '{"field": "value"}'
```
```

---

## 产出

1. `docs/API.md` - 完整的 API 使用文档

---

## 约束

- 所有请求/响应示例必须与实际代码一致
- 使用简洁清晰的语言
- 提供可直接运行的 cURL 示例
- 错误码需与 `internal/types/errors.go` 一致

---

## 验收标准

1. 文档覆盖所有 API 端点
2. 请求/响应格式与代码实现一致
3. 包含可运行的示例
4. 格式清晰、易于阅读

---

*任务创建时间: 2026-01-15*
