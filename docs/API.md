# MuxueTools - API 使用文档

> **版本**: 1.1  
> **最后更新**: 2026-01-19

---

## 目录

- [概述](#概述)
- [OpenAI 兼容端点](#openai-兼容端点)
- [Key 管理 API](#key-管理-api)
- [会话管理 API](#会话管理-api)
- [统计 API](#统计-api)
- [配置 API](#配置-api)
- [数据管理 API](#数据管理-api)
- [更新检测 API](#更新检测-api)
- [使用示例](#使用示例)

---

## 概述

### 基础信息

- **基础 URL**: `http://localhost:8080` (默认配置，可通过 `config.yaml` 修改)
- **认证方式**: 本地部署无需认证，远程部署建议通过反向代理添加认证
- **响应格式**: JSON
- **字符编码**: UTF-8

### 响应格式规范

#### 成功响应

管理 API 统一使用以下格式：

```json
{
  "success": true,
  "data": { /* 具体数据 */ }
}
```

OpenAI 兼容端点直接返回 OpenAI 格式的数据，不额外包装。

#### 错误响应

所有错误响应遵循统一格式：

```json
{
  "error": {
    "code": 40001,
    "message": "Invalid request format",
    "type": "invalid_request_error",
    "param": "model"
  }
}
```

### 错误码说明

| 错误码 | HTTP 状态 | 类型 | 描述 |
|--------|----------|------|------|
| 40001 | 400 | `invalid_request_error` | 请求格式错误 |
| 40002 | 400 | `invalid_request_error` | 不支持的模型 |
| 40003 | 400 | `invalid_request_error` | 消息格式错误 |
| 40101 | 401 | `authentication_error` | API 密钥无效 |
| 40301 | 403 | `permission_error` | 访问被拒绝 |
| 40401 | 404 | `not_found_error` | 资源不存在 |
| 42901 | 429 | `rate_limit_error` | 所有密钥均达到速率限制 |
| 50001 | 500 | `server_error` | 服务器内部错误 |
| 50201 | 502 | `upstream_error` | 上游 API 错误 |
| 50301 | 503 | `service_unavailable` | 服务暂时不可用 |

---

## OpenAI 兼容端点

### `POST /v1/chat/completions`

**描述**: 创建对话补全，支持流式和非流式响应。兼容 OpenAI Chat Completions API。

**请求头**:
```
Content-Type: application/json
```

**请求体**:

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `model` | string | 是 | 模型名称，如 `gpt-4`、`gemini-pro` 等 |
| `messages` | array | 是 | 对话消息数组 |
| `temperature` | number | 否 | 温度参数 (0-2)，默认 1.0 |
| `top_p` | number | 否 | 核采样参数 (0-1) |
| `max_tokens` | integer | 否 | 最大生成 token 数 |
| `stream` | boolean | 否 | 是否使用流式响应，默认 false |
| `stop` | string/array | 否 | 停止序列 |
| `presence_penalty` | number | 否 | 存在惩罚 (-2.0 到 2.0) |
| `frequency_penalty` | number | 否 | 频率惩罚 (-2.0 到 2.0) |
| `n` | integer | 否 | 生成响应数量，默认 1 |
| `user` | string | 否 | 用户标识 |

**消息格式**:

```json
{
  "role": "user|assistant|system",
  "content": "文本内容"
}
```

或多模态格式：

```json
{
  "role": "user",
  "content": [
    {
      "type": "text",
      "text": "这是什么？"
    },
    {
      "type": "image_url",
      "image_url": {
        "url": "data:image/jpeg;base64,/9j/4AAQ...",
        "detail": "high"
      }
    }
  ]
}
```

**非流式响应**:

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gemini-1.5-pro-latest",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "您好！有什么我可以帮助您的吗？"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 12,
    "total_tokens": 21
  }
}
```

**流式响应** (SSE):

每个事件格式为：

```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gemini-1.5-pro-latest","choices":[{"index":0,"delta":{"role":"assistant","content":"您"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gemini-1.5-pro-latest","choices":[{"index":0,"delta":{"content":"好"},"finish_reason":null}]}

data: [DONE]
```

**示例 - 非流式请求**:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "system", "content": "你是一个有帮助的助手。"},
      {"role": "user", "content": "你好！"}
    ],
    "temperature": 0.7
  }'
```

**示例 - 流式请求**:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-pro",
    "messages": [
      {"role": "user", "content": "讲一个笑话"}
    ],
    "stream": true
  }'
```

**示例 - 多模态请求**:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-vision-preview",
    "messages": [
      {
        "role": "user",
        "content": [
          {"type": "text", "text": "这张图片里有什么？"},
          {
            "type": "image_url",
            "image_url": {
              "url": "https://example.com/image.jpg"
            }
          }
        ]
      }
    ]
  }'
```

---

### `GET /v1/models`

**描述**: 获取可用模型列表。

**响应体**:

```json
{
  "object": "list",
  "data": [
    {
      "id": "gemini-1.5-pro-latest",
      "object": "model",
      "created": 1677610602,
      "owned_by": "google"
    },
    {
      "id": "gemini-1.5-flash-latest",
      "object": "model",
      "created": 1677610602,
      "owned_by": "google"
    },
    {
      "id": "gpt-4",
      "object": "model",
      "created": 1677610602,
      "owned_by": "google"
    }
  ]
}
```

**示例**:

```bash
curl http://localhost:8080/v1/models
```

---

### `GET /health`

**描述**: 健康检查端点，返回服务状态和 Key 池统计信息。

**响应体**:

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": 3600,
  "keys": {
    "total": 5,
    "active": 4,
    "rate_limited": 1,
    "disabled": 0
  }
}
```

**字段说明**:

- `status`: `ok` 或 `degraded`
- `uptime`: 服务运行时间（秒）
- `keys.total`: 总密钥数
- `keys.active`: 可用密钥数
- `keys.rate_limited`: 冷却中的密钥数
- `keys.disabled`: 禁用的密钥数

**示例**:

```bash
curl http://localhost:8080/health
```

---

## Key 管理 API

### `GET /api/keys`

**描述**: 获取所有 API 密钥列表（脱敏显示）。

**响应体**:

```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "key": "AIzaSy...xyz",
      "name": "生产环境密钥",
      "status": "active",
      "enabled": true,
      "tags": ["production", "high-priority"],
      "stats": {
        "request_count": 1250,
        "success_count": 1200,
        "error_count": 50,
        "prompt_tokens": 45000,
        "completion_tokens": 30000,
        "last_used_at": "2026-01-15T10:30:00Z"
      },
      "cooldown_until": null,
      "created_at": "2026-01-10T08:00:00Z",
      "updated_at": "2026-01-15T10:30:00Z"
    }
  ],
  "total": 5
}
```

**字段说明**:

- `status`: `active` | `rate_limited` | `disabled`
- `key`: 脱敏的 API 密钥（格式：`前6位...后3位`）
- `stats`: 使用统计（仅内存状态，重启后重置）

**示例**:

```bash
curl http://localhost:8080/api/keys
```

---

### `POST /api/keys`

**描述**: 添加单个 API 密钥。

**请求体**:

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `key` | string | 是 | Gemini API 密钥 |
| `name` | string | 否 | 密钥名称（用于标识） |
| `tags` | array | 否 | 标签数组 |
| `provider` | string | 否 | 供应商标识，默认 `google_aistudio` |
| `default_model` | string | 否 | 默认模型名称 |

```json
{
  "key": "AIzaSyABC123XYZ...",
  "name": "开发环境密钥",
  "tags": ["dev", "test"],
  "provider": "google_aistudio",
  "default_model": "gemini-1.5-pro-latest"
}
```

**响应体**:

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "key": "AIzaSy...XYZ",
    "name": "开发环境密钥",
    "status": "active",
    "enabled": true,
    "tags": ["dev", "test"],
    "provider": "google_aistudio",
    "default_model": "gemini-1.5-pro-latest",
    "stats": {
      "request_count": 0,
      "success_count": 0,
      "error_count": 0,
      "prompt_tokens": 0,
      "completion_tokens": 0,
      "last_used_at": null
    },
    "cooldown_until": null,
    "created_at": "2026-01-15T12:00:00Z",
    "updated_at": "2026-01-15T12:00:00Z"
  }
}
```

**示例**:

```bash
curl -X POST http://localhost:8080/api/keys \
  -H "Content-Type: application/json" \
  -d '{
    "key": "AIzaSyABC123XYZ...",
    "name": "开发环境密钥",
    "tags": ["dev"]
  }'
```

---

### `POST /api/keys/validate`

**描述**: 验证 API 密钥有效性并获取可用模型列表。用于在添加密钥前验证其有效性。

**请求体**:

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `key` | string | 是 | 待验证的 Gemini API 密钥 |
| `provider` | string | 否 | 供应商标识，默认 `google_aistudio` |

```json
{
  "key": "AIzaSyABC123XYZ...",
  "provider": "google_aistudio"
}
```

**成功响应** (有效密钥):

```json
{
  "success": true,
  "data": {
    "valid": true,
    "latency_ms": 245,
    "models": [
      "gemini-1.5-pro-latest",
      "gemini-1.5-flash-latest",
      "gemini-2.0-flash-exp",
      "gemini-2.0-flash-thinking-exp"
    ]
  }
}
```

**失败响应** (无效密钥):

```json
{
  "success": true,
  "data": {
    "valid": false,
    "latency_ms": 120,
    "models": null,
    "error": "API key not valid. Please pass a valid API key."
  }
}
```

**字段说明**:

- `valid`: 密钥是否有效
- `latency_ms`: API 延迟（毫秒）
- `models`: 可用模型列表（无效时为空）
- `error`: 错误消息（有效时为空）

**示例**:

```bash
# 验证有效密钥
curl -X POST http://localhost:8080/api/keys/validate \
  -H "Content-Type: application/json" \
  -d '{"key": "AIzaSyABC123XYZ..."}'

# 验证无效密钥
curl -X POST http://localhost:8080/api/keys/validate \
  -H "Content-Type: application/json" \
  -d '{"key": "invalid-key-here"}'
```

---

### `GET /api/models`

**描述**: 获取当前可用的 Gemini 模型列表。使用 Key Pool 中的有效 Key 调用 Gemini API 实时获取。

**响应体**:

```json
{
  "success": true,
  "data": [
    "gemini-2.0-flash",
    "gemini-2.0-flash-lite",
    "gemini-1.5-pro",
    "gemini-1.5-flash",
    "gemini-1.5-flash-8b",
    "gemini-1.0-pro"
  ]
}
```

**字段说明**:

- `data`: 支持 `generateContent` 方法的模型名称列表

**错误处理**:

- 如果没有可用的 API Key，返回空数组
- 如果 Gemini API 调用失败，返回空数组

**示例**:

```bash
curl http://localhost:8080/api/models
```

---

### `DELETE /api/keys/:id`

**描述**: 删除指定的 API 密钥。

**路径参数**:

- `id`: 密钥 ID（UUID）

**响应体**:

```json
{
  "success": true,
  "message": "Key deleted successfully"
}
```

**错误响应**:

```json
{
  "error": {
    "code": 40401,
    "message": "Resource not found: Key",
    "type": "not_found_error"
  }
}
```

**示例**:

```bash
curl -X DELETE http://localhost:8080/api/keys/550e8400-e29b-41d4-a716-446655440000
```

---

### `POST /api/keys/:id/test`

**描述**: 测试指定密钥的有效性和延迟。

**路径参数**:

- `id`: 密钥 ID（UUID）

**响应体**:

```json
{
  "success": true,
  "data": {
    "valid": true,
    "latency_ms": 245,
    "models": [
      "gemini-1.5-pro-latest",
      "gemini-1.5-flash-latest",
      "gemini-2.0-flash"
    ]
  }
}
```

**失败响应**:

```json
{
  "success": true,
  "data": {
    "valid": false,
    "latency_ms": 0,
    "error": "Invalid API key"
  }
}
```

**示例**:

```bash
curl -X POST http://localhost:8080/api/keys/550e8400-e29b-41d4-a716-446655440000/test
```

---

### `POST /api/keys/import`

**描述**: 批量导入 API 密钥（换行分隔）。

**请求体**:

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `keys` | string | 是 | 换行分隔的密钥列表 |
| `tag` | string | 否 | 通用标签（应用于所有导入的密钥） |

```json
{
  "keys": "AIzaSyABC123...\nAIzaSyDEF456...\nAIzaSyGHI789...",
  "tag": "batch-import-2026-01"
}
```

**响应体**:

```json
{
  "success": true,
  "data": {
    "imported": 2,
    "skipped": 1,
    "errors": [
      "Line 2: Invalid API key format"
    ]
  }
}
```

**字段说明**:

- `imported`: 成功导入数量
- `skipped`: 跳过数量（重复密钥）
- `errors`: 错误消息数组

**示例**:

```bash
curl -X POST http://localhost:8080/api/keys/import \
  -H "Content-Type: application/json" \
  -d '{
    "keys": "AIzaSyABC123...\nAIzaSyDEF456...",
    "tag": "production"
  }'
```

---

### `GET /api/keys/export`

**描述**: 导出所有密钥为文本格式（每行一个密钥，完整密钥未脱敏）。

**响应头**:
```
Content-Type: text/plain
Content-Disposition: attachment; filename="keys-export-20260115.txt"
```

**响应体**:
```
AIzaSyABC123XYZ...
AIzaSyDEF456ABC...
AIzaSyGHI789DEF...
```

**示例**:

```bash
curl http://localhost:8080/api/keys/export -o keys-export.txt
```

---

## 会话管理 API

> **注意**: 会话管理功能需要配置数据库路径（`database.path`），默认启用。

### `GET /api/sessions`

**描述**: 获取会话列表，支持分页。

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| `limit` | integer | 否 | 20 | 每页数量（最大 100） |
| `offset` | integer | 否 | 0 | 偏移量 |

**响应体**:

```json
{
  "success": true,
  "sessions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "关于 AI 的讨论",
      "model": "gemini-1.5-pro-latest",
      "message_count": 12,
      "total_tokens": 5400,
      "created_at": "2026-01-15T09:00:00Z",
      "updated_at": "2026-01-15T10:30:00Z"
    }
  ],
  "total": 25
}
```

**示例**:

```bash
# 获取前 10 个会话
curl "http://localhost:8080/api/sessions?limit=10&offset=0"
```

---

### `POST /api/sessions`

**描述**: 创建新会话。

**请求体**:

| 参数 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| `title` | string | 否 | "New Chat" | 会话标题 |
| `model` | string | 否 | "gemini-1.5-flash" | 模型名称 |

```json
{
  "title": "代码审查助手",
  "model": "gpt-4"
}
```

**响应体**:

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "代码审查助手",
    "model": "gemini-1.5-pro-latest",
    "message_count": 0,
    "total_tokens": 0,
    "created_at": "2026-01-15T12:00:00Z",
    "updated_at": "2026-01-15T12:00:00Z"
  }
}
```

**示例**:

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Python 编程助手",
    "model": "gpt-4"
  }'
```

---

### `GET /api/sessions/:id`

**描述**: 获取会话详情和所有消息。

**路径参数**:

- `id`: 会话 ID（UUID）

**响应体**:

```json
{
  "success": true,
  "session": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "代码审查助手",
    "model": "gemini-1.5-pro-latest",
    "message_count": 4,
    "total_tokens": 1250,
    "created_at": "2026-01-15T09:00:00Z",
    "updated_at": "2026-01-15T10:30:00Z"
  },
  "messages": [
    {
      "id": "msg-001",
      "session_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "user",
      "content": "请帮我审查这段代码...",
      "prompt_tokens": 150,
      "completion_tokens": 0,
      "created_at": "2026-01-15T09:01:00Z"
    },
    {
      "id": "msg-002",
      "session_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "assistant",
      "content": "这段代码有以下问题...",
      "prompt_tokens": 0,
      "completion_tokens": 200,
      "created_at": "2026-01-15T09:01:05Z"
    }
  ]
}
```

**示例**:

```bash
curl http://localhost:8080/api/sessions/550e8400-e29b-41d4-a716-446655440000
```

---

### `PUT /api/sessions/:id`

**描述**: 更新会话信息（标题或模型）。

**路径参数**:

- `id`: 会话 ID（UUID）

**请求体**:

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `title` | string | 否 | 新标题 |
| `model` | string | 否 | 新模型 |

```json
{
  "title": "重命名的会话"
}
```

**响应体**:

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "重命名的会话",
    "model": "gemini-1.5-pro-latest",
    "message_count": 4,
    "total_tokens": 1250,
    "created_at": "2026-01-15T09:00:00Z",
    "updated_at": "2026-01-15T12:05:00Z"
  }
}
```

**示例**:

```bash
curl -X PUT http://localhost:8080/api/sessions/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"title": "重命名的会话"}'
```

---

### `DELETE /api/sessions/:id`

**描述**: 删除会话及其所有消息。

**路径参数**:

- `id`: 会话 ID（UUID）

**响应体**:

```json
{
  "success": true,
  "message": "Session deleted successfully"
}
```

**示例**:

```bash
curl -X DELETE http://localhost:8080/api/sessions/550e8400-e29b-41d4-a716-446655440000
```

---

### `POST /api/sessions/:id/messages`

**描述**: 向会话添加消息（手动记录对话历史）。

**路径参数**:

- `id`: 会话 ID（UUID）

**请求体**:

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `role` | string | 是 | 消息角色：`user` | `assistant` | `system` |
| `content` | string | 是 | 消息内容 |
| `prompt_tokens` | integer | 否 | 提示词 token 数 |
| `completion_tokens` | integer | 否 | 补全 token 数 |

```json
{
  "role": "user",
  "content": "什么是量子计算？",
  "prompt_tokens": 15
}
```

**响应体**:

```json
{
  "success": true,
  "data": {
    "id": "msg-003",
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "role": "user",
    "content": "什么是量子计算？",
    "prompt_tokens": 15,
    "completion_tokens": 0,
    "created_at": "2026-01-15T12:10:00Z"
  }
}
```

**示例**:

```bash
curl -X POST http://localhost:8080/api/sessions/550e8400-e29b-41d4-a716-446655440000/messages \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "解释一下机器学习",
    "prompt_tokens": 20
  }'
```

---

## 统计 API

### `GET /api/stats`

**描述**: 获取总体使用统计。

**响应体**:

```json
{
  "success": true,
  "data": {
    "period": {
      "start": "2026-01-15T00:00:00Z",
      "end": "2026-01-15T12:30:00Z"
    },
    "requests": {
      "total": 1500,
      "success": 1450,
      "error": 40,
      "rate_limited": 10
    },
    "tokens": {
      "prompt": 125000,
      "completion": 87000,
      "total": 212000
    },
    "avg_latency_ms": 320.5
  }
}
```

**字段说明**:

- `period`: 统计时间范围（当前会话启动至今）
- `requests`: 请求统计
- `tokens`: Token 消耗统计
- `avg_latency_ms`: 平均响应延迟（毫秒）

**示例**:

```bash
curl http://localhost:8080/api/stats
```

---

### `GET /api/stats/keys`

**描述**: 获取各 API 密钥的详细统计。

**响应体**:

```json
{
  "success": true,
  "data": [
    {
      "key_id": "550e8400-e29b-41d4-a716-446655440000",
      "key_name": "生产环境密钥",
      "request_count": 750,
      "success_rate": 96.5,
      "token_usage": 105000,
      "avg_latency_ms": 315.2
    },
    {
      "key_id": "550e8400-e29b-41d4-a716-446655440001",
      "key_name": "开发环境密钥",
      "request_count": 500,
      "success_rate": 98.2,
      "token_usage": 68000,
      "avg_latency_ms": 298.7
    }
  ]
}
```

**字段说明**:

- `success_rate`: 成功率百分比 (0-100)
- `token_usage`: 总 token 消耗（prompt + completion）

**示例**:

```bash
curl http://localhost:8080/api/stats/keys
```

---

### `GET /api/stats/trend`

**描述**: 获取请求趋势数据，用于生成折线图。

**查询参数**:

| 参数 | 类型 | 必填 | 默认值 | 描述 |
|------|------|------|--------|------|
| `range` | string | 否 | `7d` | 时间范围：`24h` \| `7d` \| `30d` |

**响应体**:

```json
{
  "success": true,
  "data": [
    {
      "timestamp": "2026-01-12T00:00:00Z",
      "requests": 0,
      "tokens": 0,
      "errors": 0
    },
    {
      "timestamp": "2026-01-13T00:00:00Z",
      "requests": 0,
      "tokens": 0,
      "errors": 0
    },
    {
      "timestamp": "2026-01-18T00:00:00Z",
      "requests": 150,
      "tokens": 25000,
      "errors": 3
    }
  ],
  "time_range": "7d"
}
```

**字段说明**:

- `data`: 时间序列数据点数组
  - `timestamp`: 数据点时间戳
  - `requests`: 该时段的请求数
  - `tokens`: 该时段的 token 消耗
  - `errors`: 该时段的错误数
- `time_range`: 当前查询的时间范围

**时间范围对应数据点数**:

| 范围 | 数据点数 | 粒度 |
|------|----------|------|
| `24h` | 24 | 每小时 |
| `7d` | 7 | 每天 |
| `30d` | 30 | 每天 |

**示例**:

```bash
# 获取过去 7 天趋势
curl http://localhost:8080/api/stats/trend?range=7d

# 获取过去 24 小时趋势
curl http://localhost:8080/api/stats/trend?range=24h
```

---

### `GET /api/stats/models`

**描述**: 获取模型使用分布，用于生成饼图。

**响应体**:

```json
{
  "success": true,
  "data": [
    {
      "model": "gemini-1.5-pro-latest",
      "request_count": 850,
      "token_usage": 125000,
      "percentage": 56.67
    },
    {
      "model": "gemini-1.5-flash-latest",
      "request_count": 500,
      "token_usage": 45000,
      "percentage": 33.33
    },
    {
      "model": "unknown",
      "request_count": 150,
      "token_usage": 15000,
      "percentage": 10.0
    }
  ]
}
```

**字段说明**:

- `data`: 模型使用统计数组（按请求数降序排列）
  - `model`: 模型名称（Key 的 `default_model` 字段，若为空则显示 `unknown`）
  - `request_count`: 该模型的请求数
  - `token_usage`: 该模型的 token 消耗
  - `percentage`: 该模型的请求占比 (0-100)

**示例**:

```bash
curl http://localhost:8080/api/stats/models
```

---

## 配置 API

### `GET /api/config`

**描述**: 获取当前配置（脱敏后的配置，不包含 API 密钥）。配置值优先从 SQLite 读取，若无则使用 config.yaml 默认值。

**响应体**:

```json
{
  "success": true,
  "data": {
    "server": {
      "port": 8080,
      "host": "0.0.0.0"
    },
    "pool": {
      "strategy": "round_robin",
      "cooldown_seconds": 3600,
      "max_retries": 3
    },
    "logging": {
      "level": "info"
    },
    "update": {
      "enabled": true,
      "check_interval": "24h",
      "source": "mxln"
    },
    "security": {
      "ip_whitelist_enabled": false,
      "whitelist_ip": "",
      "proxy_key": "sk-mxln-proxy-local"
    },
    "advanced": {
      "request_timeout": 120
    }
  }
}
```

**字段说明**:

| 字段 | 类型 | 描述 |
|------|------|------|
| `pool.strategy` | string | 密钥选择策略 |
| `pool.cooldown_seconds` | int | Rate Limit 冷却时间（秒） |
| `pool.max_retries` | int | 请求失败重试次数 |
| `update.source` | string | 更新检查源：`mxln` 或 `github` |
| `security.ip_whitelist_enabled` | bool | 是否启用 IP 白名单 |
| `security.whitelist_ip` | string | 白名单 IP 地址 |
| `security.proxy_key` | string | 代理访问密钥 |
| `advanced.request_timeout` | int | HTTP 请求超时时间（秒） |
| `model_settings.system_prompt` | string | 全局系统提示词 |
| `model_settings.temperature` | float | 温度参数 (0-2) |
| `model_settings.max_output_tokens` | int | 最大输出 token 数 |
| `model_settings.top_p` | float | Top-P 采样参数 (0-1) |
| `model_settings.top_k` | int | Top-K 采样参数 (1-100) |
| `model_settings.thinking_level` | string | 思考等级: `LOW`\|`MEDIUM`\|`HIGH` |
| `model_settings.media_resolution` | string | 媒体分辨率 |

---

### `PUT /api/config`

**描述**: 更新配置。配置保存到 SQLite 并支持热更新（无需重启）。

**请求体**:

```json
{
  "pool": {
    "strategy": "weighted",
    "cooldown_seconds": 1800,
    "max_retries": 5
  },
  "logging": {
    "level": "debug"
  },
  "update": {
    "enabled": true,
    "source": "github"
  },
  "security": {
    "ip_whitelist_enabled": true,
    "whitelist_ip": "192.168.1.100",
    "proxy_key": "sk-mxln-custom-key"
  },
  "advanced": {
    "request_timeout": 180
  }
}
```

**参数说明**:

| 参数 | 类型 | 范围/选项 | 描述 |
|------|------|----------|------|
| `pool.strategy` | string | `round_robin` \| `random` \| `least_used` \| `weighted` | 选择策略 |
| `pool.cooldown_seconds` | int | ≥ 0 | 冷却时间 |
| `pool.max_retries` | int | ≥ 0 | 重试次数 |
| `logging.level` | string | `debug` \| `info` \| `warn` \| `error` | 日志级别 |
| `update.source` | string | `mxln` \| `github` | 更新源 |
| `security.ip_whitelist_enabled` | bool | - | 启用白名单 |
| `security.whitelist_ip` | string | - | 白名单 IP |
| `advanced.request_timeout` | int | 30-600 | 超时时间 |
| `model_settings.system_prompt` | string | - | 系统提示词 |
| `model_settings.temperature` | float | 0-2 | 温度参数 |
| `model_settings.max_output_tokens` | int | 1-65536 | 最大输出 token |
| `model_settings.top_p` | float | 0-1 | Top-P |
| `model_settings.top_k` | int | 1-100 | Top-K |
| `model_settings.thinking_level` | string | `LOW`\|`MEDIUM`\|`HIGH` | 思考等级 |
| `model_settings.media_resolution` | string | `MEDIA_RESOLUTION_LOW`\|`MEDIUM`\|`HIGH` | 媒体分辨率 |

**响应体**:

```json
{
  "success": true,
  "message": "Configuration updated successfully",
  "data": {
    "updated": {
      "pool.strategy": "weighted",
      "pool.cooldown_seconds": 1800,
      "logging.level": "debug"
    }
  }
}
```

---

### `POST /api/config/regenerate-proxy-key`

**描述**: 重新生成代理访问密钥。

**响应体**:

```json
{
  "success": true,
  "data": {
    "proxy_key": "sk-mxln-a1b2c3d4e5f6g7h8"
  }
}
```

**示例**:

```bash
curl -X POST http://localhost:8080/api/config/regenerate-proxy-key
```

---

## 数据管理 API

### `DELETE /api/sessions`

**描述**: 清空所有聊天会话和消息记录。**此操作不可恢复！**

**响应体**:

```json
{
  "success": true,
  "message": "All sessions deleted successfully",
  "data": {
    "deleted": 15
  }
}
```

**字段说明**:

| 字段 | 描述 |
|------|------|
| `deleted` | 删除的会话数量 |

**示例**:

```bash
curl -X DELETE http://localhost:8080/api/sessions
```

---

### `DELETE /api/stats/reset`

**描述**: 重置所有 API 密钥的统计数据。**此操作不可恢复！**

**响应体**:

```json
{
  "success": true,
  "message": "All key statistics have been reset",
  "data": {
    "keys_affected": 5
  }
}
```

**字段说明**:

| 字段 | 描述 |
|------|------|
| `keys_affected` | 受影响的密钥数量 |

**示例**:

```bash
curl -X DELETE http://localhost:8080/api/stats/reset
```

---

## 更新检测 API

### `GET /api/update/check`

**描述**: 检查 GitHub 最新版本更新。

**响应体**:

```json
{
  "success": true,
  "data": {
    "current_version": "1.0.0",
    "latest_version": "1.1.0",
    "has_update": true,
    "download_url": "https://github.com/muxueliunian/MuxueTools/releases/tag/v1.1.0",
    "changelog": "## What's Changed\n- 新增会话管理功能\n- 优化 Key 池性能",
    "published_at": "2026-01-14T10:00:00Z"
  }
}
```

**字段说明**:

- `has_update`: 是否有更新
- `changelog`: Markdown 格式的更新日志
- `published_at`: 发布时间

**无更新时**:

```json
{
  "success": true,
  "data": {
    "current_version": "1.0.0",
    "latest_version": "1.0.0",
    "has_update": false,
    "download_url": "",
    "changelog": "",
    "published_at": ""
  }
}
```

**示例**:

```bash
curl http://localhost:8080/api/update/check
```

---

## 使用示例

### 场景一：基础对话流程

1. **创建会话**

```bash
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"title": "编程助手", "model": "gpt-4"}' | jq -r '.data.id')

echo "会话 ID: $SESSION_ID"
```

2. **发送对话请求**

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"gpt-4\",
    \"messages\": [
      {\"role\": \"user\", \"content\": \"如何在 Python 中读取 JSON 文件？\"}
    ]
  }"
```

3. **保存消息到会话**

```bash
curl -X POST "http://localhost:8080/api/sessions/$SESSION_ID/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "如何在 Python 中读取 JSON 文件？",
    "prompt_tokens": 25
  }'
```

---

### 场景二：Key 管理流程

1. **批量导入密钥**

```bash
curl -X POST http://localhost:8080/api/keys/import \
  -H "Content-Type: application/json" \
  -d '{
    "keys": "AIzaSyABC123...\nAIzaSyDEF456...\nAIzaSyGHI789...",
    "tag": "production"
  }'
```

2. **测试所有密钥**

```bash
# 获取所有密钥 ID
KEY_IDS=$(curl -s http://localhost:8080/api/keys | jq -r '.data[].id')

# 逐个测试
for id in $KEY_IDS; do
  echo "测试密钥: $id"
  curl -s -X POST "http://localhost:8080/api/keys/$id/test" | jq '.data'
done
```

3. **查看统计**

```bash
curl http://localhost:8080/api/stats/keys | jq
```

---

### 场景三：流式响应处理

**JavaScript 示例**:

```javascript
const response = await fetch('http://localhost:8080/v1/chat/completions', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    model: 'gpt-4',
    messages: [
      { role: 'user', content: '写一首诗' }
    ],
    stream: true
  })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
  const { value, done } = await reader.read();
  if (done) break;
  
  const chunk = decoder.decode(value);
  const lines = chunk.split('\n').filter(line => line.trim() !== '');
  
  for (const line of lines) {
    if (line.startsWith('data: ')) {
      const data = line.slice(6);
      if (data === '[DONE]') {
        console.log('流结束');
        break;
      }
      
      try {
        const json = JSON.parse(data);
        const content = json.choices[0]?.delta?.content;
        if (content) {
          process.stdout.write(content);
        }
      } catch (e) {
        console.error('解析错误:', e);
      }
    }
  }
}
```

**Python 示例**:

```python
import requests
import json

url = 'http://localhost:8080/v1/chat/completions'
data = {
    'model': 'gpt-4',
    'messages': [
        {'role': 'user', 'content': '写一首诗'}
    ],
    'stream': True
}

response = requests.post(url, json=data, stream=True)

for line in response.iter_lines():
    if line:
        line = line.decode('utf-8')
        if line.startswith('data: '):
            data = line[6:]
            if data == '[DONE]':
                print('\n流结束')
                break
            
            try:
                chunk = json.loads(data)
                content = chunk['choices'][0]['delta'].get('content', '')
                if content:
                    print(content, end='', flush=True)
            except json.JSONDecodeError:
                pass
```

---

### 场景四：多模态图片识别

```bash
# Base64 编码图片
IMAGE_BASE64=$(base64 -w 0 image.jpg)

curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"gpt-4-vision-preview\",
    \"messages\": [
      {
        \"role\": \"user\",
        \"content\": [
          {\"type\": \"text\", \"text\": \"这张图片里有什么？请详细描述。\"},
          {
            \"type\": \"image_url\",
            \"image_url\": {
              \"url\": \"data:image/jpeg;base64,$IMAGE_BASE64\"
            }
          }
        ]
      }
    ]
  }"
```

---

### 场景五：监控和健康检查

**简单健康检查脚本**:

```bash
#!/bin/bash

while true; do
  HEALTH=$(curl -s http://localhost:8080/health)
  STATUS=$(echo $HEALTH | jq -r '.status')
  ACTIVE=$(echo $HEALTH | jq -r '.keys.active')
  TOTAL=$(echo $HEALTH | jq -r '.keys.total')
  
  echo "[$(date)] 状态: $STATUS, 可用密钥: $ACTIVE/$TOTAL"
  
  if [ "$STATUS" != "ok" ] || [ "$ACTIVE" -lt 1 ]; then
    echo "⚠️  警告：服务状态异常！"
    # 发送告警通知
  fi
  
  sleep 60
done
```

---

## 附录

### 支持的模型名称

| 请求模型 | 映射到 Gemini 模型 |
|---------|------------------|
| `gpt-4` | `gemini-1.5-pro-latest` |
| `gpt-4-turbo` | `gemini-1.5-pro-latest` |
| `gpt-4-vision-preview` | `gemini-1.5-pro-latest` |
| `gpt-4o` | `gemini-1.5-flash-latest` |
| `gpt-4o-mini` | `gemini-1.5-flash-8b-latest` |
| `gpt-3.5-turbo` | `gemini-1.5-flash-latest` |
| `gemini-pro` | `gemini-1.5-pro-latest` |
| `gemini-flash` | `gemini-1.5-flash-latest` |
| `gemini-2.0-flash` | `gemini-2.0-flash` |
| `gemini-2.5-pro` | `gemini-2.5-pro-preview` |

> **提示**: 可在 `config.yaml` 中的 `model_mappings` 部分自定义模型映射。

### Key 池选择策略

| 策略 | 描述 | 适用场景 |
|------|------|---------|
| `round_robin` | 轮询选择 | 均衡负载 |
| `random` | 随机选择 | 简单场景 |
| `least_used` | 选择使用次数最少的 Key | 优化配额消耗 |
| `weighted` | 按成功率加权选择 | 优化成功率 |

### 常见问题

**Q: 会话数据存储在哪里？**  
A: 默认存储在 SQLite 数据库 `data/MuxueTools.db`，可通过 `database.path` 配置修改。

**Q: Key 统计数据会持久化吗？**  
A: 目前统计数据仅保存在内存中，服务重启后会重置（后续版本将支持持久化）。

**Q: 如何处理速率限制？**  
A: 系统会自动将触发速率限制的 Key 标记为 `rate_limited` 并进入冷却期（默认 60 秒），期间不会使用该 Key。

**Q: 配置更新需要重启吗？**  
A: `PUT /api/config` 可热更新部分配置（如日志级别、Key 池策略），但服务器配置（端口、主机等）需要重启服务才能生效。

---

**文档版本**: 1.0  
**最后更新**: 2026-01-15  
**项目地址**: https://github.com/muxueliunian/MuxueTools
