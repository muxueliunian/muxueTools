# 任务：实现 API 层（阶段 3）

## 角色
Developer (senior-golang skill)

## 背景
阶段 2 核心逻辑已全部完成：
- ✅ 配置加载 (`internal/config/`)
- ✅ Key 池 (`internal/keypool/`)
- ✅ 格式转换 (`internal/gemini/converter.go`)
- ✅ Gemini 客户端 (`internal/gemini/client.go`)

现在进入 **阶段 3：API 层实现**，将核心逻辑组装成可用的 HTTP 服务。

## 步骤

### 1. 阅读规范
- `.agent/skills/senior-golang/SKILL.md` - Go 开发规范（Gin 框架部分）
- `docs/ARCHITECTURE.md` - API 设计规范
- `docs/IMPLEMENTATION_PLAN.md` - API 端点定义

### 2. 创建文件结构
```
internal/api/
├── router.go           # 路由配置
├── openai_handler.go   # OpenAI 兼容端点处理
├── admin_handler.go    # 管理端点处理
├── middleware.go       # 中间件（CORS、日志、Recovery）
└── response.go         # 统一响应格式
```

### 3. 功能要求

#### 3.1 路由配置 (router.go)

```go
func NewRouter(cfg *config.Config, pool *keypool.Pool, client *gemini.Client) *gin.Engine

// 路由结构：
// /v1/chat/completions  POST  - OpenAI 兼容端点
// /v1/models            GET   - 模型列表
// /health               GET   - 健康检查
//
// /api/keys             GET/POST/DELETE - Key 管理
// /api/keys/:id/test    POST  - 测试 Key
// /api/keys/import      POST  - 批量导入
// /api/keys/export      GET   - 导出
// /api/stats            GET   - 使用统计
// /api/config           GET/PUT - 配置管理
// /api/update/check     GET   - 检查更新
```

#### 3.2 OpenAI 兼容端点 (openai_handler.go)

**核心功能**：

```go
// POST /v1/chat/completions
func (h *OpenAIHandler) ChatCompletions(c *gin.Context)
```

**处理流程**：
1. 解析请求体 (`ChatCompletionRequest`)
2. 判断是否流式请求 (`stream: true`)
3. 调用 `gemini.Client` 获取响应
4. 返回 OpenAI 格式响应

**流式响应**：
- 设置 `Content-Type: text/event-stream`
- 使用 `c.Stream()` 发送 SSE 数据
- 格式：`data: {json}\n\n`
- 结束：`data: [DONE]\n\n`

**模型列表**：
```go
// GET /v1/models
func (h *OpenAIHandler) ListModels(c *gin.Context)
```

#### 3.3 管理端点 (admin_handler.go)

**Key 管理**：
```go
// GET    /api/keys         - 列出所有 Key（脱敏显示）
// POST   /api/keys         - 添加 Key
// DELETE /api/keys/:id     - 删除 Key
// POST   /api/keys/:id/test - 测试 Key 可用性
// POST   /api/keys/import  - 批量导入
// GET    /api/keys/export  - 导出
```

**统计**：
```go
// GET /api/stats - 返回使用统计
```

**配置**：
```go
// GET /api/config - 获取当前配置（脱敏）
// PUT /api/config - 更新配置
```

#### 3.4 中间件 (middleware.go)

```go
// CORS 中间件
func CORSMiddleware() gin.HandlerFunc

// 日志中间件
func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc

// Recovery 中间件（已内置，可自定义）
func RecoveryMiddleware() gin.HandlerFunc

// 请求 ID 中间件
func RequestIDMiddleware() gin.HandlerFunc
```

### 4. 依赖注入

```go
type Server struct {
    engine  *gin.Engine
    config  *config.Config
    pool    *keypool.Pool
    client  *gemini.Client
    logger  *logrus.Logger
}

func NewServer(cfg *config.Config) (*Server, error) {
    // 1. 初始化 KeyPool
    // 2. 初始化 Gemini Client
    // 3. 初始化 Router
    // 4. 返回 Server
}

func (s *Server) Run() error {
    return s.engine.Run(s.config.Server.Addr())
}
```

### 5. 测试要求

创建 `internal/api/` 下的测试文件：
- `router_test.go` - 路由集成测试
- `openai_handler_test.go` - OpenAI 端点测试

**测试场景**：
- 健康检查返回 200
- 简单文本请求成功
- 流式请求返回正确 SSE 格式
- 无效请求返回 400
- 无可用 Key 返回 429

### 6. 更新 main.go

```go
// cmd/server/main.go
func main() {
    // 1. 加载配置
    cfg, err := config.Load()
    
    // 2. 初始化服务器
    server, err := api.NewServer(cfg)
    
    // 3. 优雅关闭
    go func() {
        if err := server.Run(); err != nil {
            log.Fatal(err)
        }
    }()
    
    // 4. 等待信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // 5. 关闭服务
    server.Shutdown()
}
```

## 产出
- `internal/api/router.go`
- `internal/api/openai_handler.go`
- `internal/api/admin_handler.go`
- `internal/api/middleware.go`
- `internal/api/response.go`
- `internal/api/server.go`
- `internal/api/router_test.go`
- `internal/api/openai_handler_test.go`
- 更新后的 `cmd/server/main.go`

## 约束
- 使用 `github.com/gin-gonic/gin` 框架
- 使用 `github.com/gin-contrib/cors` 处理 CORS
- 使用 `github.com/sirupsen/logrus` 日志
- 流式响应必须符合 OpenAI SSE 规范
- 错误响应使用 `internal/types/errors.go` 中的格式
- Key 显示必须脱敏（使用 `types.MaskAPIKey`）

## 验收标准

1. **服务启动**：`go run ./cmd/server` 可正常启动
2. **健康检查**：`curl http://localhost:8080/health` 返回 200
3. **模型列表**：`curl http://localhost:8080/v1/models` 返回模型列表
4. **文本请求**：发送 OpenAI 格式请求，返回正确响应
5. **流式请求**：`stream: true` 返回 SSE 格式响应

---

*任务创建时间: 2026-01-15*
