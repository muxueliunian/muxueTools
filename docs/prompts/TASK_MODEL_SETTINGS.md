# 任务：实现模型专业设置功能

## 角色
Developer (senior-golang) + Frontend (ui-ux-pro-max)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`
- `.agent/skills/ui-ux-pro-max/SKILL.md`

## 背景

MuxueTools 是一个 AI Studio (Gemini) → OpenAI 格式的反向代理工具，目前已完成核心代理功能和前端 Settings 页面。

**用户需求**：在 Settings 页面添加模型专业设置，类似 Google AI Studio 的配置面板，支持全局默认的生成参数配置。

**已完成的依赖模块**（参见 `docs/DEVELOPMENT.md`）：
- `internal/types/config.go` - 配置类型定义
- `internal/types/gemini.go` - Gemini API 类型定义
- `internal/gemini/converter.go` - OpenAI ↔ Gemini 格式转换
- `web/src/views/SettingsView.vue` - Settings 页面 UI

**当前状态**：
- `GeminiGenerationConfig` 已支持 temperature、topP、topK、maxOutputTokens
- 需要扩展支持 thinkingConfig 和 mediaResolution
- 需要添加全局 System Prompt 配置

## 目标

| Phase | 目标 | 难度 |
|-------|------|------|
| **Phase 1** | 后端类型定义扩展 | ⭐ 低 |
| **Phase 2** | 后端配置 API 扩展 | ⭐ 低 |
| **Phase 3** | 设置应用到 Gemini 请求 | ⭐⭐ 中 |
| **Phase 4** | 前端 Settings UI 实现 | ⭐⭐ 中 |
| **Phase 5** | 验证测试 | ⭐ 低 |

**待实现的 7 项设置**：
1. System Prompt（系统提示词）
2. Temperature（温度 0-2）
3. Max Output Tokens（最大输出长度）
4. Top-P（核采样 0-1）
5. Top-K（Top-K 采样）
6. Thinking Level（思考等级：Low/Medium/High）
7. Media Resolution（媒体分辨率）

## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范**
   - `.agent/skills/senior-golang/SKILL.md`
   - `.agent/skills/ui-ux-pro-max/SKILL.md`

2. **项目文档**
   - `docs/ARCHITECTURE.md` - 系统架构
   - `docs/DEVELOPMENT.md` - 开发工作流
   - `docs/FRONTEND_WORKFLOW.md` - 前端开发流程
   - `docs/gemini/API_REFERENCE.md` - Gemini API 参考

3. **相关代码**
   - `internal/types/config.go` - 配置类型定义
   - `internal/types/gemini.go` - Gemini 类型定义
   - `internal/gemini/converter.go` - 格式转换器
   - `web/src/views/SettingsView.vue` - Settings 页面
   - `web/src/api/config.ts` - 配置 API 类型

### Phase 1：后端类型定义

1. 扩展 `internal/types/gemini.go`：
   - 添加 `ThinkingConfig` 结构体
   - 扩展 `GeminiGenerationConfig` 添加 `ThinkingConfig` 和 `MediaResolution` 字段

2. 扩展 `internal/types/config.go`：
   - 添加 `ModelSettingsConfig` 类型
   - 添加 `DefaultModelSettingsConfig()` 函数
   - 扩展 `Config` 结构体添加 `ModelSettings` 字段
   - 扩展 `ConfigData` 添加 `ModelSettings` 字段

### Phase 2：后端配置 API

1. 扩展 `internal/api/admin_handler.go`：
   - 更新 `GetConfig` 返回 model_settings
   - 更新 `UpdateConfig` 接受 model_settings 更新

### Phase 3：设置应用到 Gemini 请求

1. 修改 `internal/gemini/converter.go`：
   - 修改 `ConvertOpenAIRequest` 函数，接受全局配置参数
   - 修改 `convertGenerationConfig` 函数，应用全局默认值
   - 实现 System Prompt 注入逻辑

2. 更新调用处（如 `internal/api/openai_handler.go`）传入全局配置

### Phase 4：前端 Settings UI

1. 扩展 `web/src/api/config.ts`：
   - 添加 `ModelSettingsConfig` 接口
   - 扩展 `ConfigInfo` 添加 `model_settings` 字段

2. 修改 `web/src/views/SettingsView.vue`：
   - 添加 "Model" 标签页
   - 实现 System Prompt 多行文本输入
   - 实现 Temperature 滑块 + 输入框
   - 实现 Top-P、Top-K、Max Tokens 控件
   - 实现 Thinking Level、Media Resolution 下拉选择
   - 集成保存逻辑

### Phase 5：验证测试

1. 运行现有单元测试：`go test ./...`
2. 手动验证 Settings 页面 UI
3. 验证设置保存和应用到 Chat 请求

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/types/gemini.go` | **MODIFY** | 添加 ThinkingConfig、扩展 GeminiGenerationConfig |
| `internal/types/config.go` | **MODIFY** | 添加 ModelSettingsConfig 类型 |
| `internal/api/admin_handler.go` | **MODIFY** | 扩展配置读写 API |
| `internal/gemini/converter.go` | **MODIFY** | 应用全局设置到请求 |
| `internal/api/openai_handler.go` | **MODIFY** | 传入全局配置 |
| `web/src/api/config.ts` | **MODIFY** | 添加 ModelSettingsConfig 类型 |
| `web/src/views/SettingsView.vue` | **MODIFY** | 添加 Model 标签页 UI |

## 约束

### 技术约束
- Go 版本 1.22+
- 使用指针类型 (`*float64`, `*int`, `*string`) 区分「未设置」和「设置为默认值」
- 前端使用 Naive UI 组件库

### 质量约束
- 遵循 `.agent/skills/senior-golang/SKILL.md` 代码规范
- 遵循 `.agent/skills/ui-ux-pro-max/SKILL.md` UI 设计规范
- 保持现有项目 UI 风格一致（Claude-like 暗色主题）

### 兼容性约束
- 保持现有 `/api/config` API 向后兼容
- 新增字段使用 `omitempty` 标签
- 不影响现有 Chat 和 Proxy 功能

## 验收标准

- [ ] `go test ./...` 所有测试通过
- [ ] `go build ./...` 编译成功
- [ ] Settings 页面出现 "Model" 标签页
- [ ] System Prompt 可编辑并保存
- [ ] Temperature 滑块和输入框联动正常
- [ ] 所有设置项可保存并在刷新后保留
- [ ] Chat 发送消息时后端日志显示配置已应用
- [ ] Thinking Level 和 Media Resolution 下拉选择正常

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/README.md` | 更新开发进度 |
| `docs/API.md` | 文档化 model_settings 配置字段 |
| `docs/gemini/API_REFERENCE.md` | 添加 ThinkingConfig 说明 |

## 开发流程

- **后端**：遵循 `docs/DEVELOPMENT.md` 中的 TDD 开发流程
- **前端**：遵循 `docs/FRONTEND_WORKFLOW.md` 中的 Design-First + Component-Driven 流程

---

*任务创建时间: 2026-01-20*
