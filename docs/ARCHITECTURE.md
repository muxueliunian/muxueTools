# MxlnAPI 系统架构设计

> 版本: 1.1  
> 最后更新: 2026-01-16  
> 状态: 实施中

---

## 1. 系统架构图

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                   MxlnAPI                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│    ┌─────────────────────────────────────────────────────────────────────────┐  │
│    │                           HTTP 入口层                                   │  │
│    │  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────┐   │  │
│    │  │  OpenAI 兼容端点  │  │   管理 API 端点   │  │   静态文件服务 (UI)  │   │  │
│    │  │  /v1/*           │  │   /api/*         │  │   /* (Vue3 嵌入)     │   │  │
│    │  └────────┬─────────┘  └────────┬─────────┘  └──────────────────────┘   │  │
│    └───────────┼─────────────────────┼───────────────────────────────────────┘  │
│                │                     │                                          │
│    ┌───────────┼─────────────────────┼───────────────────────────────────────┐  │
│    │           ▼                     ▼            核心业务层                 │  │
│    │  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────┐   │  │
│    │  │    Converter     │  │   Key Manager    │  │    Config Manager    │   │  │
│    │  │  OpenAI ↔ Gemini │  │   增删改查配额    │  │    配置热加载        │   │  │
│    │  └────────┬─────────┘  └────────┬─────────┘  └──────────┬───────────┘   │  │
│    │           │                     │                       │               │  │
│    │           ▼                     ▼                       ▼               │  │
│    │  ┌──────────────────────────────────────────────────────────────────┐   │  │
│    │  │                        Key Pool (核心)                            │   │  │
│    │  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐               │   │  │
│    │  │  │ Round Robin │  │   Random    │  │ Least Used  │  策略引擎     │   │  │
│    │  │  └─────────────┘  └─────────────┘  └─────────────┘               │   │  │
│    │  │  ┌─────────────────────────────────────────────────┐             │   │  │
│    │  │  │     Circuit Breaker (熔断器 - Rate Limit)       │             │   │  │
│    │  │  └─────────────────────────────────────────────────┘             │   │  │
│    │  └──────────────────────────────────────────────────────────────────┘   │  │
│    └─────────────────────────────────────────────────────────────────────────┘  │
│                │                                                                │
│    ┌───────────┼─────────────────────────────────────────────────────────────┐  │
│    │           ▼                      基础设施层                             │  │
│    │  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────┐   │  │
│    │  │   Gemini Client  │  │  SQLite Storage  │  │   GitHub Updater     │   │  │
│    │  │  HTTP/SSE 客户端  │  │    统计持久化    │  │    版本检测更新      │   │  │
│    │  └────────┬─────────┘  └──────────────────┘  └──────────────────────┘   │  │
│    └───────────┼─────────────────────────────────────────────────────────────┘  │
│                │                                                                │
└────────────────┼────────────────────────────────────────────────────────────────┘
                 │
                 ▼
    ┌────────────────────────┐
    │   Google AI Studio     │
    │   (Gemini API)         │
    │   generativelanguage.  │
    │   googleapis.com       │
    └────────────────────────┘
```

### 1.2 数据流图

```
┌──────────┐    OpenAI Format      ┌─────────────┐    Gemini Format     ┌────────────┐
│  Client  │ ──────────────────▶   │   MxlnAPI   │ ──────────────────▶  │  Gemini    │
│ (Chatbox,│  POST /v1/chat/       │             │   POST /v1beta/      │    API     │
│  SillyT) │  completions          │  Converter  │   models/.../        │            │
└──────────┘                       │             │   generateContent    └────────────┘
     ▲                             │ ◀───────────│◀──────────────────────────┘
     │                             │   Response  │   Response (Stream/Block)
     │      OpenAI Response        │   Transform │
     └─────────────────────────────┴─────────────┘
```

### 1.3 Key 池状态机

```
                    ┌─────────────┐
                    │   初始化    │
                    └──────┬──────┘
                           │ 加载配置
                           ▼
                    ┌─────────────┐
          ┌────────▶│   Active    │◀────────┐
          │         │   (可用)    │         │
          │         └──────┬──────┘         │
          │                │                │
          │    Rate Limit  │     冷却完成   │
          │    429 Error   │     Cooldown   │
          │                ▼     Expired    │
          │         ┌─────────────┐         │
          │         │ Rate Limited│─────────┘
          ▲         │  (冷却中)   │  cooldown_seconds
          │         └─────────────┘
          │
    手动启用        ┌─────────────┐
          └─────────│  Disabled   │
                    │  (已禁用)   │ ◀── 手动禁用 / 多次失败
                    └─────────────┘
```

---

## 2. 目录结构

```
mxlnapi/
├── cmd/                                # 应用程序入口
│   ├── server/
│   │   └── main.go                     # 服务端入口
│   └── desktop/
│       ├── main.go                     # 桌面应用入口 (WebView)
│       └── rsrc_windows_amd64.syso     # Windows 图标资源 (编译时嵌入)
│
├── internal/                           # 私有应用代码（不可被外部导入）
│   │
│   ├── api/                            # HTTP API 层
│   │   ├── router.go                   # Gin 路由配置
│   │   ├── middleware/
│   │   │   ├── cors.go                 # CORS 中间件
│   │   │   ├── logger.go               # 请求日志中间件
│   │   │   ├── recovery.go             # Panic 恢复
│   │   │   └── ratelimit.go            # 本地限流（可选）
│   │   ├── handler/
│   │   │   ├── openai_handler.go       # OpenAI 兼容端点处理器
│   │   │   ├── admin_handler.go        # 管理 API 处理器
│   │   │   └── health_handler.go       # 健康检查端点
│   │   └── dto/                        # 数据传输对象
│   │       ├── openai_request.go       # OpenAI 请求结构
│   │       ├── openai_response.go      # OpenAI 响应结构
│   │       └── admin_dto.go            # 管理接口 DTO
│   │
│   ├── gemini/                         # Gemini API 客户端模块
│   │   ├── client.go                   # HTTP 客户端封装
│   │   ├── models.go                   # Gemini 请求/响应结构体
│   │   ├── stream.go                   # SSE 流式响应处理
│   │   └── errors.go                   # Gemini 错误类型定义
│   │
│   ├── converter/                      # 格式转换模块（纯逻辑，无 IO）
│   │   ├── request_converter.go        # OpenAI Request → Gemini Request
│   │   ├── response_converter.go       # Gemini Response → OpenAI Response
│   │   ├── stream_converter.go         # 流式响应转换
│   │   ├── model_mapper.go             # 模型名称映射
│   │   └── converter_test.go           # 单元测试
│   │
│   ├── keypool/                        # Key 池管理模块
│   │   ├── pool.go                     # Key 池核心逻辑
│   │   ├── key.go                      # Key 实体定义
│   │   ├── strategy/                   # 选择策略
│   │   │   ├── interface.go            # 策略接口
│   │   │   ├── round_robin.go          # 轮询策略
│   │   │   ├── random.go               # 随机策略
│   │   │   ├── least_used.go           # 最少使用
│   │   │   └── weighted.go             # 加权（按成功率）
│   │   ├── circuit_breaker.go          # 熔断器（Rate Limit 处理）
│   │   └── stats.go                    # 使用统计聚合
│   │
│   ├── config/                         # 配置管理模块
│   │   ├── config.go                   # 配置结构体定义
│   │   ├── loader.go                   # Viper 配置加载
│   │   ├── watcher.go                  # 配置热更新监听
│   │   └── validator.go                # 配置校验
│   │
│   ├── storage/                        # 持久化存储模块
│   │   ├── database.go                 # 数据库初始化
│   │   ├── models.go                   # GORM 模型定义
│   │   ├── key_repository.go           # Key 数据访问层
│   │   ├── stats_repository.go         # 统计数据访问层
│   │   └── migration.go                # 数据库迁移
│   │
│   ├── updater/                        # 更新检测模块
│   │   ├── github.go                   # GitHub Release API
│   │   └── version.go                  # 版本比较逻辑
│   │
│   └── common/                         # 公共工具
│       ├── errors.go                   # 统一错误定义
│       ├── response.go                 # 统一响应格式
│       └── logger.go                   # 日志工具封装
│
├── web/                                # Vue3 前端（独立构建）
│   ├── src/
│   │   ├── views/                      # 页面组件
│   │   ├── components/                 # 通用组件
│   │   ├── stores/                     # Pinia 状态
│   │   ├── api/                        # API 调用封装
│   │   └── App.vue
│   ├── package.json
│   └── vite.config.ts
│
├── assets/                             # 静态资源
│   └── icon.ico                        # Windows 应用图标 (多尺寸)
│
├── image/                              # 图标源文件目录
│   ├── gugugaga.png                    # 原始图标
│   └── gugugaga-removebg-preview.png   # 透明背景版本 (用于生成 ICO)
│
├── bin/                                # 编译输出目录
│   └── mxlnapi.exe                     # 桌面应用 (带图标)
│
├── configs/
│   └── config.example.yaml             # 配置文件示例
│
├── scripts/
│   ├── build.ps1                       # Windows 构建脚本
│   └── convert-icon.py                 # PNG → ICO 图标转换脚本
│
├── docs/
│   ├── IMPLEMENTATION_PLAN.md          # 实施计划
│   ├── ARCHITECTURE.md                 # 本文档
│   └── API.md                          # API 详细文档
│
├── .goreleaser.yaml                    # 多平台发布配置
├── go.mod
├── go.sum
└── README.md
```

---

## 3. API 接口契约

> **完整 API 文档请参阅 [API.md](./API.md)**

本节提供 API 端点概览，详细的请求/响应格式、示例和错误处理请查阅完整 API 文档。

### 3.1 端点概览

#### OpenAI 兼容端点
| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/chat/completions` | POST | 对话补全（支持流式 SSE） |
| `/v1/models` | GET | 获取可用模型列表 |
| `/health` | GET | 健康检查 |

#### Key 管理 API
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/keys` | GET | 获取 Key 列表 |
| `/api/keys` | POST | 添加 Key |
| `/api/keys/:id` | DELETE | 删除 Key |
| `/api/keys/:id/test` | POST | 测试 Key 可用性 |
| `/api/keys/import` | POST | 批量导入 Key |
| `/api/keys/export` | GET | 导出 Key 列表 |

#### 会话管理 API
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/sessions` | GET | 获取会话列表（分页） |
| `/api/sessions` | POST | 创建新会话 |
| `/api/sessions/:id` | GET | 获取会话详情（含消息） |
| `/api/sessions/:id` | PUT | 更新会话 |
| `/api/sessions/:id` | DELETE | 删除会话及消息 |
| `/api/sessions/:id/messages` | POST | 添加消息到会话 |

#### 统计与配置 API
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/stats` | GET | 获取总体统计 |
| `/api/stats/keys` | GET | 获取各 Key 使用统计 |
| `/api/config` | GET | 获取当前配置 |
| `/api/config` | PUT | 更新配置 |
| `/api/update/check` | GET | 检查 GitHub 更新 |

### 3.2 响应格式规范

**成功响应**:
```json
{
  "success": true,
  "data": { ... },
  "message": "可选消息"
}
```

**错误响应**:
```json
{
  "error": {
    "code": 40001,
    "message": "错误描述",
    "type": "error_type"
  }
}
```

详细的错误码定义请参阅 [API.md - 错误处理](./API.md#错误处理)。

---

## 4. 数据模型

### 4.1 核心实体

```go
// internal/keypool/key.go
type Key struct {
    ID              string            `json:"id"`
    APIKey          string            `json:"-"`                    // 不序列化到 JSON
    MaskedKey       string            `json:"key"`                  // 脱敏显示
    Name            string            `json:"name"`
    Status          KeyStatus         `json:"status"`
    Enabled         bool              `json:"enabled"`
    Tags            []string          `json:"tags"`
    Stats           KeyStats          `json:"stats"`
    CooldownUntil   *time.Time        `json:"cooldown_until"`
    CreatedAt       time.Time         `json:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at"`
}

type KeyStatus string

const (
    KeyStatusActive      KeyStatus = "active"
    KeyStatusRateLimited KeyStatus = "rate_limited"
    KeyStatusDisabled    KeyStatus = "disabled"
)

type KeyStats struct {
    RequestCount      int64      `json:"request_count"`
    SuccessCount      int64      `json:"success_count"`
    ErrorCount        int64      `json:"error_count"`
    PromptTokens      int64      `json:"prompt_tokens"`
    CompletionTokens  int64      `json:"completion_tokens"`
    LastUsedAt        *time.Time `json:"last_used_at"`
}
```

```go
// internal/storage/models.go
type DBKey struct {
    ID              string    `gorm:"primaryKey;type:varchar(36)"`
    APIKey          string    `gorm:"type:text;not null"`           // 加密存储
    Name            string    `gorm:"type:varchar(100)"`
    Tags            string    `gorm:"type:text"`                    // JSON 数组
    Enabled         bool      `gorm:"default:true"`
    RequestCount    int64     `gorm:"default:0"`
    SuccessCount    int64     `gorm:"default:0"`
    ErrorCount      int64     `gorm:"default:0"`
    PromptTokens    int64     `gorm:"default:0"`
    CompletionTokens int64    `gorm:"default:0"`
    LastUsedAt      *time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type DBRequestLog struct {
    ID                string    `gorm:"primaryKey;type:varchar(36)"`
    KeyID             string    `gorm:"type:varchar(36);index"`
    RequestModel      string    `gorm:"type:varchar(50)"`
    ActualModel       string    `gorm:"type:varchar(50)"`
    PromptTokens      int       `gorm:"default:0"`
    CompletionTokens  int       `gorm:"default:0"`
    LatencyMs         int       `gorm:"default:0"`
    StatusCode        int
    ErrorCode         *int
    IsStream          bool      `gorm:"default:false"`
    CreatedAt         time.Time `gorm:"index"`
}
```

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Keys     []KeyConfig    `mapstructure:"keys"`
    Pool     PoolConfig     `mapstructure:"pool"`
    Models   ModelMappings  `mapstructure:"model_mappings"`
    Logging  LoggingConfig  `mapstructure:"logging"`
    Update   UpdateConfig   `mapstructure:"update"`
    Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
    Port int    `mapstructure:"port" default:"8080"`
    Host string `mapstructure:"host" default:"0.0.0.0"`
}

type KeyConfig struct {
    Key     string   `mapstructure:"key"`
    Name    string   `mapstructure:"name"`
    Enabled bool     `mapstructure:"enabled" default:"true"`
    Tags    []string `mapstructure:"tags"`
}

type PoolConfig struct {
    Strategy        string `mapstructure:"strategy" default:"round_robin"`
    CooldownSeconds int    `mapstructure:"cooldown_seconds" default:"60"`
    MaxRetries      int    `mapstructure:"max_retries" default:"3"`
}

type ModelMappings map[string]string

type LoggingConfig struct {
    Level string `mapstructure:"level" default:"info"`
    File  string `mapstructure:"file"`
}

type UpdateConfig struct {
    Enabled       bool   `mapstructure:"enabled" default:"true"`
    CheckInterval string `mapstructure:"check_interval" default:"24h"`
    GithubRepo    string `mapstructure:"github_repo"`
}

type DatabaseConfig struct {
    Path string `mapstructure:"path" default:"data/mxlnapi.db"`
}
```

---

## 5. 错误码规范

### 5.1 错误码表

| 错误码 | HTTP Status | 类型 | 描述 | 触发场景 |
|--------|-------------|------|------|----------|
| `40001` | 400 | `invalid_request_error` | 无效的请求格式 | JSON 解析失败、必填字段缺失 |
| `40002` | 400 | `invalid_request_error` | 不支持的模型 | 请求的模型名称无法映射 |
| `40003` | 400 | `invalid_request_error` | 消息格式错误 | messages 数组为空或格式错误 |
| `40101` | 401 | `authentication_error` | 认证失败 | API Key 无效（Gemini 返回） |
| `40301` | 403 | `permission_error` | 权限不足 | Key 被禁用或无权访问模型 |
| `40401` | 404 | `not_found_error` | 资源不存在 | Key ID 不存在 |
| `42901` | 429 | `rate_limit_error` | 请求频率超限 | 所有 Key 均处于冷却状态 |
| `50001` | 500 | `server_error` | 内部服务器错误 | 未预期的服务器异常 |
| `50201` | 502 | `upstream_error` | 上游 API 错误 | Gemini API 返回非预期错误 |
| `50301` | 503 | `service_unavailable` | 服务暂不可用 | Key 池为空、数据库连接失败 |

### 5.2 错误响应格式

```go
// internal/common/errors.go
type APIError struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code       int         `json:"code"`                // 错误码
    Message    string      `json:"message"`             // 用户可读消息
    Type       string      `json:"type"`                // 错误类型
    Param      string      `json:"param,omitempty"`     // 相关参数（如有）
    RetryAfter int         `json:"retry_after,omitempty"` // 重试等待秒数（429 时）
}

// 预定义错误
var (
    ErrInvalidRequest = &APIError{
        Error: ErrorDetail{
            Code:    40001,
            Message: "Invalid request format",
            Type:    "invalid_request_error",
        },
    }
    
    ErrUnsupportedModel = &APIError{
        Error: ErrorDetail{
            Code:    40002,
            Message: "The specified model is not supported",
            Type:    "invalid_request_error",
        },
    }
    
    ErrRateLimited = &APIError{
        Error: ErrorDetail{
            Code:       42901,
            Message:    "All API keys are currently rate limited",
            Type:       "rate_limit_error",
            RetryAfter: 60,
        },
    }
    
    ErrNoAvailableKeys = &APIError{
        Error: ErrorDetail{
            Code:    50301,
            Message: "No available API keys in the pool",
            Type:    "service_unavailable",
        },
    }
)
```

---

## 6. 配置文件 Schema

### 6.1 完整配置示例

```yaml
# config.yaml - MxlnAPI 配置文件
# 完整配置说明请参考文档

# ========================
# 服务器配置
# ========================
server:
  port: 8080                              # 监听端口
  host: "0.0.0.0"                         # 绑定地址（0.0.0.0 表示所有接口）

# ========================
# API Key 配置
# ========================
keys:
  - key: "AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXX"
    name: "主要 Key"
    enabled: true
    tags:
      - "primary"
      - "personal"
  
  - key: "AIzaSyYYYYYYYYYYYYYYYYYYYYYYYYYY"
    name: "备用 Key"
    enabled: true
    tags:
      - "backup"

# ========================
# Key 池策略配置
# ========================
pool:
  # 选择策略：round_robin | random | least_used | weighted
  strategy: "round_robin"
  
  # 触发 Rate Limit 后的冷却时间（秒）
  cooldown_seconds: 60
  
  # 单次请求最大重试次数（换 Key 重试）
  max_retries: 3

# ========================
# 模型映射
# ========================
model_mappings:
  # OpenAI 模型名 → Gemini 模型名
  "gpt-4": "gemini-1.5-pro-latest"
  "gpt-4-turbo": "gemini-1.5-pro-latest"
  "gpt-4-vision-preview": "gemini-1.5-pro-latest"
  "gpt-4o": "gemini-1.5-flash-latest"
  "gpt-4o-mini": "gemini-1.5-flash-8b-latest"
  "gpt-3.5-turbo": "gemini-1.5-flash-latest"
  
  # Gemini 原生模型名（透传）
  "gemini-pro": "gemini-1.5-pro-latest"
  "gemini-flash": "gemini-1.5-flash-latest"
  "gemini-2.0-flash": "gemini-2.0-flash"
  "gemini-2.5-pro": "gemini-2.5-pro-preview"

# ========================
# 日志配置
# ========================
logging:
  # 日志级别：debug | info | warn | error
  level: "info"
  
  # 日志文件路径（留空则仅输出到控制台）
  file: "logs/mxlnapi.log"
  
  # 日志格式：json | text
  format: "text"

# ========================
# 数据库配置
# ========================
database:
  # SQLite 数据库文件路径
  path: "data/mxlnapi.db"

# ========================
# 更新检测配置
# ========================
update:
  enabled: true
  check_interval: "24h"                   # 检查间隔
  github_repo: "muxueliunian/mxlnapi"     # GitHub 仓库

# ========================
# 高级配置（通常无需修改）
# ========================
advanced:
  # 请求超时（秒）
  request_timeout: 120
  
  # 流式响应刷新间隔（毫秒）
  stream_flush_interval: 100
  
  # 统计数据保留天数
  stats_retention_days: 30
```

### 6.2 配置 Schema 定义

```yaml
# JSON Schema for config.yaml
type: object
required:
  - server
properties:
  server:
    type: object
    required:
      - port
    properties:
      port:
        type: integer
        minimum: 1
        maximum: 65535
        default: 8080
      host:
        type: string
        default: "0.0.0.0"
  
  keys:
    type: array
    items:
      type: object
      required:
        - key
      properties:
        key:
          type: string
          pattern: "^AIzaSy[a-zA-Z0-9_-]{33}$"
        name:
          type: string
          maxLength: 100
        enabled:
          type: boolean
          default: true
        tags:
          type: array
          items:
            type: string
  
  pool:
    type: object
    properties:
      strategy:
        type: string
        enum:
          - round_robin
          - random
          - least_used
          - weighted
        default: round_robin
      cooldown_seconds:
        type: integer
        minimum: 10
        maximum: 3600
        default: 60
      max_retries:
        type: integer
        minimum: 1
        maximum: 10
        default: 3
  
  model_mappings:
    type: object
    additionalProperties:
      type: string
  
  logging:
    type: object
    properties:
      level:
        type: string
        enum:
          - debug
          - info
          - warn
          - error
        default: info
      file:
        type: string
      format:
        type: string
        enum:
          - json
          - text
        default: text
  
  database:
    type: object
    properties:
      path:
        type: string
        default: "data/mxlnapi.db"
  
  update:
    type: object
    properties:
      enabled:
        type: boolean
        default: true
      check_interval:
        type: string
        pattern: "^\\d+[smhd]$"
        default: "24h"
      github_repo:
        type: string
```

---

## 7. 技术决策记录 (ADR)

### ADR-001: 选择 SQLite 作为持久化存储

**状态**: 已采纳

**背景**: 
需要持久化存储 Key 使用统计、请求日志等数据。

**决策**: 
使用 SQLite 作为唯一的持久化存储方案。

**理由**:
- ✅ 无需独立数据库进程，简化部署（单文件分发）
- ✅ 文件级存储，跨平台兼容
- ✅ 对于 5-10 用户规模，读写性能完全足够
- ✅ GORM 原生支持

**后果**:
- 不支持多实例并发写入（单机部署场景无影响）
- 需要定期 VACUUM 优化数据库文件大小

---

### ADR-002: Converter 模块采用纯函数设计

**状态**: 已采纳

**背景**: 
OpenAI 与 Gemini 格式转换是核心功能，需要高可测试性。

**决策**: 
`internal/converter/` 包的所有函数均为纯函数，不包含任何 IO 操作。

**理由**:
- ✅ 纯函数易于单元测试，无需 Mock
- ✅ 明确的输入输出，便于调试
- ✅ 可独立复用

**后果**:
- HTTP 请求、日志记录等 IO 操作需在调用方（handler/client）处理

---

### ADR-003: Key 池使用内存 + 持久化双层架构

**状态**: 已采纳

**背景**: 
Key 的选择需要极低延迟，但统计数据需要持久化。

**决策**: 
- 运行时 Key 池完全在内存中（`sync.Map` + `RWMutex`）
- 统计数据异步批量持久化到 SQLite
- 启动时从数据库恢复状态

**理由**:
- ✅ 选择 Key 零 IO 延迟
- ✅ 统计数据不丢失
- ✅ 程序崩溃后可恢复

**后果**:
- 需要实现优雅关闭，确保统计数据落盘
- 需要处理内存与数据库同步逻辑

---

## 8. 架构检查清单

### 模块化

- [ ] **循环依赖检查**: 使用 `go mod graph | grep cycle` 验证无循环导入
- [ ] **配置解耦**: 所有配置通过 Viper 读取，不硬编码
- [ ] **秘密管理**: API Key 在日志中脱敏显示

### 并发安全

- [ ] **Key 池线程安全**: 使用 `sync.RWMutex` 保护共享状态
- [ ] **统计计数器**: 使用 `atomic` 操作或带锁更新
- [ ] **配置热更新**: 使用读写锁保护配置读取

### 可维护性

- [ ] **Godoc 注释**: 所有导出函数/类型添加文档注释
- [ ] **目录结构**: 遵循 Standard Go Project Layout
- [ ] **错误处理**: 统一使用自定义错误类型，携带上下文

### 可扩展性

- [ ] **策略接口**: Key 选择策略通过 `Strategy` 接口扩展
- [ ] **中间件链**: Gin 中间件支持按需组合
- [ ] **模型映射**: 通过配置文件动态添加新模型

---

## 9. 图标与品牌资源

### 9.1 图标设计

**设计主题**: MyGO 高松灯 × API 工具

**视觉元素**:
- 戴耳麦的黑白企鹅（象征 API 代理/通讯）
- 手持金色钥匙（象征 API Key）
- 灰蓝色背景（取自高松灯发色）

### 9.2 图标文件结构

| 文件 | 用途 | 格式 |
|------|------|------|
| `image/gugugaga.png` | 原始图标素材 | PNG |
| `image/gugugaga-removebg-preview.png` | 透明背景版本 | PNG (499x500) |
| `assets/icon.ico` | Windows 应用图标 | ICO (多尺寸: 16-256px) |
| `cmd/desktop/rsrc_windows_amd64.syso` | Go 编译资源 | SYSO (自动嵌入) |

### 9.3 图标生成完整流程

#### 前置条件

```powershell
# 安装 Python 图像处理库
pip install Pillow

# 安装 Go 资源编译工具
go install github.com/akavel/rsrc@latest
```

#### 步骤 1: 准备图标源文件

1. 准备一张 PNG 图片（建议尺寸 512x512 以上）
2. 使用在线抠图工具去除背景（推荐 [remove.bg](https://www.remove.bg/)）
3. 将透明背景的 PNG 保存到 `image/` 目录

#### 步骤 2: 生成 ICO 图标

```powershell
# 运行图标转换脚本
python scripts/convert-icon.py

# 输出:
# ✓ Icon created: assets\icon.ico
# Size: ~90 KB (包含 6 个尺寸)
```

转换脚本会自动生成以下尺寸:
- 256x256 (Windows 大图标)
- 128x128
- 64x64
- 48x48 (Windows 任务栏)
- 32x32 (Windows 标题栏)
- 16x16 (Windows 小图标)

#### 步骤 3: 嵌入到 Go 程序

```powershell
# 生成 .syso 资源文件
rsrc -ico assets/icon.ico -o cmd/desktop/rsrc_windows_amd64.syso
```

#### 步骤 4: 编译桌面应用

```powershell
# 编译带图标的 Windows 桌面应用
go build -ldflags="-s -w -H windowsgui" -o bin/mxlnapi.exe ./cmd/desktop

# 参数说明:
# -s -w: 去除调试信息，减小体积
# -H windowsgui: 隐藏控制台窗口
```

#### 一键构建命令

```powershell
# 完整流程 (从 PNG 到 EXE)
python scripts/convert-icon.py && `
rsrc -ico assets/icon.ico -o cmd/desktop/rsrc_windows_amd64.syso && `
go build -ldflags="-s -w -H windowsgui" -o bin/mxlnapi.exe ./cmd/desktop
```

### 9.4 依赖工具

| 工具 | 用途 | 安装方式 |
|------|------|----------|
| Python 3 | 运行转换脚本 | 系统自带或 [python.org](https://www.python.org/) |
| Pillow | PNG → ICO 转换 | `pip install Pillow` |
| rsrc | ICO → SYSO 编译 | `go install github.com/akavel/rsrc@latest` |

### 9.5 常见问题

**Q: 图标不显示怎么办？**
A: 确保 `.syso` 文件在 `cmd/desktop/` 目录下，且文件名包含 `_windows_amd64`。

**Q: 如何更换图标？**
A: 替换 `image/` 下的 PNG 文件，然后重新执行步骤 2-4。

**Q: 图标背景不透明？**
A: 使用 [remove.bg](https://www.remove.bg/) 等工具去除背景后再转换。

---

*文档结束*

**变更记录**:
- 2026-01-16: 完善图标生成流程文档，更新目录结构
- 2026-01-16: 添加图标资源章节 (Section 9)
- 2026-01-15: 初始版本

