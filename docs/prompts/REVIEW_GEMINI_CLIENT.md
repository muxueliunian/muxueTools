# 任务：Gemini API 客户端 Code Review

## 角色
QA Engineer (qa-automation skill)

## 背景
Developer Agent 已完成 Gemini API 客户端的 TDD 开发。该模块负责与 Google AI Studio API 通信，支持阻塞式和流式响应。

## 审核范围
- `internal/gemini/client.go` - Gemini API 客户端实现
- `internal/gemini/client_test.go` - 客户端测试

## 步骤

### 1. 阅读规范
- `.agent/skills/qa-automation/SKILL.md` - QA 审核规范
- `.agent/skills/senior-golang/SKILL.md` - Go 代码规范（SSE 处理部分）
- `docs/ARCHITECTURE.md` - Gemini Client 设计

### 2. 功能正确性审核

| 检查项 | 说明 |
|--------|------|
| KeyPool 集成 | 是否正确获取/释放 Key？ |
| 请求转换 | 是否正确调用 converter？ |
| HTTP 请求 | URL、Headers、Body 是否正确？ |
| 响应解析 | 阻塞响应是否正确解析？ |
| 流式处理 | SSE 格式解析是否正确？ |
| 错误处理 | 各种错误是否正确处理？ |
| 熔断集成 | 成功/失败是否正确报告给 KeyPool？ |

### 3. 流式处理审核（重点）

| 检查项 | 说明 |
|--------|------|
| SSE 解析 | `data: ` 前缀处理正确？ |
| Chunk 处理 | 每个 chunk 是否正确转换？ |
| 结束检测 | finishReason 检测正确？ |
| Channel 关闭 | goroutine 是否正确退出？channel 是否正确关闭？ |
| Context 取消 | 是否响应 context.Done()？ |
| 资源泄漏 | Response.Body 是否正确关闭？ |

### 4. 并发安全审核

| 检查项 | 说明 |
|--------|------|
| Client 并发使用 | 多个 goroutine 同时调用是否安全？ |
| Stream Channel | 是否存在竞态条件？ |
| Context 传播 | 是否正确传播 Context？ |

### 5. 错误处理审核

| 错误场景 | 期望行为 |
|----------|---------|
| KeyPool 无可用 Key | 返回 `ErrNoAvailableKeys` |
| API 返回 429 | 调用 `ReportFailure`，触发熔断 |
| API 返回 400/401/500 | 返回适当的 `AppError` |
| 网络超时 | 返回超时错误 |
| 响应解析失败 | 返回解析错误 |
| 流中途断开 | 正确关闭 channel，返回错误 |

### 6. 测试完整性审核

| 测试类型 | 说明 |
|----------|------|
| Mock Server | 是否正确模拟 Gemini API？ |
| 阻塞请求测试 | 简单文本、多模态 |
| 流式请求测试 | 多 chunk、finishReason、中途断开 |
| 错误场景测试 | 429、500、超时、解析失败 |
| KeyPool 集成测试 | Success/Failure 报告 |
| Context 取消测试 | 流式请求中途取消 |

## 产出

创建审核报告 `docs/CodeReviewReport/GEMINI_CLIENT_REVIEW_REPORT.md`：

```markdown
# Gemini API 客户端审核报告

> **审核员**: QA Engineer
> **审核日期**: YYYY-MM-DD
> **审核范围**: `internal/gemini/client.go`, `client_test.go`

---

## 审核结果：✅ 通过 / ⚠️ 需修复 / ❌ 严重问题

| 指标 | 数量 |
|------|------|
| 严重问题 | X |
| 警告 | X |
| 建议改进 | X |
| 测试覆盖率 | X% |

---

## 各功能审核

### 阻塞式请求 (ChatCompletion)
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

### 流式请求 (ChatCompletionStream)
- 状态：✅ / ⚠️ / ❌
- SSE 解析：✅ / ⚠️ / ❌
- Channel 管理：✅ / ⚠️ / ❌
- 问题（如有）：

### KeyPool 集成
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

### 错误处理
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
- 重点关注流式处理的正确性和资源管理
- 验证 KeyPool 集成的正确性
- 检查是否有 goroutine 泄漏风险
- 如发现严重问题，需提供具体修复方案

---

*任务创建时间: 2026-01-15*
