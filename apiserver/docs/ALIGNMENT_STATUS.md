# APIServer 业务细节对齐状态

本文档记录 Go captain/apiserver 与 Python tgo-api 的业务逻辑对齐情况。

## 对齐状态总览

| 模块 | Python tgo-api | Go apiserver | 状态 | 说明 |
|------|---------------|--------------|------|------|
| **Auth** | ✅ login/refresh/me | ✅ | 🟢 完成 | JWT 认证 |
| **Staff** | ✅ CRUD | ✅ | 🟢 完成 | 员工管理 |
| **Projects** | ✅ CRUD + regenerate-key | ✅ | 🟢 完成 | 项目管理 |
| **Visitors** | ✅ CRUD + tags + block | ✅ | 🟢 完成 | 访客管理 |
| **Tags** | ✅ CRUD | ✅ | 🟢 完成 | 标签管理 |
| **Sessions** | ✅ CRUD + close + transfer | ✅ | 🟢 完成 | 会话管理 |
| **Chat** | ✅ send + messages + revoke | ✅ | 🟢 完成 | 聊天消息 |
| **Conversations** | ✅ | ✅ | 🟢 完成 | 会话列表 |
| **Channels** | ✅ CRUD + members | ✅ | 🟢 完成 | 频道管理 |
| **Queue** | ✅ CRUD + assign | ✅ | 🟢 完成 | 排队系统 |
| **Assignment Rules** | ✅ get + update | ✅ | 🟢 完成 | 分配规则 |
| **Search** | ✅ | ✅ | 🟢 完成 | 搜索功能 |
| **Email** | ✅ test-connection | ✅ | 🟢 完成 | 邮件测试 |
| **WuKongIM** | ✅ 完整 Client | ✅ | 🟢 完成 | 消息服务 (完整对齐) |
| **System** | ✅ info + health | ✅ | 🟢 完成 | 系统信息 |
| **AI Proxy** | ✅ agents/teams/tools/providers/models | ✅ | 🟢 完成 | AI 中心代理 |
| **RAG Proxy** | ✅ collections/files/websites/qa-pairs | ✅ | 🟢 完成 | RAG 服务代理 |
| **Platforms** | ✅ 完整 | ✅ | 🟢 完成 | 平台接入 |
| **Onboarding** | ✅ | ✅ | 🟢 完成 | 引导流程 |
| **Setup** | ✅ | ✅ | 🟢 完成 | 初始化设置 |
| **MCP Tools** | ✅ project-tools | ✅ | 🟢 完成 | MCP 工具管理 |
| **Utils** | ✅ | ✅ | 🟢 完成 | 工具接口 |
| **Docs** | ✅ | ✅ | 🟢 完成 | 文档接口 |

---

## 已完成模块

### 核心业务
- Auth (JWT 登录/刷新/当前用户)
- Staff (员工 CRUD)
- Projects (项目 CRUD + API Key 重生成)
- Visitors (访客 CRUD + 标签 + 拉黑)
- Tags (标签 CRUD)
- Sessions (会话 CRUD + 关闭 + 转接)
- Chat (发送消息 + 历史 + 撤回)
- Conversations (会话列表)
- Channels (频道 CRUD + 成员管理)
- Queue (排队 CRUD + 分配)
- Assignment Rules (获取/更新分配规则)
- Search (全局搜索)

### 外部服务代理
- AI Proxy → aicenter (agents/teams/tools/providers/models)
- RAG Proxy → rag-service (collections/files/websites/qa-pairs)
- WuKongIM (路由 + Webhook)

### 其他
- Email (测试连接)
- System (系统信息/健康检查)

---

## 待实现模块

### 1. Platforms (平台接入) 🔴 [P1]

**Python 实现：**
- 平台类型管理 (微信/网页/API 等)
- 平台 CRUD
- 平台配置管理
- OAuth 回调处理
- 平台消息转发

**文件：** `app/api/v1/endpoints/platforms.py` (32646 bytes)

### 2. Onboarding (引导流程) 🔴 [P2]

**Python 实现：**
- 新用户引导
- 项目初始化向导
- 配置检查

**文件：** `app/api/v1/endpoints/onboarding.py` (9340 bytes)

### 3. Setup (初始化设置) 🔴 [P2]

**Python 实现：**
- 系统初始化
- 管理员创建
- 首次配置

**文件：** `app/api/v1/endpoints/setup.py` (30332 bytes)

### 4. MCP Project Tools 🔴 [P3]

**Python 实现：**
- 项目级 MCP 工具管理
- 工具绑定/解绑

**文件：** `app/api/v1/endpoints/mcp_project_tools.py` (8427 bytes)

### 5. Utils (工具接口) 🟡 [P4]

**Python 实现：**
- 通用工具方法
- 文件处理等

**文件：** `app/api/v1/endpoints/utils.py` (5695 bytes)

---

## 服务层对比

| Python Service | Go Service | 状态 |
|---------------|------------|------|
| `chat_service.py` | `chat_svc.go` | ✅ |
| `session_service.py` | `session_svc.go` | ✅ |
| `visitor_service.py` | `visitor_svc.go` | ✅ |
| `wukongim_client.py` | `wukongim/client.go` | ✅ |
| `rag_client.py` | `rag/client.go` | ✅ |
| `ai_client.py` | `aicenter/client.go` | ✅ |
| `transfer_service.py` | 部分 | 🟡 |
| `platform_sync.py` | ❌ | 🔴 |
| `onboarding_service.py` | ❌ | 🔴 |
| `geoip_service.py` | ❌ | 🔴 |
| `run_registry.py` | ❌ | 🔴 |
| `queue_trigger_service.py` | ❌ | 🔴 |

---

## 实现优先级

### P1 - 高优先级
- [ ] Platforms (平台接入) - 核心业务功能

### P2 - 中优先级
- [ ] Setup (初始化设置) - 首次部署必需
- [ ] Onboarding (引导流程) - 用户体验

### P3 - 低优先级
- [ ] MCP Project Tools - 增强功能

### P4 - 可选
- [ ] Utils - 辅助功能
- [ ] Docs - 文档接口

---

*最后更新: 2026-01-01*
