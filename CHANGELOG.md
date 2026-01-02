# Changelog

本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/) 规范。

## [v1.0.0] - 2026-01-02

### 首个版本发布

基于 [TGO](https://github.com/tgoai/tgo) 开源项目使用 Go 语言重构的智能客服平台。

### 新增功能

#### 后端服务 (Go)

- **apiserver** - API 网关服务
  - 用户认证 (JWT)
  - 流式代理端点 (`/v1/chat/team/stream`)
  - WuKongIM 集成
  - 会话管理

- **aicenter** - AI Agent 执行引擎
  - 基于 [eino](https://github.com/cloudwego/eino) ADK 框架
  - 多 Agent 协作 (Parallel/Sequential/Hierarchical)
  - 流式输出 (SSE)
  - 记忆持久化 (Redis + PostgreSQL)

- **rag** - RAG 知识库服务
  - 文档处理和向量化
  - 向量检索
  - 多集合支持

- **platform** - 第三方平台集成
  - 微信公众号
  - 自定义平台接入

#### 前端 (复用 TGO)

- **web** - 管理后台
  - 流式对话支持
  - Agent/Team 配置
  - 知识库管理

- **widget** - 访客聊天组件
  - 嵌入式聊天窗口
  - 多主题支持

### 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.22+, Gin, GORM |
| AI 框架 | eino ADK |
| 数据库 | PostgreSQL 15 |
| 缓存 | Redis 7 |
| 即时通讯 | WuKongIM |
| 前端 | React, TypeScript, Vite |

### 致谢

- [TGO](https://github.com/tgoai/tgo) - 原始项目
- [eino](https://github.com/cloudwego/eino) - AI Agent 框架
- [WuKongIM](https://github.com/WuKongIM/WuKongIM) - 即时通讯服务

---

[v1.0.0]: https://github.com/ashwinyue/captain/releases/tag/v1.0.0
