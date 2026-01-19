# 任务：SQLite 存储层测试与审核（阶段 3.5）

## 角色
QA Engineer (qa-automation skill)

## 背景
Developer Agent 已完成 SQLite 存储层实现：
- Keys 存储（CRUD）
- Sessions 存储（CRUD）
- Messages 存储
- KeyPool 集成 DB
- Session Handler

现在需要进行全面的功能测试和代码审核。

## 审核范围

```
internal/
├── types/
│   └── session.go           # Session/Message 类型
├── storage/
│   ├── sqlite.go            # SQLite 初始化
│   ├── keys.go              # Keys 存储
│   ├── sessions.go          # Sessions 存储
│   └── storage_test.go      # 存储层测试
├── keypool/
│   └── pool.go              # 修改后的 Key 池
└── api/
    ├── admin_handler.go     # 修改后的 Key API
    ├── session_handler.go   # 新增的 Session API
    └── router.go            # 更新后的路由
```

## 测试步骤

### 1. 阅读规范
- `.agent/skills/qa-automation/SKILL.md` - QA 审核规范
- `.agent/skills/senior-golang/SKILL.md` - Go 代码规范

### 2. 单元测试验证

```bash
# 存储层测试
go test ./internal/storage/... -v

# KeyPool 测试（确保修改后仍通过）
go test ./internal/keypool/... -v

# 全部测试
go test ./... -v

# 静态分析
go vet ./...
```

### 3. 存储层代码审核

#### sqlite.go
| 检查项 | 说明 |
|--------|------|
| 初始化 | 数据库连接是否正确？ |
| AutoMigrate | 表结构是否正确创建？ |
| Close | 是否正确关闭连接？ |
| 错误处理 | 错误是否正确返回？ |

#### keys.go
| 检查项 | 说明 |
|--------|------|
| CreateKey | 是否正确插入？重复 Key 处理？ |
| GetKey | 不存在时返回什么？ |
| ListKeys | 是否正确查询所有？ |
| UpdateKey | 是否正确更新？ |
| DeleteKey | 是否正确删除？ |
| ImportKeys | 批量导入是否使用事务？ |

#### sessions.go
| 检查项 | 说明 |
|--------|------|
| CreateSession | ID 生成正确？ |
| GetSession | 是否包含 Messages？ |
| ListSessions | 分页是否正确？ |
| DeleteSession | 是否级联删除 Messages？ |
| AddMessage | SessionID 验证？ |

### 4. KeyPool 集成审核

| 检查项 | 说明 |
|--------|------|
| LoadFromDB | 是否正确加载 Keys？ |
| ReportSuccess | 是否同步更新 DB？ |
| ReportFailure | 是否同步更新 DB？ |
| 向后兼容 | 配置文件 Key 是否正确同步？ |

### 5. API 功能测试

启动服务器后测试：

```bash
# 启动服务
go run ./cmd/server
```

#### Key API 测试

```bash
# 获取 Keys
curl -X GET http://localhost:8080/api/keys

# 添加 Key
curl -X POST http://localhost:8080/api/keys \
  -H "Content-Type: application/json" \
  -d '{"key": "AIzaTest123456789", "name": "Test Key"}'

# 删除 Key
curl -X DELETE http://localhost:8080/api/keys/{id}

# 重启服务后验证 Key 仍存在
```

#### Session API 测试

```bash
# 创建会话
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Chat", "model": "gpt-4"}'

# 获取会话列表
curl -X GET http://localhost:8080/api/sessions

# 获取会话详情
curl -X GET http://localhost:8080/api/sessions/{id}

# 添加消息
curl -X POST http://localhost:8080/api/sessions/{id}/messages \
  -H "Content-Type: application/json" \
  -d '{"role": "user", "content": "Hello!"}'

# 更新会话标题
curl -X PUT http://localhost:8080/api/sessions/{id} \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Title"}'

# 删除会话
curl -X DELETE http://localhost:8080/api/sessions/{id}
```

### 6. 持久化验证

1. **添加数据**：通过 API 添加 Key 和 Session
2. **重启服务**：停止并重新启动服务
3. **验证数据**：确认数据仍然存在

### 7. 边界条件测试

| 场景 | 期望结果 |
|------|---------|
| 空数据库启动 | 正常启动，表自动创建 |
| 重复 API Key | 返回错误 |
| 不存在的 Session ID | 返回 404 |
| 删除不存在的 Key | 返回 404 或幂等成功 |
| 空标题创建会话 | 自动生成标题或使用默认值 |
| 配置文件 Key 同步 | 首次启动时写入 DB |

### 8. 安全检查

| 检查项 | 说明 |
|--------|------|
| API Key 暴露 | Key API 响应中是否只返回 `masked_key`？ |
| SQL 注入 | GORM 参数化查询是否正确使用？ |
| 文件权限 | SQLite 文件路径是否安全？ |

## 产出

创建审核报告 `docs/CodeReviewReport/STORAGE_LAYER_REVIEW_REPORT.md`：

```markdown
# SQLite 存储层审核报告

> **审核员**: QA Engineer
> **审核日期**: YYYY-MM-DD
> **审核范围**: `internal/storage/`, `internal/api/session_handler.go`

---

## 审核结果：✅ 通过 / ⚠️ 需修复 / ❌ 严重问题

| 指标 | 数量 |
|------|------|
| 严重问题 | X |
| 警告 | X |
| 建议改进 | X |
| 测试覆盖率 | X% |

---

## 单元测试结果

| 模块 | 测试数 | 通过 | 状态 |
|------|--------|------|------|
| storage | X | X | ✅/❌ |
| keypool | X | X | ✅/❌ |
| api | X | X | ✅/❌ |

---

## 功能测试结果

| API | 状态 | 说明 |
|-----|------|------|
| GET /api/keys | ✅/❌ | |
| POST /api/keys | ✅/❌ | |
| DELETE /api/keys/:id | ✅/❌ | |
| GET /api/sessions | ✅/❌ | |
| POST /api/sessions | ✅/❌ | |
| GET /api/sessions/:id | ✅/❌ | |
| POST /api/sessions/:id/messages | ✅/❌ | |
| DELETE /api/sessions/:id | ✅/❌ | |

---

## 持久化验证

| 测试项 | 状态 |
|--------|------|
| Key 重启后存在 | ✅/❌ |
| Session 重启后存在 | ✅/❌ |
| 配置文件 Key 同步 | ✅/❌ |

---

## 代码审核

### storage/sqlite.go
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

### storage/keys.go
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

### storage/sessions.go
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

### api/session_handler.go
- 状态：✅ / ⚠️ / ❌
- 问题（如有）：

---

## 问题详情

### ❌ 严重问题（如有）

### ⚠️ 警告（如有）

### 💡 改进建议（如有）

---

## 总结

**结论**: 继续开发 / 需先修复

---

*报告生成时间: YYYY-MM-DD HH:MM*
```

## 约束

- 优先进行单元测试验证
- 功能测试需要启动真实服务
- 持久化验证需要重启服务
- 安全检查重点关注 API Key 暴露

## 验收标准

1. 所有单元测试通过
2. 所有 API 端点可正常访问
3. 数据重启后仍然存在
4. 无 API Key 明文暴露
5. 代码质量符合规范

---

*任务创建时间: 2026-01-15*
