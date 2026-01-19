# 任务：Settings 页面完善

## 角色
Frontend (ui-ux-pro-max) + Developer (senior-golang)

## Skills 依赖
- `.agent/skills/ui-ux-pro-max/SKILL.md`
- `.agent/skills/senior-golang/SKILL.md`

---

## 背景

Settings 页面已有基础骨架（`web/src/views/SettingsView.vue`），包含以下功能：
- **General 标签**：服务器配置（Host/Port）、Key 选择策略、日志级别、自动更新
- **Security 标签**：仅占位，无内容
- **Advanced 标签**：仅占位，无内容

**当前问题：**
1. System Configuration 区块冗余（Dashboard 已展示 API Endpoint）
2. Security 和 Advanced 标签页无内容
3. 配置修改无法持久化（后端未实现）
4. 更新检查为 Mock 实现

**已修复：**
- ✅ Port 显示为 0 的 Bug（`cmd/desktop/main.go` 更新实际端口到配置）

**已完成的依赖模块：**（参见 `docs/DEVELOPMENT.md`）
- `internal/config/` - 配置加载
- `internal/storage/` - SQLite 存储层
- `web/src/api/config.ts` - 配置 API

---

## 目标

| # | 目标 | 优先级 |
|---|------|--------|
| 1 | 删除 General 页 System Configuration 区块 | P1 |
| 2 | 实现配置 SQLite 持久化 + 热更新 | P1 |
| 3 | 完善 General 页更新服务（双轨并行）| P2 |
| 4 | 实现 Security 标签页 | P2 |
| 5 | 实现 Advanced 标签页 | P2 |

---

## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范**
   - `.agent/skills/ui-ux-pro-max/SKILL.md`
   - `.agent/skills/senior-golang/SKILL.md`

2. **项目文档**
   - `docs/FRONTEND_WORKFLOW.md` - 前端开发流程
   - `docs/API.md` - 配置相关 API

3. **相关代码**
   - `web/src/views/SettingsView.vue` - 当前实现
   - `web/src/api/config.ts` - 配置 API
   - `internal/api/admin_handler.go` - 后端配置处理器

---

### 阶段 1：General 标签页

#### 1.1 UI 调整
- [x] ~~System Configuration 区块~~ → **删除**（信息已在 Dashboard 展示）

#### 1.2 功能实现

| 功能 | 需要做的 | 状态 |
|------|----------|------|
| Selection Strategy | SQLite 持久化 + 热更新 | ✅ 已完成 |
| Log Level | SQLite 持久化 + 热更新 | ✅ 已完成 |
| Automatic Updates 开关 | SQLite 持久化 | ✅ 已完成 |
| 更新源选择 | 新增 UI + SQLite 持久化 | ✅ 已完成 |
| Check Now 按钮 | 双轨更新检查 | ✅ 已完成 |

#### 1.3 更新服务设计

采用**双轨并行**方案，用户可自行选择更新源：

| 更新源 | URL |
|--------|-----|
| **mxln 服务器** (默认) | `https://mxlnuma.space/muxueTools/update/latest.json` |
| **GitHub Releases** | `https://api.github.com/repos/muxueliunian/muxueTools/releases/latest` |

**支持平台：**
- Windows x64 (amd64)
- Windows x86 (386)
- Android APK

**服务器配置：**
- 域名：`mxlnuma.space`
- 服务器路径：`/www/wwwroot/mxlnuma.space/muxueTools/`
- GitHub 仓库：`https://github.com/muxueliunian/muxueTools`

**服务器部署步骤：**
1. [x] 创建 GitHub 仓库 `muxueliunian/muxueTools`
2. [x] 在宝塔创建目录：`/www/wwwroot/mxlnuma.space/muxueTools/update/`
3. [x] 通过 GitHub Actions 自动上传 `latest.json`
4. [x] 配置 Nginx location 规则 + CORS
5. [x] 测试访问 `https://mxlnuma.space/muxueTools/update/latest.json` ✅
6. [x] GitHub Actions CI/CD 自动化部署 (`.github/workflows/release.yml`)

---

### 阶段 2：Security 标签页

**状态**: ✅ 已完成 (2026-01-19)

#### 2.1 功能需求

| 功能 | 说明 | UI 设计 | 状态 |
|------|------|---------|------|
| IP 白名单开关 | 启用/禁用 IP 白名单验证 | Switch 开关 | ✅ |
| 白名单 IP 地址 | 自定义一个允许访问的 IP | Input 输入框 | ✅ |
| 自定义代理密钥 | 替换默认的 `sk-mxln-proxy-local` | Input + 重新生成按钮 | ✅ |
| 标签页切换 | General/Security/Advanced | 侧边栏导航 | ✅ |

#### 2.2 已完成工作

**后端**:
- [x] IP 白名单中间件 (`middleware.go`)
- [x] 代理密钥生成函数 (`GenerateProxyKey`)
- [x] GetConfig/UpdateConfig 扩展支持 security
- [x] `POST /api/config/regenerate-proxy-key` API

**前端**:
- [x] Security 标签页 UI
- [x] 标签页切换逻辑
- [x] 密钥显示/隐藏切换

**技术说明：**
- IP 白名单：后端中间件校验请求来源 IP
- 默认行为：关闭 IP 白名单时，允许所有 IP 访问
- 始终允许 localhost 访问，防止锁死
- 代理密钥格式：`sk-mxln-{16字符随机}`
- **中间件应用**: 已应用到 `/v1/*` 路由 (2026-01-19)

---

### 阶段 3：Advanced 标签页

**状态**: ✅ 已完成 (2026-01-19)

#### 3.1 功能需求

| 功能 | 说明 | UI 设计 | 默认值 | 状态 |
|------|------|---------|--------|------|
| **Cooldown 时间** | Key 触发 Rate Limit 后的冷却时间 | InputNumber（秒）| 3600 | ✅ |
| **最大重试次数** | 单次请求失败后换 Key 重试次数 | InputNumber | 3 | ✅ |
| **请求超时** | HTTP 请求超时时间 | InputNumber（秒）| 120 | ✅ |
| **调试模式** | 启用详细后端日志输出 | Switch 开关 | 关闭 | ✅ |
| **数据库路径** | 当前数据库文件位置（只读）| 只读 Input | - | ✅ |
| **删除聊天记录** | 清空所有聊天会话数据 | 危险按钮 | - | ✅ |
| **删除统计数据** | 清空所有统计数据 | 危险按钮 | - | ✅ |

#### 3.2 已完成工作

**后端**:
- [x] GetConfig/UpdateConfig 支持 `advanced.request_timeout`
- [x] `DELETE /api/sessions` 清空聊天记录
- [x] `DELETE /api/stats/reset` 重置统计数据

**前端**:
- [x] Performance Tuning UI (Cooldown/MaxRetries/Timeout)
- [x] Debug Mode 开关
- [x] Database Location 显示
- [x] 危险操作确认弹窗

**默认值变更：**
- Cooldown: 60秒 → **3600秒 (1小时)** ✅ 已实现

---

### 阶段 4：后端实现

#### 4.1 配置持久化
- SQLite 表结构设计（`app_config` 表）
- 启动时加载：config.yaml → 覆盖 SQLite 存储值
- 前端保存：写入 SQLite

#### 4.2 热更新支持
- Strategy：KeyPool 支持运行时切换
- LogLevel：logrus 支持动态修改
- Cooldown/MaxRetries：KeyPool 参数更新

#### 4.3 新增 API
- `POST /api/config/security` - 安全配置
- `DELETE /api/sessions` - 清空聊天记录
- `DELETE /api/stats` - 清空统计数据
- `GET /api/update/check` - 真正的更新检查（双源）

---

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `web/src/views/SettingsView.vue` | **MODIFY** | 完善三个标签页 |
| `web/src/api/config.ts` | **MODIFY** | 新增配置相关 API |
| `internal/api/admin_handler.go` | **MODIFY** | 配置持久化、更新检查 |
| `internal/storage/sqlite.go` | **MODIFY** | 新增配置表和方法 |
| `internal/types/config.go` | **MODIFY** | 新增 Security/Advanced 配置类型 |
| `configs/update/latest.json` | **NEW** | 更新服务 JSON 模板 ✅ 已创建 |

---

## 约束

### 技术约束
- 前端：Vue 3 + Naive UI
- 后端：Go 1.22+ / Gin / SQLite
- 配置持久化使用 SQLite（不写回 YAML）

### 质量约束
- 遵循 `.agent/skills/ui-ux-pro-max/SKILL.md` UI 设计规范
- 遵循 `.agent/skills/senior-golang/SKILL.md` 代码规范
- 暗色/亮色主题兼容

### 兼容性约束
- 保持现有 API 不变
- 向后兼容现有配置文件

---

## 验收标准

### General 标签页
- [ ] System Configuration 区块已删除
- [ ] Selection Strategy 保存后生效
- [ ] Log Level 保存后生效
- [ ] 更新源选择 UI 正常
- [ ] Check Now 按钮能正确检查更新（双源）

### Security 标签页
- [ ] IP 白名单开关正常
- [ ] 白名单 IP 配置保存生效
- [ ] 自定义代理密钥保存后，Dashboard 同步更新

### Advanced 标签页
- [ ] Cooldown/重试/超时配置保存生效
- [ ] 调试模式开关正常
- [ ] 数据库路径正确显示
- [ ] 删除聊天记录/统计数据按钮正常（含二次确认）

### 通用
- [ ] 标签页切换正常
- [ ] 暗色/亮色主题下显示正常
- [ ] 配置重启后仍保留（SQLite 持久化）

---

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/FRONTEND_TASKS.md` | 更新 Settings 页面任务状态 |
| `docs/API.md` | 新增配置相关 API 文档 |

---

## 开发流程

- **前端**：遵循 `docs/FRONTEND_WORKFLOW.md` 中的 Design-First + Component-Driven 流程
- **后端**：遵循 `docs/DEVELOPMENT.md` 中的 TDD 开发流程

---

## 讨论记录

| 日期 | 讨论内容 | 决策 |
|------|----------|------|
| 2026-01-18 | System Configuration 区块的用途 | **删除**：信息已在 Dashboard 页面展示 |
| 2026-01-18 | Port 显示为 0 的 Bug | **已修复**：`cmd/desktop/main.go` |
| 2026-01-18 | General 页面功能确认 | 确认 5 项功能需实现 |
| 2026-01-19 | 配置持久化方式 | **SQLite 数据库** |
| 2026-01-19 | 更新服务设计 | **双轨并行**：mxln 服务器 + GitHub |
| 2026-01-19 | 更新服务器配置 | mxlnuma.space/muxueTools/，GitHub: muxueliunian/muxueTools |
| 2026-01-19 | 支持平台 | Windows x64/x86 + Android APK |
| 2026-01-19 | Security 标签页 | IP 白名单（开关+单个IP）+ 自定义代理密钥 |
| 2026-01-19 | Advanced 标签页 | Cooldown(1h)/重试(3)/超时(120s) + 调试模式 + 删除按钮 |
| 2026-01-19 | 默认值变更 | Cooldown 从 60秒 改为 3600秒（1小时）|
| 2026-01-19 | **阶段一完成** | General 标签页配置持久化 + 热更新已实现 |
| 2026-01-19 | **阶段二完成** | Security 标签页 IP 白名单 + 代理密钥已实现 |
| 2026-01-19 | **阶段三完成** | Advanced 标签页性能调优 + 数据清理已实现 |
| 2026-01-19 | Request Timeout 可编辑 | 因 exe 打包用户无法编辑 config.yaml，改为前端可编辑 |
| 2026-01-19 | **BUG修复**: 配置保存格式 | handleSave 发送格式与后端不匹配，已修复 |
| 2026-01-19 | **BUG修复**: GetConfig 读取问题 | GetConfig 返回 config.yaml 值而非 SQLite 存储值，已修复 |
| 2026-01-19 | **BUG修复**: 暗色模式 Select 背景 | 下拉框背景不随主题切换，通过添加 dark 类到组件根元素修复 |
| 2026-01-19 | **IP 白名单中间件应用** | 应用到 `/v1/*` 路由，保护 OpenAI 兼容端点 |
| 2026-01-19 | **GitHub Actions CI/CD** | 配置自动构建和 FTP 上传，推送 tag 自动发布 |
| 2026-01-19 | **双源更新检查 API** | 实现 mxln 服务器和 GitHub Releases 双源检查 |
| 2026-01-19 | **API 文档更新** | 添加配置 API 和数据管理 API 文档 |

---

*任务创建时间: 2026-01-18*
*需求讨论完成时间: 2026-01-19*
*阶段一完成时间: 2026-01-19*
*阶段二完成时间: 2026-01-19*
*阶段三完成时间: 2026-01-19*
*状态: ✅ Settings 页面全部完成*

