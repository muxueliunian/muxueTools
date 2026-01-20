# 任务：实现国际化 (i18n) 支持

## 角色
Frontend (ui-ux-pro-max skill)

## Skills 依赖
- `.agent/skills/ui-ux-pro-max/SKILL.md`

## 背景

MuxueTools 桌面应用当前仅支持中文界面。为了扩大用户群体，需要实现多语言支持。

**已完成的依赖模块：**（参见 `docs/FRONTEND_PROJECT.md`）
- 所有页面视图已开发完成
- 前端项目结构稳定
- 桌面应用打包流程已建立

**当前问题：**
1. 所有 UI 文本均为硬编码中文
2. 无语言切换机制
3. 国际用户无法使用

## 目标

| Phase | 目标 | 难度 |
|-------|------|------|
| **Phase 1** | 配置 vue-i18n 框架 | ⭐ 低 |
| **Phase 2** | 创建三语言翻译文件（中/英/日） | ⭐⭐ 中 |
| **Phase 3** | 实现语言切换器 UI | ⭐ 低 |
| **Phase 4** | 所有页面文本国际化 | ⭐⭐ 中 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范**
   - `.agent/skills/ui-ux-pro-max/SKILL.md`

2. **项目文档**
   - `docs/ARCHITECTURE.md` - 系统架构
   - `docs/FRONTEND_PROJECT.md` - 前端项目结构
   - `docs/FRONTEND_WORKFLOW.md` - 前端开发流程

3. **相关代码**
   - `web/src/App.vue` - 根组件
   - `web/src/layouts/MainLayout.vue` - 主布局
   - `web/src/views/*.vue` - 所有页面视图

---

### Phase 1: 配置 vue-i18n 框架

1. 安装依赖：
   ```bash
   cd web
   npm install vue-i18n@9
   ```

2. 创建 i18n 配置文件 `web/src/i18n/index.ts`：
   - 初始化 vue-i18n
   - 实现浏览器语言自动检测
   - localStorage 持久化
   - 设置 fallback 语言为英语

3. 在 `main.ts` 中集成 i18n：
   ```typescript
   import { i18n } from './i18n'
   app.use(i18n)
   ```

---

### Phase 2: 创建翻译文件

创建三个语言文件，按模块组织翻译内容：

| 文件 | 语言 |
|------|------|
| `web/src/i18n/locales/zh-CN.json` | 简体中文 |
| `web/src/i18n/locales/en-US.json` | 英语 |
| `web/src/i18n/locales/ja-JP.json` | 日语 |

**翻译内容结构：**
```json
{
  "common": { /* 通用词汇 */ },
  "sidebar": { /* 侧边栏 */ },
  "chat": { /* 对话页面 */ },
  "dashboard": { /* 仪表盘 */ },
  "keys": { /* API Keys 管理 */ },
  "stats": { /* 统计页面 */ },
  "settings": { /* 设置页面 */ },
  "modelSettings": { /* 模型设置 */ }
}
```

---

### Phase 3: 实现语言切换器

1. 在 `MainLayout.vue` 侧边栏底部添加语言切换器
2. 使用 Naive UI 的 `NDropdown` 或 `NSelect` 组件
3. 显示语言图标或当前语言名称
4. 切换时调用 `setLocale()` 并立即生效

**可选语言：**
| 语言 | 代码 | 显示名称 |
|------|------|----------|
| 简体中文 | zh-CN | 简体中文 |
| English | en-US | English |
| 日本語 | ja-JP | 日本語 |

---

### Phase 4: 页面国际化

按优先级国际化以下页面：

| 优先级 | 文件 | 说明 |
|--------|------|------|
| P0 | `MainLayout.vue` | 侧边栏导航项 |
| P0 | `ChatView.vue` | 对话页面 |
| P0 | `chat/*.vue` | 对话组件 |
| P1 | `DashboardView.vue` | 仪表盘 |
| P1 | `KeyManagerView.vue` | Key 管理 |
| P2 | `StatsView.vue` | 统计页面 |
| P2 | `SettingsView.vue` | 设置页面 |
| P2 | `ModelSettingsView.vue` | 模型设置 |

**替换规则：**
- 硬编码文本 → `{{ $t('key') }}`
- script 中使用 `const { t } = useI18n()`
- 动态文本使用参数 `$t('key', { name: value })`

---

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `web/src/i18n/index.ts` | **NEW** | vue-i18n 配置 |
| `web/src/i18n/locales/zh-CN.json` | **NEW** | 中文翻译 |
| `web/src/i18n/locales/en-US.json` | **NEW** | 英语翻译 |
| `web/src/i18n/locales/ja-JP.json` | **NEW** | 日语翻译 |
| `web/src/main.ts` | **MODIFY** | 集成 i18n |
| `web/src/layouts/MainLayout.vue` | **MODIFY** | 添加语言切换器 |
| `web/src/views/*.vue` | **MODIFY** | 文本国际化 |
| `web/src/components/chat/*.vue` | **MODIFY** | 文本国际化 |

---

## 约束

### 技术约束
- 使用 vue-i18n v9 (Composition API)
- 保持与 Naive UI 的兼容性
- 翻译文件使用 JSON 格式

### 质量约束
- 遵循 `.agent/skills/ui-ux-pro-max/SKILL.md` 代码规范
- 所有用户可见文本必须国际化
- 无硬编码文本残留

### 兼容性约束
- 保持现有功能不受影响
- 语言切换 **无需刷新页面**
- 支持浏览器语言自动检测

---

## 验收标准

- [ ] `npm run dev` 启动无错误
- [ ] 侧边栏底部显示语言切换器
- [ ] 可切换中文/英语/日语
- [ ] 切换语言后立即生效，无需刷新
- [ ] 刷新页面后语言选择保持不变
- [ ] 首次访问自动检测浏览器语言
- [ ] 所有页面无硬编码中文残留
- [ ] 三种语言翻译完整
- [ ] `npm run build` 构建成功

---

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/FRONTEND_PROJECT.md` | 新增 i18n 模块说明 |
| `docs/FRONTEND_TASKS.md` | 标记任务 6 为已完成 |
| `docs/README.md` | 更新待开发任务列表 |

---

## 开发流程

遵循 `docs/FRONTEND_WORKFLOW.md` 中的 Design-First + Component-Driven 流程。

---

## 参考资料

| 资源 | 链接 |
|------|------|
| vue-i18n 官方文档 | https://vue-i18n.intlify.dev/ |
| Naive UI i18n | https://www.naiveui.com/zh-CN/os-theme/docs/i18n |
| Vue 3 Composition API | https://vuejs.org/guide/extras/composition-api-faq.html |

---

*任务创建时间: 2026-01-20*
