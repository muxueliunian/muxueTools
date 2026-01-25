# 任务：实现跨平台数据路径标准化（Phase 1）

## 角色
Developer (senior-golang skill)

## Skills 依赖
- `.agent/skills/senior-golang/SKILL.md`

## 背景

MuxueTools 计划支持 Windows 11 和 Ubuntu 24 LTS 双平台。当前存在以下问题：

1. **数据路径硬编码**: 数据库路径使用相对路径 `data/MuxueTools.db`，不符合各操作系统的用户数据目录规范
2. **配置搜索路径有限**: 仅搜索当前目录和 `./configs`，未包含系统标准配置目录
3. **无迁移机制**: 旧版本用户升级后无法检测和迁移旧数据

**目标路径规范**（参考 `docs/UBUNTU_ADAPTATION_PLAN.md`）：

| 平台 | 类型 | 路径 |
|------|------|------|
| **Windows** | 用户数据 | `%APPDATA%\MuxueTools\data\` |
| | 配置文件 | `%APPDATA%\MuxueTools\config.yaml` |
| | 日志目录 | `%APPDATA%\MuxueTools\logs\` |
| **Linux** | 用户数据 | `~/.local/share/muxuetools/` |
| | 配置文件 | `~/.config/muxuetools/config.yaml` |
| | 日志目录 | `~/.local/share/muxuetools/logs/` |

**已完成的依赖模块：**（参见 `docs/DEVELOPMENT.md`）
- `internal/config/loader.go` - 配置加载
- `internal/types/config.go` - 配置类型定义
- `internal/storage/sqlite.go` - SQLite 存储层

## 目标

| 任务 | 目标 | 优先级 |
|------|------|--------|
| **1.1** | 新增 `paths.go` 跨平台路径工具函数 | ⭐⭐⭐ 高 |
| **1.2** | 修改 `loader.go` 使用新路径函数 | ⭐⭐⭐ 高 |
| **1.3** | 修改 `config.go` 更新默认值 | ⭐⭐ 中 |
| **1.4** | 修改 `cmd/server/main.go` 调用目录初始化 | ⭐⭐ 中 |
| **1.5** | 修改 `cmd/desktop/main.go` 同上 | ⭐⭐ 中 |
| **1.6** | 新增 `migration.go` 旧数据迁移检测 | ⭐ 低 |

## 步骤

### 阶段 0：阅读规范 (必须)

1. **Skills 规范**
   - `.agent/skills/senior-golang/SKILL.md`

2. **项目文档**
   - `docs/ARCHITECTURE.md` - 系统架构
   - `docs/UBUNTU_ADAPTATION_PLAN.md` - 完整适配计划

3. **相关代码**
   - `internal/config/loader.go` - 现有配置加载逻辑
   - `internal/types/config.go` - 配置类型定义
   - `cmd/server/main.go` - Server 启动入口
   - `cmd/desktop/main.go` - Desktop 启动入口

### 阶段 1：创建路径工具函数

**新增文件**: `internal/config/paths.go`

```go
package config

import (
    "os"
    "path/filepath"
    "runtime"
)

const AppName = "MuxueTools"

// GetDataDir 返回用户数据目录（自动扫描，不可自定义）
func GetDataDir() string {
    switch runtime.GOOS {
    case "windows":
        appData := os.Getenv("APPDATA")
        if appData == "" {
            home, _ := os.UserHomeDir()
            appData = filepath.Join(home, "AppData", "Roaming")
        }
        return filepath.Join(appData, AppName, "data")
    case "linux":
        if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
            return filepath.Join(xdgData, "muxuetools")
        }
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".local", "share", "muxuetools")
    default:
        return "data"
    }
}

// GetConfigDir 返回配置目录
func GetConfigDir() string {
    switch runtime.GOOS {
    case "windows":
        appData := os.Getenv("APPDATA")
        if appData == "" {
            home, _ := os.UserHomeDir()
            appData = filepath.Join(home, "AppData", "Roaming")
        }
        return filepath.Join(appData, AppName)
    case "linux":
        if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
            return filepath.Join(xdgConfig, "muxuetools")
        }
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".config", "muxuetools")
    default:
        return "."
    }
}

// GetLogDir 返回日志目录
func GetLogDir() string {
    return filepath.Join(GetDataDir(), "logs")
}

// GetDatabasePath 返回数据库文件路径
func GetDatabasePath() string {
    return filepath.Join(GetDataDir(), "muxuetools.db")
}

// EnsureDirectories 确保所有必需目录存在
func EnsureDirectories() error {
    dirs := []string{
        GetConfigDir(),
        GetDataDir(),
        GetLogDir(),
    }
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return err
        }
    }
    return nil
}
```

### 阶段 2：修改配置加载

**修改文件**: `internal/config/loader.go`

在 `Load()` 函数中更新搜索路径：

```diff
func (l *Loader) Load() (*types.Config, error) {
    l.setupDefaults()
    l.setupEnvBindings()

-   // Add default search paths
-   l.v.AddConfigPath(".")
-   l.v.AddConfigPath("./configs")
+   // Add default search paths (优先级从高到低)
+   l.v.AddConfigPath(GetConfigDir())  // 用户配置目录
+   l.v.AddConfigPath(".")             // 当前目录 (便携模式)
+   l.v.AddConfigPath("./configs")     // 开发模式

    // ... rest of the code
}
```

### 阶段 3：更新默认配置

**修改文件**: `internal/types/config.go`

```diff
func DefaultConfig() *Config {
    return &Config{
        // ...
        Database: DatabaseConfig{
-           Path: "data/MuxueTools.db",
+           Path: "", // 留空表示使用 config.GetDatabasePath()
        },
        // ...
    }
}
```

**注意**: 需要在使用 `DatabaseConfig.Path` 的地方添加逻辑：如果为空则调用 `config.GetDatabasePath()`

### 阶段 4：更新启动入口

**修改文件**: `cmd/server/main.go` 和 `cmd/desktop/main.go`

在配置加载前调用目录初始化：

```go
func main() {
    // ... flag parsing

    // 确保系统目录存在
    if err := config.EnsureDirectories(); err != nil {
        log.Fatalf("Failed to create directories: %v", err)
    }

    // 检测旧数据迁移
    if oldPath := config.CheckLegacyData(); oldPath != "" {
        config.LogMigrationHint(logger, oldPath)
    }

    // 加载配置
    if err := config.Init(); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // ... rest of startup
}
```

### 阶段 5：新增迁移检测

**新增文件**: `internal/config/migration.go`

```go
package config

import (
    "os"

    "github.com/sirupsen/logrus"
)

// LegacyPaths 定义旧版本使用的相对路径
var legacyPaths = []string{
    "data/MuxueTools.db",
    "data/muxuetools.db",
}

// CheckLegacyData 检查是否存在旧版本数据，返回旧路径（如存在）
func CheckLegacyData() string {
    newDBPath := GetDatabasePath()

    // 如果新路径已存在数据，无需迁移
    if fileExists(newDBPath) {
        return ""
    }

    // 检查旧路径
    for _, oldPath := range legacyPaths {
        if fileExists(oldPath) {
            return oldPath
        }
    }

    return ""
}

// LogMigrationHint 记录迁移提示日志
func LogMigrationHint(logger *logrus.Logger, oldPath string) {
    newPath := GetDatabasePath()
    logger.Warn("========================================")
    logger.Warnf("发现旧版本数据: %s", oldPath)
    logger.Warnf("新数据路径: %s", newPath)
    logger.Warn("请手动迁移数据或继续使用旧路径")
    logger.Warn("========================================")
}

func fileExists(path string) bool {
    info, err := os.Stat(path)
    return err == nil && !info.IsDir()
}
```

## 产出文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/config/paths.go` | **NEW** | 跨平台路径工具函数 |
| `internal/config/migration.go` | **NEW** | 旧数据迁移检测与提示 |
| `internal/config/paths_test.go` | **NEW** | 路径函数单元测试 |
| `internal/config/loader.go` | **MODIFY** | 使用新路径函数 |
| `internal/types/config.go` | **MODIFY** | 更新默认配置 |
| `cmd/server/main.go` | **MODIFY** | 调用 `EnsureDirectories()` |
| `cmd/desktop/main.go` | **MODIFY** | 调用 `EnsureDirectories()` |

## 约束

### 技术约束
- Go 版本 1.22+
- 使用 `os.UserHomeDir()` 获取用户目录
- 遵循 XDG Base Directory Specification (Linux)
- 遵循 Windows APPDATA 规范

### 质量约束
- 遵循 `.agent/skills/senior-golang/SKILL.md` 代码规范
- 测试覆盖率 > 80%
- 无竞态问题

### 兼容性约束
- 保持便携模式：当前目录存在 `config.yaml` 时优先使用
- 保持向后兼容：如果配置中有 `database.path` 则尊重配置值
- 旧版本数据不自动迁移，仅提示用户

## 验收标准

- [ ] `go test ./internal/config/...` 所有测试通过
- [ ] `go build ./cmd/server` 编译成功
- [ ] `go build ./cmd/desktop` 编译成功
- [ ] Windows 上数据写入 `%APPDATA%\MuxueTools\`
- [ ] Linux 上数据写入 `~/.local/share/muxuetools/`
- [ ] 便携模式仍然可用（当前目录有 config.yaml）
- [ ] 旧数据路径检测正常工作

## 交付文档

| 文档 | 更新内容 |
|------|----------|
| `docs/ARCHITECTURE.md` | 更新目录结构说明，添加跨平台路径说明 |
| `docs/UBUNTU_ADAPTATION_PLAN.md` | 更新 Phase 1 状态为已完成 |

## 开发流程

遵循 `docs/DEVELOPMENT.md` 中的 TDD 开发流程。

## 风险与注意事项

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 路径权限问题 | 目录创建失败 | `EnsureDirectories()` 返回错误并记录日志 |
| 环境变量未设置 | 路径解析失败 | 使用 `os.UserHomeDir()` 作为备用方案 |
| 旧数据未迁移 | 用户丢失历史数据 | 仅提示不自动迁移，让用户自行决定 |

---

*任务创建时间: 2026-01-25*
