# 任务：实现格式转换模块（TDD）

## 角色
Developer (senior-golang skill)

## 背景
Key 池模块已完成并通过审核。现在需要实现格式转换模块，负责 OpenAI 和 Gemini 格式的双向转换。这是纯逻辑模块，无 IO 操作，非常适合 TDD。

## 步骤

### 1. 阅读规范
- `.agent/skills/senior-golang/SKILL.md` - Go 开发规范
- `docs/ARCHITECTURE.md` - Converter 设计
- `internal/types/openai.go` - OpenAI 类型
- `internal/types/gemini.go` - Gemini 类型

### 2. TDD 开发流程
严格遵循：先写测试 → 测试失败 → 实现代码 → 测试通过

### 3. 创建文件结构
```
internal/gemini/
├── converter.go        # 格式转换核心实现
├── converter_test.go   # 转换测试
└── models.go           # 模型映射（可选，如果需要额外定义）
```

### 4. 功能要求

#### OpenAI → Gemini 转换：
```go
// 请求转换
func ConvertOpenAIRequest(req *types.ChatCompletionRequest) (*types.GeminiRequest, error)

// 处理多模态内容（文本 + 图片）
func ConvertMessages(messages []types.Message) ([]types.GeminiContent, error)

// 模型名称映射
func MapModelName(openaiModel string) string
```

#### Gemini → OpenAI 转换：
```go
// 响应转换（阻塞模式）
func ConvertGeminiResponse(resp *types.GeminiResponse, model string) (*types.ChatCompletionResponse, error)

// 流式响应转换
func ConvertGeminiStreamChunk(chunk *types.GeminiResponse, model string, index int) (*types.ChatCompletionChunk, error)
```

#### 特殊处理：
- 图片内容转换（base64 / URL）
- system 消息处理（Gemini 的 systemInstruction）
- stop 序列转换
- Usage/Token 统计转换

### 5. 测试用例（必须覆盖）

#### 基础转换：
- 简单文本消息转换
- 多轮对话转换
- 带 system 消息的转换

#### 多模态：
- 文本 + base64 图片
- 文本 + URL 图片
- 多张图片

#### 参数转换：
- temperature, topP, maxTokens 转换
- stop 序列转换（string 和 []string）

#### 响应转换：
- 正常响应转换
- 流式 chunk 转换
- Usage 统计转换

#### 边界条件：
- 空消息列表
- 未知模型名称
- 不支持的内容类型

#### Benchmark：
- BenchmarkConvertOpenAIRequest
- BenchmarkConvertGeminiResponse

### 6. 运行测试
```powershell
go test ./internal/gemini/... -v
go test ./internal/gemini/... -bench=. -benchmem
```

## 产出
- `internal/gemini/converter.go`
- `internal/gemini/converter_test.go`

## 约束
- 纯函数实现，无副作用
- 使用 `internal/types/` 中已定义的类型
- 错误处理使用 `internal/types/errors.go`
- 测试覆盖率目标：>85%（纯逻辑模块应更高）
- 必须包含 Benchmark 测试

## 模型映射参考
```go
var defaultModelMappings = map[string]string{
    "gpt-4":              "gemini-1.5-pro-latest",
    "gpt-4-turbo":        "gemini-1.5-pro-latest",
    "gpt-4o":             "gemini-1.5-flash-latest",
    "gpt-4o-mini":        "gemini-1.5-flash-8b-latest",
    "gpt-3.5-turbo":      "gemini-1.5-flash-latest",
    // Gemini 原生名称直接透传
    "gemini-1.5-pro":     "gemini-1.5-pro-latest",
    "gemini-1.5-flash":   "gemini-1.5-flash-latest",
    "gemini-2.0-flash":   "gemini-2.0-flash",
}
```

---

*任务创建时间: 2026-01-15*
