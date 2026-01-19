# 任务：格式转换模块 Code Review

## 角色
QA Engineer (qa-automation skill)

## 背景
Developer Agent 已完成格式转换模块的 TDD 开发。该模块负责 OpenAI ↔ Gemini 格式的双向转换，是代理服务的核心逻辑。

## 审核范围
- `internal/gemini/converter.go` - 格式转换核心实现
- `internal/gemini/converter_test.go` - 转换测试

## 步骤

### 1. 阅读规范
- `.agent/skills/qa-automation/SKILL.md` - QA 审核规范
- `.agent/skills/senior-golang/SKILL.md` - Go 代码规范
- `docs/ARCHITECTURE.md` - Converter 设计规范
- `internal/types/openai.go` - OpenAI 类型定义
- `internal/types/gemini.go` - Gemini 类型定义

### 2. 功能正确性审核

| 检查项 | 说明 |
|--------|------|
| OpenAI → Gemini 请求转换 | 是否正确转换所有字段？ |
| Gemini → OpenAI 响应转换 | 是否保留所有必要信息？ |
| 多模态内容处理 | 图片（base64/URL）是否正确处理？ |
| system 消息处理 | 是否正确转换为 Gemini 的 systemInstruction？ |
| 模型名称映射 | 是否覆盖常用模型？未知模型如何处理？ |
| stop 序列转换 | string 和 []string 是否都能正确处理？ |
| Usage 统计转换 | Token 计数是否正确映射？ |
| 流式响应转换 | Chunk 格式是否符合 OpenAI SSE 规范？ |

### 3. 边界条件审核

| 检查项 | 说明 |
|--------|------|
| 空消息列表 | 是否返回合适的错误？ |
| nil 输入 | 是否有 panic 风险？ |
| 超大内容 | 是否有内存问题？ |
| 不支持的内容类型 | 是否返回明确错误？ |
| 特殊字符 | Unicode、emoji 是否正确处理？ |

### 4. 代码质量审核

| 检查项 | 说明 |
|--------|------|
| 纯函数 | 是否无副作用？是否依赖全局状态？ |
| 错误处理 | 是否使用 `types/errors.go` 中的预定义错误？ |
| 代码可读性 | 命名是否清晰？逻辑是否易懂？ |
| 避免重复 | 是否有重复代码可以抽取？ |
| Godoc 注释 | 导出函数是否有文档？ |

### 5. 测试完整性审核

| 检查项 | 说明 |
|--------|------|
| 基础转换测试 | 简单文本、多轮对话、system 消息 |
| 多模态测试 | base64 图片、URL 图片、多图片 |
| 参数转换测试 | temperature, topP, maxTokens, stop |
| 响应转换测试 | 阻塞响应、流式 chunk |
| 边界条件测试 | 空输入、nil、不支持类型 |
| Benchmark | 是否包含性能基准测试？ |
| 覆盖率 | 是否 >85%？ |

### 6. 与规范一致性

对照 `docs/ARCHITECTURE.md` 中的 Converter 设计：
- 接口签名是否一致？
- 模型映射是否完整？
- 错误处理是否符合规范？

## 产出

创建审核报告 `docs/CONVERTER_REVIEW_REPORT.md`：

```markdown
# 格式转换模块审核报告

> **审核员**: QA Engineer
> **审核日期**: YYYY-MM-DD
> **审核范围**: `internal/gemini/converter.go`, `*_test.go`

---

## 审核结果：✅ 通过 / ⚠️ 需修复 / ❌ 严重问题

| 指标 | 数量 |
|------|------|
| 严重问题 | X |
| 警告 | X |
| 建议改进 | X |

---

## 各文件审核

### converter.go

**功能正确性**: ✅ / ⚠️ / ❌

| 功能 | 状态 | 说明 |
|------|------|------|
| OpenAI → Gemini 请求 | ✅/⚠️/❌ | |
| Gemini → OpenAI 响应 | ✅/⚠️/❌ | |
| 多模态处理 | ✅/⚠️/❌ | |
| system 消息 | ✅/⚠️/❌ | |
| 模型映射 | ✅/⚠️/❌ | |
| 流式转换 | ✅/⚠️/❌ | |

**代码质量**: ✅ / ⚠️ / ❌
- 问题（如有）：
- 建议修复：

### converter_test.go

**测试覆盖**: ✅ / ⚠️ / ❌

| 测试类型 | 状态 | 说明 |
|----------|------|------|
| 基础转换 | ✅/⚠️/❌ | |
| 多模态 | ✅/⚠️/❌ | |
| 边界条件 | ✅/⚠️/❌ | |
| Benchmark | ✅/⚠️/❌ | |

**缺失的测试场景**:
- （如有）

---

## 问题详情

### ❌ 严重问题（如有）
（问题描述、影响、建议修复方案）

### ⚠️ 警告（如有）
（问题描述、建议）

### 💡 改进建议（如有）
（优化建议）

---

## 总结

**结论**: 继续开发 / 需先修复

---

*报告生成时间: YYYY-MM-DD HH:MM*
```

## 约束
- 重点关注格式转换的正确性
- 验证与 OpenAI 官方 API 规范的兼容性
- 关注边界条件和异常处理
- 如发现严重问题，需提供具体修复方案

---

*任务创建时间: 2026-01-15*
