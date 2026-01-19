# Task 文档格式规范

> 本文档规范了 MxlnAPI 项目中 Agent 任务文档（TASK_*.md）的标准格式。
> 所有新建的 Task 文档必须遵循此规范。

---

## 📚 前置阅读

在创建 Task 文档前，请先阅读以下项目文档：

| 文档 | 描述 |
|------|------|
| [ARCHITECTURE.md](../ARCHITECTURE.md) | 系统架构与目录结构 |
| [DEVELOPMENT.md](../DEVELOPMENT.md) | 开发方法论与标准工作流 |
| [FRONTEND_WORKFLOW.md](../FRONTEND_WORKFLOW.md) | 前端开发流程（Vue3） |
| [FRONTEND_PROJECT.md](../FRONTEND_PROJECT.md) | 前端项目状态与结构 |

**Skills 文件位置**: `.agent/skills/` 目录下

---

## 📋 文档结构

一个标准的 Task 文档应包含以下章节（按顺序）：

```markdown
# 任务：[任务名称]

## 角色
[Agent 角色] ([skill 名称])

## Skills 依赖
- [列出需要使用的 Skills 路径]

## 背景
[任务背景、已完成的前置工作、当前问题]

## 目标
[明确的任务目标列表]

## 步骤
### 阶段 0：阅读规范 (必须)
### [实施步骤]

## 约束
[技术约束、质量标准、兼容性要求]

## 产出文件
[列出需要创建/修改的文件]

## 验收标准
[可验证的完成条件]

## 交付文档
[任务完成后需要更新的文档]

## 开发流程
[引用 docs/DEVELOPMENT.md 或 docs/FRONTEND_WORKFLOW.md]

---
*任务创建时间: [日期]*
```

---

## 📝 各章节说明

### 1. 标题

```markdown
# 任务：实现 [功能名称]（[开发方法]）
```

示例：
- `# 任务：实现 Gemini API 客户端（TDD）`
- `# 任务：实现 SQLite 存储层（阶段 3.5）`
- `# 任务：窗口标题栏自定义`

---

### 2. 角色

明确分配执行此任务的 Agent 角色及其对应的 Skill。

```markdown
## 角色
Developer (senior-golang skill)
```

**可用 Skills：** 详见 `.agent/skills/` 目录

| 角色 | Skill 路径 | 适用场景 |
|------|------------|----------|
| Developer | `.agent/skills/senior-golang/SKILL.md` | Go 后端开发 |
| Frontend | `.agent/skills/ui-ux-pro-max/SKILL.md` | 前端开发、UI 设计 |
| Architect | `.agent/skills/architect/SKILL.md` | 架构设计、技术决策 |
| QA | `.agent/skills/qa-automation/SKILL.md` | 测试编写、质量验证 |

**组合角色示例：**
```markdown
## 角色
Developer (senior-golang) + Frontend (ui-ux-pro-max)
```

---

### 3. Skills 依赖

列出需要阅读的 Skill 文件路径。

```markdown
## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`
- `.agent/skills/ui-ux-pro-max/SKILL.md`
```

---

### 4. 背景

描述任务的上下文，包括：
- 为什么需要这个任务？
- 已完成的前置工作/依赖模块（参考 `docs/DEVELOPMENT.md` 中的开发进度）
- 当前存在的问题

```markdown
## 背景

MxlnAPI 桌面应用使用 `webview_go` 作为 WebView 封装层。当前存在以下问题：

1. **窗口图标问题**: 描述...
2. **标题栏按钮**: 描述...

**已完成的依赖模块：**（参见 `docs/DEVELOPMENT.md`）
- `internal/config/` - 配置加载
- `internal/keypool/` - Key 池管理
```

---

### 5. 目标

使用表格或列表明确任务目标。

**表格格式（适用于多阶段任务）：**
```markdown
## 目标

| Phase | 目标 | 难度 |
|-------|------|------|
| **Phase 1** | 实现功能 A | ⭐ 低 |
| **Phase 2** | 实现功能 B | ⭐⭐⭐ 高 |
```

**列表格式（适用于单阶段任务）：**
```markdown
## 目标
1. 实现 XXX 功能
2. 编写单元测试
3. 更新相关文档
```

---

### 6. 步骤

#### 6.1 阶段 0：阅读规范 (必须)

**每个任务必须以此开头**，引用需要阅读的文档：

```markdown
## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范**
   - `.agent/skills/senior-golang/SKILL.md`

2. **项目文档**
   - `docs/ARCHITECTURE.md` - 系统架构
   - `docs/DEVELOPMENT.md` - 开发工作流

3. **相关代码**
   - `internal/xxx/xxx.go` - 现有实现
```

#### 6.2 实施步骤

具体的实施步骤，可以按 Phase 或步骤编号组织。

---

### 7. 产出文件

使用表格列出所有需要创建或修改的文件（参考 `docs/ARCHITECTURE.md` 中的目录结构）：

```markdown
## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/xxx/new.go` | **NEW** | 新功能实现 |
| `internal/xxx/existing.go` | **MODIFY** | 集成新模块 |
```

操作类型：
- `NEW` - 新建文件
- `MODIFY` - 修改现有文件
- `DELETE` - 删除文件

---

### 8. 约束

列出技术、质量和兼容性约束。质量约束应引用对应的 Skill 规范。

```markdown
## 约束

### 技术约束
- 使用 `net/http` 标准库
- Go 版本 1.22+

### 质量约束
- 遵循 `.agent/skills/senior-golang/SKILL.md` 代码规范
- 测试覆盖率 > 80%

### 兼容性约束
- 保持现有 API 不变
```

---

### 9. 验收标准

使用 **可验证的** Checkbox 列表：

```markdown
## 验收标准

- [ ] `go test ./internal/xxx/...` 所有测试通过
- [ ] `go test -race ./internal/xxx/...` 无竞态问题
- [ ] 功能 A 正常工作
- [ ] 现有功能不受影响
```

---

### 10. 交付文档

列出任务完成后需要更新的文档：

```markdown
## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/ARCHITECTURE.md` | 新增模块架构说明 |
| `docs/DEVELOPMENT.md` | 更新开发进度 |
| `docs/API.md` | 新增 API 端点文档 |
```

---

### 11. 开发流程

直接引用项目标准工作流文档，不要重复描述：

**后端任务：**
```markdown
## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的 TDD 开发流程。
```

**前端任务：**
```markdown
## 开发流程

遵循 `docs/FRONTEND_WORKFLOW.md` 中的 Design-First + Component-Driven 流程。
```

---

### 12. 可选章节

根据任务复杂度，可添加以下章节：

- **风险与注意事项** - 表格列出风险、影响、缓解措施
- **技术方案评估** - 对比多个方案的优缺点
- **参考资料** - 外部链接或内部文档引用

---

## ✅ 检查清单

创建 Task 文档时，请确保：

- [ ] 标题格式正确：`# 任务：[名称]`
- [ ] 指定了 Agent 角色和 Skill
- [ ] 包含"阶段 0：阅读规范"章节
- [ ] Skills/工作流/项目结构等内容引用现有文档，而非自行描述
- [ ] 明确了所有需要创建/修改的文件
- [ ] 验收标准是可验证的（使用 Checkbox）
- [ ] 列出了需要更新的交付文档
- [ ] 包含任务创建时间

---

## 📁 文件命名规范

- 路径: `docs/prompts/TASK_*.md`
- 命名: `TASK_[功能名称].md` (大写 + 下划线分隔)
- 示例:
  - `TASK_GEMINI_CLIENT.md`
  - `TASK_STORAGE_LAYER.md`
  - `TASK_WINDOW_CUSTOMIZATION.md`

---

*文档创建时间: 2026-01-17*
