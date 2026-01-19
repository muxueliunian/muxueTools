# 任务：API 层测试与审核

## 角色
QA Engineer (qa-automation skill)

## 背景
Developer Agent 已完成阶段 3 API 层实现：
- 29 个单元测试通过
- `go vet` 静态分析通过
- 服务可正常启动

现在需要进行 **全面的功能测试和代码审核**，确保 API 层实现正确、健壮。

## 审核范围

```
internal/api/
├── response.go           # 统一响应格式
├── middleware.go         # 中间件
├── openai_handler.go     # OpenAI 兼容端点
├── admin_handler.go      # 管理端点
├── router.go             # 路由配置
├── server.go             # 服务器
├── router_test.go        # 路由测试
└── openai_handler_test.go # 端点测试

cmd/server/main.go        # 主程序入口
```

## 测试步骤

### 1. 阅读规范
- `.agent/skills/qa-automation/SKILL.md` - QA 审核规范
- `.agent/skills/senior-golang/SKILL.md` - Go 代码规范（Gin 部分）
- `docs/ARCHITECTURE.md` - API 设计规范

### 2. 代码审核

#### 2.1 OpenAI 兼容端点 (openai_handler.go)

| 检查项 | 说明 |
|--------|------|
| 请求解析 | `ChatCompletionRequest` 是否正确解析？ |
| 流式判断 | `stream: true` 是否正确处理？ |
| 阻塞响应 | 响应格式是否符合 OpenAI 规范？ |
| 流式响应 | SSE 格式是否正确？(`data: {...}\n\n`) |
| 结束标记 | 流式结束是否发送 `data: [DONE]\n\n`？ |
| 错误处理 | 错误是否使用预定义格式？ |
| 模型列表 | `/v1/models` 返回格式是否正确？ |

#### 2.2 管理端点 (admin_handler.go)

| 检查项 | 说明 |
|--------|------|
| Key 脱敏 | 返回的 Key 是否正确脱敏？ |
| CRUD 完整性 | 增删改查是否都实现？ |
| 导入/导出 | 批量操作是否正确？ |
| 配置更新 | PUT /api/config 是否生效？ |
| 统计返回 | 统计数据格式是否正确？ |

#### 2.3 中间件 (middleware.go)

| 检查项 | 说明 |
|--------|------|
| CORS | 是否正确设置跨域头？ |
| Request ID | 是否生成唯一请求 ID？ |
| Logging | 是否记录请求日志？ |
| Recovery | panic 是否被正确捕获？ |

#### 2.4 服务器 (server.go)

| 检查项 | 说明 |
|--------|------|
| 依赖注入 | 是否正确初始化所有依赖？ |
| 优雅关闭 | Shutdown 是否正确实现？ |
| 配置加载 | 是否使用配置中的端口？ |

### 3. 功能测试

#### 3.1 启动服务器

```bash
# 创建测试配置
# 启动服务器
go run ./cmd/server -config ./configs/config.example.yaml
```

#### 3.2 健康检查

```bash
curl -X GET http://localhost:8080/health
# 期望: {"status":"ok","version":"..."}
```

#### 3.3 模型列表

```bash
curl -X GET http://localhost:8080/v1/models
# 期望: {"object":"list","data":[...]}
```

#### 3.4 阻塞请求测试（Mock）

如果没有真实 API Key，验证请求格式是否正确解析：

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
# 验证: 请求是否被正确解析，错误响应格式是否正确
```

#### 3.5 流式请求测试（Mock）

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
# 验证: Content-Type 是否为 text/event-stream
```

#### 3.6 管理 API 测试

```bash
# 获取 Key 列表
curl -X GET http://localhost:8080/api/keys

# 添加 Key
curl -X POST http://localhost:8080/api/keys \
  -H "Content-Type: application/json" \
  -d '{"key": "test-api-key", "name": "Test Key"}'

# 获取配置
curl -X GET http://localhost:8080/api/config

# 获取统计
curl -X GET http://localhost:8080/api/stats
```

### 4. 边界条件测试

| 测试场景 | 期望结果 |
|----------|---------|
| 空 messages | 返回 400 Invalid Request |
| 无效 JSON | 返回 400 Invalid Request |
| 未知模型 | 透传或返回错误 |
| 无可用 Key | 返回 429 或 503 |
| 超大请求体 | 正确处理或限制 |

### 5. 并发测试（可选）

```go
// 简单并发测试
func TestConcurrentRequests(t *testing.T) {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // 发送请求
        }()
    }
    wg.Wait()
}
```

## 产出

创建审核报告 `docs/CodeReviewReport/API_LAYER_REVIEW_REPORT.md`：

```markdown
# API 层测试与审核报告

> **审核员**: QA Engineer
> **审核日期**: YYYY-MM-DD
> **审核范围**: `internal/api/`, `cmd/server/main.go`

---

## 审核结果：✅ 通过 / ⚠️ 需修复 / ❌ 严重问题

| 指标 | 数量 |
|------|------|
| 严重问题 | X |
| 警告 | X |
| 建议改进 | X |

---

## 功能测试结果

| 测试项 | 状态 | 说明 |
|--------|------|------|
| 服务启动 | ✅/❌ | |
| 健康检查 | ✅/❌ | |
| 模型列表 | ✅/❌ | |
| 阻塞请求 | ✅/❌ | |
| 流式请求 | ✅/❌ | |
| Key 管理 | ✅/❌ | |
| 配置管理 | ✅/❌ | |
| 统计 API | ✅/❌ | |

---

## 代码审核结果

### openai_handler.go
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

### admin_handler.go
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

### middleware.go
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

### server.go
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

---

## 问题详情

### ❌ 严重问题（如有）

### ⚠️ 警告（如有）

### 💡 改进建议（如有）

---

## 总结

**结论**: 继续开发 / 需先修复

---

*报告生成时间: YYYY-MM-DD HH:MM*
```

## 约束
- 优先进行功能测试，验证 API 是否可用
- 代码审核重点关注请求解析和响应格式
- 流式响应必须符合 OpenAI SSE 规范
- 如发现严重问题，需提供具体修复方案
- 如果没有真实 API Key，使用 Mock 测试验证请求格式

## 验收标准
1. 所有端点可访问（返回有意义的响应）
2. 错误响应格式符合 OpenAI 规范
3. 流式响应符合 SSE 规范
4. 无严重的代码质量问题

---

*任务创建时间: 2026-01-15*
