# Gemini 流式响应（SSE）处理

> 来源：https://ai.google.dev/gemini-api/docs/text-generation#streaming

## 概述

Gemini API 支持通过 Server-Sent Events (SSE) 返回流式响应，实现实时输出显示。

## 端点

```
POST https://generativelanguage.googleapis.com/v1beta/models/{model}:streamGenerateContent?key={API_KEY}&alt=sse
```

**关键参数**：`alt=sse` 必须添加以启用 SSE 格式

## SSE 响应格式

每个 SSE 事件格式如下：

```
data: {"candidates":[...],"usageMetadata":{...}}

```

**注意**：
- 每行以 `data: ` 开头（注意冒号后有空格）
- 事件之间用空行分隔
- 最后一个事件包含 `finishReason` 和 `usageMetadata`

## 完整示例

### 请求

```bash
curl "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:streamGenerateContent?key=$GEMINI_API_KEY&alt=sse" \
  -H 'Content-Type: application/json' \
  --no-buffer \
  -d '{
    "contents": [{
      "parts": [{"text": "Write a short poem about coding"}]
    }]
  }'
```

### 响应流

```
data: {"candidates":[{"content":{"parts":[{"text":"Lines of code"}],"role":"model"},"index":0}]}

data: {"candidates":[{"content":{"parts":[{"text":" dance and flow,"}],"role":"model"},"index":0}]}

data: {"candidates":[{"content":{"parts":[{"text":"\nBuilding dreams"}],"role":"model"},"index":0}]}

data: {"candidates":[{"content":{"parts":[{"text":" that start to grow."}],"role":"model"},"finishReason":"STOP","index":0}],"usageMetadata":{"promptTokenCount":7,"candidatesTokenCount":18,"totalTokenCount":25}}

```

## 关键处理逻辑

### 1. 解析 SSE 数据

```go
// 读取每一行
line, err := reader.ReadBytes('\n')
if err != nil {
    if err == io.EOF {
        return // 流结束
    }
    return err
}

// 跳过空行
line = bytes.TrimSpace(line)
if len(line) == 0 {
    continue
}

// 解析 data: 前缀
if !bytes.HasPrefix(line, []byte("data: ")) {
    continue
}
jsonData := line[6:] // 去掉 "data: " 前缀

// 解析 JSON
var chunk GeminiResponse
json.Unmarshal(jsonData, &chunk)
```

### 2. 检测流结束

流结束的特征：
1. `finishReason` 不为空（如 `"STOP"`、`"MAX_TOKENS"`）
2. `usageMetadata` 出现在响应中（通常在最后一个 chunk）

```go
if len(chunk.Candidates) > 0 && chunk.Candidates[0].FinishReason != "" {
    // 这是最后一个有效 chunk
}
```

### 3. 空 Candidates 处理

某些 chunk 可能有空的 `candidates`（例如安全过滤），需要处理：

```go
if len(chunk.Candidates) == 0 {
    // 可能是被安全过滤，或者是心跳
    if chunk.PromptFeedback != nil && chunk.PromptFeedback.BlockReason != "" {
        // 内容被阻止
    }
    continue
}
```

## 转换为 OpenAI SSE 格式

OpenAI 的 SSE 格式类似，但字段不同：

### Gemini Chunk → OpenAI Chunk

**Gemini 格式**：
```json
{
  "candidates": [{
    "content": {
      "parts": [{"text": "Hello"}],
      "role": "model"
    },
    "finishReason": "STOP"
  }],
  "usageMetadata": {...}
}
```

**OpenAI 格式**：
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion.chunk",
  "created": 1234567890,
  "model": "gpt-4",
  "choices": [{
    "index": 0,
    "delta": {
      "role": "assistant",  // 仅第一个 chunk
      "content": "Hello"
    },
    "finish_reason": "stop"  // 仅最后一个 chunk
  }]
}
```

### 转换要点

1. **第一个 chunk**：设置 `delta.role = "assistant"`
2. **中间 chunks**：只设置 `delta.content`
3. **最后一个 chunk**：设置 `finish_reason`（映射 Gemini 的 `finishReason`）

### finishReason 映射

| Gemini | OpenAI |
|--------|--------|
| `STOP` | `stop` |
| `MAX_TOKENS` | `length` |
| `SAFETY` | `content_filter` |
| `RECITATION` | `content_filter` |
| 其他 | `stop` |

## Go 实现示例

```go
func streamGeminiResponse(resp *http.Response, outputChan chan<- StreamEvent) {
    defer resp.Body.Close()
    defer close(outputChan)
    
    reader := bufio.NewReader(resp.Body)
    isFirstChunk := true
    
    for {
        line, err := reader.ReadBytes('\n')
        if err != nil {
            if err == io.EOF {
                return
            }
            outputChan <- StreamEvent{Err: err}
            return
        }
        
        line = bytes.TrimSpace(line)
        if len(line) == 0 || !bytes.HasPrefix(line, []byte("data: ")) {
            continue
        }
        
        jsonData := line[6:]
        
        var geminiResp types.GeminiResponse
        if err := json.Unmarshal(jsonData, &geminiResp); err != nil {
            outputChan <- StreamEvent{Err: err}
            return
        }
        
        // 转换为 OpenAI 格式
        chunk, err := ConvertGeminiStreamChunk(&geminiResp, model, isFirstChunk)
        if err != nil {
            outputChan <- StreamEvent{Err: err}
            return
        }
        
        isFirstChunk = false
        outputChan <- StreamEvent{Chunk: chunk}
        
        // 检测流结束
        if geminiResp.HasFinishReason() {
            return
        }
    }
}
```

## 注意事项

1. **使用 `--no-buffer`**：cURL 测试时必须添加此参数
2. **Content-Type**：响应的 Content-Type 是 `text/event-stream`
3. **超时处理**：长时间无响应应设置超时
4. **Context 取消**：支持 context 取消以中断流
5. **资源清理**：确保 `resp.Body.Close()` 被调用

---

*最后更新：2026-01-15*
