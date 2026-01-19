# Gemini 模型列表与映射

> 来源：https://ai.google.dev/gemini-api/docs/models/gemini

## 可用模型

### Gemini 3.0 系列（2026 最新旗舰）

| 模型名称 | API 标识符 | 特点 |
|---------|-----------|------|
| Gemini 3.0 Pro | `gemini-3.0-pro` | **最强多模态/推理**，支持 Deep Think 模式 |
| Gemini 3.0 Flash | `gemini-3.0-flash` | 极速响应，支持 1M 上下文，低延迟 |

### Gemini 2.5 系列

| 模型名称 | API 标识符 | 特点 |
|---------|-----------|------|
| Gemini 2.5 Pro | `gemini-2.5-pro-preview` | 强推理能力，支持 thinking |
| Gemini 2.5 Flash | `gemini-2.5-flash-preview` | 均衡性能与速度 |

### Gemini 2.0 系列

| 模型名称 | API 标识符 | 特点 |
|---------|-----------|------|
| Gemini 2.0 Flash | `gemini-2.0-flash` | 经典的快速多模态模型 |
| Gemini 2.0 Flash Lite | `gemini-2.0-flash-lite` | 轻量级，适合边缘计算 |

### Gemini 1.5 系列（经典）

| 模型名称 | API 标识符 | 特点 |
|---------|-----------|------|
| Gemini 1.5 Pro | `gemini-1.5-pro-latest` | 2M context window |
| Gemini 1.5 Flash | `gemini-1.5-flash-latest` | 成本效益之选 |

## OpenAI → Gemini 模型映射

MuxueTools 使用以下默认映射：

```go
var DefaultModelMappings = map[string]string{
    // === OpenAI GPT-5.2 (2025/2026) ===
    "gpt-5.2":             "gemini-3.0-pro",
    "gpt-5.2-pro":         "gemini-3.0-pro",
    "gpt-5.2-instant":     "gemini-3.0-flash",
    "gpt-5.2-thinking":    "gemini-3.0-pro", // 映射到支持 thinking 的模型

    // === OpenAI GPT-5 (2025) ===
    "gpt-5":               "gemini-2.5-pro-preview",
    "gpt-5-turbo":         "gemini-2.5-flash-preview",

    // === OpenAI GPT-4 系列 ===
    "gpt-4":               "gemini-2.0-flash", // GPT-4 性能已对应 2.0 Flash
    "gpt-4-turbo":         "gemini-1.5-pro-latest",
    "gpt-4o":              "gemini-2.0-flash",
    "gpt-4o-mini":         "gemini-1.5-flash-8b-latest",

    // === Gemini 原生名称（透传或别名） ===
    "gemini-pro":          "gemini-3.0-pro",    // 默认指向最新 Pro
    "gemini-flash":        "gemini-3.0-flash",  // 默认指向最新 Flash
    "gemini-3.0-pro":      "gemini-3.0-pro",
    "gemini-3.0-flash":    "gemini-3.0-flash",
    "gemini-2.5-pro":      "gemini-2.5-pro-preview",
    "gemini-2.5-flash":    "gemini-2.5-flash-preview",
}
```

## 映射策略

### 1. 透传（Passthrough）

未在映射表中的模型名称将直接透传给 Gemini API。

### 2. 可配置化

用户可以在 `config.yaml` 中自定义映射：

```yaml
model_mappings:
  "gpt-6-preview": "gemini-3.0-pro"
  "my-custom-agent": "gemini-1.5-flash"
```

## API 速率限制 (参考)

### 免费层 (Free Tier)

| 模型 | RPM | TPM | RPD |
|------|-----|-----|-----|
| Gemini 3.0 Flash | 20 | 2M | 2,000 |
| Gemini 3.0 Pro | 5 | 500K | 100 |
| Gemini 1.5 Flash | 15 | 1M | 1,500 |

> **注意**：具体限制请以 Google AI Studio 仪表盘为准，Google AI Plus 用户通常拥有 5x 以上的配额。

## 模型能力对比

| 能力 | 1.5 Pro | 2.5 Pro | 3.0 Flash | 3.0 Pro |
|------|---------|---------|-----------|---------|
| 文本生成 | ✅ | ✅ | ✅ | ✅ |
| 多模态 (AV) | ✅ | ✅ | ✅ | ✅ |
| 代码执行 | ✅ | ✅ | ✅ | ✅ |
| Thinking Mode | ❌ | ✅ | ❌ | ✅ (Deep Think) |
| 上下文长度 | 2M | 1M | 1M | 1M+ |
| 智能体能力 | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

## 返回模型名称

在响应中，`model` 字段应返回用户请求的原始模型名称（OpenAI 格式），而非 Gemini 模型名称，以兼容 OpenAI SDK。

---

*最后更新：2026-01-15 (已验证 Gemini 3.0 & GPT-5.2 发布信息)*
