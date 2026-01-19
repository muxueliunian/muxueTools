# 任务: 增强版 API Key 管理 - 后端实现

> **角色**: Developer (Senior Golang Engineer)  
> **技能**: `.agent/skills/senior-golang/SKILL.md`  
> **参考文档**: `docs/DEVELOPMENT.md`, `docs/ARCHITECTURE.md`, `docs/API.md`

---

## 背景

当前添加 API Key 功能仅支持保存密钥，缺乏供应商识别、模型发现和连接测试。本次迭代需扩展后端支持：
1. 验证 Key 并从 Gemini API 获取可用模型列表
2. 存储 Key 的供应商和默认模型信息

---

## 步骤

### 1. 阅读 Skill 和规范
```
.agent/skills/senior-golang/SKILL.md
docs/DEVELOPMENT.md (单模块 TDD 开发流程)
docs/ARCHITECTURE.md (系统架构)
```

### 2. 扩展类型定义

**文件**: `internal/types/key.go`

```go
// 在 Key 结构体中添加
type Key struct {
    // ...existing fields...
    Provider     string `json:"provider"`       // e.g., "google_aistudio"
    DefaultModel string `json:"default_model"`  // e.g., "gemini-1.5-pro-latest"
}

// 在 CreateKeyRequest 中添加
type CreateKeyRequest struct {
    // ...existing fields...
    Provider     string `json:"provider,omitempty"`
    DefaultModel string `json:"default_model,omitempty"`
}

// 新增 DTO
type ValidateKeyRequest struct {
    Key      string `json:"key" binding:"required"`
    Provider string `json:"provider,omitempty"` // 默认 "google_aistudio"
}

type ValidateKeyResponse struct {
    Success bool              `json:"success"`
    Data    ValidateKeyResult `json:"data"`
}

type ValidateKeyResult struct {
    Valid     bool     `json:"valid"`
    LatencyMs int64    `json:"latency_ms"`
    Models    []string `json:"models"`
    Error     string   `json:"error,omitempty"`
}
```

### 3. 实现 Key 验证 Handler

**文件**: `internal/api/admin_handler.go`

新增 `ValidateKey` 函数:

```go
// ValidateKey handles POST /api/keys/validate
// 验证 Key 有效性并返回可用模型列表
func (h *AdminHandler) ValidateKey(c *gin.Context) {
    var req types.ValidateKeyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondBadRequest(c, "Invalid request")
        return
    }

    // 调用 Gemini models.list API
    url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + req.Key
    start := time.Now()
    resp, err := http.Get(url)
    latency := time.Since(start).Milliseconds()
    
    if err != nil {
        c.JSON(http.StatusOK, types.ValidateKeyResponse{
            Success: true,
            Data: types.ValidateKeyResult{Valid: false, Error: err.Error()},
        })
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        // 解析错误响应...
        c.JSON(http.StatusOK, types.ValidateKeyResponse{
            Success: true,
            Data: types.ValidateKeyResult{Valid: false, Error: "Invalid API Key"},
        })
        return
    }

    // 解析 models 列表
    var result struct {
        Models []struct {
            Name string `json:"name"`
        } `json:"models"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    // 提取模型名称 (去掉 "models/" 前缀)
    modelNames := make([]string, 0, len(result.Models))
    for _, m := range result.Models {
        modelNames = append(modelNames, strings.TrimPrefix(m.Name, "models/"))
    }

    c.JSON(http.StatusOK, types.ValidateKeyResponse{
        Success: true,
        Data: types.ValidateKeyResult{
            Valid:     true,
            LatencyMs: latency,
            Models:    modelNames,
        },
    })
}
```

### 4. 注册新路由

**文件**: `internal/api/router.go`

```go
// 在 keys 路由组中添加
keys.POST("/validate", adminHandler.ValidateKey)
```

### 5. 修改 AddKey Handler

**文件**: `internal/api/admin_handler.go`

更新 `AddKey` 函数以处理 `provider` 和 `default_model` 字段:

```go
newKey := &types.Key{
    // ...existing fields...
    Provider:     req.Provider,
    DefaultModel: req.DefaultModel,
}

if newKey.Provider == "" {
    newKey.Provider = "google_aistudio" // 默认值
}
```

---

## 产出

| 文件 | 变更类型 |
|------|----------|
| `internal/types/key.go` | 修改 (新增字段和类型) |
| `internal/api/admin_handler.go` | 修改 (新增 ValidateKey, 修改 AddKey) |
| `internal/api/router.go` | 修改 (注册新路由) |
| `docs/API.md` | 更新 (添加新接口文档) |

---

## 验证

1. **单元测试**: 补充 `ValidateKey` 的测试用例
2. **手动测试**:
   ```bash
   # 验证有效 Key
   curl -X POST http://localhost:8080/api/keys/validate \
     -H "Content-Type: application/json" \
     -d '{"key": "YOUR_VALID_KEY"}'
   
   # 验证无效 Key
   curl -X POST http://localhost:8080/api/keys/validate \
     -H "Content-Type: application/json" \
     -d '{"key": "invalid-key"}'
   ```

---

## 约束

- 遵循 `senior-golang` Skill 中的代码质量检查清单
- 使用 `context.Context` 传递 HTTP 超时
- 错误处理使用 `fmt.Errorf("%w", err)` 包装
- 运行 `go test -race ./internal/api/...` 确保无竞态

---

## 交接给下一个 Agent

- **下一步**: 前端 Agent 实现 Add Key Modal 改造
- **依赖**: 本任务完成后前端可调用 `/api/keys/validate`
