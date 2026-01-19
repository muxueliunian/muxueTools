# 任务：实现 SQLite 存储层（阶段 3.5）

## 角色
Developer (senior-golang skill)

## 背景
MxlnAPI 需要将 API Key 和会话数据持久化到 SQLite。当前 API Key 仅在内存中管理，重启后丢失。需要实现完整的存储层。

## 目标
1. 实现 SQLite 存储层（Keys + Sessions + Messages）
2. 修改 KeyPool 使用 DB 存储
3. 修改 admin_handler 的 Key CRUD 使用 DB
4. 新增 session_handler 处理会话 API

## 步骤

### 1. 阅读规范
- `.agent/skills/senior-golang/SKILL.md` - Go 开发规范
- `docs/ARCHITECTURE.md` - 系统架构
- `internal/keypool/pool.go` - 现有 Key 池实现
- `internal/api/admin_handler.go` - 现有 Key API

### 2. 创建文件结构

```
internal/
├── types/
│   └── session.go           # 新增：Session/Message 类型
├── storage/
│   ├── sqlite.go            # 新增：SQLite 初始化
│   ├── keys.go              # 新增：Keys 存储
│   ├── sessions.go          # 新增：Sessions 存储
│   └── storage_test.go      # 新增：存储层测试
├── keypool/
│   └── pool.go              # 修改：集成 DB 存储
└── api/
    ├── admin_handler.go     # 修改：Key CRUD 使用 DB
    └── session_handler.go   # 新增：Session API
```

### 3. 类型定义 (internal/types/session.go)

```go
package types

import "time"

// Session represents a chat session
type Session struct {
    ID           string    `json:"id" gorm:"primaryKey"`
    Title        string    `json:"title"`
    Model        string    `json:"model"`
    MessageCount int       `json:"message_count"`
    TotalTokens  int       `json:"total_tokens"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// Message represents a message in a session
type Message struct {
    ID               string    `json:"id" gorm:"primaryKey"`
    SessionID        string    `json:"session_id" gorm:"index"`
    Role             string    `json:"role"` // user/assistant
    Content          string    `json:"content"`
    PromptTokens     int       `json:"prompt_tokens"`
    CompletionTokens int       `json:"completion_tokens"`
    CreatedAt        time.Time `json:"created_at"`
}

// DTOs for API
type CreateSessionRequest struct {
    Title string `json:"title"`
    Model string `json:"model"`
}

type CreateSessionResponse struct {
    Session
}

type SessionListResponse struct {
    Sessions []Session `json:"sessions"`
    Total    int       `json:"total"`
}

type SessionDetailResponse struct {
    Session  Session   `json:"session"`
    Messages []Message `json:"messages"`
}

type AddMessageRequest struct {
    Role    string `json:"role" binding:"required"`
    Content string `json:"content" binding:"required"`
}

type UpdateSessionRequest struct {
    Title *string `json:"title,omitempty"`
    Model *string `json:"model,omitempty"`
}
```

### 4. SQLite 初始化 (internal/storage/sqlite.go)

```go
package storage

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "mxlnapi/internal/types"
)

type Storage struct {
    db *gorm.DB
}

func NewStorage(dbPath string) (*Storage, error) {
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    // Auto migrate
    if err := db.AutoMigrate(&types.Key{}, &types.Session{}, &types.Message{}); err != nil {
        return nil, err
    }
    
    return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
    sqlDB, err := s.db.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}
```

### 5. Keys 存储层 (internal/storage/keys.go)

实现以下方法：
```go
func (s *Storage) CreateKey(key *types.Key) error
func (s *Storage) GetKey(id string) (*types.Key, error)
func (s *Storage) GetKeyByAPIKey(apiKey string) (*types.Key, error)
func (s *Storage) ListKeys() ([]types.Key, error)
func (s *Storage) UpdateKey(key *types.Key) error
func (s *Storage) DeleteKey(id string) error
func (s *Storage) ImportKeys(keys []types.Key) (int, error)
```

### 6. Sessions 存储层 (internal/storage/sessions.go)

实现以下方法：
```go
func (s *Storage) CreateSession(session *types.Session) error
func (s *Storage) GetSession(id string) (*types.Session, error)
func (s *Storage) ListSessions(limit, offset int) ([]types.Session, int, error)
func (s *Storage) UpdateSession(session *types.Session) error
func (s *Storage) DeleteSession(id string) error

func (s *Storage) AddMessage(message *types.Message) error
func (s *Storage) GetMessages(sessionID string) ([]types.Message, error)
func (s *Storage) DeleteMessages(sessionID string) error
```

### 7. 修改 KeyPool (internal/keypool/pool.go)

添加 Storage 依赖：
```go
type Pool struct {
    // ... 现有字段
    storage *storage.Storage  // 新增
}

func NewPool(storage *storage.Storage, opts ...PoolOption) *Pool

// 修改：从 DB 加载 Keys
func (p *Pool) LoadFromDB() error

// 修改：更新 Key 时同步到 DB
func (p *Pool) ReportSuccess(...) // 更新 DB 统计
func (p *Pool) ReportFailure(...) // 更新 DB 统计
```

**注意**：保持向后兼容，支持从配置文件初始化（首次启动时同步到 DB）

### 8. 修改 admin_handler.go

Key CRUD 方法使用 Storage：
```go
// GET /api/keys - 从 DB 读取
// POST /api/keys - 写入 DB + 添加到 Pool
// DELETE /api/keys/:id - 从 DB 删除 + 从 Pool 移除
// POST /api/keys/import - 批量写入 DB
```

### 9. 新增 session_handler.go

```go
type SessionHandler struct {
    storage *storage.Storage
}

// GET /api/sessions
func (h *SessionHandler) ListSessions(c *gin.Context)

// POST /api/sessions
func (h *SessionHandler) CreateSession(c *gin.Context)

// GET /api/sessions/:id
func (h *SessionHandler) GetSession(c *gin.Context)

// PUT /api/sessions/:id
func (h *SessionHandler) UpdateSession(c *gin.Context)

// DELETE /api/sessions/:id
func (h *SessionHandler) DeleteSession(c *gin.Context)

// POST /api/sessions/:id/messages
func (h *SessionHandler) AddMessage(c *gin.Context)
```

### 10. 更新路由 (internal/api/router.go)

添加 Session 路由组：
```go
sessions := r.Group("/api/sessions")
{
    sessions.GET("", sessionHandler.ListSessions)
    sessions.POST("", sessionHandler.CreateSession)
    sessions.GET("/:id", sessionHandler.GetSession)
    sessions.PUT("/:id", sessionHandler.UpdateSession)
    sessions.DELETE("/:id", sessionHandler.DeleteSession)
    sessions.POST("/:id/messages", sessionHandler.AddMessage)
}
```

### 11. 更新 Server 初始化 (internal/api/server.go)

```go
func NewServer(cfg *types.Config) (*Server, error) {
    // 1. 初始化 Storage
    storage, err := storage.NewStorage(cfg.Database.Path)
    
    // 2. 初始化 KeyPool (使用 Storage)
    pool := keypool.NewPool(storage, ...)
    
    // 3. 从配置文件同步 Keys 到 DB（首次）
    syncKeysFromConfig(cfg.Keys, storage)
    
    // 4. 从 DB 加载 Keys 到 Pool
    pool.LoadFromDB()
    
    // ... 其他初始化
}
```

### 12. 更新依赖

```bash
go get gorm.io/gorm
go get gorm.io/driver/sqlite
```

## 产出

| 文件 | 说明 |
|------|------|
| `internal/types/session.go` | Session/Message 类型 |
| `internal/storage/sqlite.go` | SQLite 初始化 |
| `internal/storage/keys.go` | Keys 存储层 |
| `internal/storage/sessions.go` | Sessions 存储层 |
| `internal/storage/storage_test.go` | 存储层测试 |
| `internal/keypool/pool.go` | 修改：集成 DB |
| `internal/api/admin_handler.go` | 修改：使用 DB |
| `internal/api/session_handler.go` | 新增：Session API |
| `internal/api/router.go` | 修改：Session 路由 |
| `internal/api/server.go` | 修改：初始化 Storage |

## 约束

- 使用 GORM 作为 ORM
- SQLite 文件路径从配置读取 (`config.Database.Path`)
- Key 的 `api_key` 字段不应暴露在 API 响应中（使用 `masked_key`）
- 保持向后兼容：配置文件中的 Keys 首次启动时同步到 DB
- 删除 Session 时级联删除 Messages

## 验收标准

1. **存储层**
   - `go test ./internal/storage/...` 通过

2. **Key 持久化**
   - 通过 API 添加的 Key 重启后仍存在
   - `curl POST /api/keys` 写入 DB

3. **Session 持久化**
   - `curl POST /api/sessions` 创建会话
   - `curl GET /api/sessions/:id` 返回会话 + 消息
   - 重启后会话仍在

4. **兼容性**
   - 现有单元测试通过
   - `go run ./cmd/server` 正常启动

---

*任务创建时间: 2026-01-15*
