# Windows 平台收尾任务

> **阶段**: Windows 平台开发最后阶段  
> **预计工期**: 1-2 天  
> **下一阶段**: Android 适配

---

## 📋 任务清单

### 1. 会话持久化 (优先级: P0)

**目标**: 实现对话历史的保存和恢复

#### 需求决策 (2026-01-16 已确认)

| 需求项 | 决策 | 备注 |
|--------|------|------|
| 会话标题 | 第一条消息截取 | 后续: AI ✨按钮生成摘要 |
| 删除交互 | 悬停显示删除按钮 | 预留: 移动端滑动删除 |
| 模型绑定 | 全局模型 | 所有会话共用 |

#### 后端 API (已就绪)
- [x] `GET /api/sessions` - 列出会话
- [x] `POST /api/sessions` - 创建会话
- [x] `GET /api/sessions/:id` - 获取会话详情
- [x] `PUT /api/sessions/:id` - 更新会话
- [x] `DELETE /api/sessions/:id` - 删除会话
- [x] `POST /api/sessions/:id/messages` - 添加消息

#### 前端实现 (待开发)
- [ ] `src/api/sessions.ts` - 会话 API 封装
- [ ] `src/api/types.ts` - Session/Message 类型
- [ ] `src/stores/sessionStore.ts` - 会话状态管理
- [ ] 侧边栏会话列表 UI
  - 显示历史会话 (标题=第一条消息截取)
  - 当前会话高亮
  - 悬停显示删除按钮
- [ ] 消息自动保存
  - 发送消息时同步到后端
  - 接收回复后保存助手消息
- [ ] 会话切换
  - 加载历史消息
  - 全局模型保持不变

**技术要点**:
- 使用 Pinia 管理会话状态
- 侧边栏 "New Chat" 按钮创建新会话
- 每条消息发送后立即持久化

#### 后续待开发 (不在本次范围)

| 功能 | 描述 |
|------|------|
| AI 摘要标题 | 选中会话右侧 ✨ 按钮，点击 AI 生成摘要 |
| 移动端滑动删除 | 触屏设备滑动删除会话 |
| Dashboard 高级参数 | System Prompt / Temperature / Thinking Level / Media Resolution |

---

### 2. App 图标设计 (优先级: P1)

**目标**: 设计并应用桌面应用图标

#### 设计要求
- 风格: 简洁现代，与 Claude 风格一致
- 颜色: 主色 `#D97757` (coral/橘红色)
- 尺寸要求:
  - Windows: 256x256 ICO (含多尺寸)
  - 可选: 16x16, 32x32, 48x48, 64x64, 128x128

#### 实现步骤
- [ ] 设计图标 (使用 generate_image 或外部工具)
- [ ] 转换为 ICO 格式
- [ ] 集成到 `cmd/desktop/main.go`
- [ ] 更新 WebView 窗口图标

---

### 3. UI 细节优化 (优先级: P1)

**目标**: 完善用户体验细节

#### Chat 页面
- [ ] 优化空状态提示
- [ ] 考虑添加快捷提示词
- [ ] 消息时间戳显示 (可选)
- [ ] 代码块复制按钮样式优化

#### 整体
- [ ] 检查暗色模式下的对比度
- [ ] 统一按钮 hover 效果
- [ ] 加载状态优化 (骨架屏/Spinner)
- [ ] 错误提示优化

---

## 🚀 Android 适配预研

### 技术方案评估

| 方案 | 优点 | 缺点 |
|------|------|------|
| **Go + WebView (Android)** | 代码复用度高 | 需要 gomobile/cgo |
| **纯 Web PWA** | 无需打包 | 无法访问原生功能 |
| **Flutter + 嵌入 WebView** | 成熟的移动端方案 | 需要额外学习 |

### 推荐路线
1. 首先确保前端响应式适配 (当前使用 Tailwind，基础已有)
2. 使用 Android WebView 包装现有前端
3. 后端编译为 Android ARM 二进制

### 需要调研
- [ ] Go 程序在 Android 上的运行方式
- [ ] WebView 与本地 Go 服务的通信
- [ ] 跨域问题处理
- [ ] 应用签名和分发

---

## 📁 相关文件

| 类型 | 路径 |
|------|------|
| 会话 API | `internal/api/session_handler.go` |
| 会话存储 | `internal/storage/sessions.go` |
| 桌面入口 | `cmd/desktop/main.go` |
| Chat 组件 | `web/src/views/ChatView.vue` |
| 侧边栏 | `web/src/layouts/MainLayout.vue` |

---

## ✅ 完成标准

- [ ] 可以创建、切换、删除会话
- [ ] 刷新页面后历史消息仍在
- [ ] App 有自定义图标
- [ ] UI 细节无明显问题
- [ ] 准备好 Android 适配计划文档
