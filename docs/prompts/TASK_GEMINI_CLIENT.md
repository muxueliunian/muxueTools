# 任务：实现 Gemini API 客户端（TDD）

## 角色
Developer (senior-golang skill)

## 背景
配置加载、Key 池和格式转换模块已完成并通过审核。现在需要实现 Gemini API 客户端，负责与 Google AI Studio API 通信。

**已完成的依赖模块：**
- `internal/config/` - 配置加载
- `internal/keypool/` - Key 池管理（获取可用 Key、熔断报告）
- `internal/gemini/converter.go` - 格式转换

## 步骤

### 1. 阅读规范
- `.agent/skills/senior-golang/SKILL.md` - Go 开发规范（特别是 SSE 处理部分）
- `docs/ARCHITECTURE.md` - Gemini Client 设计
- `internal/types/gemini.go` - Gemini 类型定义
- `internal/keypool/pool.go` - Key 池接口（了解如何获取/释放 Key）

**⚠️ 重要：阅读 Gemini 官方 API 文档**
- `docs/gemini/API_REFERENCE.md` - API 端点和请求/响应格式
- `docs/gemini/STREAMING.md` - 流式响应（SSE）处理细节
- `docs/gemini/VISION.md` - 图片输入处理
- `docs/gemini/MODELS.md` - 模型列表和映射

### 2. TDD 开发流程
严格遵循：先写测试 → 测试失败 → 实现代码 → 测试通过

### 3. 创建文件结构
```
internal/gemini/
├── client.go           # Gemini API 客户端（新建）
├── client_test.go      # 客户端测试（新建）
├── converter.go        # 格式转换（已存在）
└── converter_test.go   # 转换测试（已存在）
```

### 4. 功能要求

#### Client 结构
```go
type Client struct {
    httpClient    *http.Client
    pool          *keypool.Pool
    baseURL       string
    requestTimeout time.Duration
}

func NewClient(pool *keypool.Pool, opts ...ClientOption) *Client
```

#### 核心方法

```go
// 阻塞式请求（返回完整响应）
func (c *Client) ChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error)

// 流式请求（返回 channel）
func (c *Client) ChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (<-chan StreamEvent, error)

// StreamEvent 类型
type StreamEvent struct {
    Chunk *types.ChatCompletionChunk
    Err   error
    Done  bool
}
```

#### 内部流程
1. 从 KeyPool 获取可用 Key
2. 调用 Converter 转换请求格式
3. 发送 HTTP 请求到 Gemini API
4. 解析响应/流式读取
5. 调用 Converter 转换响应格式
6. 向 KeyPool 报告成功/失败（触发统计和熔断）
7. 释放 Key

### 5. API 端点
Gemini API 端点格式：
```
POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={API_KEY}
POST https://generativelanguage.googleapis.com/v1beta/models/{model}:streamGenerateContent?key={API_KEY}&alt=sse
```

### 6. 测试用例

#### Mock 设计
- 创建 Mock HTTP Server 模拟 Gemini API
- 创建 Mock KeyPool 接口

#### 测试场景

**Happy Path：**
- 简单文本请求/响应
- 多模态请求（文本 + 图片）
- 流式响应（多个 chunk）

**Error Cases：**
- 无可用 Key（KeyPool 返回错误）
- API 返回 429 Rate Limit
- API 返回 400/401/500 错误
- 网络超时
- 响应解析失败

**流式特殊场景：**
- 流中途断开
- 空 chunk 处理
- Context 取消

**集成 KeyPool：**
- 成功请求后调用 `ReportSuccess`
- 失败请求后调用 `ReportFailure`
- Rate Limit 触发熔断

### 7. 运行测试
```powershell
go test ./internal/gemini/... -v
go test ./internal/gemini/... -race
```

## 产出
- `internal/gemini/client.go`
- `internal/gemini/client_test.go`

## 约束
- 使用 `net/http` 标准库
- 流式响应使用 `bufio.Reader` 逐行读取 SSE
- 必须正确处理 Context 取消
- 使用 `internal/keypool/` 获取/释放 Key
- 使用 `internal/gemini/converter.go` 进行格式转换
- 错误处理使用 `internal/types/errors.go`
- 测试需要 Mock HTTP Server，不调用真实 API

## SSE 格式参考

Gemini streamGenerateContent 返回格式：
```
data: {"candidates":[{"content":{"parts":[{"text":"Hello"}],"role":"model"}}]}

data: {"candidates":[{"content":{"parts":[{"text":" world"}],"role":"model"}}]}

data: {"candidates":[{"content":{"parts":[{"text":"!"}],"role":"model"},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":5,"candidatesTokenCount":3,"totalTokenCount":8}}

```

注意：
- 每行以 `data: ` 开头
- 以空行分隔
- 最后一个 chunk 包含 `finishReason` 和 `usageMetadata`

---

*任务创建时间: 2026-01-15*
