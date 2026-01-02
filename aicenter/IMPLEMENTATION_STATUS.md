# AICenter 实现状态文档

> 更新时间: 2026-01-02

## 已完成功能

### 1. Orchestration 模块 ✅

基于 eino 最佳实践实现的智能查询处理模块。

| 组件 | 文件 | 功能 |
|------|------|------|
| QueryAnalyzer | `orchestration/query_analyzer.go` | 查询意图分析、Agent 选择、复杂度判断 |
| ExecutionManager | `orchestration/execution_manager.go` | 并行/串行多 Agent 执行 |
| ResultConsolidator | `orchestration/result_consolidator.go` | LLM 聚合多 Agent 结果 |
| Types | `orchestration/types.go` | 类型定义 |
| Prompts | `orchestration/prompts.go` | Prompt 模板 |

**工作流类型:**
- `single`: 单 Agent 执行
- `parallel`: 并行执行多个 Agent
- `sequential`: 串行执行多个 Agent

### 2. 自定义工具 ✅

| 工具 | 文件 | 功能 |
|------|------|------|
| RAG 检索 | `tool/rag_tool.go` | 知识库向量检索 |
| MCP 工具 | `tool/mcp_tool.go` | MCP 协议工具调用 |
| 转人工 | `tool/transfer_human.go` | 请求人工客服 |
| 访客信息 | `tool/visitor_info.go` | 更新/获取访客联系方式 |
| 访客情绪 | `tool/visitor_sentiment.go` | 记录满意度、情绪、意图 |
| 访客标签 | `tool/visitor_tag.go` | 为访客添加分类标签 |

### 3. API 路由 ✅

完整实现与原 Python 版本兼容的 API：

- **Agents**: CRUD + Run + Exists + EnableTool/Collection
- **Teams**: CRUD + GetDefault
- **LLM Providers**: CRUD + Sync + Test
- **Tools**: CRUD
- **Project AI Configs**: Sync + Upsert + Get
- **Chat Completions**: OpenAI 兼容接口

### 4. 服务层 ✅

| 服务 | 文件 | 功能 |
|------|------|------|
| AgentService | `service/agent_svc.go` | Agent CRUD |
| TeamService | `service/team_svc.go` | Team CRUD |
| RuntimeService | `service/runtime_svc.go` | Agent 运行时 + QueryAnalyzer |
| ProviderService | `service/provider_svc.go` | LLM Provider 管理 |
| ToolService | `service/tool_svc.go` | Tool CRUD |
| ProjectAIConfigService | `service/project_ai_config_svc.go` | 项目 AI 配置 |
| EmbeddingSyncService | `service/embedding_sync_svc.go` | Embedding 同步 |

### 5. Eino 集成 ✅

| 组件 | 文件 | 功能 |
|------|------|------|
| LLM Factory | `eino/llm/factory.go` | ChatModel/ToolCallingModel 创建 |
| Agent Builder | `eino/agent/builder.go` | ReAct Agent 构建 |
| Supervisor | `eino/supervisor/` | Team 协调运行 |
| Memory | `eino/memory/` | 会话记忆 (内存) |

---

## 待优化项 (TODO)

### P1: 会话记忆持久化 ✅ (已完成)

**实现方案:** 使用 PostgreSQL 持久化 (`memory.PostgresStore`)

**已完成:**
- ✅ 使用 PostgreSQL `conversation_messages` 表存储会话历史
- ✅ 支持按 session_id 和 project_id 隔离
- ✅ 支持窗口化历史 (默认保留最近 10 条消息)
- ✅ `RunWithReactAgentAndMemory` 方法集成会话记忆

**相关文件:**
- `eino/memory/postgres.go` - PostgreSQL 存储实现
- `eino/memory/manager.go` - 会话管理器
- `service/runtime_svc.go` - 运行时服务集成

**使用方式:**
```bash
curl -X POST '/api/v1/agents/run' \
  -d '{"message":"你好，我是张三","session_id":"my-session","enable_memory":true}'
```

### P2: 错误处理和重试机制

**现状:** 基础错误处理，无自动重试

**优化方案:**
- LLM 调用失败自动重试 (指数退避)
- 工具调用超时处理
- 降级策略 (fallback to simpler model)

**相关文件:**
- `service/runtime_svc.go`
- `eino/llm/factory.go`

### P3: 性能监控和日志优化

**现状:** 基础日志输出

**优化方案:**
- 结构化日志 (JSON 格式)
- Prometheus 指标采集
- 链路追踪 (OpenTelemetry)
- Agent 执行统计

**相关文件:**
- `middleware/logger.go`

### P4: 多 Agent 并行执行测试

**现状:** 并行执行已实现，但需要更多测试场景

**优化方案:**
- 创建测试用例验证并行执行
- 测试结果聚合正确性
- 压力测试

### P5: MCP 工具动态加载

**现状:** MCP 工具需要手动配置

**优化方案:**
- 从 MCP 服务动态发现工具
- 工具缓存和刷新机制

### P6: 敏感信息脱敏中间件 (参考 work_v3)

**现状:** 无敏感信息过滤

**优化方案:**
- 添加 `security_middleware.go` 中间件
- 自动脱敏密码、身份证、银行卡等敏感字段
- 支持正则匹配和字段名匹配
- 配置化敏感字段列表

**参考实现:** `work_v3/security_middleware.py`

```go
// 敏感字段配置
var sensitiveFields = []string{"password", "passwd", "secret", "token", "id_number", "card_no"}

// 正则匹配身份证、银行卡
var sensitivePatterns = []*regexp.Regexp{
    regexp.MustCompile(`\d{6}(?:19|20)\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01])\d{3}[0-9Xx]`),
    regexp.MustCompile(`(?:\d[ -]?){16,19}`),
}
```

### P7: 推荐问题生成 (参考 work_v3)

**现状:** 无追问建议功能

**优化方案:**
- 根据用户问题和 AI 回答生成 3-5 个追问建议
- 帮助用户深入了解具体细节
- 在 API 响应中返回 `suggested_questions` 字段

**Prompt 模板:**
```
基于用户问题'{question}'和客服回答'{answer}'，生成3-5个用户可能继续追问的相关问题。
要求：
- 必须与用户原问题和客服回答直接相关
- 引导用户深入了解具体细节
- 问题形式应为开放式问句
```

### P8: 意图路由优化 - 关键词前置 ✅ (已完成)

**实现方案:** 参考 eino router 模式，关键词匹配优先于 LLM 分析

**已完成:**
- ✅ `IntentRouter` 关键词快速匹配
- ✅ 支持意图类型：greeting, human, order
- ✅ 匹配成功时跳过 LLM 调用 (SkipLLM=true)
- ✅ 集成到 `QueryAnalyzer.Analyze` 方法

**相关文件:**
- `orchestration/intent_router.go` - 意图路由器
- `orchestration/query_analyzer.go` - 查询分析器集成

**测试结果:**
```
Query: "你好" → Quick match: intent=greeting, confidence=0.95
Query: "转人工客服" → Quick match: intent=human, confidence=0.95
```

**性能提升:**
- 简单意图 (打招呼/转人工) 跳过 LLM 调用
- 响应延迟从 ~2s 降低到 ~1s

### P9: 日志推送 ELK (参考 work_v3)

**现状:** 本地日志输出

**优化方案:**
- 批量推送日志到 Logstash/Elasticsearch
- 支持认证 (Basic Auth / Bearer Token)
- 断点续传 (状态持久化)
- 多线程异步推送

**参考实现:** `work_v3/log_push.py`

---

## 与原项目对比

| 功能 | 原项目 (Python) | 新项目 (Go) |
|------|----------------|-------------|
| API 兼容性 | - | ✅ 100% |
| QueryAnalyzer | ✅ | ✅ |
| WorkflowPlanner | ✅ | ✅ (集成) |
| ExecutionManager | ✅ | ✅ |
| ResultConsolidator | ✅ | ✅ |
| 自定义工具 | 6 个 | ✅ 6 个 |
| SSE 流式输出 | ✅ | ✅ |
| 会话记忆 | Redis | 内存 (待优化) |

---

## 参考资料

- eino-examples: `/Users/mervyn/go/src/github/tgo/eino-examples`
- 原项目: `/Users/mervyn/go/src/github/tgo/repos/tgo-ai`



P1	会话记忆持久化 (Redis)	中
P2	错误处理和重试机制	中
P3	性能监控和日志优化	中
P4	多 Agent 并行执行测试	低
P5	MCP 工具动态加载	中
P6	敏感信息脱敏中间件	低
P7	推荐问题生成	低
P8	意图关键词前置	低
P9	日志推送 ELK	高