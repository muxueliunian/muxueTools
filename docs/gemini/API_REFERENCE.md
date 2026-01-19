# Gemini API 参考

> 来源：https://ai.google.dev/api/generate-content

## API 端点

### generateContent（阻塞式请求）

**端点**：
```
POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={API_KEY}
```

**路径参数**：
- `model`：模型名称，格式：`models/{model}`，例如 `models/gemini-1.5-flash-latest`

### streamGenerateContent（流式请求）

**端点**：
```
POST https://generativelanguage.googleapis.com/v1beta/models/{model}:streamGenerateContent?key={API_KEY}&alt=sse
```

**注意**：必须添加 `alt=sse` 参数以启用 Server-Sent Events 格式

---

## 请求体结构

```json
{
  "contents": [
    {
      "role": "user" | "model",
      "parts": [
        { "text": "string" },
        { "inlineData": { "mimeType": "string", "data": "base64-string" } },
        { "fileData": { "mimeType": "string", "fileUri": "string" } }
      ]
    }
  ],
  "systemInstruction": {
    "parts": [
      { "text": "string" }
    ]
  },
  "generationConfig": {
    "stopSequences": ["string"],
    "responseMimeType": "text/plain" | "application/json",
    "candidateCount": 1,
    "maxOutputTokens": 8192,
    "temperature": 1.0,
    "topP": 0.95,
    "topK": 40
  },
  "safetySettings": [
    {
      "category": "HARM_CATEGORY_XXX",
      "threshold": "BLOCK_XXX"
    }
  ]
}
```

### contents[] - 对话内容

| 字段 | 类型 | 说明 |
|------|------|------|
| `role` | string | `user` 或 `model` |
| `parts[]` | array | 内容部分数组 |

### parts[] - 内容部分

**文本内容**：
```json
{ "text": "Hello, world!" }
```

**内联图片（Base64）**：
```json
{
  "inlineData": {
    "mimeType": "image/jpeg",
    "data": "base64-encoded-image-data"
  }
}
```

**文件引用**：
```json
{
  "fileData": {
    "mimeType": "image/jpeg",
    "fileUri": "https://example.com/image.jpg"
  }
}
```

### systemInstruction - 系统指令

```json
{
  "systemInstruction": {
    "parts": [
      { "text": "You are a helpful assistant." }
    ]
  }
}
```

**注意**：`systemInstruction` 不需要 `role` 字段

### generationConfig - 生成配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `stopSequences[]` | string[] | - | 停止序列（最多 5 个） |
| `responseMimeType` | string | "text/plain" | 响应格式 |
| `candidateCount` | integer | 1 | 候选响应数量 |
| `maxOutputTokens` | integer | - | 最大输出 token 数 |
| `temperature` | number | 1.0 | 温度参数 (0-2) |
| `topP` | number | 0.95 | Top-P 采样 |
| `topK` | integer | 40 | Top-K 采样 |

---

## 响应体结构

### GenerateContentResponse

```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          { "text": "response text" }
        ],
        "role": "model"
      },
      "finishReason": "STOP",
      "index": 0,
      "safetyRatings": [...]
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 10,
    "candidatesTokenCount": 50,
    "totalTokenCount": 60
  },
  "modelVersion": "gemini-1.5-flash-latest"
}
```

### candidates[] - 候选响应

| 字段 | 类型 | 说明 |
|------|------|------|
| `content` | object | 响应内容 |
| `finishReason` | string | 完成原因 |
| `index` | integer | 候选索引 |
| `safetyRatings[]` | array | 安全评分 |

### finishReason - 完成原因枚举

| 值 | 说明 | 映射到 OpenAI |
|----|------|---------------|
| `STOP` | 正常完成 | `stop` |
| `MAX_TOKENS` | 达到最大 token 限制 | `length` |
| `SAFETY` | 安全过滤 | `content_filter` |
| `RECITATION` | 引用过滤 | `content_filter` |
| `LANGUAGE` | 语言不支持 | `stop` |
| `OTHER` | 其他原因 | `stop` |
| `BLOCKLIST` | 阻止列表 | `content_filter` |
| `PROHIBITED_CONTENT` | 禁止内容 | `content_filter` |
| `SPII` | 敏感信息 | `content_filter` |
| `MALFORMED_FUNCTION_CALL` | 函数调用格式错误 | `stop` |

### usageMetadata - Token 使用统计

| 字段 | 类型 | 说明 |
|------|------|------|
| `promptTokenCount` | integer | 输入 token 数 |
| `candidatesTokenCount` | integer | 输出 token 数 |
| `totalTokenCount` | integer | 总 token 数 |
| `cachedContentTokenCount` | integer | 缓存内容 token 数（可选） |
| `thoughtsTokenCount` | integer | 思考 token 数（thinking 模型） |

---

## 错误响应

```json
{
  "error": {
    "code": 429,
    "message": "Resource exhausted",
    "status": "RESOURCE_EXHAUSTED"
  }
}
```

### 常见错误码

| HTTP 状态码 | 错误类型 | 说明 |
|------------|---------|------|
| 400 | INVALID_ARGUMENT | 请求参数错误 |
| 401 | UNAUTHENTICATED | API Key 无效 |
| 403 | PERMISSION_DENIED | 权限不足 |
| 404 | NOT_FOUND | 模型不存在 |
| 429 | RESOURCE_EXHAUSTED | 速率限制 |
| 500 | INTERNAL | 服务器内部错误 |
| 503 | UNAVAILABLE | 服务不可用 |

---

## cURL 示例

### 简单文本请求

```bash
curl "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=$GEMINI_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "contents": [{
      "parts": [{"text": "Hello, how are you?"}]
    }]
  }'
```

### 带系统指令

```bash
curl "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=$GEMINI_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "system_instruction": {
      "parts": [{"text": "You are a helpful assistant."}]
    },
    "contents": [{
      "parts": [{"text": "What is the capital of France?"}]
    }]
  }'
```

### 多模态请求（图片）

```bash
curl "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=$GEMINI_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "contents": [{
      "parts": [
        {
          "inline_data": {
            "mime_type": "image/jpeg",
            "data": "'$(base64 -w0 image.jpg)'"
          }
        },
        {"text": "Describe this image."}
      ]
    }]
  }'
```

---

*最后更新：2026-01-15*
