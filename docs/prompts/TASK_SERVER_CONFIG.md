# 任务：Desktop 服务器地址配置

## 角色
Developer (senior-golang)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`

## 背景

MuxueTools Desktop 版本当前存在以下问题：

1. **端口随机分配**：每次启动使用不同的随机端口（如 62687, 63297），导致第三方应用（如 Cursor）需要每次更新配置
2. **localhost 显示**：API 端点显示为 `http://localhost:port` 而非 `http://127.0.0.1:port`，部分应用可能有兼容性差异

**当前代码逻辑** (`cmd/desktop/main.go`):
- Line 80-81: 强制覆盖配置端口为 0（随机）
- Line 136: 监听 `:0` 获取随机端口
- Line 141: 使用 `localhost` 构建 URL

**现有配置支持**:
- `config.yaml` 已有 `server.port` 和 `server.host` 字段
- `internal/types/config.go` 已定义 `ServerConfig` 结构体
- Desktop 版本**故意忽略**这些配置值

## 目标

| Phase | 目标 | 难度 |
|-------|------|------|
| **Phase 1** | Desktop 使用配置文件中的固定端口 | ⭐ 低 |
| **Phase 2** | 端口冲突时自动回退到随机端口 | ⭐ 低 |
| **Phase 3** | 显示 `127.0.0.1` 替代 `localhost` | ⭐ 低 |
| **Phase 4** | Settings Security 页面添加端口配置 | ⭐⭐ 中 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范**
   - `.agent/skills/senior-golang/SKILL.md`

2. **项目文档**
   - `docs/ARCHITECTURE.md` - 项目结构
   - `cmd/desktop/main.go` - Desktop 入口点
   - `docs/FRONTEND_WORKFLOW.md` - 前端开发流程

### Phase 1: 使用配置端口

1. 移除 `main.go` 第 80-81 行的端口覆盖逻辑
2. 使用 `cfg.Server.Port` 作为目标端口
3. 默认端口改为 8080（若配置为 0）

**代码变更：**
```go
// Before
originalPort := cfg.Server.Port
cfg.Server.Port = 0

// After
targetPort := cfg.Server.Port
if targetPort == 0 {
    targetPort = 8080
}
```

### Phase 2: 端口冲突回退

修改第 136 行监听逻辑，添加回退机制：

```go
listener, err := net.Listen("tcp", fmt.Sprintf(":%d", targetPort))
if err != nil {
    logger.Warnf("Port %d unavailable, using random port", targetPort)
    listener, err = net.Listen("tcp", ":0")
    if err != nil {
        logger.Fatalf("Failed to create listener: %v", err)
    }
}
```

### Phase 3: 使用 127.0.0.1

修改第 141 行 URL 构建逻辑：

```go
// Before
serverAddr := fmt.Sprintf("http://localhost:%d", actualPort)

// After
displayHost := cfg.Server.Host
if displayHost == "0.0.0.0" || displayHost == "" {
    displayHost = "127.0.0.1"
}
serverAddr := fmt.Sprintf("http://%s:%d", displayHost, actualPort)
```

### Phase 4: Settings UI 端口配置

#### 4.1 后端 API 修改

**文件**: `internal/api/admin_handler.go`

修改 `UpdateConfig` 接口，支持更新 `server.port`：

```go
// UpdateConfigRequest 添加 server 字段
type UpdateConfigRequest struct {
    Server        *ServerConfigUpdate  `json:"server,omitempty"`
    // ... existing fields
}

type ServerConfigUpdate struct {
    Port *int `json:"port,omitempty"`
}
```

**文件**: `internal/types/config.go`

在 `UpdateConfigRequest` 添加 Server 配置更新支持。

#### 4.2 前端 Settings UI

**文件**: `web/src/views/SettingsView.vue`

在 Security 标签页添加端口配置：

```vue
<!-- Server Port Configuration (Security Tab) -->
<n-form-item :label="$t('settings.serverPort')">
    <n-input-number 
        v-model:value="serverPort" 
        :min="1024" 
        :max="65535"
        :placeholder="8080"
    />
    <template #feedback>
        <span class="text-xs text-claude-secondaryText">
            {{ $t('settings.serverPortDescription') }}
        </span>
    </template>
</n-form-item>
```

#### 4.3 i18n 翻译

**文件**: `web/src/i18n/locales/*.json`

添加翻译文本：
- `settings.serverPort`: "服务端口" / "Server Port" / "サーバーポート"
- `settings.serverPortDescription`: "修改后需重启应用生效" / "Restart required after change" / "変更後は再起動が必要です"

#### 4.4 保存逻辑

保存时调用 `updateConfig` API 更新 `server.port`，并显示提示：
- 保存成功后显示 Toast："端口已更新，请重启应用生效"

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `cmd/desktop/main.go` | **MODIFY** | 修改端口和地址逻辑 |
| `internal/api/admin_handler.go` | **MODIFY** | 添加 server.port 更新支持 |
| `internal/types/config.go` | **MODIFY** | 添加 ServerConfigUpdate 类型 |
| `web/src/views/SettingsView.vue` | **MODIFY** | Security 标签页添加端口配置 |
| `web/src/api/config.ts` | **MODIFY** | 更新 ConfigInfo 类型 |
| `web/src/i18n/locales/zh-CN.json` | **MODIFY** | 添加中文翻译 |
| `web/src/i18n/locales/en-US.json` | **MODIFY** | 添加英文翻译 |
| `web/src/i18n/locales/ja-JP.json` | **MODIFY** | 添加日文翻译 |

## 约束

### 技术约束
- Go 1.22+
- 保持与 `cmd/server/main.go` 的一致性

### 质量约束
- 遵循 `.agent/skills/senior-golang/SKILL.md` 代码规范
- 端口冲突时有明确的日志警告

### 兼容性约束
- 不影响现有配置文件格式
- 不影响 Server 版本行为

## 验收标准

### Phase 1-3 后端验证
- [ ] 编辑 `config.yaml` 设置 `server.port: 8888`，启动 Desktop
- [ ] Dashboard 显示 `http://127.0.0.1:8888/v1`
- [ ] `curl http://127.0.0.1:8888/health` 返回正常
- [ ] 端口 8888 被占用时，应用仍能启动（使用随机端口）
- [ ] 日志显示端口占用警告信息
- [ ] Cursor 等第三方应用可正常连接

### Phase 4 前端验证
- [x] Settings → Security 标签页显示"服务端口"输入框
- [x] 端口输入范围限制 1024-65535
- [x] 修改端口后保存成功
- [x] 保存后显示"需要重启生效"提示
- [ ] 重启应用后新端口生效
- [x] 中/英/日三语言翻译正确显示

## 交付文档

| 文档 | 更新内容 | 状态 |
|------|----------|------|
| `README.md` | 更新配置说明，说明如何设置固定端口 | ✅ 已完成 |

## 开发流程

遵循 `docs/DEVELOPMENT.md` 开发流程。

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 端口被占用 | 应用启动失败 | 自动回退到随机端口 |
| 多实例运行 | 端口冲突 | 回退机制保障 |

---
*任务创建时间: 2026-01-21*
