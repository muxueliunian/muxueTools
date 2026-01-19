# 任务：实现基于请求的模型使用统计

## 角色
Developer (senior-golang)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`

---

## 背景

当前 `/api/stats/models` 端点根据每个 API Key 的 `default_model` 字段来统计模型使用分布。这导致以下问题：

1. **数据不准确**：用户在请求时可以指定任意模型，而不是使用 Key 的 `default_model`。
2. **批量导入的 Key 无模型信息**：批量导入功能不设置 `default_model`，导致统计显示为 "unknown"。
3. **缺乏请求级粒度**：无法追踪历史请求中实际使用的模型。

**现有相关实现：**（参见 `docs/DEVELOPMENT.md`）
- `internal/api/admin_handler.go` - 现有 `GetStatsModels` 实现
- `internal/api/openai_handler.go` - 请求处理逻辑
- `internal/keypool/pool.go` - Key 池和统计管理
- `internal/types/key.go` - 统计类型定义

---

## 目标

| # | 目标 | 优先级 |
|---|------|--------|
| 1 | 在请求处理时记录实际使用的模型名称 | P0 |
| 2 | 修改 `/api/stats/models` 使用请求级模型数据 | P0 |
| 3 | 保持向后兼容（如无请求数据则回退到 `default_model`） | P1 |

---

## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范**
   - `.agent/skills/senior-golang/SKILL.md`

2. **项目文档**
   - `docs/ARCHITECTURE.md` - 系统架构
   - `docs/DEVELOPMENT.md` - 开发工作流
   - `docs/API.md` - 现有 API 文档

3. **相关代码**
   - `internal/api/openai_handler.go` - 请求处理（`ChatCompletions` 函数）
   - `internal/api/admin_handler.go` - 现有 `GetStatsModels` 实现
   - `internal/keypool/pool.go` - `IncrementStats` 方法
   - `internal/types/key.go` - `KeyStats` 结构体

---

### 步骤 1：扩展 KeyStats 结构体

**修改文件**: `internal/types/key.go`

在 `KeyStats` 结构体中添加模型使用计数器：

```go
// KeyStats holds usage statistics for a single key.
type KeyStats struct {
    RequestCount     int64      `json:"request_count"`
    SuccessCount     int64      `json:"success_count"`
    ErrorCount       int64      `json:"error_count"`
    PromptTokens     int64      `json:"prompt_tokens"`
    CompletionTokens int64      `json:"completion_tokens"`
    LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
    // 新增字段：按模型统计请求数
    ModelUsage       map[string]int64 `json:"model_usage,omitempty"`
}
```

---

### 步骤 2：修改 IncrementStats 方法

**修改文件**: `internal/types/key.go` 和/或 `internal/keypool/pool.go`

更新 `IncrementStats` 方法签名以接受模型参数：

```go
// IncrementStats updates the key's statistics after a request.
// model: 实际请求使用的模型名称
func (k *Key) IncrementStats(success bool, promptTokens, completionTokens int, model string) {
    k.Stats.RequestCount++
    if success {
        k.Stats.SuccessCount++
    } else {
        k.Stats.ErrorCount++
    }
    k.Stats.PromptTokens += int64(promptTokens)
    k.Stats.CompletionTokens += int64(completionTokens)
    now := time.Now()
    k.Stats.LastUsedAt = &now
    k.UpdatedAt = now
    
    // 记录模型使用
    if model != "" {
        if k.Stats.ModelUsage == nil {
            k.Stats.ModelUsage = make(map[string]int64)
        }
        k.Stats.ModelUsage[model]++
    }
}
```

---

### 步骤 3：更新请求处理逻辑

**修改文件**: `internal/api/openai_handler.go`

在 `ChatCompletions` 函数中，调用 `IncrementStats` 时传递实际使用的模型名称：

```go
// 在请求成功后更新统计（伪代码示意）
key.IncrementStats(true, promptTokens, completionTokens, req.Model)
```

---

### 步骤 4：修改 GetStatsModels 逻辑

**修改文件**: `internal/api/admin_handler.go`

重写 `GetStatsModels` 函数，优先使用 `KeyStats.ModelUsage` 数据：

```go
func (h *AdminHandler) GetStatsModels(c *gin.Context) {
    stats := h.pool.GetStats()
    
    // 聚合所有 Key 的 ModelUsage
    modelMap := make(map[string]*types.ModelUsageItem)
    var totalRequests int64
    
    for _, key := range stats {
        // 优先使用请求级模型统计
        if len(key.Stats.ModelUsage) > 0 {
            for model, count := range key.Stats.ModelUsage {
                if _, exists := modelMap[model]; !exists {
                    modelMap[model] = &types.ModelUsageItem{Model: model}
                }
                modelMap[model].RequestCount += count
                // TokenUsage 无法按模型分割，可统一计入或使用估算
                totalRequests += count
            }
        } else if key.DefaultModel != "" {
            // 回退：使用 default_model
            model := key.DefaultModel
            if _, exists := modelMap[model]; !exists {
                modelMap[model] = &types.ModelUsageItem{Model: model}
            }
            modelMap[model].RequestCount += key.Stats.RequestCount
            modelMap[model].TokenUsage += key.Stats.TotalTokens()
            totalRequests += key.Stats.RequestCount
        }
        // 如果既无 ModelUsage 也无 DefaultModel，则不计入（或归入 "other"）
    }
    
    // 计算百分比并排序...
}
```

---

### 步骤 5：更新单元测试

**修改/新增文件**: `internal/types/key_test.go`, `internal/api/admin_handler_test.go`

确保：
- `IncrementStats` 正确记录模型使用
- `GetStatsModels` 正确聚合 `ModelUsage` 数据
- 向后兼容逻辑正常工作

---

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/types/key.go` | **MODIFY** | 扩展 `KeyStats` 结构体，更新 `IncrementStats` |
| `internal/api/openai_handler.go` | **MODIFY** | 调用 `IncrementStats` 时传递模型参数 |
| `internal/api/admin_handler.go` | **MODIFY** | 重写 `GetStatsModels` 逻辑 |
| `internal/keypool/pool.go` | **MODIFY** | 可能需要更新 `ReportSuccess`/`ReportFailure` 接口 |
| `internal/types/key_test.go` | **MODIFY** | 新增 `IncrementStats` 模型参数测试 |
| `internal/api/admin_handler_test.go` | **MODIFY** | 新增 `GetStatsModels` 测试用例 |

---

## 约束

### 技术约束
- 使用 `map[string]int64` 存储模型使用计数
- `ModelUsage` 字段序列化时使用 `omitempty` 避免空 map 输出
- 模型名称使用原始请求中的 `model` 字段值（不做映射转换）

### 质量约束
- 遵循 `.agent/skills/senior-golang/SKILL.md` 代码规范
- `go test -race` 无竞态问题（注意 map 并发安全）
- 现有 API 行为保持向后兼容

### 兼容性约束
- 旧数据（无 `ModelUsage`）仍能正常统计（回退到 `default_model`）
- `/api/stats/models` 响应格式不变

---

## 验收标准

- [ ] `go vet ./internal/...` 无错误
- [ ] `go test ./internal/types/...` 通过（含新增测试）
- [ ] `go test ./internal/api/...` 通过
- [ ] `go test -race ./internal/...` 无竞态问题
- [ ] 发送请求后 `/api/stats/models` 返回实际模型分布
- [ ] 无历史数据时回退显示 `default_model` 或空

---

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/DEVELOPMENT.md` | 更新任务进度（统计功能增强） |
| `docs/API.md` | 更新 `/api/stats/models` 响应说明 |

---

## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的 TDD 开发流程。

---

*任务创建时间: 2026-01-18*
