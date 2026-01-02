# Captain 可观测性与 Trace 方案

## 概述

基于 `eino-examples` 的最佳实践，为 Captain 项目添加可观测性和链路追踪能力，支持 AI Agent 执行过程的监控、调试和性能分析。

## 可用方案对比

eino 框架通过 `callbacks` 机制支持多种可观测性后端：

| 方案 | 类型 | 特点 | 适用场景 |
|------|------|------|----------|
| **CozeLoop** | SaaS | Coze 官方 AI 可观测平台，专为 Agent 设计 | 生产环境、多 Agent 监控 |
| **Langfuse** | SaaS/Self-hosted | 开源 LLM 可观测平台，功能丰富 | 开发调试、成本分析 |
| **APMPlus** | SaaS | 火山云 APM 服务，集成 OpenTelemetry | 企业级监控、与火山云生态集成 |
| **Eino DevOps** | 本地 | 可视化调试工具，实时查看 Graph 执行 | 本地开发调试 |

## 推荐方案

### 方案一：CozeLoop（推荐）

CozeLoop 是 Coze 官方的 AI 可观测平台，专为 Agent 设计，提供：
- Agent 执行链路追踪
- Token 消耗统计
- 延迟分析
- 错误监控

**优点**：
- 专为 AI Agent 设计，trace 展示更友好
- 与 eino 框架深度集成
- 免费额度足够开发使用

**配置**：
```bash
# .env
COZELOOP_WORKSPACE_ID=your_workspace_id
COZELOOP_API_TOKEN=your_api_token
```

**代码实现**：
```go
// aicenter/internal/trace/cozeloop.go
package trace

import (
    "context"
    "os"
    
    clc "github.com/cloudwego/eino-ext/callbacks/cozeloop"
    "github.com/cloudwego/eino/callbacks"
    "github.com/coze-dev/cozeloop-go"
)

func InitCozeLoop(ctx context.Context) (func(context.Context), error) {
    wsID := os.Getenv("COZELOOP_WORKSPACE_ID")
    apiToken := os.Getenv("COZELOOP_API_TOKEN")
    
    if wsID == "" || apiToken == "" {
        return func(ctx context.Context) {}, nil
    }
    
    client, err := cozeloop.NewClient(
        cozeloop.WithWorkspaceID(wsID),
        cozeloop.WithAPIToken(apiToken),
    )
    if err != nil {
        return nil, err
    }
    
    handler := clc.NewLoopHandler(client)
    callbacks.AppendGlobalHandlers(handler)
    
    return client.Close, nil
}
```

**获取 CozeLoop 凭证**：
1. 访问 https://loop.coze.cn
2. 创建 Workspace
3. 获取 Workspace ID 和 API Token

### 方案二：Langfuse

Langfuse 是开源的 LLM 可观测平台，支持自托管。

**优点**：
- 开源免费，可自托管
- 详细的成本分析
- 支持 Prompt 版本管理

**配置**：
```bash
# .env
LANGFUSE_PUBLIC_KEY=your_public_key
LANGFUSE_SECRET_KEY=your_secret_key
LANGFUSE_HOST=https://cloud.langfuse.com  # 或自托管地址
```

**代码实现**：
```go
// aicenter/internal/trace/langfuse.go
package trace

import (
    "os"
    
    "github.com/cloudwego/eino-ext/callbacks/langfuse"
    "github.com/cloudwego/eino/callbacks"
)

func InitLangfuse() error {
    publicKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
    secretKey := os.Getenv("LANGFUSE_SECRET_KEY")
    
    if publicKey == "" || secretKey == "" {
        return nil
    }
    
    host := os.Getenv("LANGFUSE_HOST")
    if host == "" {
        host = "https://cloud.langfuse.com"
    }
    
    handler, _ := langfuse.NewLangfuseHandler(&langfuse.Config{
        Host:      host,
        PublicKey: publicKey,
        SecretKey: secretKey,
        Name:      "Captain AI Center",
        Release:   "v0.1.0",
        Tags:      []string{"captain", "aicenter"},
    })
    
    callbacks.AppendGlobalHandlers(handler)
    return nil
}
```

### 方案三：Eino DevOps（开发调试）

Eino 提供本地可视化调试工具，可实时查看 Graph 执行过程。

**代码实现**：
```go
// aicenter/cmd/main.go
import "github.com/cloudwego/eino-ext/devops"

func main() {
    // 开发环境启用 devops
    if os.Getenv("EINO_DEVOPS") == "true" {
        if err := devops.Init(context.Background()); err != nil {
            log.Printf("eino devops init failed: %v", err)
        }
    }
    // ...
}
```

## 实施计划

### Phase 1: 基础集成（1 天）

1. 添加依赖：
```bash
go get github.com/cloudwego/eino-ext/callbacks/cozeloop
go get github.com/coze-dev/cozeloop-go
go get github.com/cloudwego/eino-ext/devops
```

2. 创建 `aicenter/internal/trace/` 目录
3. 实现 CozeLoop 初始化
4. 在 `main.go` 中调用初始化

### Phase 2: 开发工具（0.5 天）

1. 集成 Eino DevOps 可视化调试
2. 添加环境变量开关

### Phase 3: 多后端支持（可选）

1. 添加 Langfuse 支持
2. 支持同时启用多个 trace 后端

## 环境变量汇总

```bash
# CozeLoop (推荐)
COZELOOP_WORKSPACE_ID=
COZELOOP_API_TOKEN=

# Langfuse (可选)
LANGFUSE_PUBLIC_KEY=
LANGFUSE_SECRET_KEY=
LANGFUSE_HOST=https://cloud.langfuse.com

# Eino DevOps (开发环境)
EINO_DEVOPS=true
```

## 文件结构

```
aicenter/
├── internal/
│   └── trace/
│       ├── init.go       # 统一初始化入口
│       ├── cozeloop.go   # CozeLoop 集成
│       └── langfuse.go   # Langfuse 集成
└── cmd/
    └── main.go           # 调用 trace.Init()
```

## 预期效果

1. **链路追踪**：查看完整的 Agent 执行链路，包括 LLM 调用、Tool 执行
2. **性能监控**：Token 消耗、延迟统计、错误率
3. **调试能力**：可视化 Graph 执行过程，快速定位问题
4. **成本分析**：按请求/用户统计 API 调用成本

## 参考资料

- [CozeLoop Go SDK](https://loop.coze.cn/open/docs/cozeloop/go-sdk)
- [eino-examples/adk/common/trace/coze_loop.go](https://github.com/cloudwego/eino-examples/blob/main/adk/common/trace/coze_loop.go)
- [eino-examples/quickstart/eino_assistant](https://github.com/cloudwego/eino-examples/tree/main/quickstart/eino_assistant)
- [Langfuse Documentation](https://langfuse.com/docs)
