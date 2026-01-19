# MuxueTools v1.0.0 前端改进任务

> 创建时间: 2026-01-18
> 状态: 进行中

---

## 📋 任务总览

| # | 任务 | 页面 | 优先级 | 状态 |
|---|------|------|--------|------|
| 1 | [API Keys 页面优化](#任务-1-api-keys-页面优化) | KeyManagerView | P0 | ✅ 已完成 |
| 2 | [Chat 界面改进](#任务-2-chat-界面改进) | ChatView | P0 | ✅ 已完成 |
| 3 | [Statistics 页面实现](#任务-3-statistics-页面实现) | StatsView | P1 | ✅ 已完成 |
| 4 | [Settings 页面完善](#任务-4-settings-页面完善) | SettingsView | P1 | ⏳ 待开发 |
| 5 | [Model Settings 页面](#任务-5-model-settings-页面新增) | ModelSettingsView | P2 | ⏳ 待开发 |
| 6 | [国际化 (i18n)](#任务-6-国际化-i18n) | 全局 | P1 | ⏳ 待开发 |
| 7 | [README 使用文档](#任务-7-readme-使用文档) | - | P0 | ⏳ 待开发 |
| 8 | [v1.0.0 发布](#任务-8-v100-发布) | - | P0 | ⏳ 待执行 |

---

## 任务 1: API Keys 页面优化

**文件**: `web/src/views/KeyManagerView.vue`

### 问题清单
- [x] 删除 MODEL 列（显示不完整且不必要）
- [x] 调整整体字号和间距
- [x] 实现搜索功能（按 Key 名称/标签过滤）
- [x] 实现导入功能（批量导入 Keys）
- [x] UI 细节打磨

### 参考截图
![API Keys 当前状态](../../image/api_keys_current.png)

---

## 任务 2: Chat 界面改进

**文件**: `web/src/views/ChatView.vue`, `web/src/components/chat/*`

### 问题清单
- [x] 模型选择列表显示完整模型名称（当前被截断）
- [x] 支持上传图片/文件（Gemini Vision）
- [x] 输入框增加图片上传按钮
- [x] 消息中显示已上传的图片

### 参考截图
![Chat 模型选择](../../image/chat_model_selector.png)

---

## 任务 3: Statistics 页面实现

**文件**: `web/src/views/StatsView.vue`

### 功能需求
- [x] 总请求数、成功率统计卡片
- [x] 总 Token 消耗统计
- [x] 按日期的请求趋势图 (ECharts)
- [x] 各 Key 使用分布饼图
- [x] 最近请求记录表格

### 新增文件
- `web/src/stores/statsStore.ts`
- `web/src/api/stats.ts`

### 参考截图
![Statistics 当前状态](../../image/stats_current.png)

---

## 任务 4: Settings 页面完善

**文件**: `web/src/views/SettingsView.vue`

### 问题清单
- [ ] Security 标签页面内容设计
- [ ] Advanced 标签页面内容设计
- [ ] UI 细节调整（间距、对齐等）
- [ ] 配置保存功能验证

### 参考截图
![Settings 当前状态](../../image/settings_current.png)

---

## 任务 5: Model Settings 页面（新增）

**新增文件**: `web/src/views/ModelSettingsView.vue`

### 功能需求
- [ ] 新增路由 `/model-settings`
- [ ] 侧边栏添加导航项
- [ ] 模型参数配置：
  - temperature (温度)
  - max_tokens (最大输出长度)
  - top_p
  - top_k
- [ ] 参数保存到 localStorage 或后端
- [ ] Chat 界面使用这些默认参数

### 设计说明
此页面用于设置 MuxueTools **自带 Chat 界面**的默认模型参数。
反向代理 API 不受此影响，客户端仍可传递自己的参数。

---

## 任务 6: 国际化 (i18n)

**涉及文件**: 全局

### 功能需求
- [ ] 安装 vue-i18n
- [ ] 创建语言文件 (`locales/zh-CN.json`, `locales/en-US.json`)
- [ ] 侧边栏添加语言切换按钮
- [ ] 所有页面文本国际化
- [ ] 语言偏好 localStorage 持久化

### 新增文件
- `web/src/locales/zh-CN.json`
- `web/src/locales/en-US.json`
- `web/src/i18n.ts`

---

## 任务 7: README 使用文档

**文件**: `README.md`

### 内容清单
- [ ] 项目简介
- [ ] 功能特性列表
- [ ] 快速开始指南
- [ ] 配置文件说明
- [ ] 应用截图
- [ ] 开发者说明

---

## 任务 8: v1.0.0 发布

### 发布清单
- [ ] 所有前端任务完成
- [ ] 构建最终版本 (`scripts/build.ps1`)
- [ ] 验收测试通过
- [ ] 创建 GitHub Release
- [ ] 上传 exe 文件
- [ ] 编写 Release Notes

---

## 📅 进度记录

| 日期 | 完成内容 |
|------|----------|
| 2026-01-18 | 创建任务文档 |
| 2026-01-18 | 完成任务 1: API Keys 页面优化 |
| 2026-01-18 | 完成任务 2: Chat 界面改进 |
| 2026-01-18 | 完成任务 3: Statistics 页面实现 |

---

*最后更新: 2026-01-18*
