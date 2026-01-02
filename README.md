# Captain - 智能客服平台 (Go)

<p align="center">
  <strong>基于 <a href="https://github.com/tgoai/tgo">TGO</a> 开源项目重构的 Go 语言版本</strong>
</p>

<p align="center">
  <a href="https://github.com/tgoai/tgo">
    <img src="https://img.shields.io/badge/Based%20on-TGO-blue" alt="Based on TGO">
  </a>
  <a href="https://github.com/tgoai/tgo/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-Apache%202.0-green" alt="License">
  </a>
</p>

## 关于

Captain 是基于 [TGO](https://github.com/tgoai/tgo) 开源项目使用 Go 语言重构的智能客服平台。

- 后端服务使用 **Go** 语言重写，采用 [eino](https://github.com/cloudwego/eino) ADK 框架
- 前端界面复用 TGO 原项目的 **React** 前端（`web` 和 `widget`）

### 核心特性

- 🤖 **多 Agent 协作** - 支持 Parallel/Sequential/Hierarchical 工作流
- 📚 **RAG 知识库** - 向量检索增强的智能问答
- 💬 **流式输出** - 实时 SSE 流式响应
- 👥 **人工转接** - AI 与人工客服无缝切换
- 📱 **多平台集成** - 网站、微信公众号等
- 🔌 **WuKongIM** - 高性能即时通讯

## 项目架构

| 服务 | 端口 | 说明 |
|------|------|------|
| **apiserver** | 8000 | API 网关、用户认证、流式代理 |
| **aicenter** | 8081 | AI Agent 执行引擎 (eino ADK) |
| **rag** | 8082 | RAG 文档处理和向量检索 |
| **platform** | 8083 | 第三方平台集成 |
| **web** | 3000 | 管理后台前端 (React) |
| **widget** | 3001 | 访客聊天组件 (React) |

## 快速开始

### 前置要求

- Docker & Docker Compose
- Go 1.22+ (如需本地开发)

### 使用 Docker Compose 启动

```bash
# 克隆项目
cd /path/to/captain

# 复制环境变量配置
cp .env.example .env

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 基础设施端口

| 服务 | 端口 | 说明 |
|------|------|------|
| postgres | 5432 | PostgreSQL 数据库 |
| redis | 6379 | Redis 缓存 |
| wukongim | 5001 | WuKongIM API |
| wukongim | 5100 | WuKongIM TCP |
| wukongim | 5200 | WuKongIM WebSocket |

### 健康检查

```bash
# API Server
curl http://localhost:8000/health

# AI Center
curl http://localhost:8081/health

# RAG Service
curl http://localhost:8082/health

# Platform Service
curl http://localhost:8083/health

# WuKongIM
curl http://localhost:5001/health
```

## 本地开发

### 启动基础设施

```bash
# 只启动数据库和中间件
docker-compose up -d postgres redis wukongim adminer redis-commander
```

### 本地运行服务

```bash
# API Server
cd apiserver && go run ./cmd/server

# AI Center
cd aicenter && go run ./cmd/server

# RAG Service
cd rag && go run ./cmd/server

# Platform Service
cd platform && go run ./cmd/server
```

## 项目结构

```
captain/
├── docker-compose.yml      # Docker Compose 配置
├── scripts/
│   └── init-db.sql        # 数据库初始化脚本
├── apiserver/             # API 网关服务
│   ├── cmd/server/        # 入口
│   ├── internal/          # 内部代码
│   └── Dockerfile
├── aicenter/              # AI 执行引擎
│   ├── cmd/server/
│   ├── internal/
│   └── Dockerfile
├── rag/                   # RAG 服务
│   ├── cmd/server/
│   ├── internal/
│   └── Dockerfile
└── platform/              # 平台集成服务
    ├── cmd/server/
    ├── internal/
    └── Dockerfile
```

## API 文档

启动服务后访问：

- API Server: http://localhost:8000/docs
- AI Center: http://localhost:8081/docs
- RAG Service: http://localhost:8082/docs
- Platform Service: http://localhost:8083/docs

## 配置说明

### 环境变量

参考 `.env.example` 文件配置环境变量。

### WuKongIM 配置

WuKongIM 是即时通讯服务，用于实时消息推送：

- API 端口: 5001
- TCP 端口: 5100
- WebSocket 端口: 5200
- 管理端口: 5300

### AI Provider 配置

支持 OpenAI 兼容的 API：

```env
OPENAI_API_KEY=your-api-key
OPENAI_BASE_URL=https://api.openai.com/v1
```

也支持其他兼容提供商（如 Azure OpenAI, Anthropic 等）。

## 故障排除

### 数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
docker-compose ps postgres

# 查看日志
docker-compose logs postgres
```

### WuKongIM 连接失败

```bash
# 检查 WuKongIM 状态
docker-compose ps wukongim

# 查看日志
docker-compose logs wukongim
```

### 重置数据

```bash
# 停止并删除所有容器和数据卷
docker-compose down -v

# 重新启动
docker-compose up -d
```

## 致谢

本项目基于以下开源项目：

- [TGO](https://github.com/tgoai/tgo) - 原始智能客服平台（前端代码复用）
- [eino](https://github.com/cloudwego/eino) - 字节跳动 AI Agent 开发框架
- [WuKongIM](https://github.com/WuKongIM/WuKongIM) - 高性能即时通讯服务

## License

Apache License 2.0

本项目遵循 [TGO 原项目](https://github.com/tgoai/tgo) 的 Apache 2.0 开源协议。
