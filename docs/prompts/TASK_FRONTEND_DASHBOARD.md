# 任务：前端可视化看板开发 (Dashboard)

## 角色
Frontend Visual Specialist & UI Designer

## 必备技能
- **UI/UX Pro Max**: 必须阅读并应用 `.agent/skills/ui-ux-pro-max/SKILL.md`。

## 背景
请先阅读 **`docs/FRONTEND_PROJECT.md`** 了解当前项目结构。
任务目标是创建一个直观的仪表盘 (`Dashboard.vue`)。

## 任务目标
使用 ECharts 展示数据，注重美观和响应式体验。

## 详细步骤

### 1. 可视化设计确认 (Design First) 🗣️
- **应用 Skill**: 使用 `ui-ux-pro-max` 设计高颜值的仪表盘。
- **生成预览**: 使用 `generate_image` 生成 Dashboard 的视觉 Mockup（展示暗色模式下的图表配色）。
- **沟通**: 询问用户对图表类型（折线 vs 柱状）、配色方案的偏好。
- **获得批准**: 只有在用户满意设计方案后，才开始集成 ECharts。

### 2. 依赖安装
- 确认已安装 `echarts` 和 `vue-echarts`。
- 按需引入 ECharts 组件（LineChart, PieChart, Tooltip, Grid 等）以减小体积。

### 2. Dashboard 页面结构
- **顶部卡片组**:
  - **总请求数**: 数字展示。
  - **Token 消耗**: Prompt / Completion / Total。
  - **系统状态**: 在线时长 (Uptime)、活动 Key 数量。
  - **健康度**: 根据 `/health` 接口展示 (绿色/黄色/红色)。

### 3. 图表开发
- **Token 消耗趋势图 (Line Chart)**:
  - 暂时使用模拟数据或后端提供的按时间维度的统计接口（如果 `/api/stats` 仅提供总量，则只展示今日增量或总量）。
  - *注：后端当前 `/api/stats` 可能仅提供汇总数据，若需趋势图可能需要后端支持历史查询，此处先实现总量展示即可，或展示 key 的使用比例饼图。*
- **Key 使用分布 (Pie Chart)**:
  - 根据 `/api/stats/keys` 返回的数据。
  - 展示各 Key 的请求占比或 Token 消耗占比。

### 4. 实时性优化
- 设置定时器 (每 30s) 刷新 Dashboard 数据。
- 或者实现 SSE 监听（如果后端支持实时推送统计，当前 API.md 暂未涉及，故使用轮询）。

---

## 产出物
- 一个美观、响应式的 Dashboard 页面。

## 约束
- **移动端适配**: 图表在手机端应自动调整大小或隐藏复杂细节。
- **暗色模式**: 图表颜色需适配深色主题 (ECharts dark theme)。
