# AICenter 业务细节对齐状态

本文档记录 Go captain/aicenter 与 Python tgo-ai 的业务逻辑对齐情况。

## 对齐状态总览

| 模块 | Python tgo-ai | Go aicenter | 状态 | 优先级 |
|------|--------------|-------------|------|--------|
| **API 路由** | ✅ | ✅ | 🟢 完成 | - |
| **Supervisor 模式** | ✅ | ✅ | 🟢 完成 | - |
| **PlanExecute 模式** | ✅ | ✅ | 🟢 完成 | - |
| **Memory 持久化** | ✅ | ✅ | 🟢 完成 | - |
| **Streaming** | ✅ 完整 | ✅ | 🟢 完成 | - |
| **Usage 统计** | ✅ | ✅ | 🟢 完成 | - |
| **MCP Tools** | ✅ | ✅ | 🟢 完成 | - |
| **RAG Tools** | ✅ | ✅ | 🟢 完成 | - |
| **UI Template Tools** | ✅ | ✅ | 🟢 完成 | - |

---

## 已完成模块

### 1. API 路由 ✅

所有 API 路由已对齐：
- `/api/v1/agents` - Agent CRUD
- `/api/v1/teams` - Team CRUD
- `/api/v1/llm-providers` - LLM Provider CRUD + `/sync`
- `/api/v1/chat` - Chat completions
- `/api/v1/tools` - Tools CRUD
- `/api/v1/project-ai-configs` - Project AI Config

### 2. Supervisor 模式 ✅

使用 eino ADK `supervisor.New()` 实现：
- `SupervisorBuilder` - 构建 Supervisor Agent
- `SupervisorConfig` - 配置结构
- `Runner` - 运行 Supervisor

### 3. PlanExecute 模式 ✅

使用 eino ADK `planexecute.New()` 实现：
- `PlanExecuteBuilder` - 构建 Plan-Execute Agent
- 支持 Planner/Executor/Replanner

### 4. Memory 持久化 ✅

完整实现对话历史存储：
- `Store` 接口 - 存储抽象
- `InMemoryStore` - 内存存储 (开发/测试)
- `PostgresStore` - 数据库持久化
- `Manager` - 高层 API
- 集成到 `RuntimeService.Run/Stream`

---

## 待实现模块

### 5. Streaming 增强 ✅

**已实现：**
- `streaming/manager.go` - StreamManager 会话管理
- `streaming/session.go` - StreamingSession 状态跟踪
- `streaming/events.go` - 事件类型定义
- `streaming/errors.go` - 错误定义
- 会话超时自动清理

### 6. Usage 统计 ✅

**已实现：**
- `usage/model.go` - UsageRecord 模型
- `usage/repository.go` - Usage Repository
- `usage/tracker.go` - Usage Tracker
- Token 计数和成本计算
- 按 project/agent/model 统计

### 7. MCP Tools ✅

**已实现：**
- `tool/mcp_tool.go` - MCP 工具适配器
- HTTP 和 SSE 两种传输方式
- 工具发现和调用

### 8. RAG Tools ✅

**已实现：**
- `tool/rag_tool.go` - RAG 检索工具
- Collection 管理
- 文档搜索

### 9. UI Template Tools ✅

**已实现：**
- `tool/uitpl/schema.go` - 模板 Schema 定义
- `tool/uitpl/templates.go` - Order/Product/Logistics 等模板
- `tool/uitpl/registry.go` - 模板注册表
- `tool/uitpl/tools.go` - Eino 工具适配器

### 10. RAG Embedding Sync ✅

**已实现：**
- `service/embedding_sync_svc.go` - Embedding 配置同步服务
- 支持带重试的同步到 RAG 服务
- 同步状态跟踪 (pending/success/failed)
- 自动重试失败的同步任务

### 11. 后台任务系统 ✅

**已实现：**
- `task/scheduler.go` - 任务调度器
- `task/embedding_sync_task.go` - Embedding 同步重试任务
- 启动时自动重试失败的同步
- 可选周期性执行

---

## 实现计划

### Phase 1: 核心功能 (已完成)
- [x] API 路由对齐
- [x] Supervisor 模式
- [x] PlanExecute 模式
- [x] Memory 持久化

### Phase 2: 增强功能
- [ ] Streaming 增强
- [ ] Usage 统计
- [ ] RAG Tools 增强

### Phase 3: 扩展功能
- [ ] MCP Tools 集成
- [ ] UI Template Tools

---

## 文件结构

```
captain/aicenter/internal/eino/
├── agent/
│   └── builder.go           # Agent 构建器
├── llm/
│   └── factory.go           # LLM 工厂
├── memory/
│   ├── store.go             # Store 接口
│   ├── inmem.go             # 内存存储
│   ├── postgres.go          # 数据库存储
│   └── manager.go           # Memory 管理器
├── supervisor/
│   ├── team_builder.go      # Supervisor 构建
│   ├── plan_executor.go     # PlanExecute 构建
│   └── runner.go            # 运行器
├── tool/
│   ├── registry.go          # 工具注册表
│   ├── rag_tool.go          # RAG 工具
│   ├── mcp_tool.go          # MCP 工具
│   └── builtin/             # 内置工具
├── streaming/
│   ├── manager.go           # StreamManager
│   ├── session.go           # StreamingSession
│   ├── events.go            # 事件类型
│   └── errors.go            # 错误定义
└── usage/
    ├── model.go             # UsageRecord 模型
    ├── repository.go        # Repository
    └── tracker.go           # Tracker
```

---

## 实现进度

- ✅ Phase 1: 核心功能 (完成)
  - API 路由对齐
  - Supervisor 模式
  - PlanExecute 模式
  - Memory 持久化

- ✅ Phase 2: 增强功能 (完成)
  - Streaming 增强
  - Usage 统计
  - MCP Tools
  - RAG Tools

- ✅ Phase 3: 扩展功能 (完成)
  - UI Template Tools
  - RAG Embedding Sync
  - 后台任务系统

---

*最后更新: 2026-01-01*
