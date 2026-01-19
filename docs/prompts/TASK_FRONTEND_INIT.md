## 任务：前端项目初始化

## 角色
Senior Frontend Developer (Vue3 Ecosystem Expert) & UI/UX Designer

## 必备技能
- **UI/UX Pro Max**: 必须阅读并应用 `.agent/skills/ui-ux-pro-max/SKILL.md`。

## 背景
MxlnAPI 后端已就绪。前端开发遵循 `docs/FRONTEND_WORKFLOW.md`，强调 **Design First**。
你的任务是初始化 Web 前端项目，但在写代码前，**必须先与用户沟通确认设计风格**。

## 任务目标
构建一个基于 Vite + Vue3 + Naive UI 的前端项目骨架。

## 详细步骤

### 1. 设计风格确认 (Interactive) 🗣️
- **加载 Skill**: 读取 `ui-ux-pro-max` 技能文件。
- **沟通**: 向用户询问偏好的 UI 风格（例如：主色调、圆角风格、紧凑度）。
- **确认**: 确认移动端适配和独立窗口的特殊要求（无边框等）。
- **等待用户批准**：只有在用户明确确认设计方向后，才继续进行技术初始化。

### 2. 项目脚手架
- 在 `web/` 目录下初始化项目。
- 技术栈：**Vue 3 (TypeScript) + Vite**。
- 安装依赖：
  - `naive-ui` & `vfonts` (字体)
  - `vue-router` (版本 4)
  - `pinia` (状态管理)
  - `axios` (HTTP 请求)
  - `@vicons/ionicons5` (图标库)

### 2. 工程化配置
- **Vite 配置** (`vite.config.ts`)：
  - 配置别名 `@` 指向 `src/`。
  - 配置 Server Proxy：将 `/api` 和 `/v1` 代理到 `http://localhost:8080`.
- **TypeScript 配置**：确保严格模式启用。

### 3. API 层基础 (Contract-First)
- 阅读 `docs/API.md`。
- 创建 `src/api/types.ts`：定义 API 契约中的所有 TypeScript 接口（如 `KeyInfo`, `StatsResponse` 等）。
- 创建 `src/api/client.ts`：封装 Axios 实例，配置基础 URL 和响应拦截器（统一处理错误）。

### 4. 核心架构
- **Router** (`src/router/index.ts`)：
  - 定义路由表：
    - `/` -> Dashboard
    - `/keys` -> KeyManager
    - `/stats` -> Stats
    - `/settings` -> Settings
- **Store** (`src/stores/`)：
  - 初始化 Pinia。
  - 创建 `useGlobalStore` 用于管理全局状态（如主题模式、侧边栏收折）。

### 5. 基础布局 (App Layout)
- 创建 `src/App.vue`。
- 使用 `NConfigProvider` 配置 Naive UI 全局主题（默认深色）。
- 实现 **Sidebar Layout**：
  - 左侧：导航菜单 (Dashboard, Keys, Stats, Settings)。
  - 右侧：顶部 Header + 内容区域 (`RouterView`)。
- **设计确认**：遵循 `docs/FRONTEND_WORKFLOW.md` 中的设计规范（移动端响应式）。

---

## 产出物
- `web/` 目录下的完整工程文件。
- 可运行的开发服务器 (`pnpm dev`)，显示带有导航栏的空页面。

## 约束
- 严禁使用 `any` 类型。
- 必须使用 Composition API (`<script setup lang="ts">`)。
- 样式推荐使用 CSS Modules 或 Utility Classes。
- 确保代码符合 `docs/FRONTEND_WORKFLOW.md` 规范。
