# MuxueTools 前端开发工作流

> 本文档规范了 MuxueTools 前端项目的开发流程、目录结构和代码规范。
> 采用 **Design First + Component-Driven** 开发模式。

---

## 🛠️ 技术栈
- **核心框架**: Vue 3 (Composition API) + TypeScript
- **构建工具**: Vite
- **UI 框架**: Naive UI (支持深色模式)
- **状态管理**: Pinia
- **HTTP 客户端**: Axios
- **CSS 方案**: TailwindCSS (可选) 或 CSS Modules
- **图表库**: ECharts (via vue-echarts)

---

## 📂 目录结构规范

```
web/src/
├── api/                # API 接口定义 (axios 封装)
│   ├── client.ts       # Axios 实例配置
│   ├── types.ts        # 全局 API 类型定义 (对应 API.md)
│   └── ...             # 各模块 API (keys.ts, chats.ts)
├── assets/             # 静态资源 (images, styles)
├── components/         # 通用/基础组件 (Atoms/Molecules)
│   ├── common/         # 全局通用组件
│   └── mobile/         # 移动端专用组件 (如有特殊差异)
├── composables/        # 组合式函数 (Hooks)
├── router/             # Vue Router 路由配置
├── stores/             # Pinia 状态管理
├── types/              # TypeScript 类型定义
├── views/              # 页面视图 (Page/Organisms)
│   ├── Dashboard/
│   ├── KeyManager/
│   └── ...
├── App.vue             # 根组件
└── main.ts             # 入口文件
```

---

## 🔄 开发循环 (Component-Driven)

### 1. 设计确认 (Design First) 🛑
**在编写代码之前**，必须先确认设计方案：

1. **加载 Skill**: 必须阅读并使用 `.agent/skills/ui-ux-pro-max/SKILL.md` 中的设计规范和流程。
2. **沟通确认**: 在开始设计前，**必须先询问用户**关于设计风格的偏好（如："偏好深色极简还是 Vibrant？" "针对什么设备优化？"）。
3. **Mockup 验证**:
   - 使用 Skill 指导或 `generate_image` 生成界面预览。
   - **必须获得用户明确批准**后方可编写代码。

**关注点**:
- **UI 风格**: 默认深色主题 (Dark Mode)，符合 Naive UI 风格。
- **适配检查**:
  - 📱 **移动端**: 是否使用响应式 Grid？按钮大小是否适合触控？
  - 🖥️ **独立窗口**: 是否隐藏了必要的导航？是否依赖了 `window.open` (禁止使用)？

### 2. API 类型定义
查阅 `docs/API.md`，在 `src/api/types.ts` 中定义 TypeScript 接口。
*Contract-First: 先定义类型，再写逻辑。*

```typescript
// src/api/types.ts
export interface KeyInfo {
    id: string;
    key: string;
    // ...
}
```

### 3. Store 状态实现
在 `src/stores/` 中建立对应的 Pinia Store。
- Store 负责调用 API 并管理数据状态。
- View 层**不直接调用 API**，而是调用 Store 的 Action。

### 4. 组件开发
- **基础组件**: 封装 Naive UI 组件或通用逻辑，不包含业务 Store。
- **业务组件**: 引入 Store，处理具体业务逻辑。

### 5. 页面组装
在 `src/views/` 中组装组件，配置路由。

---

## 📝 代码规范

### 组件命名
- 使用 PascalCase (大驼峰)。
- 文件名: `KeyList.vue`, `UsageChart.vue`。
- 基础组件前缀: `BaseButton.vue`, `AppCard.vue`。
- **视图组件**: 放在 `views/` 下，统一以 `View.vue` 结尾 (e.g., `ChatView.vue`)。

### TypeScript 规范
- **Props**: 使用 `defineProps<Props>()` 并在外部定义接口。
- **Emits**: 使用 `defineEmits<Emits>()` 明确事件类型。
- **No Any**: 严禁使用 `any`，必须定义清楚类型。如类型复杂，在 `types/` 下定义。
- **Strict Mode**: 必须开启 `strict: true`。

### 注释规范 (JSDoc)
为了避免“屎山代码”，关键逻辑必须包含清晰注释：

1.  **函数/方法**: 使用 JSDoc 格式，解释参数、返回值和副作用。
    ```typescript
    /**
     * 将 OpenAI 格式的消息转换为组件所需的 UI 格式
     * @param messages - 原始 API 消息数组
     * @returns 格式化后的 UI 消息对象
     */
    function formatMessages(messages: Message[]): UIMessage[] { ... }
    ```

2.  **复杂逻辑**: 在 `if/else`, `for`, 复杂计算等代码块前，用单行注释解释“为什么这么做”。
    ```typescript
    // 因为 Claude API 对 temperature 范围限制不同，这里需要做归一化处理
    const adjustedTemp = Math.min(temp, 1.0);
    ```

3.  **组件顶层**: 在 `<script setup>` 顶部简述组件职责。
    ```vue
    <!-- KeyManagerView.vue -->
    <script setup lang="ts">
    /**
     * 密钥管理视图
     * 职责: 列出 API Key，提供创建、删除和测试 Key 的功能。
     * 依赖: KeyStore
     */
    </script>
    ```

### 架构与防腐
- **逻辑抽离**: 超过 200 行的 `.vue` 文件，必须考虑将逻辑抽离到 `composables/` (Hooks)。
- **API 解耦**: UI 组件**严禁**直接调用 `axios` 或 `fetch`。必须通过 `stores/` 或 `api/` 层调用。
- **常量管理**: 魔法数字/字符串（如 API 路径、默认配置值）必须提取到 `constants.ts` 或 `config.ts`。

### 移动端与独立窗口适配指南
- **视口 (Viewport)**: 确保 `<meta name="viewport">` 正确设置。
- **布局**: 优先使用 Flex 和 Grid 实现响应式。
- **交互**: 
  - 避免 hover 交互作为唯一操作方式（触屏无 hover）。
  - 点击区域 (Tap Target) 至少 44x44px。
- **WebView 限制**:
  - 不要使用 `alert()`, `confirm()` (UI 可能会被系统拦截或样式不统一)，使用 Naive UI 的 `useMessage`, `useDialog`。
  - 不要使用 `window.open()` 打开新标签页，应在原窗口跳转或使用模态框。

---

## 🚀 常用命令

```bash
pnpm install
pnpm dev      # 启动本地开发服务器
pnpm build    # 构建生产版本 (生成的 dist 目录将被 Go embed)
pnpm type-check # TypeScript 类型检查
```
