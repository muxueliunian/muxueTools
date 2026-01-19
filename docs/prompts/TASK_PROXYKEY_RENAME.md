# 任务：同步 Proxy Key + 项目改名 muxueTools

## 角色
Developer (senior-golang) + Frontend (ui-ux-pro-max)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`
- `.agent/skills/ui-ux-pro-max/SKILL.md`

---

## 背景

### 问题 1：Proxy Key 不同步

Dashboard 页面的 API Key 显示与 Settings/Security 页面的 Proxy API Key 不一致：
- **Dashboard**: 硬编码 `sk-mxln-proxy-local`（默认值）
- **Settings/Security**: 从 SQLite 读取实际配置值

用户在 Settings 中点击 "Regenerate" 后，Dashboard 仍显示旧值。

### 问题 2：项目需要改名

项目名称从 `mxlnapi` 改为 `muxueTools`，涉及：
- Go 模块名和导入路径
- 可执行文件名
- 数据库文件名
- UI 显示文字

---

## 目标

| Phase | 目标 | 难度 | 状态 |
|-------|------|------|------|
| **Phase 1** | 同步 Proxy Key：Dashboard 从 API 获取实际值 | ⭐ 低 | ✅ 已完成 |
| **Phase 2** | 项目改名：mxlnapi → muxueTools | ⭐⭐⭐ 高 | ✅ 已完成 |

---

## 验收标准

### Phase 1 ✅
- [x] Dashboard 的 API Key 与 Settings 的 Proxy API Key 显示一致
- [x] 点击 Settings 的 Regenerate 后，刷新 Dashboard 显示新值

### Phase 2 ✅
- [x] `go build ./...` 编译通过
- [x] `npm run build` 前端构建通过
- [x] `go test ./...` 测试通过
- [x] 所有 `mxlnapi` 字符串已替换为 `muxueTools`
- [x] 所有文档已更新

---

## 已完成的修改

### Phase 1: Proxy Key 同步

| 文件 | 修改内容 |
|------|----------|
| `web/src/views/DashboardView.vue` | 从 `/api/config` 获取实际 Proxy Key |

### Phase 2: 项目改名

| 类别 | 修改文件 | 说明 |
|------|----------|------|
| Go 模块 | `go.mod` | `module mxlnapi` → `module muxueTools` |
| Go 导入路径 | `internal/**/*.go`, `cmd/**/*.go` | 约 31 个文件 |
| 构建脚本 | `scripts/build.ps1` | 输出文件名 `muxueTools.exe` |
| 数据库默认路径 | `internal/types/config.go` | `mxlnapi.db` → `muxueTools.db` |
| 配置模板 | `internal/config/loader.go` | 更新模板中的路径和仓库名 |
| 前端 UI | `web/index.html`, `DashboardView.vue`, `SettingsView.vue` | 更新显示名称 |
| 日志输出 | `cmd/server/main.go`, `cmd/desktop/main.go` | 更新日志消息 |
| Windows 资源 | `cmd/desktop/app.rc` | 更新注释 |
| 文档 | `README.md`, `docs/**/*.md` | 更新所有文档 |
| 配置示例 | `configs/config.example.yaml` | 更新注释和路径 |

---

## 注意事项

> ⚠️ 数据库默认路径已从 `mxlnapi.db` 改为 `muxueTools.db`。
> 已有用户需手动重命名数据库文件：
> ```powershell
> mv data/mxlnapi.db data/muxueTools.db
> ```

---

## 讨论记录

| 日期 | 主题 | 决定 |
|------|------|------|
| 2026-01-20 | Proxy Key 同步 | Dashboard 需从 API 获取实际值 |
| 2026-01-20 | 项目改名 | mxlnapi → muxueTools |

---

*任务创建时间: 2026-01-20*
*完成时间: 2026-01-20*
*状态: ✅ 已完成*
